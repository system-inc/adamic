# Typed arrays runtime contract

October 7, 2026: Uint8Array, Int32Array and Float64Array only. The C contract is
`internal/native/runtime/adamic.h`; reads return `adamic_maybe_number`, agreed
with the compiler so lowering can reuse plain-array read paths.

`adamic_typed_array` begins with `adamic_heap`, followed by element `kind`,
fixed element `length`, `data` and `owner`. Data is one flat buffer with 1, 4 or
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
| `adamic_typed_array_set_from(array, source, offset, has_offset)` | Same kind only; offset defaults to 0, ToIntegerOrInfinity, negative or insufficient room panics with RangeError; overlap behaves as a temporary copy |
| `adamic_typed_array_subarray(array, start, end, has_end)` | Owned view; relative clamped indexes, omitted end is length; start defaults to 0 at call site |
| `adamic_typed_array_iterate(array)` | Owned counted iterator retaining array |
| `adamic_typed_array_iterator_next(iterator, value)` | Reads next current element; true with output value, false when exhausted |

All arguments are borrowed. Constructors, subarray and iterate return one
reference that the compiler releases. Iterators must be released on normal,
early and exceptional exits. Fill returns no additional count. Optional numeric
arguments use explicit presence flags; an explicit undefined is lowered as absent.

Uint8 writes use ToUint8: truncate and wrap modulo 256, NaN and either infinity
become 0. Int32 writes truncate and wrap modulo 2^32 into signed range. Float64
stores the double unchanged, including negative zero and NaN. Out-of-range reads
are undefined. Out-of-range writes panic with the existing plain-array shape:
`adamic: panic: index <index> is outside an array of length <length>` and exit 70.
Node silently drops typed-array writes outside the buffer; the JavaScript backend
must insert this same Adamic check before its write.

ArrayBuffer, DataView, other element types, constructing views over another
view's buffer, and sharing across parallel tasks remain NotYet in the compiler.
This does not prohibit subarray on a subarray: it retains the ultimate owner
without growing an owner chain. No buffer property or buffer constructor is exposed.

## Runtime validation

`TestTypedArrayRuntime` builds the C harness in release and ASan/UBSan modes,
with LeakSanitizer enabled on the finished sanitized run. Its 4,045 stdout
lines must equal source Node's, including negative-zero spelling. The harness
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
