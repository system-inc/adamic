# Stream R, round twelve: the three branches that make integration 10

This round covers three branches:

- A2's claude/release-globals-at-exit-uqmaf6 (926feae).
- B2's claude/stage1-language-gaps-o3lbxx (7340724).
- B3's claude/cycle-finder-fresh-writes-ssz3oh (691ed9e).

Everything ran on Linux x86-64 with Go 1.27.0, clang 18.1.3 and Node v24.21.0. `review/fxspptb/probe.sh` runs
each probe four ways:

1. Node on the source.
2. Native under ASan, UBSan and LeakSanitizer.
3. Native as `adamic build` makes it (-O2).
4. The JavaScript backend.

## B3: the fresh-write relaxation (691ed9e)

### Confirmed: a constructor can push `this` and then close the cycle through a readonly field

`ctor_readonly.a`:

    class Node {
      readonly kids: Node[] = [];
      readonly parent: Node | undefined;
      ...
      constructor(name: string, parent: Node | undefined) {
        this.name = name;
        if (parent !== undefined) {
          parent.kids.push(this);
        }
        this.parent = parent;
      }
    }

`build()` makes a root and a child. The child's constructor pushes itself into the root's `kids`, then sets
its own `parent` to the root. That's a cycle, root to kids to child to root, with no Weak anywhere.

| Run | Result |
|---|---|
| main (4ff4657) | refused at compile time: `Node[]`, an array whose elements can reach back |
| B3, Node and the JavaScript backend | `total 6`, exit 0 |
| B3, native sanitized | `total 6`, then LeakSanitizer: 993 bytes in 21 allocations, every one an indirect leak (a cycle), exit 1 |
| B3, native -O2 | `total 6`, exit 0: the leak is silent |

`ctor_wrapped.a` is the same hole with `this` wrapped in a literal (`parent.links.push({ owner: this })`). It
also compiles on B3 and leaks 240 bytes in 6 allocations, all indirect. `ctor_set.a` is the Set version
(`parent.members.add(this)`), and it compiles on B3 too.

**Why it gets through.** Each half is judged correctly on its own terms, but nobody judges the second half:

- **The push is proven.** At `parent.kids.push(this)`, `this` is a literal the constructor just made. It
  reaches only itself and its own empty `kids`, so it can't reach the array it's pushed into. By the rule
  "a fresh value reaches only confined objects", that's right at that moment.
- **The closing write is never consulted.** `this.parent = parent` is the write that closes the cycle. By
  then `this` has escaped, so fresh.go marks the write unproven ("what it's written into has escaped by
  then"). But `parent` is readonly, and `slotsOf` in cycles.go skips readonly fields (`if
  l.checker.IsReadonlySymbol(field) { continue }`). So no slot ever asks about that write.

The old finder could skip readonly fields because every write into a mutable cycle-capable slot was refused,
so a half-built `this` could never be put where it reaches back. B3 now lets the push through, and the
induction in fresh.go's own comment ("a cycle needs a last write... at every write, the value can't reach the
object written into") only holds if every write is consulted. The readonly writes in constructors aren't.

This is the round-seven `early_this` family (`this` escaping before the constructor finishes), now as a leak
rather than a null read.

**A fix that holds on these probes.** I deleted the readonly skip in `slotsOf`, so a readonly cycle-capable
field also needs every write proven. A constructor's ordinary `this.parent = parent`, with `this` still
confined, stays proven, since the holder is fresh. With that change:

- `ctor_readonly.a`, `ctor_wrapped.a` and `ctor_set.a` are all refused, each naming the `this.parent` write.
- `via_class.a` is still refused as before.

The patch is `review/round12/readonly-fields.patch`. Its refusal text still says "a mutable field" for a
readonly one, which would need its own wording. Whether the patch costs any existing fixture is still being run (lower
and oracle packages); this report will say when it's back.

The mutant is the patch run backwards: without it, the three constructor probes compile and leak; with it,
they're refused. A probe that doesn't run a constructor (`via_*.a`, `param_object.a`) is refused either way,
so only the readonly check catches these three.

### Held: refused, or compiled and correct

Refused on B3, as they should be, each naming the closing write:

- **The direct shapes** (`via_call.a`, `via_literal.a`, `via_closure.a`, `via_class.a`, `via_array.a`): a
  child reaching its root through a call's result, a literal, a closure's capture, a class constructor, or an
  array or a Map, then pushed into the root.
- **`loop_pair.a`:** two nodes from one allocation site linking each other across a loop's iterations. The
  newest and older instances are told apart, and the back link is caught.
- **`read_back.a`:** an element read back out of an array and given the array's holder.
- **`spread_back.a`:** the root reached through a spread copy of an object that holds it.
- **`map_entries_back.a`:** the root reached through a Map's entries in a for...of.
- **`caught_early.a`:** a link made in a try, a throw while the variable still reaches the root, and the
  closing push in the catch. The handler sees the state from before the throw.
- **`union_receiver.a`:** a write through a receiver typed `Left | Right`, in a helper. It's still matched to
  both classes' `next` field.
- **`param_object.a`:** a helper that links a fresh object and a parameter both ways.

Compiled and correct: `caught_write.a`, where the throw comes after the variable is given a fresh object, so
the catch's push makes no cycle. It agrees on all four runs, and is leak-clean.

### Not covered on B3

- Readonly fields written anywhere but in constructors. 0.1 lets only a constructor write one, as far as I
  know; I didn't check whether a readonly field can be written through a non-readonly view.
- Map keys. I probed Set elements; the Map key relaxation is the same code path in fresh.go, but I didn't
  build a cycle through a key.
- Whether summaries reach a fixed point too slowly on a large program (the 32-round cap). I only read that
  falling back to "every call is outside" is sound.
