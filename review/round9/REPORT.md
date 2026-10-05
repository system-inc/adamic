# Stream R, round nine: stream C's integer loop counters, and C2's collections and file system

I ran everything on Linux x86-64 with Go 1.27.0, clang 18.1.3 and Node v24.21.0. `review/fxspptb/probe.sh` runs
each probe four ways:

1. Node on the source.
2. Native under ASan, UBSan and LeakSanitizer.
3. Native as `adamic build` makes it.
4. The JavaScript backend.

## Stream C's integer loop counters: not found

cloud/hx74vgw-loops at 6881bd3 carries no integer loop counters. Its commits past main are the size-class
allocator, the chunk sharing (f5ac483, 6881bd3) and a counts change. No commit or emit.go on any remote branch
mentions a loop counter. I'll review them when they're pushed.

The chunk sharing, read: a chunk is unlisted before it goes to the spares. Its free list, fresh pointer, live
count and class are reset when another class carves it. A class keeps its last chunk. Header numbers are
checked against UINT32_MAX. No finding.

## C2's collections (cloud/collections-fs; a sub-review at 2c512d0, re-checked by me at d67e4c5)

### Findings

1. **A tuple used where an array goes, read as an array: a silent miscompile, new on this branch.** `c3.a`:

       const q: readonly number[] = [3, 4] as [number, number];

   Node prints `2` for `q.length`. Native prints `94496053762432` sanitized and `94917759917104` at -O2, exit 0,
   with no sanitizer report: the array code reads the tuple object's shape pointer as its length. I re-ran this
   on d67e4c5 myself.

   Indexing or joining such a value (`c1.a`, `c2b.a`) is an ASan heap-buffer-overflow, at `array.c:73` and
   `adamic.h:259`. At -O2 it segfaults.

   A `[string, string]` variable passed to a `readonly string[]` parameter (`c2a.a`) reaches clang as bad C.

   On main all three are refused ("can't lower a value of type [string, string] where an array goes yet"). On
   this branch, `arrayLiteral` (lower/object.go:100) lowers any literal the checker types as a tuple to an
   object, even where the value flows into an array type, so main's check (object.go:163) no longer runs.
2. **Array methods on a tuple fail at run time.** `p.join(",")` and `p.push(...)` on a `[string, number]`:
   - native panics "compiler bug: a field the checker proved is there is missing", exit 70;
   - the JavaScript backend gives `TypeError: closure.code is not a function`, exit 70;
   - Node prints `x1,2`.

   Not silent, but it should be NotYet when compiled. This is the sub-review's run; I didn't re-run it.
3. **A Set or Map keyed by `string | undefined` dereferences NULL (on main too).** `a5.a` builds a Set from a
   `(string | undefined)[]` that holds undefined:
   - Node prints `2` and `2 true`;
   - native gives UBSan "member access within null pointer" at map.c:49, and -O2 segfaults.

   I re-ran this myself. `mapTypes` accepts the key type because `string | undefined` lowers to ir.String, and
   the hash dereferences the NULL. Writing `m.set(undefined, ...)` directly is refused, but values from an
   array slip through. Separately, `return undefined` from a function returning `string | undefined` (`a4.a`)
   reaches clang as bad C, on main and the branch.
4. **Memory grows with every insertion while an iteration is open.** `a7.a`, a worklist that deletes one
   element and adds one, never holds more than one:

   | Steps | Native peak | Node peak |
   |---|---|---|
   | 3M | 137 MB | 71 MB |
   | 6M | 270 MB | 71 MB |

   The output is right. `adamic_map_set` and `rebuild` grow the table rather than compact while an iteration
   is open. This matters for cohere's and tsc's worklists.
5. **Small integer keys all hash to one home bucket (on main too).** `(bits ^ bits >> 29) * prime` leaves
   their low bits zero. `a9.a` takes 0.57 s natively and 0.10 s on Node; the same program with non-integer
   keys (`a9b.a`) takes 0.037 s. Speed only.

### Mutant that survives the whole native oracle

M6, `clear` not releasing the values, passes all of `TestNativeAgreesWithNode`: no fixture clears a map whose
values are references. `a2.a` catches it with LeakSanitizer, 824 bytes in 20 allocations. A fixture needs a
`clear()` on a `Map<Box, Box>`.

Eight other mutants are caught by the branch's fixtures: clear resetting during iteration, compacting during
iteration, set keeping the passed key, delete not releasing the key, forEach not holding the key, a Set's
forEach argument order, an iterator visiting deleted entries, and identity keys not freed.

### What held

These agree on all four runs with no leak: insertion order after delete and re-add; deletes of the current,
an earlier and a later entry, and adds, during for...of and forEach; clear during iteration; nested loops over
one map while deleting; -0 and NaN as keys; forEach's (value, key, map); identity keys of every kind, including
one freed after delete with 200 allocations after; the map dying mid-loop, or reassigned in its own forEach;
the table growing 200 entries mid-iteration; return from three nested iterations.

## C2's file system (readDirectory, fileStatus)

These landed on the branch after the sub-review above ran (c57f6f2, d67e4c5). They're under review now, and
this section follows.
