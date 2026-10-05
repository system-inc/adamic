# Stream R2, round three: integrate-7 (stream A's fuzzer and what it changed) and stream C's allocator

Reviewed `origin/cloud/integrate-7` at 7774d94: A's merge 047cb0d (TimSort, the narrowing check in both backends, splice, map/find over a shrinking array), and C2's review fixes. Also `origin/cloud/hx74vgw-alloc` at 55a5755, as it stood when I read it (four commits: the size-class allocator, inline length and charCodeAt, the direct path for whole numbers, ascii_scan.a). Nothing on either branch or on main was changed. The probes are in `review/r2/i7/` and `review/r2/alloc/`, run three ways with `./review/r2/run3.sh` from a checkout of the branch named.

## Confirmed findings

### 1. A narrowed number | undefined field, put back to undefined by a call, reads as NaN natively (pre-existing; A's check misses it)

A's narrowing check covers a local of a Maybe type (`ir.Unwrap`) and a reference read through a field or a local (`ir.Defined`). A field whose type is `number | undefined`, narrowed to `number`, is neither, so its read isn't checked, and the packed undefined (a reserved NaN) reads as NaN.

`i7/narrowed_field_silent.a`:

```ts
class Counter {
	count: number | undefined = 5;
}
const counter = new Counter();
function reset(): void {
	counter.count = undefined;
}
if (counter.count !== undefined) {
	reset();
	const seen = counter.count;
	console.log(`field ${seen} ${Number.isNaN(seen)} ${seen > 0 ? 'positive' : 'not positive'}`);
}
```

```
== node     exit=0  field undefined false not positive
== native   exit=0  field NaN true not positive
== backend  exit=0  field undefined false not positive
```

The release build (`adamic build`) prints the same as native. The same probe on main at 2385966 gives the same three answers, so this is older than A's merge, and the fix didn't reach it. A `Property` read whose declared field type is `number | undefined` and whose type at the read is `number` needs the same check `Unwrap` has.

### 2. The narrowing check panics where JavaScript reads undefined and goes on (A's check, both backends)

`ir.Defined` and `ir.Unwrap` panic at every read of a narrowed variable a call has put back to undefined. That includes reads into places typed to hold undefined, where JavaScript neither throws nor computes with a value of the wrong type.

`i7/narrowed_reads_undefined.a`: after `if (chain !== undefined) { drop(); ... }`, with `chain: Link | undefined`:

| The read | Node | native and backend |
|---|---|---|
| `typeof chain` | `typeof undefined` | panic |
| `describe(chain)`, a parameter typed `Link \| undefined` | `passed none` | panic |
| `[chain]` into `(Link \| undefined)[]` | `kept 1 true` | panic |

Both backends agree, so the oracle would see this only on a fixture, and then as native and Node differing. `narrowed_number_field.a`'s last line (`${loose}` of a narrowed `number | undefined` local) panics too, where Node prints `local undefined`. That one is the stated design: "JavaScript would go on with undefined in a place typed not to hold it". The three in the table are not: each goes to a place typed to hold undefined. A fix in the spirit of `comparedWithUndefined`: no check where the read is `typeof`'s operand, or is converted to a type with undefined in it (an argument, an element, an assignment or a return of such a type).

### 3. Reading a narrowed number | undefined array element reaches clang as C it refuses (pre-existing)

`i7/narrowed_element_plain.a`, with no call at all:

```ts
const slots: (number | undefined)[] = [4];
if (slots[0] !== undefined) {
	const seen = slots[0];
	console.log(`element ${seen + 1}`);
}
```

```
error: initializing 'double' with an expression of incompatible type 'adamic_maybe_number'
```

The same happens on main at 2385966 and on integrate-7. The checker narrows an element access with a literal index, and the lowering types the read `number` while the value it emits is still the pair. It's not a miscompile, but it's the thing the doctrine says a NotYet must never do. `narrowed_other_slots.a` is the same read after a call.

### 4. The allocator keeps each size class's peak: about five times main's memory when a program moves between sizes (C)

Chunks are never given back, and a free slot only serves its own class. So a program whose working set moves from one size to another holds the peak of every class it passed through. `alloc/size_class_churn.a` builds 20,000 objects with strings at one size, checks and drops them, then does the same at the next size, through 24 sizes from 1 to 300 bytes. Both builds are release builds (`adamic build`), and their stdout is identical to Node's (checked with `cmp`):

```
churn-slabs  (cloud/hx74vgw-alloc): peak 51,504 KB, 0.76 s
churn-malloc (main, 4ddd17f):       peak 10,600 KB, 0.79 s
```

The control, `alloc/one_size_churn.a` (the same churn at one size every time), shows what the allocator is for: 10,048 KB against 9,980 KB, and 0.20 s against 0.40 s, twice as fast. It's a cost, not a wrong answer. The comment in heap.c says it ("its memory is its peak"), but it's per class, which the comment doesn't say. It's worth measuring on stage 1's slices before it lands, against the goal of a program leaner than Go's. Peaks were read with `getrusage(RUSAGE_CHILDREN)` (`ru_maxrss`), since this container has no `/usr/bin/time`.

## What turned out fine

- **TimSort against V8's (A).** `i7/timsort_sweep.a` sorts arrays of every length from 0 to 300, and of 511, 1,024, 2,049 and 5,000. There are four shapes of data (random with duplicates, mostly sorted, descending, alternating runs of 16), each under four comparators: consistent, random answers, NaN for some pairs, and one that shrinks the array while it sorts. Each sort hashes the order of the comparator's calls and the result. All 4,880 sorts agree with Node call for call, natively and in the backend. **Mutant:** `minimum_gallop` 8 instead of 7 in sort.c changed native's hashes from length 64 on, and nothing else. Round one's NaN probe (`probes/sort_nan_plain.a`) now matches too.
- **C2's review fixes**, on integrate-7. Round one's probes all pass now. writeTextFile to /dev/stdout and /dev/stderr keeps the order. The prompt is shown before the read of stdin. SIGTERM, SIGINT and SIGHUP each leave the 39 bytes Node leaves, with the same exit status Node has (143, 130, 129).
- **The narrowing check where it fires as designed**: the fixtures A added. I didn't reprobe them beyond findings 1 to 3.
- **The allocator's correctness (C).** The churn program, and a check of every place a heap value is made or freed. All nine allocation sites go through `adamic_allocate`, nothing frees a heap value but `deallocate`, emit.go allocates nothing itself, and the `slab` field fills what was padding (the header stays 16 bytes). heap_test.go's slab build catches a use after a free as use-after-poison.
- **Whole numbers written directly (C).** `alloc/whole_numbers.a`: powers of ten to 1e22 and of two to 2^60, their neighbours and negatives, 2^53 - 1 to 2^53 + 2, a half below 2^52. All three sides and the release build match Node. The sign and zero are handled before the new path, so its `(uint64_t)` conversion only ever sees values from 1 to 2^53.

## What I didn't cover

- The fuzzer itself (internal/fuzz, cmd/adamic-fuzz): what it generates and whether its shrinker keeps a failure. Read in passing, not run.
- P2's value parser slice (stage1/cohere/values) on integrate-7: not reviewed this round.
- The inline charCodeAt and length on the allocator branch: read (the guard is ASCII, counted, and an index from 0 below the length; everything else goes out of line), and covered by round one's string index probe only through main's code, not rerun on that branch.
- macOS, and the allocator's behaviour under memory pressure.
