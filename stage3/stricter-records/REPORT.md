Built readonly string-key record lowering for D129/D130/D131 through the existing runtime table and indexed presence guard.
Commits: base dfb82dabbf7b548de84a334ae5e72935bd08ce90; implementation cb0e4d61dd1a2be5401f59cfe6eaf75f8232e3f1; final proof/report is this branch-head commit.
Commands and outputs: record proofs 20.491s, lowering 15.145s, native package 179.381s, 36 uncached oracle fixtures 19.480s, vet exit 0.
Mutants: six single-site guard erasures caught in both native modes (12 builds/runs), one JS prototype-guard erasure caught, and the existing nine native record runtime mutants pass their detection tests.
Not covered: mutable records, recursive/nullable/tagged payloads, computed construction keys, reflection/spread, full repository gate and whole-program compiler build.

## Scope and merge contract

Branch codex/stricter-records starts at the newest fetched
origin/codex/stricter-options-checks, dfb82dab. The report and indexed guard at
5aef91ae were read before implementation. No second indexed guard or record
runtime was introduced. Sparse arrays and typed arrays remain with that worker.

The three blocked rows in codex/stricter-indexed-d's report are D129 at
src/compiler/program.ts:4157 and D130/D131 at :4160. They share
`options.paths[key][i]`, where paths is a readonly string index signature
holding string arrays. The witness reproduces d's receiver and key/index forms.

**d must expect two indexed-presence guards for this chained expression:** the
record key and the inner array element each need presence. Both AST reads start
at the same source column, so explain lists that location twice and reports
indexed-presence=2, trusted=0. The empty-array witness proves the inner guard;
a separate array-valued-record missing-key witness proves the outer guard.
Their erase mutants are independent. d's prior blocked flag and one-guard
assertion need updating when merging. This base also includes published sparse
support, so d's prior hole-refusal expectations need updating independently.

## Representation

IR Record distinguishes table objects from fixed-shape objects across aliases.
MapNew selects the existing adamic_record_new/define operations; MapGet selects
adamic_record_get. Counting, table ordering, slot lookup and iterative destruction
remain in runtime/record.c and the existing Map implementation. The shared
checkedIndexedRead adds a Record MapGet case and emits the same Coalesce panic
used for arrays and strings, evaluating receiver and key once.

JS construction uses Object.fromEntries and its lookup observes own keys. A
missing Object.prototype name stops with the existing runtime's own-only
contract, including the member-specific message. Own shadow keys and the
shorthand __proto__ data-property form work; prototype-setting literal syntax
refuses. RegExp named-group dictionaries retain their existing representation.

Conservative scope: one readonly string index signature, no named fields or
call/construct signatures, with nonnullable number/boolean/string values or
arrays of those scalars. Mutable tables and recursive payloads are refused until
index-slot invariance and reachability proofs support them. Arrays stored as
values still retain their ordinary mutable-slot invariance. Fixed-object/record
views, including nested views and casts, cannot reinterpret one representation
as the other. Mixed record unions, intersections and reflection remain unsupported.

## Observations

Seven source templates are materialized as .ts interoperability probes under
their own strict tsconfig with noUncheckedIndexedAccess off. Node executes that
same source with type stripping. These are requested interoperability witnesses,
not newly authored Adamic programs.

Number, boolean, string and array-valued records cover present and absent keys.
The shared D129-D131 template covers a present record key whose array element
is present or missing. Own toString data and shorthand __proto__ data properties
cover prototype boundary construction. Each required guard is checked by its
independently computed source location, CLI explain output and actual IR count.
Native C contains exactly one record lookup per read. JS and release/sanitized
native match Node on successful cases, and required misses have exact named
stderr and exit 70 where Node prints undefined. Successful counted sanitizer
builds have allocations equal frees, with zero regions; LeakSanitizer also runs.

An additional function/alias witness uses an allocated string payload and a
computed lookup key, preserves record identity, and matches Node in both
backends and native modes. Fifteen boundary witnesses observe Node first and
then require explicit refusal: mutable, nullable, undefined and recursive
records; mixed named fields; fixed/record and nested views; mutable array-value
widening; reflection; named reads; prototype-setting literal syntax; spread;
and direct/nested casts.

The inherited toString miss independently observes Node's typeof result
"function". Both compiled backends stop with exactly:
`adamic: panic: record member 'toString' is missing; records hold own keys only`.
This is the published runtime interface's explicit own-only boundary, not a
claim that inherited Node values are undefined.

## Every mutant run

| Erased guard | Release / sanitized mutant observation | Catcher |
|---|---|---|
| number record miss | exit 0, prints 0 | exact site stderr and exit 70 |
| boolean record miss | exit 0, prints false | same |
| string record miss | signal / sanitizer exit 1 | same |
| D129-D131 inner string-array miss | signal / sanitizer exit 1 | same; record guard retained |
| shorthand __proto__ record's ordinary miss | signal / sanitizer exit 1 | same |
| array-valued record miss | exit 0, prints undefined | same; no array lookup involved |
| JS prototype-member miss | exit 0, prints undefined | exact own-only stderr and exit 70 |

All erased native guards compile successfully with the same release/sanitized
options. Compiler/runtime source is untouched by these emitted-artifact mutants.

The full native package also runs the existing production-runtime mutants:
unsorted integer keys; UINT32_MAX classified as an array index; deleted iterator
keys yielded; overwrite-key release erased; stored key freed early; null own-slot
access; inherited membership restored; missing-member get silently returning
NULL; and member guarding moved onto an own-hit path. Node comparisons, exact
stop assertions and sanitizer/count checks catch them as documented in
docs/records.md. No new runtime implementation was copied.

## Commands, logs and limits

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/stricter-records-setup.log 2>&1
source /workspace/adamic-tools/env.sh
go test -count=1 -v ./stage3/stricter-records > /tmp/stricter-records-proof-final.log 2>&1
go test -count=1 -timeout 10m ./internal/lower > /tmp/stricter-records-lower-final.log 2>&1
go test -count=1 -timeout 10m ./internal/native > /tmp/stricter-records-native-final.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/(indexing|string_index|maps_and_text|objects|library_object_order|library_object_own|library_map|map_)' -count=1 -timeout 10m -v > /tmp/stricter-records-oracle.log 2>&1
go vet ./internal/load ./internal/lower ./internal/ir ./internal/native ./internal/javascript ./cmd/adamic ./stage3/stricter-options ./stage3/stricter-records > /tmp/stricter-records-vet.log 2>&1
git diff --check > /tmp/stricter-records-whitespace.log 2>&1
```

The other affected packages passed in the combined run: load 15.215s, ir 2.052s,
cmd/adamic 3.829s and stricter-options 28.718s; javascript has no package tests.
That first run exposed the RegExp group conflict and therefore failed native;
the conflict was fixed and the complete native package rerun successfully.
A concurrent retry had lowering killed and clang killed compiling the existing
24,038,583-byte random-RegExp fixture. cgroup memory.events recorded OOM kills.
Lowering and native were then rerun separately and passed, as listed above.
No resource-killed run is counted as verification.

Setup timing lines: Node ready 0.017s, Go 0.018s, submodules 0.053s, markdown
skip step 0.006s / ready 0.058s, clang ready 0.146s, Go build 39.389s, test
binaries deferred 39.494s, build cache warm 39.495s, done 39.519s. nproc=5,
cpu.max=400000 100000; Go 1.27.1, clang 20.1.8, Node v24.19.0.

Observed: the three blocked record rows' shared shape is supported and proven.
Inference: d can complete those rows after merging and updating its harness
expectations. This report does not claim d's unchanged harness passes, that the
whole TypeScript compiler lowers, or that mutable Record<string,T> is supported.
