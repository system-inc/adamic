Held group eighteen's five original diagnostic and string array pairs, including join consumers; no compiler change was needed.
Commits: follows group seventeen 1ebb3c58; this report and fixtures are in the next group commit on codex/views-arrays-callables-parser.
Validation: 36 probes pass Node, sanitized/release native and JavaScript; all 18 finishing probes pass leaks; 36 count rows measured; vet passes.
Mutants: native and JavaScript numeric-field bypasses and join element-check bypasses all caught during execution; unused native stringification mutant survived and receives no credit.
Limits: production reachability unmeasured; tuples and union-target casts skipped pending their lane; integrator owns graph_regions_regression_06.a and process_exit.a baseline counts failures.

| Original pair | Type id | Static reads | Original declaration |
| --- | ---: | ---: | --- |
| JsonSourceFile.parseDiagnostics | 9847 | 4 | `DiagnosticWithLocation[]` |
| TsConfigSourceFile.extendedSourceFiles | 9852 | 4 | `string[] \| undefined` |
| TsConfigSourceFile.parseDiagnostics | 9852 | 4 | `DiagnosticWithLocation[]` |
| ResolvedModuleWithFailedLookupLocations.affectingLocations | 9934 | 4 | `string[] \| undefined` |
| ResolvedModuleWithFailedLookupLocations.failedLookupLocations | 9934 | 4 | `string[] \| undefined` |

The original microsoft/TypeScript pin remains 050880ce59e30b356b686bd3144efe24f875ebc8. The adapter emits 78 complete declarations, validates original member types and all twenty static source spans, and pins emitted hashes and complete field lists. The oracle checks full receiver fields, plus DiagnosticWithLocation on selected diagnostic element reads. No reduced target schema or cohere code is substituted.

Each pair covers valid arrays, an unread wrong array, a reached wrong array, selected wrong scalar/element kind, selected missing scalar/undefined element readiness, and an unread bad second element. Optional properties additionally cover missing and explicit undefined. Each of the three string arrays covers valid join and a reached bad join. DiagnosticWithLocation.start remains required number. The ResolvedModuleWithFailedLookupLocations base retains resolvedModule explicitly undefined and reads it lazily in the unread-array control.

The initial heterogeneous string/number literal fixtures hit the existing raw-array lowering refusal. Their replacements use a homogeneous numeric array for bad join (Node prints 7;8, checked backends reject the first number), and string/undefined arrays with an unread second element for laziness. No refusal was removed to make these fixtures run.

```sh
source /workspace/adamic-tools/env.sh
node stage3/interface-downcasts/lane2/original18/prepare.cjs /workspace/lane2-original-archive /workspace/lane2-original-declarations18 > /tmp/lane2-group18-prepare.log 2>&1
ADAMIC_ARRAY18_ORIGINAL_DECLS=/workspace/lane2-original-declarations18 go test ./internal/oracle -run '^TestCheckedViewRanked18OriginalArrays$' -count=1 -v -timeout 10m > /tmp/lane2-group18-oracle-corrected.log 2>&1
ADAMIC_ARRAY18_ORIGINAL_DECLS=/workspace/lane2-original-declarations18 go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts > /tmp/lane2-group18-required-counts.log 2>&1
ADAMIC_ARRAY18_ORIGINAL_DECLS=/workspace/lane2-original-declarations18 python3 stage3/interface-downcasts/lane2/original18/run-mutants.py > /tmp/lane2-group18-mutants-corrected.log 2>&1
ADAMIC_ARRAY18_ORIGINAL_DECLS=/workspace/lane2-original-declarations18 go test ./internal/oracle -run '^TestCheckedViewRanked18ArrayCounts$' -count=1 -timeout 10m -args -update-counts > /tmp/lane2-group18-counts.log 2>&1
go vet ./internal/oracle > /tmp/lane2-group18-vet.log 2>&1
ADAMIC_ARRAY18_ORIGINAL_DECLS=/workspace/lane2-original-declarations18 go test ./internal/oracle -run '^TestCheckedViewRanked18OriginalArrays$/^(json-diagnostics-wrong-element|config-files-join-bad|config-files-join|config-files-lazy-element)$' -count=1 -v -timeout 10m > /tmp/lane2-group18-restored.log 2>&1
```

Complete oracle passed in 86.231s; group counts passed in 27.282s; restored witnesses passed in 5.492s; vet passed with no output. Tests write logs retained under original18/evidence. No whole-package tests or full gate were run.

| Mutant | Witness | Executed result |
| --- | --- | --- |
| Native numeric field accepts any slot | json-diagnostics-wrong-element | Both native executions print unchecked numeric values and exit zero |
| JavaScript numeric field always true | json-diagnostics-wrong-element | Prints bad and exits zero |
| Native string element kind guard bypass | config-files-join-bad | Sanitized execution catches invalid pointer dereference in adamic_retain; release aborts, replacing the expected boundary rejection |
| JavaScript join omits element checks | config-files-join-bad | Prints 7;8 and exits zero |

All four executed kills restore temporary source edits in finally; the restored witnesses pass. An initial native mutant bypassed adamic_view_array_string's checked lookup and survived. Ordinary join uses emitViewArrayJoin and adamic_view_array_at instead; that mutant does not exercise this consumer and is explicitly not credited. Its survivor log is retained. The successful native join mutant bypasses the string-kind guard used by the real join path. Neither successful kill relies on compiler failure.

The required counts attempt failed in 90.156s before the lane registry and wrote no table changes. Named baseline failures: internal/oracle/testdata/graph_regions_regression_06.a (invalid pointer free) and internal/oracle/testdata/process_exit.a (unsupported process.exit value lowering). Both reproduce at baseline 0f47b23c; evidence is under original15/evidence/counts-baseline.log. Per the user, the integrator owns them and this lane continues. We wrote only the 36 measured group rows, verifying no previous row changed. Missing external declaration inputs explicitly skip the declaration-dependent tests and retain recorded rows without claiming remeasurement.

| Family | Candidate pairs / reads | Cumulative fixture obligations held | Remainder |
| --- | ---: | ---: | ---: |
| Array contracts | 334 / 3189 | 169 / 2797 | 165 / 392 |
| Element or consumer reads | 251 / 1602 | 0 / 0 production credit | 251 / 1602 |
| Intrinsic contracts | 179 / 794 | 0 / 0 production credit | 179 / 794 |
| Own array fields | 30 / 72 | 3 / 26 | 27 / 46 |

These static ledger totals combine inherited representative obligations with 24 pairs / 96 candidate reads held using complete original declarations in groups fifteen through eighteen. They do not measure production execution. Remaining four-read candidates include FileWatcherWithModifiedTime.callbacks; NodeBuilderContext.trackedSymbols has tuple elements and is skipped, as is IncrementalMultiFileEmitBuildInfo.affectedFilesPendingEmit's tuple element branch. IncrementalBuildInfo.fileNames has a union target. Toolchain and integration provenance remain in group fifteen's report.

Next priority, per the latest user steering, is merging codex/views-mixed-unions-2 at 94b200ef and certifying FunctionLikeDeclaration.parameters and ClassDeclaration | ClassExpression.members. The next ordinary ranked profiles (thirteen three-read pairs) are prepared but unvalidated and receive no ledger credit yet.
