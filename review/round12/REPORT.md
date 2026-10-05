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
(`parent.members.add(this)`). It compiles on B3 and leaks 560 bytes in 6 allocations, all indirect. Node, -O2
and the JavaScript backend all print `2` with exit 0.

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

## A2: regions, values in a region left uncounted, regions ended on a throw (926feae)

A sub-review did most of this half, under my instructions. Its probes are in `review/round12/a2/`. I checked
its central claims myself on 926feae:

- `s1_spread.a` and `t1_consumer_throws.a` give the same bytes as Node on all three other runs, with exit 0
  and no leak.
- I re-ran its main mutant. The mutant turns off the region end in `checkThrown`
  (internal/native/exceptions.go:66, `if false && ...`). Against `t1_consumer_throws.a` it's caught only by
  LeakSanitizer: 39,384 bytes in 128 allocations. Node, -O2 and the JavaScript backend still agree with it on
  (output and exit 0), so nothing but the leak check catches it.

The branch's own commits are 0222103 (regions), f9af53f (no retain or release on values in a region) and the
merge 2bed985 (a region ends on the throw path). The spread-of-undefined fix came in through main at 4ff4657;
it was probed here anyway.

### Confirmed: a compiler crash, already on main

`a2/x_optional_method.a` calls a method through optional chaining on a class value that may be undefined:

    const p = maybe(1);              // Point | undefined
    console.log(p?.describe() ?? 'none');

`adamic c` panics with a nil pointer dereference: instantiate.go:29, from class.go:68, from class.go:282. I
reproduced it on my branch, which carries main 4ff4657. A2's branch doesn't touch internal/lower, so it's
main's. It should compile, or say NotYet. It's loud, not silent.

### Not a bug, but a limit for probes

An uncaught throw leaves through `adamic_panic`'s `_exit(70)`, so LeakSanitizer never runs. A probe that ends
uncaught can't catch a leak. The sub-review's first `t1` ended that way, and the region-end mutant passed it
until the trailing throw was removed.

### Held, on all four runs, leak-clean, counted builds balanced (allocations = frees + in regions)

- **`t1_consumer_throws.a`:** a consumer throws after reading region nodes that hold heap strings, in a try,
  in a loop, on 4 of 6 rounds. The same again one call down, caught by the caller, with a finally.
- **`t2_nested_regions.a`:** a fresh function with its own region statement, recursing and throwing at
  varied depths.
- **`e2_finally.a`:** region statements in a catch and a finally while an exception unwinds, including a throw
  from the finally's own region statement. Also caught by the region-end mutant: 12,543 bytes leaked.
- **`f1_fresh_try.a`:** a fresh function returning from a try, catch and finally. One field is half made when
  a later one throws, and the finally throws.
- **`c1_versions.a`:** a function compiled in both versions, heap and region, called from both kinds of
  statement. Strings and array elements are read out of a region and kept after it ends. A mutant that
  stores a region literal's heap string fields without a retain is caught by ASan
  (heap-use-after-free in `adamic_string_concat`).
- **`s1_spread.a`:** these forms of spread, including the reuse path (whose `!= NULL &&` guard the C shows
  twice):
  - `{...u}`, `{...u, x}`, `{...u, label}` and `{...u, leaf: {...}}`;
  - `{...holder.point}`, `{...u?.leaf}` and `{...u?.leaf, name}`.

  A mutant that leaves an undefined spread's number fields zero prints `x=0 tag=0` where Node prints
  `x=u tag=u`.
- **`s2_spread_class.a`** and **`r1_reuse_region.a`:** a spread of a class value that may be undefined, and a
  spread of a maybe-undefined local in the same statement as a region.
- **The earlier rounds' probes** (round7, round8, round8b) behave as reported then. `early_this.a`,
  `half_built.a`, `defined_global.a` and `defined_try.a` are still open, unchanged.

### Not covered on A2

- **Mutual recursion between fresh functions:** stage 0 says NotYet.
- **Not probed:**
  - region statements in a for-loop update or a while condition;
  - default parameters that make a fresh value;
  - a Weak to a region object, which escape analysis keeps out by reading alone.
- **No gate:** the full gate wasn't run on 926feae.
