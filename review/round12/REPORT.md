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

## Summary, worst first

1. **B2, silent:** a readonly view turns back into a mutable one (`b2_silent.a`). A Cat lands in a Dog's
   slot, and native reads Cat's string `weight` as a number. It prints `cat4 weighs 1 1.00` with exit 0,
   sanitizers clean, where Node throws.
2. **B3, a silent leak:** a constructor pushes `this` into its parent's array (or Set), then sets a readonly
   `parent` (`ctor_readonly.a`, `ctor_wrapped.a`, `ctor_set.a`). The cycle compiles. LeakSanitizer catches
   it; the release build doesn't. A one-line patch holds and costs no fixture.
3. **B2, loud:** 16 more unsound programs compile and fail at run time, all a Cat read as a Dog. Their routes:
   - views `viewSite` doesn't list: shorthand, default parameters, class field initializers, object spread,
     `as`;
   - union sources and targets;
   - method signatures;
   - tuple tails;
   - class targets;
   - method bivariance.

   Four more are masked by NotYet today.
4. **Main, loud:**
   - `p?.describe()` on a `Point | undefined` crashes the compiler (`a2/x_optional_method.a`);
   - a class called through an interface it implements panics natively (`b2/c02_class_as_interface_sound.a`).
5. **B2, too strict:** fresh copies are refused (`slice()`, `map`, a conditional of literals,
   `new Map<string, Dog>()`).
6. **A2:** nothing found. The region end on a throw is held by a mutant only LeakSanitizer catches.

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
readonly one, which would need its own wording. With the patch on B3:

- `go test ./internal/lower` passes.
- `go test -timeout 30m ./internal/oracle` passes every fixture except the six input fixtures
  (`read_files.a`, `walk.a`, `write_files.a`, `read_arguments.a`, `arguments.a`, `utf8_sweep.a`).
  - All 12 of those failures, in TestInputAgreesWithNode and TestCountsAreRecorded, are
    `fork/exec /opt/node24/bin/node: permission denied`. That's the input tests dropping to uid 65534 in my
    worktree, the environmental failure from earlier rounds. It isn't the patch.
- So no existing program relied on a readonly field closing a cycle. That includes `fresh_parser.a`,
  `fresh_writes.a` and the 22 refusal probes.

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

## B2: invariance refusals (7340724)

A sub-review hunted this half under my instructions. Its probes are in `review/round12/b2/`, and each starts
with `prelude.txt`: Animal, Dog and Cat with mutable fields; `dogs: Dog[]`; `cat`; and `report(list: Dog[])`,
which reads `bark`. I re-ran its headline findings myself on 7340724 and added one probe, `b2_silent.a`.

The core rule works: the direct shapes (`Dog[]` seen as `Animal[]`, `{ pet: Dog }` as `{ pet: Animal }`) are
refused. What's left is in where the rule is applied, plus one hole in the design.

### Confirmed, worst first

**1. A readonly view turned back into a mutable one, and it can be silent.** The rule keeps readonly
properties covariant, which is right on its own. But tsc ignores `readonly` when it relates properties, so
the readonly view converts straight back to a mutable type. Both sides of that second step say `Animal`, so
the walk finds nothing:

    const kennel: { pet: Dog } = { pet: makeDog(1) };
    const view: { readonly pet: Animal } = kennel;   // allowed: readonly is covariant
    const pen: { pet: Animal } = view;                // allowed: tsc ignores readonly here
    pen.pet = cat;                                    // kennel.pet is now a Cat

In `b2_silent.a`, Dog's `weight` is a number and Cat's `weight` is a string:

| Run | Result |
|---|---|
| Node and the JavaScript backend | `TypeError: total.toFixed is not a function`, exit 70 |
| Native, sanitized | `cat4 weighs 1 1.00`, exit 0, sanitizers clean |
| Native, -O2 | `cat4 weighs 1 1.00`, exit 0 |

That's a silent miscompile: a string read as a number. `b2_direct.a`, the same program without the readonly
hop, is refused on B2, so the readonly view is the only thing letting it through. On main (4ff4657) both
compile, since main has no invariance rule.

When the Cat simply lacks the Dog's field, it's loud instead. `b53_readonly_laundering.a` (I re-ran it)
panics natively with `compiler bug: a field the checker proved is there is missing`, where Node and the
JavaScript backend give the TypeError. Exit is 70 everywhere. Two more versions of the hole compile:

- `b54_minus_readonly.a`: the same through a `-readonly` mapped type.
- `b55_readonly_laundering_array_field.a`: the same with a readonly array field.

A fix: refuse a mutable target property whose source property is readonly, unless the two types are
identical.

Every finding below fails the same loud way as `b53` when run: a Cat read as a Dog, the native panic, and
Node's TypeError.

**2. Places a value enters a typed slot that `viewSite` doesn't list:**

- `b01_shorthand.a` (I re-ran it): a shorthand property, `{ pets }`.
- `b05_default_param.a`: a parameter default, `animals: Animal[] = dogs`.
- `b06_class_field.a`: a class field's initializer.
- `b08_spread_object.a`: an object spread, `{ ...kennel }`. The copy is shallow, so the array is shared.
- `b04_as.a`: `dogs as Animal[]`. cast.go treats an upcast as the value itself, and nothing checks it.

**3. Union targets and union sources are skipped.** `withoutUndefined` gives up when more than one member is
left besides undefined:

- `b09_union_target.a`: `Animal[] | string`.
- `b10_union_source.a` (I re-ran it): `Dog[] | Cat[]` seen as `Animal[]`.
- `b31_conditional_union.a`: `dogs.length > 0 ? dogs : cats`.
- `b49_ternary_union_obj.a`: the same shape with objects.
- `b72_function_union_return.a`: a function returning `Dog[] | Cat[]`, passed as `Animal[]`.

**4. A method signature in the target is skipped.** `widened` passes over a target property that's a method,
so a method's return type is never compared:

- `b51_method_signature_target.a`: `{ list(): Animal[] }`.
- `b36_class_method_returns.a`: the same with a class instance as the source.

**5. A tuple is compared only up to the target's type argument count.** `b11_tuple_tail.a` compares index 0
of `[Animal, Dog]` and never the Dog at index 1. Natively it can't show this yet, because a tuple seen as an
array is already broken (the c3.a finding from earlier rounds).

**6. Classes:**

- `b17_class_to_class.a` (`Box<Dog>` seen as `Box<Animal>`) and `b38_param_property_class_target.a`. The code
  says class-to-class belongs to the nominal rule stage 0 doesn't check yet, so this is known. Still, today
  it's a wrong run, not a refusal.
- `b18_object_to_class.a` (I re-ran it): a plain object seen as a class. `isClassInstance(to)` returns
  before the walk.

**7. Holes masked by NotYet today.** The refusal pass lets these through; lowering then stops with NotYet.
They'll compile wrong once stage 0 lowers them:

- `b12_constraint_push.a`: `T extends Animal[]`, then `list.push(cat)`.
- `b13_constraint_assign.a`: `T extends Animal[]` assigned to `Animal[]`.
- `b14_constraint_field.a`: `T extends { pet: Animal }`, then `holder.pet = cat`.
- `b66_spread_argument.a`: `adopt(...args)`.

**8. Method bivariance**, in `b33_class_method_bivariance.a`. This is docs/0.1.md's `method-signature-style`,
not this commit's rule, but it compiles and fails at run time.

### Sound programs it now refuses

All four are conservative rather than wrong, but each refuses a value nothing else holds:

- `a19_conditional_fresh.a`: `dogs.length > 5 ? [cat] : []`. The array-literal exemption doesn't reach a
  conditional's branches.
- `a20_fresh_copies.a`: `dogs.slice()`. It stops at that first line, so `filter` and `Array.from` in the same
  file weren't reached.
- `a13_map_method.a`: `dogs.map((dog) => dog)`.
- `a14_fresh_new_map.a`: `new Map<string, Dog>()` into `Map<string, Animal>`.

`a17_return_fresh.a`, a local `Dog[]` returned as `Animal[]`, is refused too. That one would need escape
analysis to allow.

### Confirmed on main, not this branch: a sound program panics

`c02_class_as_interface_sound.a` calls a class through an interface it implements:

    const handler: Handler = new AnimalHandler();
    handler.handle(cat);

On main (4ff4657):

- Node prints `handled dog1` and `handled cat2`, exit 0.
- Both native builds panic with `compiler bug: a field the checker proved is there is missing`.
- The JavaScript backend panics reading `code`.

It's loud, but it's a sound everyday program that doesn't run. It may share a root with `method_value.a`,
already routed to B2.

`c01_tuple_same_type.a` is the open c3.a again: a tuple seen as an array. ASan reports a heap-buffer-overflow
in `adamic_array_set`, and -O2 segfaults.

### Held

- **Refused, correctly:**
  - conditionals and `??` whose type is `Dog[]`;
  - a spread element;
  - Map and Set;
  - a callback parameter, either way;
  - a field assignment, and `return`;
  - an optional field;
  - an intersection source;
  - a readonly-to-mutable `as`;
  - function-typed fields;
  - `Weak<Dog[]>`;
  - a class seen as a structural type;
  - an explicit type argument;
  - the arguments of `push`, `Map.set`, `fill`, `unshift` and `splice`;
  - a nested object;
  - a closure returning the alias;
  - an element write;
  - an array of functions;
  - destructuring assignment.
- **Compiled, not refused:** readonly arrays, ReadonlyMap and ReadonlySet, covariant as they should be. Of
  those, `a21_callbacks.a` and `a23_readonly_map_of_arrays.a` were run all four ways and agree.
- **Existing programs:**
  - All ten of docs/0.1.md's compile programs still compile.
  - No oracle fixture is refused.
  - `go test ./internal/lower` passes.
  - In the oracle, TestNativeAgreesWithNode passes, including `invariance_readonly.a`. The input fixtures
    failed only on the worktree permission issue described under B3.

### Not covered on B2

- No mutant was run against the new checks.
- The full gate wasn't run on 7340724.
- Stopped by NotYet, or refused for another reason, so the rule wasn't reached:
  - overloads past the first signature;
  - destructuring declarations and binding defaults;
  - `??=`;
  - `concat`;
  - arrays of a union element type.
- Not probed: index signatures and getters (both refused anyway), and `satisfies`.

### Fixed along the way

`review/fxspptb/probe.sh` lacked `-Wno-unused-parameter -Wno-self-assign`, which native.Build passes, so a
class method that doesn't use `this` failed `-Werror` under it. It now copies native.Build's flags again.
