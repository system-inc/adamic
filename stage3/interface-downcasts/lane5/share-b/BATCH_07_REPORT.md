Certified: 20 new pairs / 40 reads / 93 fixtures; step 09.
Commits: base 83429e12; batch 07 commit on codex/views-callables-b.
Commands: TestCheckedViewCallableShareB, TestCheckedViewCallableShareBCounts, verify.cjs, run-mutants.py, go vet ./internal/oracle; outputs in logs/batch-07.
Mutants: 20 arity mutants; declaration mutant 979 and array carrier mutant 1000 exit 1.
Uncovered: 120 code-needed pairs / 1,568 reads; Set dependencies at codex/views-set-receiver 058635b9; optional-host, bind/condition, and remaining ranked pairs.

certified-pairs.json: 120 pairs / 402 reads. families.json: 564 fixtures. Original lane ledger plus share b: 276 pairs / 3,101 reads.

batch-07-families.json:

| Rank | Pair | Reads |
| --- | --- | ---: |
| 970 | NodeFactory.updateIndexSignature | 2 |
| 973 | NodeFactory.updateJsxOpeningElement | 2 |
| 979 | NodeFactory.updateNonNullExpression | 2 |
| 982 | NodeFactory.updateReturnStatement | 2 |
| 985 | NodeFactory.updateTypeAliasDeclaration | 2 |
| 988 | NodeFactory.updateTypeQueryNode | 2 |
| 997 | ParenthesizerRules.parenthesizeCheckTypeOfConditionalType | 2 |
| 1000 | ParenthesizerRules.parenthesizeConstituentTypesOfIntersectionType | 2 |
| 1003 | ParenthesizerRules.parenthesizeExpressionOfExportDefault | 2 |
| 1006 | ParenthesizerRules.parenthesizeExtendsTypeOfConditionalType | 2 |
| 1009 | ParenthesizerRules.parenthesizeOperandOfTypeOperator | 2 |
| 1024 | Program.getConfigFileParsingDiagnostics | 2 |
| 1027 | Program.getProgramDiagnostics | 2 |
| 1033 | Program \| undefined.getSourceFiles | 2 |
| 1036 | ProgramDiagnostics.getFileProcessingDiagnostics | 2 |
| 1045 | ResolutionCache.invalidateResolutionsOfFailedLookupLocations | 2 |
| 1048 | ResolutionCacheHost.onChangedAutomaticTypeDirectiveNames | 2 |
| 1054 | ResolveModuleNameResolutionHost.getCanonicalFileName | 2 |
| 1057 | RuntimeTypeSerializer.serializeParameterTypesOfNode | 2 |
| 1093 | SourceMapGenerator.toString | 2 |

Original TypeScript checkout: 050880ce59e30b356b686bd3144efe24f875ebc8. Fixtures retain complete original callable declarations and original read expressions. verify.cjs checks declarations, UTF-16 spans, source hashes, and carrier declarations. Adjacent object carriers use the existing lane reductions. Rank 1000 retains NodeArray<T> extends ReadonlyArray<T> and an actual array result. Array result variants retain original element types and optionality. generate.cjs now creates nested receiver scaffolds for state.program reads.

TestCheckedViewCallableShareBFamilies passes Node comparisons in native release, ASan/UBSan, and JavaScript; positive leak checks; negative exit 70 and exact field/expected-type/found-type messages. Array wrong-result variants produce Node TypeError and the pinned Adamic incompatible-result exit 70.

New code-needed pairs: 2 pairs / 4 reads.

| Rank | Read | Reads | Stop | Needed |
| --- | --- | ---: | --- | --- |
| 976 | context.factory.updateMethodSignature | 2 | (node: MethodSignature, modifiers: readonly Modifier[] \| undefined, name: PropertyName, questionToken: QuestionToken \| undefined, typeParameters: NodeArray<...> \| undefined, parameters: NodeArray<...>, type: TypeNode \| undefined) => MethodSignature | complete expected callable type display with both NodeArray type arguments |
| 994 | moduleResolutionCache.getPackageJsonInfoCache().getInternalMap | 2 | a primitive brand member __pathBrand whose type is not void | original Path brand admission preserving __pathBrand: any |

Rank 976 retains its original NodeArray declarations and array parameters. The inventory's expected type already abbreviates its NodeArray arguments. The complete declaration names TypeParameterDeclaration and ParameterDeclaration; TestCheckedViewCallableShareBDiagnosticNames compares the complete expected type with the observed abbreviated display. Native release, ASan/UBSan, JavaScript and positive leaks pass. This pair is code-needed and has one executable counts row.

Rank 994 retains Map<Path, PackageJsonInfoCacheEntry> and the original Path primitive brand. Its original nested getter receiver is retained. The probe stops on __pathBrand: any before callable certification and adds no executable counts row.

code-dependencies.json retains 22 Set pairs / 48 reads naming codex/views-set-receiver 058635b9. No Set branch was merged or certified here.

Commands, with source /workspace/adamic-tools/env.sh and TMPDIR=/tmp/adamic-gate:

- ADAMIC_CALLABLE_SHARE_B_FAMILIES=batch-07-families.json ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableShareB' -count=1 -v -timeout 10m: PASS, 106.859s.
- ADAMIC_CALLABLE_SHARE_B_FAMILIES=batch-07-families.json python3 stage3/interface-downcasts/lane5/share-b/run-mutants.py: 20 arity mutants fail original exit/message pins; native and JavaScript execution controls pass. Per-rank controls and failed mutants are in controls.log and arity.log.
- ADAMIC_CALLABLE_SHARE_B_FAMILIES=batch-07-families.json node stage3/interface-downcasts/lane5/share-b/verify.cjs /tmp/lane5-b-original: 20 pairs / 40 reads / 93 fixtures.
- node stage3/interface-downcasts/lane5/share-b/verify.cjs /tmp/lane5-b-original: 120 pairs / 402 reads / 564 fixtures.
- ADAMIC_CALLABLE_SHARE_B_FAMILIES=batch-07-families.json ADAMIC_CALLABLE_SHARE_B_DECLARATION_MUTANT=979 node stage3/interface-downcasts/lane5/share-b/verify.cjs /tmp/lane5-b-original: exit 1, original declaration changed. The expression parameter changes from Expression to number in memory.
- ADAMIC_CALLABLE_SHARE_B_FAMILIES=batch-07-families.json ADAMIC_CALLABLE_SHARE_B_ARRAY_MUTANT=1000 node stage3/interface-downcasts/lane5/share-b/verify.cjs /tmp/lane5-b-original: exit 1, carrier declaration changed. NodeArray changes from a ReadonlyArray subtype to an object carrier in memory.
- ADAMIC_CALLABLE_SHARE_B_FAMILIES=batch-07-families.json go test ./internal/oracle -run '^TestCheckedViewCallableShareBCounts$' -count=1 -args -update-counts: PASS, 12.348s. counts.md adds 94 rows and removes 0: 93 certified fixtures and one executable diagnostic probe.
- go test ./internal/oracle -run TestCountsAreRecorded -args -update-counts: FAIL, 50.259s, same 60 existing fixture failures as batch 06.
- go vet ./internal/oracle: exit 0.

Toolchain setup from batch 04 reused: GOPROXY=https://proxy.golang.org|direct, nproc=5; setup done 37.944s. Timing lines and output remain in logs/batch-04/setup.log.

Files changed: share-b fixtures, JSON ledgers, fixture harnesses, internal/oracle/checked_views_callable_share_b_test.go, appended internal/oracle/counts.md rows. No compiler/runtime edits, full package test, full gate, branch merge, or PR.

All share-b .cjs files pass node --check. Repository log copies remove trailing whitespace; raw outputs remain under /tmp/lane5-b-*07.log and /tmp/lane5-b-mutants/batch-07-families/.
