Certified: 20 new pairs / 28 reads / 89 fixtures; step 09.
Commits: base 5ebb8087; batch 08 commit on codex/views-callables-b.
Commands: TestCheckedViewCallableShareB, TestCheckedViewCallableShareBCounts, verify.cjs, run-mutants.py, go vet ./internal/oracle; outputs in logs/batch-08.
Mutants: 20 arity mutants; declaration mutant 1117 exits 1.
Uncovered: 134 code-needed pairs / 1,590 reads; Set dependencies at codex/views-set-receiver 058635b9; optional-host, bind/condition, indexed aliases, and remaining ranked pairs.

certified-pairs.json: 140 pairs / 430 reads. families.json: 653 fixtures. Original lane ledger plus share b: 296 pairs / 3,129 reads.

batch-08-families.json:

| Rank | Pair | Reads |
| --- | --- | ---: |
| 1099 | SymbolTrackerImpl.reportCyclicStructureError | 2 |
| 1108 | SyntacticNodeBuilder.serializeReturnTypeForSignature | 2 |
| 1111 | SyntacticTypeNodeBuilderResolver.serializeTypeName | 2 |
| 1114 | SyntaxCursor.currentNode | 2 |
| 1117 | System.getDirectories | 2 |
| 1135 | TransformationContext.suspendLexicalEnvironment | 2 |
| 1141 | TypeChecker.getEmitResolver | 2 |
| 1144 | TypeCheckerHost.getCompilerOptions | 2 |
| 1288 | BuilderProgram.getProgram | 1 |
| 1297 | CachedDirectoryStructureHost.clearCache | 1 |
| 1315 | CompilerHost.getDefaultLibFileName | 1 |
| 1360 | CustomTransformer.transformSourceFile | 1 |
| 1378 | DirectoryWatcher.close | 1 |
| 1381 | EmitHelperFactory.createAddDisposableResourceHelper | 1 |
| 1384 | EmitHelperFactory.createClassPrivateFieldInHelper | 1 |
| 1387 | EmitHelperFactory.createGeneratorHelper | 1 |
| 1390 | EmitHelperFactory.createRestHelper | 1 |
| 1393 | EmitHelperFactory.getUnscopedHelperName | 1 |
| 1405 | EmitResolver.getJsxFragmentFactoryEntity | 1 |
| 1417 | EmitTextWriter.getColumn | 1 |

Original TypeScript checkout: 050880ce59e30b356b686bd3144efe24f875ebc8. Fixtures retain complete original callable declarations and original read expressions. verify.cjs checks declarations, UTF-16 spans, source hashes, and carrier declarations. Adjacent object carriers use the existing lane reductions. Rank 1117 retains string[] and returns an actual string array; its elements are observed as strings. generate.cjs now creates the original nested getter receiver for flattenContext.context.getEmitHelperFactory().createRestHelper.

Rank 1108 retains SignatureDeclaration and JSDocSignature in the original declaration. The checker displays JSDocSignature | SignatureDeclaration; the exact runtime message pin follows that display order. No union member is removed or changed.

TestCheckedViewCallableShareBFamilies passes Node comparisons in native release, ASan/UBSan, and JavaScript; positive leak checks; negative exit 70 and exact field/expected-type/found-type messages. The string-array wrong-result fixture produces Node TypeError and the pinned Adamic incompatible-result exit 70.

New code-needed pairs: 14 pairs / 22 reads.

| Rank | Read | Reads | Stop | Needed |
| --- | --- | ---: | --- | --- |
| 1102 | getPropertiesOfType(target)             .filter | 2 | adamic/no-type-predicate | original predicate callable contracts with a proven predicate return |
| 1105 | symlinkCache.getSymlinkedDirectories | 2 | a primitive brand member __pathBrand whose type is not void | original Path brand admission preserving __pathBrand: any |
| 1150 | type.target.localTypeParameters?.map | 2 | checked view read of field map with unsupported callable contract | original rest parameter or generic callback callable contracts |
| 1165 | awaitedTypeStack.lastIndexOf | 2 | native array-to-object argument conversion clang error; JavaScript exit 70 with unknown intrinsic signature | array receiver conversion, immutable intrinsic callable metadata and receiver-preserving checked calls |
| 1168 | simpleOptions.filter | 2 | adamic/no-type-predicate | original predicate callable contracts with a proven predicate return |
| 1171 | sourceFiles.indexOf | 2 | native array-to-object argument conversion clang error; JavaScript exit 70 with unknown intrinsic signature | array receiver conversion, immutable intrinsic callable metadata and receiver-preserving checked calls |
| 1174 | referenceType.resolvedTypeArguments?.map | 2 | checked view read of field map with unsupported callable contract | original rest parameter or generic callback callable contracts |
| 1177 | array.map | 2 | stage 0 can't lower a value of type unknown yet | original unknown-valued callback parameter lowering |
| 1306 | commonOptionsWithBuild.forEach | 1 | checked view read of field forEach with unsupported callable contract | original rest parameter or generic callback callable contracts |
| 1309 | commentDirectives.push | 1 | checked view read of field push with unsupported callable contract | original rest parameter or generic callback callable contracts |
| 1363 | overloadSignatures.indexOf | 1 | native array-to-object argument conversion clang error; JavaScript exit 70 with unknown intrinsic signature | array receiver conversion, immutable intrinsic callable metadata and receiver-preserving checked calls |
| 1366 | input.symbol.declarations?.filter | 1 | adamic/no-type-predicate | original predicate callable contracts with a proven predicate return |
| 1372 | relatedInformation.map | 1 | checked view read of field map with unsupported callable contract | original rest parameter or generic callback callable contracts |
| 1375 | getTypeChecker().getGlobalDiagnostics().slice | 1 | native array-to-object argument conversion clang error; JavaScript exit 70 with unknown intrinsic signature | array receiver conversion, immutable intrinsic callable metadata and receiver-preserving checked calls |

batch-08-array-probes.json retains original generic Array/ReadonlyArray declarations and uses actual arrays. Exact JavaScript unknown-signature messages are pinned; native array-to-object conversion errors are recorded separately. batch-08-signature-probes.json retains all filter overloads, original generic callbacks, and unknown parameter types. Rank 1177 stops on unknown-value lowering before callable admission. Rank 1105 retains ReadonlyMap<Path, false | SymlinkedDirectory> and the original Path brand.

Uncovered original contexts include the Collator bound method read at rank 1303, logical assignment receivers at ranks 1096 and 1132, and the indexed Program member alias at rank 1396. No certification or code-needed observation is claimed for those contexts in this batch. The rank 1102 fixture normalizes the original read's CRLF to LF; its AST expression and original UTF-16 evidence remain pinned.

code-dependencies.json retains 22 Set pairs / 48 reads naming codex/views-set-receiver 058635b9. No Set branch was merged or certified here.

Commands, with source /workspace/adamic-tools/env.sh and TMPDIR=/tmp/adamic-gate:

- ADAMIC_CALLABLE_SHARE_B_FAMILIES=batch-08-families.json ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableShareB' -count=1 -v -timeout 10m: PASS, 102.176s.
- go test ./internal/oracle -run '^TestCheckedViewCallableShareB(CollectionReceivers|AdmissionProbes)$' -count=1 -v -timeout 5m: PASS, 8.581s, after exact collection runtime message pins.
- ADAMIC_CALLABLE_SHARE_B_FAMILIES=batch-08-families.json python3 stage3/interface-downcasts/lane5/share-b/run-mutants.py: 20 arity mutants fail original exit/message pins; native and JavaScript execution controls pass. Per-rank controls and failed mutants are in controls.log and arity.log.
- ADAMIC_CALLABLE_SHARE_B_FAMILIES=batch-08-families.json node stage3/interface-downcasts/lane5/share-b/verify.cjs /tmp/lane5-b-original: 20 pairs / 28 reads / 89 fixtures.
- node stage3/interface-downcasts/lane5/share-b/verify.cjs /tmp/lane5-b-original: 140 pairs / 430 reads / 653 fixtures.
- ADAMIC_CALLABLE_SHARE_B_FAMILIES=batch-08-families.json ADAMIC_CALLABLE_SHARE_B_DECLARATION_MUTANT=1117 node stage3/interface-downcasts/lane5/share-b/verify.cjs /tmp/lane5-b-original: exit 1, original declaration changed. The path parameter changes from string to number in memory.
- ADAMIC_CALLABLE_SHARE_B_FAMILIES=batch-08-families.json go test ./internal/oracle -run '^TestCheckedViewCallableShareBCounts$' -count=1 -args -update-counts: PASS, 13.250s. counts.md adds 89 rows and removes 0. Refusal and receiver-conversion probes have no executable counts rows.
- go test ./internal/oracle -run TestCountsAreRecorded -args -update-counts: FAIL, 56.077s, same 60 existing fixture failures as batch 07.
- go vet ./internal/oracle: exit 0.

Toolchain setup from batch 04 reused: GOPROXY=https://proxy.golang.org|direct, nproc=5; setup done 37.944s. Timing lines and output remain in logs/batch-04/setup.log.

Files changed: share-b fixtures, JSON ledgers, fixture harnesses, internal/oracle/checked_views_callable_share_b_test.go, appended internal/oracle/counts.md rows. No compiler/runtime edits, full package test, full gate, branch merge, or PR.

All share-b .cjs files pass node --check. Repository log copies remove trailing whitespace; raw outputs remain under /tmp/lane5-b-*08.log and /tmp/lane5-b-mutants/batch-08-families/.
