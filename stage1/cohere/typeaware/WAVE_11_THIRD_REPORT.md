Built native .a ports of no-eval, no-extend-native and no-func-assign.
Claim 5e818c1c was pushed before implementation 9c9cfa57; evidence is committed separately.
Agreement PASS 77.282s: 211 controls, 191 findings, frozen 287 repository and 77 compiler roots, normal and sanitized.
Three rule mutants, two judgment mutants and one released-handle mutant were caught.
Not covered: full repository gate, full upstream option matrix, and emitted-JavaScript rule comparison.

## Scope and selection

All six earlier claims were completed and pushed through 617850b4 before this
batch. Fetch audited 330 origin refs, 325 distinct trees and 33 Markdown claim
documents. Combined descending compiler/repository finding volume with lexical
ties selected these three zero-volume rules after excluding base ports and every
rule named in any origin claim, including released and skipped reservations.
Main and bridge references to these names were inventory/count entries, not
ports. The preclaim snapshot is recorded in validation-wave-11-third/third-selection.json.

Each production rule is in its own .a file. The support and runner files are
also .a. Existing declaration-lineage and binding-declarations checker facts
suffice; this batch adds no bridge questions or registrations and changes no
shared generator, shared harness, compiler implementation, or submodule.
Its independent Go oracle invokes unchanged production rules from cohere
715ba94f3608a6500086b1076ce5cb7e51b836db. TypeScript compiler roots are from
050880ce59e30b356b686bd3144efe24f875ebc8 (v6.0.3).

NoEval preserves direct-call wrapper ranges, shadowed direct calls, indirect
reference filtering and same-name global member chains. NoExtendNative preserves
the 49 builtin names, first-declaration provenance, assignment operator coverage,
and Object.defineProperty/defineProperties first-argument behavior. Updates,
deletes and deeper property writes stay silent exactly where Go stays silent.
NoFuncAssign uses first local declaration identity, including shorthand value
symbols, named function expressions, hoisting and overload declarations. Anonymous
functions and arrow functions do not create function-name anchors.

## Complete-byte comparison

The comparator checks serialized ranges, rule/message identity, full message,
fixes, suggestions and duplicate findings. These three default-configured Go
rules offer no fixes or suggestions, so their empty fields are included. Nothing
is normalized except deterministic sorting shared by the independent runners.

| Population | Roots | Findings | Equal bytes | Normal / ASan / UBSan / LSan |
| --- | ---: | ---: | ---: | --- |
| Positive and negative controls | 211 | 191 | 119509 | PASS |
| Frozen repository corpus | 287 | 0 | 18485 | PASS |
| TypeScript src/compiler | 77 | 0 | 5318 | PASS |

Controls include 22 hand-written cases and 189 literal source cases extracted
from upstream Go test tables. Upstream expectations are not copied: production
Go computes every expected byte independently. Options on those tables are not
replayed; both runners use default options. Computed/joined upstream source
expressions are not extracted. Findings by rule: no-eval 79, no-extend-native 49,
no-func-assign 63. Zero findings on the two large corpora are observations, not
proof of all behavior; positive controls and mutants exercise the report paths.

## Mutants

All five diagnostic mutants compile, exit zero and emit empty stderr. Only
comparison with independent Go findings detects their changed behavior.

| Mutant | Comparison catches first changed byte |
| --- | ---: |
| no-eval range end +1 | 58 |
| no-extend-native range end +1 | 19650 |
| no-func-assign range end +1 | 27889 |
| Invert global declaration-file provenance | 1748 |
| Replace declaration identity with function-name membership | 35269 |

A separate scratch Go overlay keeps released programs in the bridge registry.
The real released-handle query exits with panic 70, `invalid or released checker
handle`; the mutant exits zero and is caught by the required panic check. The
production bridge is unchanged. Initial validation also exposed an unsupported
long optional chain and a missing anonymous-function name guard; both were
corrected before the passing run and are not counted as mutants.

## Native timing against Go

Three alternating rounds ran without competing builds/tests. Every timed output
was compared with the complete expected bytes. Medians include process startup,
program loading, rule work, output and teardown; per-phase timing and checker
query counts are retained in third-timing.json.

| Population | Go median | Native median | Native / Go |
| --- | ---: | ---: | ---: |
| Repository | 126.228 ms | 285.873 ms | 2.265x |
| Compiler | 386.204 ms | 2091.649 ms | 5.416x |

Native is slower in both measurements. Instrumented validation observed 633
repository queries and 9190 compiler queries. These are observed timings on this
worker, not a general performance claim.

## Commands and evidence

Toolchain setup was reused from this branch: cloud/setup.sh ready 0s, cache warm
86s, total 86s; nproc 5. Commands sourced /workspace/adamic-tools/env.sh. Test
output was redirected to log files, never piped.

```
ADAMIC_WAVE_11_THIRD_ARTIFACTS=/workspace/wave-11-third-validation ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-11-typescript go test -v -count=1 -timeout 30m ./stage1/cohere/typeaware -run '^TestWave11ThirdAgreementAndMutants$'
# PASS, 77.282s; third-agreement.log

go test -v -count=1 -timeout 10m ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(sorting|string_index|functions|closures)\.a$'
# PASS, 13.415s; third-node.log

go vet ./...
# PASS, empty third-vet.log

gofmt -l stage1/cohere/typeaware/wave_11_third_test.go stage1/cohere/typeaware/testdata/oracle_wave_11_third.go
# Empty third-gofmt.log

git diff --check
# Empty third-diff-check.log
```

validation-wave-11-third contains full compressed Go/native/sanitized outputs,
output hashes, source hashes, exact relative corpus manifests, claim/push logs,
selection snapshot, passing logs and the quiet timing script/results. Build
archives and executables remain in scratch, not the repository.

## Limits

No full repository gate or full upstream option matrix was run. The pinned
cohere CLI still cannot resolve .a modules for the own-lint/format gate; the
native compiler does compile these .a sources. Shared .a/profile/suggestion and
emitted-JavaScript comparison work remains on codex/lint-harness-dot-a and was
not edited or integrated here. The filtered Node oracle tests native compiler
behavior; it is not an emitted-JavaScript comparison of these rules. This batch
has no unresolved blocker for the requested native default-rule comparisons.
The full bridge package check reported in WAVE_11_NEXT_REPORT.md is prior
validation of the unchanged bridge, not a newly rerun check for this batch.
