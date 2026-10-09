Built generic indexOf.call / lastIndexOf.call, proven Array.isArray and undefined copyWithin ends; heterogeneous searches now refuse before C emission.
Claim commits: d5643a4 (reproduction helpers) and c4fb1aa (complete classification and feature claim), pushed before implementation. The implementation is the commit containing this report.
Commands: real npm tsc classified every refusal; final Array measurement has 191 pass, 0 disagreement, 2281 refused, 0 crashed, 610 skipped; the final complete lower / flow / oracle gate and vet pass; repository-wide attempts hit other-package execution limits and were stopped.
Mutants: forward boundary, backward negative sign, isArray boolean and undefined-end clamp; all four caught only by Node stdout comparison.
Uncovered: sparse / boxed constructors, descriptors, prototypes, dynamic object coercion, structural hidden shapes and other language gaps handed to @system_adamic.

## Observations

Linux, Go 1.27.1, clang 20.1.8, Node 24.19.0, npm TypeScript 6.0.3; test262 c8c798898646638cd0c24879f8e0374e847e7d74. No cmd/adamic-test262 files changed.

| measurement | pass | disagreement | refused | crashed | skipped | total |
|---|---:|---:|---:|---:|---:|---:|
| before, main 5d4c801 | 125 | 0 | 2339 | 8 | 610 | 3082 |
| after | 191 | 0 | 2281 | 0 | 610 | 3082 |

66 newly passing tests all agree with Node; all 125 baseline passes remain passes. The eight crashes now refuse with a slot-representation reason. before.json and after.json contain both full directory tables and test paths; new-passes.json identifies every new pass and its original checker bucket.

Every one of the 2,339 baseline refusals was checked using real npm TypeScript on the complete adapted source, with Adamic checker options and prelude. Bucket (a): 1,845 matching-code TypeScript rejections. Bucket (b): 282 missing library features. Bucket (c): 212 language gaps. No diagnostic-code mismatches. The claim and complete ledger were pushed before code was written. See refusals.md and refusals.json.

| original first-blocker feature | newly passing tests |
|---|---:|
| generic Array.prototype.indexOf.call | 28 |
| generic Array.prototype.lastIndexOf.call | 25 |
| Array.isArray observations | 11 |
| Array.isArray or prototype method observations | 1 |
| copyWithin explicit undefined end | 1 |

## Implementation and limits

The largest family was the 100 generic search refusals (51 indexOf and 49 lastIndexOf). The lowering ports V8 Runtime_ArrayIndexOf in src/runtime/runtime-array.cc and GetFromIndex / GenericArrayLastIndexOf in src/builtins/array-lastindexof.tq, version 13.6.233.17. THIRD_PARTY_NOTICES.md names both files.

Supported receivers are dense arrays, UTF-16 strings, primitive numbers / booleans and fixed plain literal shapes or their unannotated, unreassigned bindings. Searches bind arguments before reading length; preserve strict equality, absent-key skipping, primitive ToLength and bounds, negative indexes, omitted versus explicit undefined fromIndex, huge lengths and canonical numeric keys. Only proven present numeric keys are visited; absent iterations are elided because the supported shapes have no accessors or inherited numeric properties.

Array.isArray proves the result from supported representations and preserves operand effects and ownership. Object views that may hide arrays, mixed / optional array representations and detached or shadowed intrinsics refuse. copyWithin now accepts literal and optional-number undefined ends and takes length after argument effects.

The crash was reproduced alone as built-ins/Array/prototype/indexOf/15.4.4.14-5-10.js. Before: one crash, clang rejects a double passed to pointer retain. After: one explicit refusal, no crash. The reduced negative oracle is internal/oracle/testdata/library_array_refused/library_array_heterogeneous.a, checked directly by TestLibraryArrayHeterogeneousCrashIsRefused. A top-level negative fixture initially failed the positive graph globs; it was moved into the conventional refusal subdirectory. During development the old floating key comparator also failed the new NaN-name Node case; integer key sorting fixes it. The rerun completed lower, flow, native and oracle packages successfully, but the full repository attempts did not complete. A strict-identity audit added conservative optional-value / structural-view refusals; complete lower / flow / oracle packages were then rerun on the final frozen source.

For fixed-shape generic searches, optional/nullable or mixed search values and fields refuse. Object fields need a plain literal identity proof; object search views that could hide another representation refuse. Numeric fields with an empty object type can hide primitives and refuse, as do null / undefined generic receivers requiring catchable ToObject errors, descriptors, dynamic coercion, unproven structural shapes and boxed receivers. Constructors remain refused: Array(n) needs sparse presence, and boxed primitives need identity and prototypes. Faithful support requires shared representation / emission work outside this unit. Language handoff and one-line reproducers are in refusals.md.

All new lowering is in internal/lower/library_array*.go. It uses ordinary existing IR and runtime primitives; no new runtime or emitter implementation was needed. Shared hooks: object.go builtin dispatches to libraryArrayGenericCall; objectLiteral obtains canonical numeric field names from libraryArrayLikeFieldName. The only other shared-file edits are generated oracle/counts.md rows and the required V8 notice. lower.go, native.go, emit.go, oracle_test.go and cmd/adamic-test262 are unchanged.

## Mutants and controls

| mutant | independent catcher | controls |
|---|---|---|
| generic indexOf key < length changed to <= length | Node excludes the key at length | native compiled; exit 0; empty stderr under ASan, UBSan and leak detection |
| generic lastIndexOf negative start length + n changed to length - n | Node relative-index results | same controls |
| Array.isArray proven return boolean inverted | Node true / false identity results | same controls |
| copyWithin undefined-end +Infinity clamp changed to zero | Node copying to the actual length | same controls; test uses only the new undefined-end fixture segment |

Implemented in TestLibraryArrayNewFamilyMutants, not a manually inspected claim. Each test first requires a clean native run, then requires its stdout to differ from a separate Node execution. mutants-and-refusals.log contains all four catches and the refusal probes.

## Linux gate and commands

All output was written to log files. Exact measurement and checker reproduction commands are in reproduce.md. Final commands:

```sh
source /workspace/adamic-tools/env.sh
go run ./cmd/adamic-test262 -adapt -json -test262 /workspace/test262 -work /tmp/library-array-complete built-ins/Array > /tmp/library-array-record.json 2> /tmp/library-array-record.log
go test ./internal/lower -run TestLibraryArray -count=1 -v -timeout 10m > /tmp/library-array-tests-final.log 2>&1
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/library_array_' -count=1 -timeout 30m > /tmp/library-array-oracles-final.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts > /tmp/library-array-counts.log 2>&1
go vet ./... > /tmp/library-array-vet-final.log 2>&1
TMPDIR=/tmp/adamic-array-gate go test -count=1 -timeout 30m ./... > /tmp/library-array-gate-final.log 2>&1
TMPDIR=/tmp/adamic-array-gate go test -count=1 -timeout 30m ./internal/lower ./internal/flow ./internal/oracle > /tmp/library-array-gate-guard-final.log 2>&1
```

Final measurement, direct Array checks, complete lower / flow / oracle packages and vet exited 0. linux-guard-gate.log records the final affected-package gate, including every native / JavaScript / Node oracle, sanitizer / ownership checks, graph traces and recorded counts. Native runtime files are unchanged. The complete native package passed in a repository-wide attempt (627.339s).

Full repository attempts are preserved in linux-first-gate.log and linux-gate.log. Observed: the first attempt failed the development fixture / comparator issues described above; its Unicode canonicalization Node child was killed, and the JSON native child exited -1 after a roughly 900s test run. The second attempt passed completed lower, flow, native and oracle packages, then was stopped while remaining long-running stage1 / Unicode work was pending. No full-repository pass is claimed. The final source uses the user-authorized complete affected-package fallback; other-package execution limits were not resolved in this slice.

Setup bash cloud/setup.sh: go ready 0s; clang ready 1s; node ready 1s; submodules ready 1s; build cache warm 100s; done 100s. nproc: 5; cgroup cpu.max 400000 100000; 17.6 GB. Sourced /workspace/adamic-tools/env.sh.
