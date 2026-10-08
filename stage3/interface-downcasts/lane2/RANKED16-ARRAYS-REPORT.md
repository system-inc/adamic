Held group sixteen's seven array pairs against complete original declarations; no compiler change was needed.
Commits: follows 19b40cd487e497a007bf6467a6daa038bc96492c; this report and fixtures are in the next group commit on codex/views-arrays-callables-parser.
Validation: 42 probes passed against Node, sanitized native, release native and JavaScript; all 21 finishing probes passed leak checks; 42 count rows measured; go vet passed.
Mutants: native and JavaScript numeric field bypasses were both caught by function-parameters-wrong-pos; restored witness passes.
Limits: static witnesses are not production reachability; tuples and union-target casts skipped; the two named baseline counts failures belong to the integrator.

| Original pair | Type id | Static candidate reads | Original declaration |
| --- | ---: | ---: | --- |
| SourceFile.patternAmbientModules | 6998 | 4 | `PatternAmbientModule[] \| undefined` |
| PropertyAssignment.modifiers | 7294 | 4 | `NodeArray<ModifierLike> \| undefined` |
| ShorthandPropertyAssignment.modifiers | 7295 | 4 | `NodeArray<ModifierLike> \| undefined` |
| InterfaceDeclaration.modifiers | 7303 | 4 | `NodeArray<ModifierLike> \| undefined` |
| IndexSignatureDeclaration.modifiers | 7323 | 4 | `NodeArray<ModifierLike> \| undefined` |
| FunctionTypeNode.parameters | 7324 | 4 | `NodeArray<ParameterDeclaration>` |
| TypeAliasDeclaration.modifiers | 7636 | 4 | `NodeArray<ModifierLike> \| undefined` |

Original revision: microsoft/TypeScript 050880ce59e30b356b686bd3144efe24f875ebc8. The existing lane 4b declaration emitter produces 78 full declaration files, including dependencies; the adapter validates all original pair types and 28 exact read spans. The manifest pins emitted digests and full field lists. The oracle asserts complete receiver field lists; it also asserts PatternAmbientModule, Symbol and ParameterDeclaration field lists wherever those contracts are read. ModifierLike remains the original tagged union, with all branches declared.

Each pair has a good control (optional fields additionally absent and explicit undefined), an unread malformed array, wrong array kind, wrong descendant scalar, missing descendant readiness, and a malformed second element left unread. SourceFile's selected descendant is symbol.flags, checked against the original SymbolFlags enum; the other pairs select pos. Full declarations include callable text members, but these probes do not read text and do not trigger that known frontier.

Commands and observations (test output is in original16/evidence/):

```sh
source /workspace/adamic-tools/env.sh
node stage3/interface-downcasts/lane2/original16/prepare.cjs /workspace/lane2-original-archive /workspace/lane2-original-declarations > /tmp/lane2-group16-prepare.log 2>&1
ADAMIC_ARRAY16_ORIGINAL_DECLS=/workspace/lane2-original-declarations go test ./internal/oracle -run '^TestCheckedViewRanked16OriginalArrays$' -count=1 -v -timeout 10m > /tmp/lane2-group16-oracle-corrected.log 2>&1
ADAMIC_ARRAY15_ORIGINAL_DECLS=/workspace/lane2-original-declarations ADAMIC_ARRAY16_ORIGINAL_DECLS=/workspace/lane2-original-declarations go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts > /tmp/lane2-group16-required-counts.log 2>&1
ADAMIC_ARRAY16_ORIGINAL_DECLS=/workspace/lane2-original-declarations python3 stage3/interface-downcasts/lane2/original16/run-mutants.py > /tmp/lane2-group16-mutants.log 2>&1
ADAMIC_ARRAY16_ORIGINAL_DECLS=/workspace/lane2-original-declarations go test ./internal/oracle -run '^TestCheckedViewRanked16ArrayCounts$' -count=1 -timeout 10m -args -update-counts > /tmp/lane2-group16-counts.log 2>&1
go vet ./internal/oracle > /tmp/lane2-group16-vet.log 2>&1
ADAMIC_ARRAY16_ORIGINAL_DECLS=/workspace/lane2-original-declarations go test ./internal/oracle -run '^TestCheckedViewRanked16OriginalArrays$/^function-parameters-wrong-pos$' -count=1 -v -timeout 10m > /tmp/lane2-group16-restored-mutant-witness.log 2>&1
```

The complete corrected oracle passed in 65.420s; counts passed in 37.385s; restored mutant witness passed in 2.079s. The initial run caught fixture mistakes (excess-property literals, a generated malformed numeric token, and diagnostic names); those were corrected before certification. No compiler change masked them.

Native-number-field mutant returned any numeric slot without checking its stored kind. Sanitized and release executions printed unchecked numeric values and exited zero instead of rejecting. JavaScript-number-field mutant made the numeric predicate unconditional, printed bad and exited zero. Both failures come from executed checks, not compiler rejection; the runner restores the temporary production edits in finally. Logs retain the exact witnesses.

The required global counts command failed in 77.645s before reaching the additional lane registry and made no counts.md change. Named baseline failures: internal/oracle/testdata/graph_regions_regression_06.a (free(): invalid pointer) and internal/oracle/testdata/process_exit.a (stage 0 cannot lower process.exit as a value). Their baseline reproductions are recorded in original15/evidence/counts-baseline.log. These belong to the integrator and do not stop lane 2, per the user. We separately wrote the 42 measured group rows and asserted that removing those new rows gives the previous counts.md byte for byte; no other row changes remain.

Declaration-dependent tests explicitly skip without generated inputs; the count registry retains already recorded rows without claiming remeasurement in that case. With inputs supplied it measures them. No whole-package test or full gate was run. Toolchain setup timings and integration merge are in the group fifteen report.

| Family | Candidate pairs / reads | Cumulative fixture obligations held | Remainder |
| --- | ---: | ---: | ---: |
| Array contracts | 334 / 3189 | 155 / 2741 | 179 / 448 |
| Element or consumer reads | 251 / 1602 | 0 / 0 production credit | 251 / 1602 |
| Intrinsic contracts | 179 / 794 | 0 / 0 production credit | 179 / 794 |
| Own array fields | 30 / 72 | 3 / 26 | 27 / 46 |

The cumulative array total comprises inherited representative obligations plus ten pairs held using complete original declarations in groups fifteen and sixteen. Remaining pairs include skipped union-target casts and previously refused obligations; tuples belong to their worker. Continue at JSDocTemplateTag.comment and the remaining four-read array contracts.
