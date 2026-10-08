# Typed arrays runtime contract

October 7, 2026: Uint8Array, Uint16Array, Int32Array and Float64Array. The C contract is
`internal/native/runtime/adamic.h`; reads return `adamic_maybe_number`, agreed
with the compiler so lowering can reuse plain-array read paths.

`adamic_typed_array` begins with `adamic_heap`, followed by element `kind`,
fixed element `length`, `data` and `owner`. Data is one flat buffer with 1, 2, 4 or
8 bytes per element, never boxed. An owning header has NULL owner; a subarray
has its own count, points into the original buffer, and retains its ultimate
owning header. Releasing the last header/view releases the buffer. Unlike
string sharing, even short or empty subarrays are views, never copies.

| Entry point | Contract |
|---|---|
| `adamic_typed_array_new(kind, length)` | Owned, zero-filled; JavaScript ToIndex of length, with RangeError panic for invalid length |
| `adamic_typed_array_from_numbers(kind, numbers)` | Owned; borrowed plain number[] copied with element conversion |
| `adamic_typed_array_get(array, index)` | Present number at an existing integer index; absent (`present = false`, number initialized to 0) otherwise; numeric -0 addresses index 0 |
| `adamic_typed_array_check_write(array, index)` | Compiler write check; rejects absent indexes |
| `adamic_typed_array_set(array, index, value)` | Checks index itself, then converts and writes |
| `adamic_typed_array_length(array)` | Element count as double |
| `adamic_typed_array_fill(array, value, start, end, has_start, has_end)` | Converts value, uses relative clamped indexes; returns borrowed self |
| `adamic_typed_array_sort(array)` | Stable numeric ascending, in place within the receiver range; borrowed self; Float64 -0 precedes +0 and NaNs come last |
| `adamic_typed_array_set_from(array, source, offset, has_offset)` | Same kind only; offset defaults to 0, ToIntegerOrInfinity, negative or insufficient room panics with RangeError; overlap behaves as a temporary copy |
| `adamic_typed_array_subarray(array, start, end, has_end)` | Owned view; relative clamped indexes, omitted end is length; start defaults to 0 at call site |
| `adamic_typed_array_iterate(array)` | Owned counted iterator retaining array |
| `adamic_typed_array_iterator_next(iterator, value)` | Reads next current element; true with output value, false when exhausted |

All arguments are borrowed. Constructors, subarray and iterate return one
reference that the compiler releases. Iterators must be released on normal,
early and exceptional exits. Fill returns no additional count. Optional numeric
arguments use explicit presence flags; an explicit undefined is lowered as absent.

Uint8 writes use ToUint8: truncate and wrap modulo 256, NaN and either infinity
become 0. Uint16 writes use ToUint16 with the same rules modulo 65536. Int32 writes truncate and wrap modulo 2^32 into signed range. Float64
stores the double unchanged, including negative zero and NaN. Out-of-range reads
are undefined. Out-of-range writes panic with the existing plain-array shape:
`adamic: panic: index <index> is outside an array of length <length>` and exit 70.
Node silently drops typed-array writes outside the buffer; the JavaScript backend
must insert this same Adamic check before its write.

ArrayBuffer, DataView, other element types, constructing views over another
view's buffer, and sharing across parallel tasks remain NotYet in the compiler.
This does not prohibit subarray on a subarray: it retains the ultimate owner
without growing an owner chain. No buffer property or buffer constructor is exposed.

## Original three-kind runtime validation

`TestTypedArrayRuntime` builds the C harness in release and ASan/UBSan modes,
with LeakSanitizer enabled on the finished sanitized run. Its stdout must
equal source Node's, including negative-zero spelling. The harness
also asserts one counted allocation per constructor and balanced heap values.
It covers all requested conversion edges through construction, writes, fill
and same-kind set; 121 relative start/end pairs per kind; overlapping copies
in both directions; nested, empty and surviving views; live iteration and
early iterator release. Invalid reads, write panic text and exit 70, constructor
ranges, set offsets, and same-kind rejection are checked separately. Node also
confirms that the numeric constructor and set failure probes are RangeErrors.

Run with the setup environment sourced:

```sh
go test ./internal/native -run '^TestTypedArrayRuntime$' -count=1 -timeout 10m > /tmp/typed-arrays-focused-final.log 2>&1
python3 internal/native/testdata/typed-arrays/run-mutants.py > /tmp/typed-arrays-mutants-final.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./internal/native ./internal/oracle > /tmp/typed-arrays-packages.log 2>&1
```

The runner restores every source mutation and writes individual logs under
`/tmp/adamic-typed-array-mutants`. All twelve independent mutants were caught,
with no build-warning failures:

| Mutant | Observed catcher |
|---|---|
| Missing conversion wrap | Node output difference and UBSan float-to-integer overflow |
| Missing write bounds | Exit-code assertion and ASan buffer overflow |
| Subarray copies | Node output difference, parent element 2 instead of 44 |
| Missing view retain | ASan use after free |
| Missing iterator retain | ASan use after free |
| Missing buffer free | LeakSanitizer |
| Overlap uses memcpy | ASan overlapping-copy diagnostic |
| Missing same-kind check | Expected panic absent |
| Missing set offset check | UBSan float-to-integer overflow |
| Missing set capacity check | ASan buffer overflow |
| Missing constructor length check | UBSan float-to-integer overflow |
| Fractional index accepted | Absent-read assertion |

This runtime harness does not compile typed-array Adamic programs: lowering,
native emission, JavaScript write checks and their NotYet cases belong to the
compiler unit. Allocation exhaustion, platform allocation ceilings, and parallel
sharing are not exercised. The size multiplication guard is defensive; valid
ToIndex lengths on the tested 64-bit target cannot overflow at these widths.

Toolchain: Go 1.27.1, clang 20.1.8, Node 24.19.0, Linux, `nproc` 5,
CPU quota 4. The initial setup cache-warming run failed with
`vet: internal/native/emit.go:86:30: emitter.requestABI undefined (type *emitter has no field or method requestABI)`.
Rerunning `bash cloud/setup.sh` on the stable runtime checkout passed: Go,
clang, Node and submodules ready at 0s, build cache warm at 97s, done in 97s.
Every build/test shell sourced `/workspace/adamic-tools/env.sh`. Setup logs
are `/tmp/typed-arrays-setup.log` and `/tmp/typed-arrays-setup-final.log`.

Final observations, after restoring all mutants:

```text
focused runtime: ok github.com/system-inc/adamic/internal/native 0.726s
mutant runner: all 12 caught; exit 0
complete native package, uncached: ok github.com/system-inc/adamic/internal/native 315.219s
complete oracle package, uncached: ok github.com/system-inc/adamic/internal/oracle 145.456s
focused owner/array oracle, uncached: ok github.com/system-inc/adamic/internal/oracle 1.032s
go vet ./...: exit 0, no diagnostics
gofmt -l cmd internal: exit 0, no output
git diff --check: exit 0, no output
```

The complete package gate includes the existing recorded-counts check; no
existing counts rows changed. The whole repository gate was not run under the
worker-gate allowance. Its other packages were compiled by setup and vetted.
Complete logs are `/tmp/typed-arrays-packages.log`,
`/tmp/typed-arrays-oracle-focused.log`, `/tmp/typed-arrays-vet.log`,
`/tmp/typed-arrays-format.log`, and the focused/mutant logs in the commands above.

## Uint16Array runtime and compiler extension

Uint16Array uses the generic header unchanged apart from its new kind enum.
The runtime width is two bytes and every numeric store uses ToUint16. Lowering
and both emitters recognize the kind through the existing typed-array paths.
`byteLength` is the fixed view length multiplied by element width for all four
supported kinds. Optional byteLength access remains NotYet.

The C harness and the two `typed_arrays_uint16*.a` fixtures hold construction,
stores, fill and set to Node at 0, 65535, 65536, -1, 1.5, NaN, -0, Infinity and
2^32 + 1. They also cover numeric reads, zero initialization, fixed length and
byteLength, bidirectional aliasing, relative fill, overlapping set in both
directions, live iteration, and a view surviving its owner's scope. The write
fixture pins the Adamic panic and independently verifies Node drops the write.

The Uint16-specific runner restores the runtime after each mutation:

```sh
python3 internal/native/testdata/typed-arrays/run-uint16-mutants.py > /tmp/uint16-mutants.log 2>&1
```

All three mutants were caught independently by the C harness and backend
oracles. Saturating stores at 65536 instead of wrapping and copying subarrays
produce stdout differences from Node. Removing bounds checks fails the panic
exit assertion in release and triggers an ASan heap-buffer-overflow under the
sanitizers. Individual logs are in `/tmp/adamic-uint16-mutants`.

Linux counts regeneration adds only the two new fixture rows. The successful
fixture finishes with 173 allocations and 173 frees; the stopped write records
5 allocations and 4 frees at its panic. This extension adds no mutable static
runtime storage; see `runtime-statics.md`. Allocation exhaustion and parallel
sharing are not exercised; the previously unsupported facilities stay NotYet.

Setup passed with tools and submodules ready at 0s, cache warm at 120s and done
at 120s, using the same tool versions and CPU settings recorded above. Every
build shell sourced `/workspace/adamic-tools/env.sh`.

Final Uint16 validation, after restoring the mutants:

```sh
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./internal/native ./internal/lower ./internal/oracle > /tmp/uint16-packages.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts > /tmp/uint16-counts.log 2>&1
go vet ./... > /tmp/uint16-vet.log 2>&1
gofmt -l cmd internal > /tmp/uint16-format.log 2>&1
git diff --check > /tmp/uint16-diffcheck.log 2>&1
```

Observed: native passed in 306.686s, lower in 25.986s and oracle in 136.195s.
The whole oracle package compares source Node, JavaScript emission, sanitized
native and release native, and checks Linux leaks on finished programs. The C
harness also runs release and ASan/UBSan with leak detection enabled. Counts
regeneration passed in 20.396s; vet, format and whitespace checks had no output.
The whole repository test suite was not run; the three requested packages were
run in full, uncached, and `go vet ./...` covered the repository.

## Numeric sort, October 8, 2026

All four kinds support `sort()` without arguments, mutating their own range and
returning the same header. Subarray sorting writes into the shared buffer and
leaves elements outside that range unchanged. Sort takes and returns a borrowed
reference, like fill; a temporary receiver and a returned view use the existing
compiler ownership paths. Comparator arguments remain NotYet with the precise
name `typed array sort with a comparator`. For this unit, every supplied argument,
including explicit undefined, stays behind that gap; the admitted form is `sort()`.

The implementation uses stable bottom-up merges specialized at each element
width, with one temporary unboxed buffer for a nontrivial sort. It chooses the
left run on equality, preserving equal elements, including NaN payload order.
The Float64 comparison distinguishes signed zero and puts all NaNs last. Neither
sort nor its workspace creates counted heap values or mutable static storage.
The workspace is freed before return; empty and singleton sorts allocate none.

The four `typed_arrays_sort_*.a` fixtures hold both backends to source Node. Each
covers duplicates and kind-specific limits, empty and singleton arrays, a view
and nested view with parent sentinels, identity, write-through, temporary
receivers, escaping views, and 1025 deterministic numbers. Float64 additionally
prints reciprocal values so signed zero order is observable. The C harness holds
all four kinds to Node at 13 lengths spanning 0 through 1025, along with existing
conversion edges, parent/view sorting, surviving views and heap-count balance.

The WASI oracle and emission checks run over these four new fixtures, with the
byte/exit runner mutants also enabled. That filtered WASI run passed in 7.874s,
with no opt-in skips. It does not rerun every older WASI fixture.

Workers' million-number sort benchmark and reproducible commands are in
`internal/native/performance/typed-arrays-sort/REPORT.md`. This unit does not
change Workers' route. Allocation exhaustion and buffer/shared-task facilities
that are already NotYet remain outside this unit.

All four requested sort mutants were caught independently by release/sanitized
C and backend oracles. Equal signed zeros gave `0` where Node gives `-0`;
NaNs first gave `NaN` where Node gives `-Infinity`; whole-parent sorting moved
sentinels; string-order comparison reordered 2 and 10. No compile failures or
sanitizer accidents count as these catches: each failed stdout agreement. A
fifth mutant bypassing comparator refusal is caught by the named NotYet assertion
for named and arrow comparators in every kind.

```sh
python3 internal/native/testdata/typed-arrays/run-sort-mutants.py > /tmp/typed-sort-mutants.log 2>&1
```

The four behavior mutants were run together; the comparator bypass was added
and run with `--comparator-only`. The complete runner now reproduces all five.
It restores sources even on failure and writes individual logs under
`/tmp/adamic-typed-sort-mutants`.

Linux counts regeneration passed in 97.885s. Only the four new rows are added;
every previous row is unchanged. Allocations/frees are 2245/2245 for Uint8 and
Uint16, 2251/2251 for Int32, and 2275/2275 for Float64; all have peak 14 and no
region values. These counts include the fixture's formatted output; sort itself
adds no counted allocation, as the C harness asserts.
