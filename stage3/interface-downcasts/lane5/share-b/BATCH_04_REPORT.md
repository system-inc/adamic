Certified: 19 new pairs / 63 reads / 90 fixtures; step 09.
Commits: base edf888d7; batch 04 commit on codex/views-callables-b.
Commands: TestCheckedViewCallableShareB, TestCheckedViewCallableShareBCounts, verify.cjs, run-mutants.py, go vet ./internal/oracle; outputs in logs/batch-04.
Mutants: 19 new arity mutants, 2 replacement arity mutants, rank 469 declaration and enum mutants, rank 466 array carrier mutant.
Uncovered: 71 code-needed pairs / 1,435 reads, including 22 Set pairs / 48 reads at codex/views-set-receiver 058635b9; optional-host and bind/condition families.

certified-pairs.json: 63 pairs / 279 reads. families.json: 295 current fixtures. Original lane ledger plus share b: 219 pairs / 2,978 reads.

batch-04-families.json:

| Rank | Pair | Reads |
| --- | --- | ---: |
| 373 | Program.getResolvedProjectReferences | 5 |
| 403 | EmitResolver.getReferencedExportContainer | 4 |
| 409 | HostForUseSourceOfProjectReferenceRedirect.getResolvedProjectReferences | 4 |
| 469 | Printer.writeNode | 4 |
| 490 | System.writeFile | 4 |
| 532 | EmitHelperFactory.createRewriteRelativeImportExtensionsHelper | 3 |
| 580 | NodeFactory.createAssignmentTargetWrapper | 3 |
| 583 | NodeFactory.createConstructorTypeNode | 3 |
| 586 | NodeFactory.createImportAttribute | 3 |
| 589 | NodeFactory.createLoopVariable | 3 |
| 595 | NodeFactory.createPostfixIncrement | 3 |
| 598 | NodeFactory.createThisTypeNode | 3 |
| 601 | NodeFactory.updateCaseBlock | 3 |
| 604 | NodeFactory.updateExportAssignment | 3 |
| 607 | NodeFactory.updateModuleBlock | 3 |
| 610 | NodeFactory.updateQualifiedName | 3 |
| 613 | NodeFactory.updateSwitchStatement | 3 |
| 649 | SyntacticTypeNodeBuilderResolver.getAllAccessorDeclarations | 3 |
| 661 | TypeCheckerHost.isSourceFileDefaultLibrary | 3 |

batch-04-array-carrier-families.json replaces ranks 457 and 466: 2 existing pairs / 8 reads / 10 fixtures. NodeArray<T> extends ReadonlyArray<T>; parameter and result fixtures use arrays. These replacements add no pair or read counts. Existing counts rows remain; replacement directories have new rows.

batch-04-check-families.json: 21 pairs / 71 reads / 100 fixtures. TestCheckedViewCallableShareBFamilies compares good fixtures with Node in native release, ASan/UBSan, and JavaScript; checks successful-run leaks; pins negative exits and field/type messages. Array wrong-result variants for ranks 373, 409, and 466 produce Node TypeError and Adamic exit 70. TestCheckedViewCallableShareBByteViewModule checks rank-169/node-byte-view.a in the same backends with leak checks.

Source: independent TypeScript checkout 050880ce59e30b356b686bd3144efe24f875ebc8. Original declarations, read expressions, UTF-16 spans, and source hashes are checked by verify.cjs. Adjacent object carriers use the existing lane reductions. EmitHint keeps its complete original enum; NodeArray retains its ReadonlyArray base. Rank 403 retains every original union member; the expected message uses the checker's displayed union order. Tracing probes retain original Phase and recursive Args declarations. fs.writeSync retains both @types/node 25.3.3 overloads; byte-view carriers are reduced in a separate .a module, and the probe tests signature admission. No fs host execution is certified.

New code-needed pairs:

| Rank | Read | Reads | Stop | Needed |
| --- | --- | ---: | --- | --- |
| 64 | `tracing?.push` | 26 | /workspace/adamic/stage3/interface-downcasts/lane5/share-b/rank-64/contract-probe.a:16:90: Adamic 0.1 refuses checked view read of field push with unsupported callable contract; prove or implement the callable contract before reading this field | callable parameter contracts for Phase and the recursive Args index signature |
| 97 | `topNamespace.parent.parent.symbol.exports?.get` | 18 | native Map-to-object argument conversion clang error; JavaScript exit 70 with unknown intrinsic signature | Map receiver conversion and immutable intrinsic signatures for the original get contract |
| 169 | `fs.writeSync` | 11 | /workspace/adamic/stage3/interface-downcasts/lane5/share-b/rank-169/contract-probe.a:30:85: Adamic 0.1 refuses checked view read of field writeSync with unsupported callable contract; prove or implement the callable contract before reading this field | overloaded external callable contracts and Node byte-view receiver/signature metadata |
| 229 | `optionsNameMap.get` | 7 | native Map-to-object argument conversion clang error; JavaScript exit 70 with unknown intrinsic signature | Map receiver conversion and immutable intrinsic signatures for the original get contract |
| 265 | `tracing?.instant` | 7 | /workspace/adamic/stage3/interface-downcasts/lane5/share-b/rank-265/contract-probe.a:16:90: Adamic 0.1 refuses checked view read of field instant with unsupported callable contract; prove or implement the callable contract before reading this field | callable parameter contracts for Phase and the recursive Args index signature |
| 280 | `parsedConfigs?.get` | 6 | /workspace/adamic/stage3/interface-downcasts/lane5/share-b/rank-280/actual-map.a:3:13: Adamic 0.1 refuses a primitive brand member __pathBrand whose type is not void; make __pathBrand void (or optional and typed undefined) so the brand is phantom | original Path brand admission or a proven checked adapter preserving its any-typed __pathBrand declaration |

code-dependencies.json contains the complete pair/read/change list. The Set rows retain codex/views-set-receiver 058635b9; that branch was not merged. Rank 655 belongs to the excluded bind family and adds no certification.

Commands, with /workspace/adamic-tools/env.sh sourced and TMPDIR=/tmp/adamic-gate:

- ADAMIC_CALLABLE_SHARE_B_FAMILIES=batch-04-check-families.json ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableShareB' -count=1 -v -timeout 10m
- ADAMIC_CALLABLE_SHARE_B_FAMILIES=batch-04-families.json python3 stage3/interface-downcasts/lane5/share-b/run-mutants.py: 19 arity mutants fail the original exit/message pins; native and JavaScript execution controls pass.
- ADAMIC_CALLABLE_SHARE_B_FAMILIES=batch-04-array-carrier-families.json python3 stage3/interface-downcasts/lane5/share-b/run-mutants.py: 2 arity mutants fail the original exit/message pins; native and JavaScript execution controls pass.
- ADAMIC_CALLABLE_SHARE_B_FAMILIES=batch-04-check-families.json node stage3/interface-downcasts/lane5/share-b/verify.cjs /tmp/lane5-b-original: 21 pairs / 71 reads / 100 fixtures.
- node stage3/interface-downcasts/lane5/share-b/verify.cjs /tmp/lane5-b-original: 63 pairs / 279 reads / 295 fixtures.
- ADAMIC_CALLABLE_SHARE_B_DECLARATION_MUTANT=469 with the batch-04-check-families.json verifier: exit 1, original declaration changed.
- ADAMIC_CALLABLE_SHARE_B_ENUM_MUTANT=469 with the same verifier: exit 1, original enum representation changed.
- ADAMIC_CALLABLE_SHARE_B_ARRAY_MUTANT=466 with the same verifier: exit 1, carrier declaration changed.
- go test ./internal/oracle -run TestCountsAreRecorded -args -update-counts: FAIL, 60 existing fixture failures, 58.756s.
- ADAMIC_CALLABLE_SHARE_B_FAMILIES=batch-04-check-families.json go test ./internal/oracle -run '^TestCheckedViewCallableShareBCounts$' -count=1 -args -update-counts: PASS, 30.820s. counts.md adds 101 rows and removes 0: 90 new certified fixtures, 10 replacements, 1 module. Refusal and receiver-conversion probes have no executable counts.
- go vet ./internal/oracle: exit 0.

Toolchain: export GOPROXY='https://proxy.golang.org|direct'; bash cloud/setup.sh; source /workspace/adamic-tools/env.sh. Setup exit 0; nproc=5. Timing lines: node 0.022s, go 0.022s, submodules 0.071s, markdown 0.075s, clang 0.198s, go build 37.805s, test binaries deferred 37.917s, build cache 37.919s, done 37.944s. Full output: logs/batch-04/setup.log.

Files changed: share-b fixtures and harnesses, internal/oracle/checked_views_callable_share_b_test.go, appended internal/oracle/counts.md rows. Compiler/runtime files unchanged. No full package test, full gate, branch merge, or PR.

Final scoped output: PASS, 106.507s. git diff --check: exit 0. Repository log copies remove trailing whitespace; raw outputs remain under /tmp/lane5-b-*04.log and /tmp/lane5-b-mutants/.
