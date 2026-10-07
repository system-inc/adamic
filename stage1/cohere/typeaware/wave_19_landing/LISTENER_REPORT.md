Built: numeric listener declarations for all six owned rules; per-node execution is blocked on the shared numeric parser and dispatcher.
Commits: this change follows f8f29123aa9b9ac27e8b4f2c6272d2669ea088e7 and contains origin/main e8ba3d5d81de4d3773c723914fccd4c76248b965; pushed only to codex/typeaware-wave-19.
Checks: six native rule oracles PASS in 538.287s, enum declarations PASS, checker PASS, vet PASS, filtered Node oracle PASS; normal and sanitizer corpora agree.
Mutants: ten existing byte-comparison mutants, three retained-handle mutants and the new numeric listener metadata mutant were caught.
Uncovered: numeric dispatch and handed-node execution, three production question registrations, lint emitted-JavaScript comparison and the full gate; no new claims.

The six owned .a classes now expose `readonly listenerKinds: readonly number[]`. The numeric IDs are pinned against the independent Go frontend enum, using imported Go ast.Kind constants in TestWave19NumericListenerDeclarations:

| Rule | Listener kinds |
| --- | --- |
| consistent-generic-constructors | VariableDeclaration 261, PropertyDeclaration 173, Parameter 170 |
| dot-notation | ElementAccessExpression 213 |
| no-array-constructor | CallExpression 214, NewExpression 215 |
| correctness-no-uncleared-race-timeout | CallExpression 214 |
| correctness-no-process-exit-after-output | CallExpression 214 |
| correctness-require-blocking-standard-streams | SourceFile 307 |

These are Go parser IDs, not assumed stock TypeScript enum values. The compilation-unit blocking rule declares SourceFile because its analysis covers the complete file.

The shared parser's stage1/typescript/parser/nodes.ts:5 exposes only `readonly kind: string`. The current driver supplies no numeric node-kind field or handed-node visitor API. Ahra's correction prohibits shared parser/harness edits. Existing rule run methods therefore still inspect string kinds and scan/refetch nodes. This change does not meet the execution part of the speed rule, and does not claim a performance improvement. Integration must provide the numeric parser/driver and then convert the existing owned visitors to consume handed nodes. No adapter translating string kinds was added.

The declaration test reads the actual owned source and compares the arrays to the external frontend enum. A separate source copy changed the process rule listener from CallExpression 214 to NewExpression 215. Running the same test with ADAMIC_WAVE19_LISTENER_SOURCE pointing at that copy failed with `got [215], want [214]`. This proves the metadata check can fail; the current driver ignores metadata, so this mutant is not a findings-comparison proof.

Commands (all output saved in evidence/listeners):

```
source /workspace/adamic-tools/env.sh
go test ./stage1/cohere/typeaware -run '^TestWave19NumericListenerDeclarations$' -count=1 -v
ADAMIC_WAVE19_LISTENER_SOURCE=/tmp/wave19-listeners/metadata-mutant go test ./stage1/cohere/typeaware -run '^TestWave19NumericListenerDeclarations$' -count=1 -v
ADAMIC_WAVE19_ARTIFACTS=/tmp/wave19-listeners/original ADAMIC_WAVE19_TIMEOUT_ARTIFACTS=/tmp/wave19-listeners/timeout ADAMIC_WAVE19_STREAM_ARTIFACTS=/tmp/wave19-listeners/process ADAMIC_WAVE19_BLOCKING_ARTIFACTS=/tmp/wave19-listeners/blocking ADAMIC_WAVE19_RELEASE_ARTIFACTS=/tmp/wave19-listeners/released ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave19-typescript go test ./stage1/cohere/typeaware -run '^TestWave19(AgreementAndMutants|TimeoutPendingRegistration|ProcessPendingRegistration|BlockingPendingRegistration|StreamReleasedHandles)$' -v -count=1 -timeout 30m
go test ./bridge/tsgo/checker -count=1 -v
go vet ./stage1/cohere/typeaware
go test ./internal/oracle -run 'TestTheOracleCatchesOneByte|TestNativeAgreesWithNode/internal/oracle/testdata/(closures|method_closures|generic_functions|regions|regions_throw|call_targets_.*|devirtualize)\.a$' -count=1 -timeout 15m -v
```

The Node oracle may reuse its supported compiler cache. The full bridge gate and the five raw-fact contract mutants passed in the preceding landing checks and were not rerun for this metadata-only change. The rule oracle reran all controls, byte mutants, released-handle refusal checks and ASAN/UBSAN/LSAN runs on TypeScript compiler and repository corpus roots. The three pending questions use scratch registration overlays; their production dispatch remains explicitly unavailable.

Three alternating count-only benchmark rounds after tests, median whole-process seconds:

| Runner | Corpus | Go seconds | Native seconds | Native / Go |
| --- | --- | --- | --- | --- |
| original | compiler | 0.479801 | 2.356452 | 4.91 |
| original | repository | 0.302096 | 0.446881 | 1.48 |
| timeout | compiler | 0.421566 | 2.986782 | 7.08 |
| timeout | repository | 0.199700 | 0.378864 | 1.90 |
| process | compiler | 0.420190 | 1.979270 | 4.71 |
| process | repository | 0.244301 | 0.419083 | 1.72 |
| blocking | compiler | 0.428409 | 2.223158 | 5.19 |
| blocking | repository | 0.208030 | 0.353485 | 1.70 |

Raw results, stdout/stderr and manifests are in streams.tar.gz with per-file SHA-256 hashes. Benchmarks measure total process time, including program loading and the existing driver, not isolated listener cost. Corpus roots and prior exclusions remain those documented in WAVE_19_REPORT.md and QUESTION_COLLISION_REPORT.md.
