Built finite nested readonly records for D119 using the existing representation and presence guard; chained reads check each level once.
Commits: parent 28d30cd3b4924f975049aa686523de975f6a9cdb; extension 96e1eafc2d9c50fca730102305b368e3483b342f; final report is this branch-head commit.
Commands and outputs: record proofs 23.896s, all lowering tests 15.557s, record runtime tests 30.502s, eight uncached Node oracle fixtures 1.318s, vet and diff check exit 0.
Mutants: outer record, inner record and inner array guard erasures each caught in release and sanitized builds; all six print undefined with exit 0; previous guard and runtime mutants also pass.
Not covered: mutable or cyclic records, expanding generic graphs, more than 64 record levels, nullable/tagged payloads, d's own harness, full repository gate or whole-program compiler build.

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
call/construct signatures, with finite nested readonly records, nonnullable number/boolean/string values or
arrays of those scalars. Nesting beyond 64 record levels is conservatively refused.
Mutable tables and cyclic or expanding generic payloads are refused until
index-slot invariance and reachability proofs support them. Arrays stored as
values still retain their ordinary mutable-slot invariance. Fixed-object/record
views, including nested views and casts, cannot reinterpret one representation
as the other. Mixed record unions, intersections and reflection remain unsupported.

## Observations

The initial seven source templates are materialized as .ts interoperability probes under
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


## D119 nested payload extension, October 8

The latest request pins D119 on codex/stricter-indexed-d at 71897d7e.
Its read `typesVersions[key]` has receiver type
`{ readonly [key: string]: { readonly [path: string]: string[] } }`.
The witness preserves that receiver and string-variable index form, including
its record-valued result. Present `{ version: { entry: ["7"] } }` prints 7;
absent `{}` prints undefined under source Node. Both compiled backends instead
stop at the read with independently located, exact
`adamic: panic: indexed read is absent: <file>:<line>:<column>` stderr and exit 70.
CLI explain lists one checked indexed-presence site and trusted=0.

The extension only changes record type validation and record view validation.
Every nested level must satisfy the existing readonly string index signature
constraints; scalars and scalar arrays remain the leaves. The active type path
rejects direct and mutual recursive types. A 64-level bound also rejects
expanding generics that instantiate a different checker type at every edge.
This is a conservative implementation limit, not a claim that such finite
JavaScript objects cannot exist. No IR kind, backend emitter, runtime table or
guard helper was duplicated or changed. Nested views recurse through the existing
keeping and mutable-array widening checks, so accepting a nested record cannot
silently expose its literal-string array as a writable string array.

Two supplemental witnesses isolate the next levels:

| Prefix | Missing value | Checks / record gets / array gets | Mutant |
|---|---|---|---|
| `typesVersions[key]` (D119) | outer version key | 1 / 1 / 0 | erase outer record guard |
| `typesVersions[key][path]` | inner entry key | 2 / 2 / 0 | erase inner record guard; outer retained |
| `typesVersions[key][path][i]` | final array element | 3 / 2 / 1 | erase array guard; both record guards retained |

Each prefix's present control matches source Node in JS and native release and
ASan/UBSan. Each missing-last-level source Node control prints undefined, while
both compiled backends stop with the pinned site stderr and exit 70. The IR
and explain assertions require exactly one check per read; C assertions require
exactly one lookup per read. Chained AST nodes share their starting location,
so explain repeats that location two or three times with the corresponding
checked count. Each emitted-C mutant builds and runs in both native modes,
exits 0 and prints undefined: all six are caught by the exact stderr/exit assertion.
Earlier prefix guards remain present. The successful counted sanitizer controls
for all three prefixes each allocate and free five objects, with zero regions.

The expanded function/alias witness returns the nested record through a
side-effecting receiver, then calls side-effecting outer key, inner key and
array index functions. Source Node and both compiled backends print each label
once in that order, then vv, vv, true, true. This preserves outer and inner
identity and uses an allocated string payload under native sanitizers.
Six added Node-held refusal controls cover nested mutable tables, nested null,
mutual recursion, expanding generics, nested mutable-array widening and nested
record-to-fixed views. All original boundary controls still pass (21 total).

The full record witness suite reran all nine native single-site erasures in
both modes (18 mutant builds/runs) and the JS prototype guard erasure. The
focused runtime suite reran its nine existing runtime mutants: unsorted keys,
UINT32_MAX key classification, deleted iterator keys, overwrite-key leak,
stored-key premature free, null own-slot read, restored prototype membership,
silent missing read, and own-hit falsely guarded. All were caught by their
Node, named-stop or sanitizer assertions. Raw output is in evidence/nested-*.log.

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/nested-records-setup.log 2>&1
source /workspace/adamic-tools/env.sh
go test -count=1 -v ./stage3/stricter-records > /tmp/nested-records-proof-final.log 2>&1
go test -count=1 -timeout 10m ./internal/lower > /tmp/nested-records-lower.log 2>&1
go test -count=1 -v ./internal/native -run '^TestRecordsAgainstNode$|^TestRecordMutants$|^TestRecordReadMutants$' > /tmp/nested-records-runtime.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 10m -v ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/(indexing|string_index|maps_and_text|objects)\.a$' > /tmp/nested-records-oracle.log 2>&1
go vet ./internal/lower ./stage3/stricter-records > /tmp/nested-records-vet.log 2>&1
git diff --check > /tmp/nested-records-whitespace.log 2>&1
```

The oracle selection ran eight fixtures (including four object foreach fixtures
matched by Go's slash-separated filtering), with native zero hits/24 misses,
Node zero hits/16 misses. All passed. The first nested proof run also passed in
28.202s; the final run above includes the expanded alias/evaluation-order control.
No failed or resource-killed test run occurred in this extension.
Setup timings: Go and Node ready 0.023s, submodules 0.084s, markdown skip step
0.031s / ready 0.096s, clang 0.208s, Go build 40.268s, deferred test binaries
40.405s, build cache warm 40.407s, done 40.439s. nproc=5, cpu.max=400000 100000;
Go 1.27.1, clang 20.1.8, Node v24.19.0.

Observed: D119's nested shape now lowers and its guard can fail. Inference: d can
merge this branch and finish its own row by removing the refusal expectation;
D119 itself requires one guard. Its supplemental full chain would require three.
This branch does not claim d's unchanged harness passes or that nullable/typed-array
work from its later dependency merges is part of this records base.
