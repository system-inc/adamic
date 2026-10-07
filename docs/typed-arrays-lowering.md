# Typed arrays: compiler claim

Compiler owner: codex/typed-arrays. Runtime owns docs/typed-arrays.md and
internal/native/runtime/adamic.h on codex/typed-arrays-runtime. This claim
precedes implementation. Native emission waits for runtime's interface commit;
its published names and signatures supersede the provisional names below.

## IR and lowering

Add counted representations Uint8Array, Int32Array and Float64Array. Each
identifies the width and conversion rule without boxing elements. Add
TypedArrayNew (kind, length or number[] source), TypedArrayFill (receiver,
value, optional start/end), TypedArraySet (receiver, same-kind source, optional
offset), and TypedArraySubarray (receiver, optional begin/end). Preserve
JavaScript evaluation order, evaluating each receiver and argument once.

Reuse ArrayIndex with Element Number, SetIndex with Element Number, Length,
and ForOf. Index reads return MaybeNumber, including undefined past the end,
and reuse the existing optional-number narrowing and array-read paths.
Writes keep the existing loud index check and its text:
`index <index> is outside an array of length <length>`, exit 70.

## Provisional seam claimed before runtime published its interface

Expect a counted `adamic_typed_array *` for each representation, with a flat
buffer of its element width. Proposed operations (runtime's header decides
final spelling):

- `adamic_typed_array_new(kind, double length)`: owned, zero-filled result;
  JavaScript ToIndex length conversion and validation.
- `adamic_typed_array_from(kind, adamic_array *source)`: borrowed number[]
  input, owned result, conversion per element.
- `adamic_typed_array_at(array, double index)`: borrowed receiver,
  `adamic_maybe_number` result by value, no element allocation.
- `adamic_typed_array_set_index(array, double index, double value)`: borrowed
  receiver, checks bounds before storing with the kind's JavaScript conversion.
- `adamic_typed_array_length(array)`: borrowed receiver, numeric length.
- `adamic_typed_array_fill(array, value, start, end)`: borrowed receiver,
  mutates, JavaScript relative-index normalization; compiler returns receiver
  with ordinary ownership when the expression's result is kept.
- `adamic_typed_array_set(array, source, offset)`: both receivers borrowed,
  same kind, overlap-safe copy; JavaScript offset conversion and range failure.
- `adamic_typed_array_subarray(array, begin, end)`: borrowed receiver, owned
  counted view retaining the buffer owner, always shares, including nested and
  empty views. Omitted end means length; compiler preserves omitted arguments.

All ordinary reference paths retain/release the view through adamic_retain and
adamic_release. ForOf holds the receiver for the loop, loads doubles without
boxing, and releases it on normal and abrupt exits. Integer stores use ToUint8
or ToInt32; Float64 stores the double, including negative zero and infinities.

## Published runtime seam

Runtime published interface `3a30161c889dd3d4a4d3641c70d1356ba7cad55b`
and implementation `4f47cc42234989247d23319744a1f764efcb0fdc`. Native
emission began after those commits were fetched and read. Their changes alone
were cherry-picked onto this compiler branch as `5a1db80` and `c829511`;
the committed compiler implementation does not alter runtime's header,
contract or implementation.

The agreed header replaces the provisional names above:

| IR operation | C call and result |
|---|---|
| TypedArrayNew, length | `adamic_typed_array_new(kind, double length)`, owned pointer |
| TypedArrayNew, number[] | `adamic_typed_array_from_numbers(kind, const adamic_array *source)`, owned pointer |
| ArrayIndex, numeric element | `adamic_typed_array_get(const adamic_typed_array *array, double index)`, `adamic_maybe_number` by value |
| SetIndex | `adamic_typed_array_set(adamic_typed_array *array, double index, double value)`, void; performs the loud check itself |
| Length | `adamic_typed_array_length(const adamic_typed_array *array)`, double |
| TypedArrayFill | `adamic_typed_array_fill(array, value, start, end, has_start, has_end)`, borrowed self |
| TypedArraySet | `adamic_typed_array_set_from(array, source, offset, has_offset)`, void |
| TypedArraySubarray | `adamic_typed_array_subarray(array, start, end, has_end)`, owned pointer |
| ForOf | `adamic_typed_array_iterate(array)`, owned iterator; `adamic_typed_array_iterator_next(iterator, double *value)`, bool |

The kinds are `adamic_typed_array_uint8`, `adamic_typed_array_int32` and
`adamic_typed_array_float64`. All receiver/source arguments are borrowed.
Optional offsets are evaluated once into MaybeNumber temporaries; their
presence flags distinguish omitted/undefined from zero. Subarray's absent
begin becomes zero. Ordinary reference scopes retain/release arrays, views
and loop iterators, including return, break, continue and throw.

Freshness and alias analysis recognize all four new expression nodes. A view
conservatively aliases its parent's identity, and fill/set mutate that identity.
Numeric stores cannot introduce reference cycles. Structural assignments and
casts that erase typed-array storage are NotYet; set's library ArrayLike
parameter is a protocol handled directly, with an exact same-kind check.

## JavaScript backend

TypedArrayNew emits `new Uint8Array(source)`, `new Int32Array(source)` or
`new Float64Array(source)`, where source is length or number[]. ArrayIndex
emits ordinary bracket access. SetIndex uses existing `adamicSetIndex`, which
checks the index then lets the JavaScript typed array convert the store.
Length emits `.length`; fill, set and subarray emit the corresponding built-in
method with the original optional arguments. ForOf uses the existing backend loop over a captured receiver and its fixed
length, reading current elements each step.
Built-in subarray preserves shared writes in both directions.

## Refusals and proof

Give separate NotYet messages for ArrayBuffer, DataView, construction over
another view's buffer, resizable buffers, unsupported element types, and
sharing typed arrays across parallel tasks. Unsupported members and detached
methods must also stop during lowering, never become generic object reads.

Run Node/native/JavaScript oracles for zero fill, conversion edges (300, -1,
NaN, 2^31, -0, Infinity), reads, fill, overlapping same-kind set, relative
subarrays, nested-view lifetimes, alias writes and for...of. Pin the deliberate
out-of-range write stop separately, and pin Node's silently dropped write.
Run Workers primes and stats with Uint8Array and Float64Array using the sieve
at origin/codex/workers-bench-handler 0a3e651. Run missing-wrap,
missing-bounds-check and copying-subarray mutants and report what caught each.

## Proof boundaries

The Uint8Array sieve is adapted from `bench/workers/handler/sieve.a` at
`0a3e6511691f98bdadba30a6792d18d3a0e54b76`. The small fixture covers limits
0, 1, 2, 10 and 100. The large oracle also covers 10,000 and 100,000. The
flow trace gate excludes only the large fixture because printing the entire
buffer at every marked point exceeded its five-minute limit; it traces the
same algorithm in the small fixture.

That Workers revision contains `statistics.mjs`, a best/median timing helper,
not a numeric stats handler. The Float64Array fixture ports its statistics
operation, using insertion sort because typed-array sort is outside this unit.
Odd/even samples, fractions, negatives, repeated samples, preserved inputs and
subarray inputs are checked. This proves the port, not a deployed Workers
handler or a performance comparison.

Runtime's agreed interface panics on invalid constructor lengths and set
ranges. Like the existing library policy, catches around potentially failing
length conversion or set are NotYet. Explicit user throws inside typed-array
loops are supported and checked. Main has no parallel-task capture API to
integrate; SharedArrayBuffer construction and annotations get the named
sharing-across-parallel-tasks NotYet. Actual task capture remains untested.

The pinned cohere CLI cannot include `.a` in its program, despite the repo's
sourceExtensions setting. Its named-fixture gate exits 1 with "nothing to
check" and "not a TypeScript or JavaScript file". Adamic's own checker and
Node oracle do load and check these `.a` fixtures. No cohere source was copied.

## Verification commands and observed results

Toolchain setup was `bash cloud/setup.sh`, followed by
`source /workspace/adamic-tools/env.sh`. Its timing lines were: Go ready 0s,
clang ready 1s, Node ready 1s, submodules ready 1s, build cache warm 117s,
done 117s. Versions: Go 1.27.1, clang 20.1.8, Node 24.19.0. `nproc` printed
5; cgroup cpu.max was `400000 100000`.

All test commands wrote directly to log files. The final commands were:

```sh
gofmt -l cmd internal
go vet ./...
go test ./internal/lower ./internal/flow ./internal/fresh -count=1 -timeout 30m
go test ./internal/ir ./internal/native ./internal/javascript -count=1 -timeout 30m
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestTypedArrayWriteStopIsPinned|TestNativeAgreesWithNode/internal/oracle/testdata/typed_arrays_' -count=1 -timeout 30m
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m
```

The filtered oracle passed in 3.540s, count recording in 12.836s, and the normal counts comparison in 13.119s. Complete
lowering, flow and freshness packages passed in 26.517s, 58.444s and 34.188s.
IR passed in 0.530s and native in 73.698s; JavaScript has no standalone
package tests and is exercised by every typed-array oracle. Formatting and
vet printed nothing and exited 0. The entire repository `go test ./...` gate
was not run; the complete touched packages and filtered uncached oracle were run instead. Stage-1 suites and
performance comparisons are outside this proof. Logs live in
`cloud/reports/typed-arrays/`.

| Temporary mutant | Actual catcher and observation |
|---|---|
| Uint8 stores clamp instead of wrap | Node differential fixture: 300 produced native 255 instead of Node 44; -1 produced 0 instead of 255; both exited 0 with no sanitizer/compiler failure |
| Native invalid writes return silently | Independent stop pin: sanitized and release native exited 0 and printed after the write instead of exit 70 and the pinned diagnostic |
| JavaScript typed writes bypass adamicSetIndex | Same independent pin: JavaScript exited 0 and printed after the write |
| Subarray allocates/copies | Node alias fixture: native `alias 2 3 3` instead of `alias 44 255 3`, and parent Float64 write remained NaN instead of 3.25; both exited 0 |
| Typed-array structural representation guard removed | Refusal test accepted all three erased views, including cast and nested readonly field, with nil error |
| Length/set catch guards removed | Gap test accepted both catches with nil error |

Every mutant exited the test command with status 1. The original runtime and
compiler files were restored after each mutant; the final uncached oracle and
package runs use the restored files. The deliberate write-stop fixture has its
own assertion, and separately pins Node's silent drop, so agreement between
two broken backends cannot validate the exception.
