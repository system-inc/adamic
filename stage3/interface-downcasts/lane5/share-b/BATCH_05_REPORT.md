Certified: 17 new pairs / 43 reads / 81 fixtures; step 09.
Commits: base 51762724; batch 05 commit on codex/views-callables-b.
Commands: TestCheckedViewCallableShareB, TestCheckedViewCallableShareBCounts, verify.cjs, run-mutants.py, go vet ./internal/oracle; outputs in logs/batch-05.
Mutants: 17 arity mutants; rank 508 declaration, rank 529 enum, and rank 658 alias mutants.
Uncovered: 90 code-needed pairs / 1,500 reads; Set dependencies at codex/views-set-receiver 058635b9; optional-host, bind/condition, and remaining ranked pairs.

certified-pairs.json: 80 pairs / 322 reads. families.json: 376 fixtures. Original lane ledger plus share b: 236 pairs / 3,021 reads.

batch-05-families.json:

| Rank | Pair | Reads |
| --- | --- | ---: |
| 268 | DebugType.__debugTypeToString | 6 |
| 508 | BaseNodeFactory.createBaseNode | 3 |
| 529 | EmitHelperFactory.createClassPrivateFieldSetHelper | 3 |
| 625 | Program.getSourceFile | 3 |
| 658 | TypeCheckerHost.getDefaultResolutionModeForFile | 3 |
| 664 | Version.compareTo | 3 |
| 754 | EmitHelperFactory.createImportDefaultHelper | 2 |
| 757 | EmitHelperFactory.createValuesHelper | 2 |
| 760 | EmitResolver.createTypeOfDeclaration | 2 |
| 763 | EmitResolver.getJsxFactoryEntity | 2 |
| 766 | EmitResolver.getTypeReferenceSerializationKind | 2 |
| 769 | EmitResolver.isDeclarationWithCollidingName | 2 |
| 772 | EmitResolver.isValueAliasDeclaration | 2 |
| 775 | EmitTextWriter.write | 2 |
| 781 | EmitTextWriter.writeSymbol | 2 |
| 787 | FormatDiagnosticsHost.getCanonicalFileName | 2 |
| 790 | HasCurrentDirectory.getCurrentDirectory | 2 |

Original TypeScript checkout: 050880ce59e30b356b686bd3144efe24f875ebc8. verify.cjs checks complete member declarations, read expressions, UTF-16 spans, source hashes, enum values, and carrier declaration provenance. prepare.cjs now extracts type-literal members and class method headers. DebugType.__debugTypeToString uses the original cast expression inside a fixture class. Comparison, SyntaxKind, PrivateIdentifierKind, and TypeReferenceSerializationKind retain complete original enum declarations. ResolutionMode retains ModuleKind.ESNext | ModuleKind.CommonJS | undefined and the complete ModuleKind enum. Adjacent object carriers use the existing lane reductions.

TestCheckedViewCallableShareBFamilies checks Node, native release, ASan/UBSan, JavaScript, positive leak results, negative exit 70, and exact field/expected-type/found-type messages. The SyntaxKind wrong-members fixture supplies a string parameter; a number has the same runtime representation as this numeric enum and does not test an incompatible representation.

New code-needed pairs: 19 pairs / 65 reads.

| Rank | Read | Reads | Stop | Needed |
| --- | --- | ---: | --- | --- |
| 337 | state.seenEmittedFiles?.get | 5 | a primitive brand member __pathBrand whose type is not void | original Path brand admission preserving __pathBrand: any |
| 340 | sourceFileToPackageName.set | 5 | a primitive brand member __pathBrand whose type is not void | original Path brand admission preserving __pathBrand: any |
| 415 | builderProgram.state.fileInfos.has | 4 | a primitive brand member __pathBrand whose type is not void | original Path brand admission preserving __pathBrand: any |
| 418 | temporaryMarks.set | 4 | a primitive brand member __pathBrand whose type is not void | original Path brand admission preserving __pathBrand: any |
| 421 | context.remappedSymbolReferences.set | 4 | native collection-to-object argument conversion clang error; JavaScript exit 70 with unknown intrinsic signature | collection receiver conversion, immutable intrinsic callable metadata and receiver-preserving checked calls |
| 424 | fileDiagnostics.get | 4 | native collection-to-object argument conversion clang error; JavaScript exit 70 with unknown intrinsic signature | collection receiver conversion, immutable intrinsic callable metadata and receiver-preserving checked calls |
| 472 | programDiagnostics.getFileReasons | 4 | a primitive brand member __pathBrand whose type is not void | original Path brand admission preserving __pathBrand: any |
| 475 | state.referencedMap.getValues | 4 | a primitive brand member __pathBrand whose type is not void | original Path brand admission preserving __pathBrand: any |
| 544 | typeReferenceDirectives.push | 3 | checked view read of field push with unsupported callable contract | original rest parameter or generic callback callable contracts |
| 577 | node.elements.slice | 3 | native collection-to-object argument conversion clang error; JavaScript exit 70 with unknown intrinsic signature | collection receiver conversion, immutable intrinsic callable metadata and receiver-preserving checked calls |
| 592 | factory.createMethodSignature | 3 | (modifiers: readonly Modifier[] \| undefined, name: string \| PropertyName, questionToken: QuestionToken \| undefined, typeParameters: readonly TypeParameterDeclaration[] \| undefined, parameters: ..., type: TypeNode \| undefined) => MethodSignature | complete expected callable type display without parameters: ... |
| 619 | queue.pop | 3 | a primitive brand member __pathBrand whose type is not void | original Path brand admission preserving __pathBrand: any |
| 646 | parents!.map | 3 | checked view read of field map with unsupported callable contract | original rest parameter or generic callback callable contracts |
| 670 | secondaryRootErrors.unshift | 3 | stage 0 can't lower a tuple element of type string \| number yet | original tuple rest element string \| number lowering before callable certification |
| 673 | diagnostic.diagnostics.forEach | 3 | checked view read of field forEach with unsupported callable contract | original rest parameter or generic callback callable contracts |
| 676 | host.getSourceFiles().map | 3 | checked view read of field map with unsupported callable contract | original rest parameter or generic callback callable contracts |
| 679 | this.prerelease.join | 3 | native collection-to-object argument conversion clang error; JavaScript exit 70 with unknown intrinsic signature | collection receiver conversion, immutable intrinsic callable metadata and receiver-preserving checked calls |
| 739 | (this.mapper1 as unknown as DebugTypeMapper).__debugToString | 2 | adamic/no-unchecked-cast at (this.mapper1 as unknown as DebugTypeMapper) | checked admission for the original cast through unknown before the callable read |
| 751 | emitHelpers().createAwaiterHelper | 2 | (hasLexicalThis: boolean, argumentsExpression: Expression \| undefined, promiseConstructor: EntityName \| Expression \| undefined, parameters: ... \| undefined, body: Block) => Expression | complete expected callable type display without parameters: ... |

batch-05-map-probes.json and batch-05-array-probes.json use actual collections. Path aliases retain the original any-typed __pathBrand. ResolvedConfigFilePath retains ResolvedConfigFileName & Path. SortedArray retains its original Array base and brand. JavaScript unknown-signature messages are pinned for ranks 421, 424, 577, and 679. Native receiver-conversion errors are recorded separately from runtime checks.

batch-05-signature-probes.json retains original generic/rest declarations and calls the methods with their receivers. Rank 670 stops on the original tuple rest element string | number before callable certification. Rank 616's constructor signature admitted in an exploratory probe; constructor-producing positive fixtures and mutants remain uncovered, and no code-needed row is claimed for it.

batch-05-diagnostic-probes.json: ranks 592 and 751 execute against Node in all backends with sanitizers and leak checks. Their expected callable display replaces the parameters type with ...; TestCheckedViewCallableShareBDiagnosticNames pins that observation. These two pairs are code-needed and add two executable counts rows. Rank 739 retains the original cast through unknown and stops at adamic/no-unchecked-cast.

code-dependencies.json retains all 22 Set pairs / 48 reads naming codex/views-set-receiver 058635b9. No Set branch was merged or certified here.

Commands, with source /workspace/adamic-tools/env.sh and TMPDIR=/tmp/adamic-gate:

- ADAMIC_CALLABLE_SHARE_B_FAMILIES=batch-05-families.json ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableShareB' -count=1 -v -timeout 15m
- ADAMIC_CALLABLE_SHARE_B_FAMILIES=batch-05-families.json python3 stage3/interface-downcasts/lane5/share-b/run-mutants.py: 17 arity mutants fail original exit/message pins; native and JavaScript execution controls pass. Per-rank controls and mutant failures are in controls.log and arity.log.
- ADAMIC_CALLABLE_SHARE_B_FAMILIES=batch-05-families.json node stage3/interface-downcasts/lane5/share-b/verify.cjs /tmp/lane5-b-original: 17 pairs / 43 reads / 81 fixtures.
- node stage3/interface-downcasts/lane5/share-b/verify.cjs /tmp/lane5-b-original: 80 pairs / 322 reads / 376 fixtures.
- ADAMIC_CALLABLE_SHARE_B_FAMILIES=batch-05-families.json ADAMIC_CALLABLE_SHARE_B_DECLARATION_MUTANT=508 node stage3/interface-downcasts/lane5/share-b/verify.cjs /tmp/lane5-b-original: exit 1, original declaration changed.
- ADAMIC_CALLABLE_SHARE_B_FAMILIES=batch-05-families.json ADAMIC_CALLABLE_SHARE_B_ENUM_MUTANT=529 node stage3/interface-downcasts/lane5/share-b/verify.cjs /tmp/lane5-b-original: exit 1, original enum representation changed.
- ADAMIC_CALLABLE_SHARE_B_FAMILIES=batch-05-families.json ADAMIC_CALLABLE_SHARE_B_ALIAS_MUTANT=658 node stage3/interface-downcasts/lane5/share-b/verify.cjs /tmp/lane5-b-original: exit 1, carrier declaration changed.
- ADAMIC_CALLABLE_SHARE_B_FAMILIES=batch-05-families.json go test ./internal/oracle -run '^TestCheckedViewCallableShareBCounts$' -count=1 -args -update-counts: PASS, 12.344s. counts.md adds 83 rows and removes 0: 81 certified fixtures and 2 diagnostic probes. Refused probes and receiver conversion errors have no executable counts rows.
- go test ./internal/oracle -run TestCountsAreRecorded -args -update-counts: FAIL, 55.247s, same 60 existing fixture failures as batch 04.
- go vet ./internal/oracle: exit 0.

Toolchain setup from batch 04 reused: GOPROXY=https://proxy.golang.org|direct, nproc=5; setup done 37.944s. Timing lines and output remain in logs/batch-04/setup.log.

Files changed: share-b fixtures, JSON ledgers, fixture harnesses, internal/oracle/checked_views_callable_share_b_test.go, appended internal/oracle/counts.md rows. No compiler/runtime edits, full package test, full gate, branch merge, or PR.

Final scoped output: PASS, 218.452s. All share-b .cjs files pass node --check. git diff --check: exit 0. Repository log copies remove trailing whitespace; raw logs remain under /tmp/lane5-b-*05.log and /tmp/lane5-b-mutants/batch-05-families/.
