# Stream R, rounds three and four

Round three is cloud/integrate-5 (7cefeda): A2's control-flow graph and reuse in place, and B2's Weak
handles and cycle finder. Round four is cloud/integrate-6 (adbacc7): C's inline field and array reads, and
the fixes for my round-two findings, each checked against what it was meant to close.

integrate-6 doesn't contain integrate-5. It has B2's cycles, but not A2's flow and reuse. So the reuse
findings are integrate-5's, and the Weak findings were re-run on both.

I ran everything on Linux x86-64 with Go 1.27.0, clang 18.1.3 and Node v24.21.0. `review/fxspptb/probe.sh`
runs each probe four ways:

1. Node on the source.
2. Native under ASan, UBSan and LeakSanitizer.
3. Native as `adamic build` makes it (-O2), the binary that ships. This run is new in this round.
4. The JavaScript backend.

The probes in `review/integrate5` need integrate-5 or integrate-6 to compile. Weak and reuse aren't on this
branch's base.

The Weak and cycle-finder half was a sub-review, which I gave my own instructions. I re-ran its findings on
integrate-6 myself (below). Where only the sub-review ran something, it says so.

## Round three: confirmed findings, worst first

### Reuse in place (A2), on integrate-5

1. **A borrowed parameter is moved into a consumed one (`forward.a`).** `forward(p) { return bump(p); }`,
   where `bump` reuses its parameter through a spread. `movable` (reuse.go) moves `forward`'s borrowed `p`,
   whose count is the caller's. `bump` then sees count 1, reuses the caller's object, and releases it.
   - Sanitized: ASan heap-use-after-free.
   - -O2: prints `2 start1 2 start1`, exit 0. Node prints `1 start1 2 start1`.
   - Adding `local.Borrowed` to `movable`'s refusals makes the probe agree. I didn't run that change through
     the whole gate.
2. **A moved global is read by a later argument (`global_sibling.a`).** `tree = insert(tree, size())`,
   where `size()` reads `tree`. `tree` is NULL by the time `size()` runs: `touches` checks only the callee,
   never the call's other arguments.
   - Sanitized: UBSan null dereference at object.c:27.
   - -O2: segfault (exit 139).
3. **A Weak reaches an object without counting it (`weak_during_spread.a`).** `{ ...item, label: touch() }`,
   where `touch()` writes `count = 100` through a Weak to `item`, while a strong local still holds `item`.
   The plan's premise, "count 1 means nothing else reaches it", is false with Weak.
   - Node: `1 touched100`. JavaScript copies the spread's fields first.
   - Native: `100 touched100`, exit 0, sanitizers silent.
4. **After a takeover, the Weak reads the new object (`weak_after_reuse.a`).** Native prints `new2 true`:
   the Weak is now `===` the result. That is neither undefined (the documented native behavior after a free)
   nor Node's `old1`. Taking an object over never tells its handle.
   - A fix for 3 and 4: no reuse of an object that has a weak handle, or a takeover that tells the handle
     the way a free does.

These agree across all four runs, with no leaks:

- `consumed_paths.a`: a consumed parameter on early returns, in loops and in chains.
- `alias_local.a`: a local alias of a borrowed parameter takes its own count, as do for-of elements.
- `liveness_paths.a`: the source read again only after a switch's default, around a continue, past a
  break, or on a do-while's next iteration.

A spread that adds a field is NotYet. Casts that would show a subtype's extra field are refused.

### Weak and the cycle finder (B2): the sub-review's findings, every one re-run by me on integrate-6

5. **One object seen through two types whose same-named field is Weak in one and strong in the other.**
   tsc allows the assignment both ways (a plain Box assigns into a Weak). Lowering checks this mismatch only
   for arrays and maps (`keepsWeakly`, expression.go). For object fields and function types it checks
   nothing, so the same slot is read as a handle through one type and as a target through the other.

   | Probe | Node | Native, sanitized and -O2 |
   |---|---|---|
   | `weak/a_objview3.a` | `true` | `false`, exit 0, sanitizers silent (it compares the handle's address) |
   | `weak/a_objview.a` | `b1` | SEGV in `adamic_retain`; -O2 exit 139 |
   | `weak/a_fnview.a`, `(x: Weak<Box>) => ...` seen as `(x: Box) => ...` | `true` | SEGV in `adamic_retain`; -O2 exit 139 |

   I didn't re-run `weak/a_objview2.a` or `weak/a_fnview2.a` (the return direction), which the sub-review
   reports as SEGVs.
6. **A Weak array narrowed to non-Weak elements.** tsc infers a type predicate for `x => x !== undefined`.
   The element type becomes `Box & WeakBrand`, which lowers as a plain Object, though the elements are still
   handles.

   | Probe | Node | Native |
   |---|---|---|
   | `weak/a_filter2.a`, through `filter` | `1 0 true` | `1 -1 false`, exit 0 |
   | `weak/a_every.a`, narrowing the array in place through `every` | `true` | `false`, exit 0 |
   | `weak/a_filter.a`, reading a field after the filter | `2`, `a1`, `b1` | SEGV in `adamic_object_find` |

7. **The cycle finder misses cycles through generic classes, and they leak.** The finder walks the types
   written in the source. Inside a generic class those are `T`, which reaches nothing. B2's
   every-instantiation fix in integrate-6 didn't close these two:

   | Probe | Shape | Output | Leak |
   |---|---|---|---|
   | `weak/b_genericclass2.a` | node → `Box<Node>` → node, through a generic maker | `3` on every side | LeakSanitizer 130 bytes in 3 allocations; `--count` 9 allocations, 6 frees (sub-review's) |
   | `weak/b_generictie.a` | a captured `let cell: T` instantiated with a closure | `1` on every side | 306 bytes in 8 allocations |

8. **The compiler crashes on a narrowed `Weak<Map>` (`weak/a_map4.a`).** `mm.has('k1')` after narrowing
   panics in `checker.getTypeArguments`, from `mapTypes` (object.go:903). The type there is
   `Map<...> & WeakBrand`. The array case is a proper NotYet.
9. **integrate-5's own gate is red.** `go test ./internal/flow`: TestEveryMutationIsInItsRange fails on
   B2's `doubly_linked.a` with "List_insertAfter: this$5 was mutated at instruction 2, outside its range
   [6, 11)". A2's mutation ranges don't see B2's program; I didn't dig into why. This may be the aliasing
   finding you already have from stream B.

### What held, from the sub-review's runs

- Freeing a target, then allocating 2000 objects of the same size, never let a stale handle read one.
- Handles dropped and remade, two handles to one target, a target dying together with the holder of its
  handle, and a narrowed value kept across a call that frees the target: all agreed.
- Pointer hiding is sound on 64-bit, by reasoning; it was not run on 32-bit.
- The finder rightly refused: a link through parameters, `this` captured by a stored callback, a union
  field, a spread, a readonly field holding a mutable array, and a concrete `Box<Node>`.

Two mutants:

- One survives the weak tests: `remove_entry` writing EMPTY instead of TOMBSTONE. The sub-review's
  `a_probechain.a` catches it under UBSan.
- Reads of a Weak after its target is freed differ from Node by design. Only weak_test.go holds them, and
  no compiler message marks a program that does it.

## Round four: integrate-6

### C2's fixes, checked against what each was meant to close

| Round-two finding | On integrate-6 |
|---|---|
| String() of 46 powers of two | **Closed.** `power_of_two_string.a` agrees on all four runs. |
| Array.from passing 0 for undefined | **Closed.** `from_undefined.a` agrees. |
| Tail call with no stack check | **Closed.** `stack_forever.a` now panics like Node, sanitized and -O2. My probe.sh had copied native.Build's flags and lacked the new `-fno-optimize-sibling-calls`, which made this look unfixed until I corrected it. |
| normalize and padStart with an undefined string argument | **Closed.** Both agree. Every optional number argument left undefined (lastIndexOf, endsWith, slice, splice, toString, toFixed, fill, split and more) is NotYet, so no trap opens there (`review/integrate6/optional_numbers.a`). |
| Globals' leaks hidden by keeping their pointers | **Closed** (cleared after release). `global_cycle.a` is now refused by the finder. |
| counts.md's stack_overflow.a row moving with `ulimit -s` | **Closed** (pinned to 8 MiB). |
| Concatenation's length check untested | **Closed** (concat_too_long.a). |
| Large arguments segfaulting | **Closed at 8 MiB, but the fix opened a regression at 1 MiB** (finding 10). |
| Native recursion deeper or shallower than Node's | **Still open.** Not claimed fixed: `stack_deep.a` and `stack_tail_call.a` still print results where Node panics. |
| readTextFile's limit, which Node counts in bytes | **Still open.** input.c has no check. |
| A method read as a value | **Still open.** `method_value.a` still panics "compiler bug". |

### Confirmed findings

10. **The stack fix turned the check off at 1 MiB (`review/integrate6/stack_small.a`).** stack.c now keeps
    `size / 4 + 256 KiB` back, and sets no limit at all unless `size > 2 * reserved`. At 1 MiB that's
    1 MiB > 1 MiB: false. Under `ulimit -s 1024`, a recursion with no end:

    | Run | Result |
    |---|---|
    | integrate-5's native build | panics, exit 70 |
    | integrate-6's native build | segfaults, exit 139 |
    | Node | panics, exit 70 |

    At 1100 KiB and above, all agree.

    One more limit that reserve doesn't cover: Linux allows arguments and environment
    `max(stack / 4, 128 KiB)`, so below 512 KiB they can take more than a quarter.
11. **integrate-6's own gate is red: C's class_layouts.a is refused by B2's cycle finder.** Its
    `Point.next: Point | undefined` is a cycle-capable field. Each stream's branch was green alone. Merged,
    `TestNativeAgreesWithNode/class_layouts.a` fails to lower, so the fixture holds nothing.
    `review/integrate6/layouts.a` covers the same ground with no cyclic field:
    - shapes from two classes in different field orders;
    - literals in both orders;
    - a packed `number | undefined` field read through the class.

    It agrees on all four runs on integrate-6. With the shape compare removed from `fieldSlot`, it fails
    (`5 string 10 4` where Node prints `6 string 4 3`). class_layouts.a can't show that while it doesn't
    lower.

### What held

- C's `adamic_array_at`: past the bounds check, `(size_t)index` back as a double is trunc(index) exactly.
  NaN and -0 are right.
- C's `fieldSlot`: shapes are keyed by field names and which fields are references, so a key shared by
  two layouts means the same order. Anything else goes by name.

### Gate on integrate-6

- gofmt clean, vet clean.
- load, lower and native pass.
- `./internal/oracle` hit Go's 10-minute timeout. The sub-review's runs were loading the machine; nothing
  had failed by then.
- stage1's gitignore test failed in my worktree only. In a real checkout of integrate-6 (a real submodule,
  not my symlink) it passes: exit 0, 23.7 s.
- class_layouts.a fails as in finding 11.

## Not covered

- Nothing ran on macOS, arm64 or 32-bit.
- The flow package wasn't read line by line, and I didn't find why finding 9's range is wrong.
- Multi-module Weak programs, destructuring into captured cells, and Weak inside a generic instantiation
  were not tried.
- The JavaScript backend's new code was only run, not read.
