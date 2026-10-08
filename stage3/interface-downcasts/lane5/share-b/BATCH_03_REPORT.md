Built 10 additional original-member callable certifications / 47 candidate reads, serving step 09.
Commit: this checkpoint on codex/views-callables-b; previous checkpoint 3df4c66e.
Checks: 47 certified fixtures held to Node, native release, ASan/UBSan, JavaScript, and successful-run leak checks; scoped commands below.
Mutants: all 10 arity mutants killed by original exit/message pins with valid native and JavaScript execution controls; declaration and enum provenance mutants killed.
Uncovered: 23 newly observed code-needed pairs / 248 reads; Set dependencies retained; remaining ranked pairs, optional-host methods, and bind/condition families not certified here.

Share b now carries 44 certified pairs / 216 candidate reads in 205 certified fixtures. With the original lane ledger, this branch carries 200 certified pairs / 2,915 reads. No compiler/runtime edits, protected-file edits, other lane merges, or PRs.

Certified ranks:

| Rank | Pair | Reads |
| --- | --- | ---: |
| 175 | NodeFactory.createModuleDeclaration | 10 |
| 214 | NodeFactory.updateForOfStatement | 8 |
| 292 | NodeFactory.createImportEqualsDeclaration | 6 |
| 442 | NodeFactory.createPropertyDescriptor | 4 |
| 466 | ParenthesizerRules.parenthesizeExpressionsOfCommaDelimitedList | 4 |
| 526 | EmitHelperFactory.createAssignHelper | 3 |
| 535 | EmitResolver.isBindingCapturedByNode | 3 |
| 538 | EmitTextWriter.getTextPos | 3 |
| 541 | EmitTextWriter.writeLiteral | 3 |
| 547 | HostForUseSourceOfProjectReferenceRedirect.getSymlinkCache | 3 |

Fixtures retain complete original member declarations and original reads. Adjacent object carriers remain reductions under the lane's existing convention. Rank 175 retains the entire original NodeFlags enum with its scalar representation and values; the verifier pins it to the independent original types.ts. Removed export/const qualifiers permit a standalone runtime fixture without changing enum values or parameter representation. Rank 466 preserves the original parenthesizerRules() receiver call. The independent source remains pinned at 050880ce59e30b356b686bd3144efe24f875ebc8.

Conservative assumption for rank 535: the original declaration displays BindingElement | VariableDeclaration while the ranked ledger displays VariableDeclaration | BindingElement. Both members are retained unchanged. The pinned runtime message follows the original declaration's display order; this does not change the union's types.

Namespace probes preserve full original Debug function headers and all overloads. Namespace members are represented as callable properties, matching the existing lane's Debug fixture convention. The original assertNever message default is represented as optional string in the property signature because default initializers are not allowed in interface declarations. Original headers and source hashes are included in the evidence. The original AnyFunction alias remains (...args: never[]) => void.

New code-needed pairs:

| Rank | Pair | Reads | Needed |
| --- | --- | ---: | --- |
| 19 | typeof Debug.assertNever | 62 | never parameter/result callable contracts and optional AnyFunction rest-never callback contract |
| 43 | typeof Debug.assertIsDefined | 39 | proven original assertion-predicate callable admission, including generic overloads |
| 46 | typeof Debug.assertNode | 38 | proven original assertion-predicate callable admission, including generic overloads |
| 112 | Map<string, boolean>.get | 15 | Map receiver conversion, immutable intrinsic callable metadata, and receiver-preserving checked method dispatch |
| 184 | NodeArray<Statement>.slice | 9 | array receiver conversion, immutable intrinsic callable metadata, and receiver-preserving checked method dispatch |
| 232 | Map<string, string>.has | 7 | Map receiver conversion, immutable intrinsic callable metadata, and receiver-preserving checked method dispatch |
| 259 | number[].pop | 7 | array receiver conversion, immutable intrinsic callable metadata, and receiver-preserving checked method dispatch |
| 262 | typeof Debug.assertLessThanOrEqual | 7 | nested optional AnyFunction callback contract with rest-never parameters |
| 283 | Map<string, number>.clear | 6 | Map receiver conversion, immutable intrinsic callable metadata, and receiver-preserving checked method dispatch |
| 343 | Map<number, Type>.set | 5 | Map receiver conversion, immutable intrinsic callable metadata, and receiver-preserving checked method dispatch |
| 346 | Map<string, boolean>.has | 5 | Map receiver conversion, immutable intrinsic callable metadata, and receiver-preserving checked method dispatch |
| 370 | Path[].push | 5 | original rest-parameter or callback callable descriptors and immutable intrinsic metadata |
| 382 | Statistic[].push | 5 | original rest-parameter or callback callable descriptors and immutable intrinsic metadata |
| 388 | string[].sort | 5 | original rest-parameter or callback callable descriptors and immutable intrinsic metadata |
| 406 | FileIncludeReason[]  /  undefined.forEach | 4 | original rest-parameter or callback callable descriptors and immutable intrinsic metadata |
| 439 | NodeFactory.createImportAttributes | 4 | original overloaded callable descriptors including indexed result-token types |
| 448 | NodeFactory.updateArrowFunction | 4 | complete untruncated expected callable type in checked-view failure messages |
| 463 | ObjectConstructor.create | 4 | checked representation for the original any-returning overloaded callable |
| 496 | number[].map | 4 | original rest-parameter or callback callable descriptors and immutable intrinsic metadata |
| 499 | readonly string[].includes | 4 | array receiver conversion, immutable intrinsic callable metadata, and receiver-preserving checked method dispatch |
| 511 | CaseClause[].push | 3 | original rest-parameter or callback callable descriptors and immutable intrinsic metadata |
| 520 | Declaration[].slice | 3 | array receiver conversion, immutable intrinsic callable metadata, and receiver-preserving checked method dispatch |
| 523 | Diagnostic[].forEach | 3 | original rest-parameter or callback callable descriptors and immutable intrinsic metadata |

The full list is code-dependencies.json: 65 pairs / 1,360 reads, including previous blockers. Every entry names its original read and observed stop. The nine actual collection receiver fixtures retain the original generic Map/Array declarations and instantiate their parameters for actual collections. Native stops on the collection-to-object argument conversion; JavaScript stops at exit 70 with an unknown intrinsic signature. Exact stderr is pinned. The original polymorphic set result `this` is represented as the Target<number, Type> view in that probe; no certification of it is claimed.

Set rows remain user-provided dependencies, not observations against the fix: 22 pairs / 48 reads naming codex/views-set-receiver 058635b9. No Set branch was merged. The Set arm of the mixed ReadonlyMap/ReadonlySet receiver still has other obligations open.

Rank 448 executes correctly in all backends and leaks nothing, but its expected-type diagnostic abbreviates the required parameter type as `parameters: ...`. Its complete original-type display pin failed, so it is recorded as code-needed for an untruncated diagnostic, not certified. That executable fixture receives a counts row.

Commands (test output is in logs/batch-03):

- ADAMIC_CALLABLE_SHARE_B_FAMILIES=batch-03-families.json ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableShareB' -count=1 -v -timeout 10m
- ADAMIC_CALLABLE_SHARE_B_FAMILIES=batch-03-families.json python3 stage3/interface-downcasts/lane5/share-b/run-mutants.py: 10 arity mutants caught; both-backend valid-execution controls pass. Each rank above has its own successful wrong execution control and failed exit-70/message-pin test. No clang/sanitizer failure is credited as a mutant kill.
- ADAMIC_CALLABLE_SHARE_B_FAMILIES=batch-03-families.json node stage3/interface-downcasts/lane5/share-b/verify.cjs /tmp/lane5-b-original: Verified 10 pairs / 47 candidate reads in 47 original-member fixtures.
- node stage3/interface-downcasts/lane5/share-b/verify.cjs /tmp/lane5-b-original: Verified 44 pairs / 216 candidate reads in 205 original-member fixtures.
- ADAMIC_CALLABLE_SHARE_B_FAMILIES=batch-03-families.json ADAMIC_CALLABLE_SHARE_B_DECLARATION_MUTANT=535 node stage3/interface-downcasts/lane5/share-b/verify.cjs /tmp/lane5-b-original: rejected changing the original Node parameter annotation to number.
- ADAMIC_CALLABLE_SHARE_B_FAMILIES=batch-03-families.json ADAMIC_CALLABLE_SHARE_B_ENUM_MUTANT=175 node stage3/interface-downcasts/lane5/share-b/verify.cjs /tmp/lane5-b-original: rejected changing the original NodeFlags.Let value. Both mutations are in-memory only; disk fixtures stay unchanged.
- go test ./internal/oracle -run TestCountsAreRecorded -args -update-counts: required global attempt failed on 60 existing fixture subtests (41.694s), including the existing fresh_refused/set_add.a invalid-pointer failure. No global table rewrite is credited.
- ADAMIC_CALLABLE_SHARE_B_FAMILIES=batch-03-families.json go test ./internal/oracle -run '^TestCheckedViewCallableShareBCounts$' -count=1 -args -update-counts: executable fixtures only. The initial attempt to count the actual receiver gaps failed at the already observed Map conversion clang error; those nine cannot yield valid native counts until their listed code change lands. The refusal probes also cannot produce executable counts. No rows were fabricated for either group.
- go vet ./internal/oracle

The existing successful toolchain setup is reused; GOPROXY=https://proxy.golang.org|direct, nproc=5. Setup commands and timing lines remain in REPORT.md. No full package test or full gate was run.

Exploratory failures were corrected before final checks: receiver scaffolds for context.enclosingSymbolTypes and conditions?.includes; namespace assertion targets needing explicit annotations; and generator control flow for optional/array observations. A numeric-to-never cast was rejected before the callable read, so the final assertNever probe reads the original namespace function as a value. None of these exploratory failures was credited as a code blocker. Rank 460 admits its constructor-return signature and remains un-certified, recorded in batch-03-admitted-pending.json. The remaining prior tracked admitted pairs are ranks 229 and 280. Tracing namespace signatures, SymbolTable branding, external fs declarations, and other remaining higher-read ranks still need original-receiver probes before the next certification batch.

Push only to codex/views-callables-b after restored scoped tests, counts verification, and diff checks.

Final observed outputs: restored scoped check PASS (40.996s); scoped executable counts PASS (6.209s); go vet exit 0; source verifier 44 pairs / 216 reads / 205 certified fixtures. Counts diff is 48 added rows / 0 removed rows: 47 certified fixtures plus the executable rank 448 diagnostic gap.

Staged whitespace review normalized copied TypeScript enum line endings to LF without changing declarations or values. Rank 175 checks and counts were rerun afterward. Repository log copies remove trailing line whitespace; raw command outputs remain under /tmp/lane5-b-*03.log.

Affected-input verification after LF normalization: ADAMIC_CALLABLE_SHARE_B_FAMILIES=batch-03-families.json ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableShareB(Families|ArityMutants|Counts)$/^rank-175' -count=1 -v -timeout 5m passed (11.561s). The rank-175 wrong-arity test with ADAMIC_CALLABLE_SHARE_B_MUTANT=arity failed its original message/exit pin as required, with no clang/build failure. All owned .cjs files pass node --check.
