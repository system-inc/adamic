Certified: 20 new pairs / 40 reads / 95 fixtures; step 09.
Commits: base 8dae7880; batch 06 commit on codex/views-callables-b.
Commands: TestCheckedViewCallableShareB, TestCheckedViewCallableShareBCounts, verify.cjs, run-mutants.py, go vet ./internal/oracle; outputs in logs/batch-06.
Mutants: 20 arity mutants; rank 799 declaration mutant exits 1.
Uncovered: 118 code-needed pairs / 1,564 reads; Set dependencies at codex/views-set-receiver 058635b9; optional-host, bind/condition, and remaining ranked pairs.

certified-pairs.json: 100 pairs / 362 reads. families.json: 471 fixtures. Original lane ledger plus share b: 256 pairs / 3,061 reads.

batch-06-families.json:

| Rank | Pair | Reads |
| --- | --- | ---: |
| 799 | IterationTypesResolver.getGlobalIteratorObjectType | 2 |
| 886 | ModuleResolutionCache.clear | 2 |
| 889 | ModuleResolutionCache \| undefined.clear | 2 |
| 901 | ModuleSpecifierResolutionHost.useCaseSensitiveFileNames | 2 |
| 913 | NodeFactory.copyStandardPrologue | 2 |
| 916 | NodeFactory.createBigIntLiteral | 2 |
| 919 | NodeFactory.createContinueStatement | 2 |
| 922 | NodeFactory.createGlobalMethodCall | 2 |
| 925 | NodeFactory.createJSDocFunctionType | 2 |
| 928 | NodeFactory.createJSDocNullableType | 2 |
| 931 | NodeFactory.createJsxClosingElement | 2 |
| 937 | NodeFactory.createNamespaceExportDeclaration | 2 |
| 943 | NodeFactory.createShorthandPropertyAssignment | 2 |
| 949 | NodeFactory.createTemplateSpan | 2 |
| 952 | NodeFactory.createTypeOfExpression | 2 |
| 955 | NodeFactory.getNamespaceMemberName | 2 |
| 958 | NodeFactory.updateAsExpression | 2 |
| 961 | NodeFactory.updateConditionalExpression | 2 |
| 964 | NodeFactory.updateDeleteExpression | 2 |
| 967 | NodeFactory.updateExternalModuleReference | 2 |

Original TypeScript checkout: 050880ce59e30b356b686bd3144efe24f875ebc8. Fixtures retain complete original callable declarations and original read expressions. verify.cjs checks declarations, UTF-16 spans, source hashes, and primitive carrier provenance. Adjacent object carriers use the existing lane reductions. generate.cjs now handles number | undefined arguments and observations of optional string/object unions.

TestCheckedViewCallableShareBFamilies passes Node comparisons in native release, ASan/UBSan, and JavaScript; positive leak checks; negative exit 70 and exact field/expected-type/found-type messages. Original optional receiver reads such as moduleResolutionCache?.clear remain in the fixtures.

New code-needed pairs: 28 pairs / 64 reads.

| Rank | Read | Reads | Stop | Needed |
| --- | --- | ---: | --- | --- |
| 550 | fileInfos.set | 3 | a primitive brand member __pathBrand whose type is not void | original Path brand admission preserving __pathBrand: any |
| 553 | isSymlinkCache.clear | 3 | a primitive brand member __pathBrand whose type is not void | original Path brand admission preserving __pathBrand: any |
| 556 | projectReferenceRedirects.set | 3 | a primitive brand member __pathBrand whose type is not void | original Path brand admission preserving __pathBrand: any |
| 559 | skipOutputs?.has | 3 | a primitive brand member __pathBrand whose type is not void | original Path brand admission preserving __pathBrand: any |
| 562 | projectPendingBuild.set | 3 | a primitive brand member __pathBrand whose type is not void | original Path brand admission preserving __pathBrand: any |
| 565 | multiMap.get | 3 | native collection-to-object argument conversion clang error; JavaScript exit 70 with unknown intrinsic signature | collection receiver conversion, immutable intrinsic callable metadata and receiver-preserving checked calls |
| 568 | enumRelation.get | 3 | native collection-to-object argument conversion clang error; JavaScript exit 70 with unknown intrinsic signature | collection receiver conversion, immutable intrinsic callable metadata and receiver-preserving checked calls |
| 571 | durations.delete | 3 | native collection-to-object argument conversion clang error; JavaScript exit 70 with unknown intrinsic signature | collection receiver conversion, immutable intrinsic callable metadata and receiver-preserving checked calls |
| 802 | optionsToRedirectsKey.get | 2 | a primitive brand member __compilerOptionsKey whose type is not void | original RedirectsCacheKey brand admission preserving __compilerOptionsKey: any |
| 805 | state.affectedFilesPendingEmit!.delete | 2 | a primitive brand member __pathBrand whose type is not void | original Path brand admission preserving __pathBrand: any |
| 808 | packageDirWatcher.dirPathToWatcher.get | 2 | a primitive brand member __pathBrand whose type is not void | original Path brand admission preserving __pathBrand: any |
| 811 | emitSignatures.set | 2 | a primitive brand member __pathBrand whose type is not void | original Path brand admission preserving __pathBrand: any |
| 814 | state.fileInfos.forEach | 2 | a primitive brand member __pathBrand whose type is not void | original Path brand admission preserving __pathBrand: any |
| 817 | packageDirWatchers.get | 2 | a primitive brand member __pathBrand whose type is not void | original Path brand admission preserving __pathBrand: any |
| 820 | mapOutputFileToResolvedRef?.set | 2 | a primitive brand member __pathBrand whose type is not void | original Path brand admission preserving __pathBrand: any |
| 823 | seenFileNamesMap.set | 2 | a primitive brand member __pathBrand whose type is not void | original Path brand admission preserving __pathBrand: any |
| 826 | fileExistsCache.delete | 2 | a primitive brand member __pathBrand whose type is not void | original Path brand admission preserving __pathBrand: any |
| 829 | dirPathToSymlinkPackageRefCount.get | 2 | a primitive brand member __pathBrand whose type is not void | original Path brand admission preserving __pathBrand: any |
| 832 | state.semanticDiagnosticsPerFile.set | 2 | a primitive brand member __pathBrand whose type is not void | original Path brand admission preserving __pathBrand: any |
| 835 | oldProgram.sourceFileToPackageName.get | 2 | a primitive brand member __pathBrand whose type is not void | original Path brand admission preserving __pathBrand: any |
| 862 | context.remappedSymbolNames!.has | 2 | native collection-to-object argument conversion clang error; JavaScript exit 70 with unknown intrinsic signature | collection receiver conversion, immutable intrinsic callable metadata and receiver-preserving checked calls |
| 868 | sourceFilesCache.delete | 2 | native collection-to-object argument conversion clang error; JavaScript exit 70 with unknown intrinsic signature | collection receiver conversion, immutable intrinsic callable metadata and receiver-preserving checked calls |
| 871 | packageIdToSourceFile.set | 2 | native collection-to-object argument conversion clang error; JavaScript exit 70 with unknown intrinsic signature | collection receiver conversion, immutable intrinsic callable metadata and receiver-preserving checked calls |
| 874 | reverseHomomorphicMappedCache.get | 2 | native collection-to-object argument conversion clang error; JavaScript exit 70 with unknown intrinsic signature | collection receiver conversion, immutable intrinsic callable metadata and receiver-preserving checked calls |
| 877 | activeTypeMappersCaches[activeTypeMappersCount].clear | 2 | native collection-to-object argument conversion clang error; JavaScript exit 70 with unknown intrinsic signature | collection receiver conversion, immutable intrinsic callable metadata and receiver-preserving checked calls |
| 880 | durations.forEach | 2 | checked view read of field forEach with unsupported callable contract | original callback callable contract admission |
| 907 | node.elements.indexOf | 2 | native collection-to-object argument conversion clang error; JavaScript exit 70 with unknown intrinsic signature | collection receiver conversion, immutable intrinsic callable metadata and receiver-preserving checked calls |
| 910 | obj.properties.some | 2 | checked view read of field some with unsupported callable contract | original callback callable contract admission |

batch-06-map-probes.json uses actual Map receivers with complete generic Map declarations. Path and ResolvedConfigFilePath retain their original primitive brands. RedirectsCacheKey retains __compilerOptionsKey: any. EmitSignature, RelationComparisonResult, and ProgramUpdateLevel retain original declarations. Generic rank 565 instantiates THash as string and TElement as a structural object for its receiver probe; no generic certification is claimed.

Rank 877 retains activeTypeMappersCaches[activeTypeMappersCount].clear on an array, with an undefined guard for the fixture's indexed element. Its observed native stop is Map-to-object argument conversion; JavaScript exits 70 with an unknown intrinsic signature. An exploratory tuple scaffold stopped at ElementAccessExpression lowering and was replaced with the array scaffold before final checks. That tuple stop is not a code-needed row.

batch-06-array-probes.json retains the original indexOf declaration and node.elements.indexOf read on an actual array. batch-06-signature-probes.json retains the original some callback declaration, including its unknown return type. Exact JavaScript unknown-signature messages are pinned wherever the original contracts lower. Native conversion failures and lowering refusals remain separate observations; neither is credited as a runtime mutant.

code-dependencies.json retains 22 Set pairs / 48 reads naming codex/views-set-receiver 058635b9. No Set branch was merged or certified here.

Commands, with source /workspace/adamic-tools/env.sh and TMPDIR=/tmp/adamic-gate:

- ADAMIC_CALLABLE_SHARE_B_FAMILIES=batch-06-families.json ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableShareB' -count=1 -v -timeout 10m: PASS, 99.726s.
- go test ./internal/oracle -run '^TestCheckedViewCallableShareB(CollectionReceivers|AdmissionProbes)$' -count=1 -v -timeout 5m: PASS, 6.271s, after exact runtime message pins.
- ADAMIC_CALLABLE_SHARE_B_FAMILIES=batch-06-families.json python3 stage3/interface-downcasts/lane5/share-b/run-mutants.py: 20 arity mutants fail original exit/message pins; native and JavaScript execution controls pass. Per-rank controls and failed mutants are in controls.log and arity.log.
- ADAMIC_CALLABLE_SHARE_B_FAMILIES=batch-06-families.json node stage3/interface-downcasts/lane5/share-b/verify.cjs /tmp/lane5-b-original: 20 pairs / 40 reads / 95 fixtures.
- node stage3/interface-downcasts/lane5/share-b/verify.cjs /tmp/lane5-b-original: 100 pairs / 362 reads / 471 fixtures.
- ADAMIC_CALLABLE_SHARE_B_FAMILIES=batch-06-families.json ADAMIC_CALLABLE_SHARE_B_DECLARATION_MUTANT=799 node stage3/interface-downcasts/lane5/share-b/verify.cjs /tmp/lane5-b-original: exit 1, original declaration changed. The reportErrors parameter changes from boolean to number in memory.
- ADAMIC_CALLABLE_SHARE_B_FAMILIES=batch-06-families.json go test ./internal/oracle -run '^TestCheckedViewCallableShareBCounts$' -count=1 -args -update-counts: PASS, 13.820s. counts.md adds 95 rows and removes 0. Refusal and receiver-conversion probes have no executable counts rows.
- go test ./internal/oracle -run TestCountsAreRecorded -args -update-counts: FAIL, 56.588s, same 60 existing fixture failures as batch 05.
- go vet ./internal/oracle: exit 0.

Toolchain setup from batch 04 reused: GOPROXY=https://proxy.golang.org|direct, nproc=5; setup done 37.944s. Timing lines and output remain in logs/batch-04/setup.log.

Files changed: share-b fixtures, JSON ledgers, fixture harnesses, internal/oracle/checked_views_callable_share_b_test.go, appended internal/oracle/counts.md rows. No compiler/runtime edits, full package test, full gate, branch merge, or PR.

All share-b .cjs files pass node --check. Repository log copies remove trailing whitespace; raw outputs remain under /tmp/lane5-b-*06.log and /tmp/lane5-b-mutants/batch-06-families/.
