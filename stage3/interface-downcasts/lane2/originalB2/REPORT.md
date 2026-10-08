Certified 20 original array-field pairs toward roadmap step 09.
Commits: previous delivery 57fa2c98ed282dbe86afb3263ab68237670b55e4; this delivery SHA is reported with the push.
Checks: 144 controls and 22 mutants pass in 76.796s; 144 scoped count rows pass in 149.695s; required global counts remain red; vet passes.
Mutants: 20 reached element or descendant contract replacements and two missing-field admissions are caught in all three execution modes.
Not covered: delegated mixed-member and union-target work, unread synthesized-comment text, later pairs and runtime corpus reachability.

This is the second 20-pair delivery on codex/views-arrays-b. It uses the requested 70522aa1d base with only this worker's commits. No production code changed and no lane branch was merged. The newest worker report was read at 5dcbad439f5246ae43ceafc644e29858d5ed5052. It has reached three-read modifier and JSX arrays; this batch remains entirely within the one-read tail. progress.json records every scanned pair and the conservative handoffs retained from the first delivery. The next pair is Deferred TypeMapper.targets.

| Original pair | Type ID | Static reads | Result |
| --- | ---: | ---: | --- |
| `ConfigFileSpecs.validatedFilesSpecBeforeSubstitution` | 9853 | 1 | certified |
| `ConfigFileSpecs.validatedFilesSpec` | 9853 | 1 | certified |
| `Bundle.syntheticTypeReferences` | 9844 | 1 | certified |
| `Bundle.syntheticLibReferences` | 9844 | 1 | certified |
| `Bundle.syntheticFileReferences` | 9844 | 1 | certified |
| `Block | undefined.statements` | 9227 | 1 | certified |
| `AnonymousType.properties` | 8635 | 1 | certified |
| `AnonymousType.callSignatures` | 8635 | 1 | certified |
| `SignatureDeclaration.typeParameters` | 7331 | 1 | certified |
| `IndexSignatureDeclaration.typeParameters` | 7323 | 1 | certified |
| `SourceFile.jsDocDiagnostics` | 6998 | 1 | certified |
| `Identifier.jsDoc` | 6997 | 1 | certified |
| `EmitNode | undefined.trailingComments` | 6995 | 1 | certified |
| `EmitNode | undefined.tokenSourceMapRanges` | 6995 | 1 | certified |
| `EmitNode | undefined.leadingComments` | 6995 | 1 | certified |
| `EmitNode | undefined.identifierTypeArguments` | 6995 | 1 | certified |
| `EmitNode | undefined.annotatedNodes` | 6995 | 1 | certified |
| `EmitNode.trailingComments` | 6994 | 1 | certified |
| `EmitNode.leadingComments` | 6994 | 1 | certified |
| `EmitNode.annotatedNodes` | 6994 | 1 | certified |

Preparation invokes the existing declaration adapter by reference against pristine TypeScript 050880ce59e30b356b686bd3144efe24f875ebc8. All 79 complete declaration files are hash checked, including complete private watch declarations. Every original span and declared array-field type is checked. Receiver and selected element field sets are checked against the lowered descriptors. Original union arms remain declared; controls select one arm and make no claim of executing every arm.

Every pair has a valid read, lazy malformed array, valid first element followed by an unread malformed element, a stopping read and a source-alias mutation. Optional fields and receivers retain absent, explicit undefined and undefined receiver controls. Required ConfigFileSpecs fields whose types admit undefined retain separate missing-field stopping messages. SynthesizedComment reads use its original CommentKind, with valid kind 2 changing through the source alias to valid kind 3. The original text member remains present and unread. An initial text-demand probe observed an unsupported callable-contract refusal despite the declared string type; that member needs demand-scope work and receives no member-read claim. The first run is retained. No declaration was reduced.

The original TypeNode | TypeParameterDeclaration controls initially lacked a required kind. They were corrected to provide the original kind before the structural union selection, and the filtered check passed in 4.773s. The final complete run passes: all 144 sources agree with Node, all 144 compiled probes meet their pinned normal or stopping outcomes in ASan/UBSan native, release native and JavaScript. All 122 finishing controls and all 22 finishing mutants pass native leak checks.

mutants.json records each independent execution and exact stopping catcher. String consumer mutants widen to string | number and print 42. Numeric descendant mutants replace one reached read and its conversion with a string read and print bad. Missing-field mutants admit one absent required ConfigFileSpecs member and print absent. Each mutation changes exactly one operation, executes successfully with empty stderr and exit zero, and is caught by the original exit-70 message. No clang rejection or sanitizer failure is credited.

The scoped count refresh adds exactly 144 own rows. count-rows.json explains every addition. Removing originalB2 rows reproduces the previous delivery's counts.md byte for byte. The required global command failed in 60.022s in existing graph-region invalid frees, process.exit and fs fixtures. Its full log is retained; this is not a successful global refresh. No whole package test or full gate ran.

The worker's newest ledger records 192 pairs / 2927 static reads. With this worker's disjoint 39 certified pairs / 39 reads across the first two deliveries, the combined scheduling ledger is 231 / 2966, leaving 103 / 223 of 334 / 3189. These are static candidate obligations; runtime reachability remains unmeasured. Consumer, intrinsic and own-array-field totals receive no new credit.

Exact commands, with test output sent directly to logs:

```sh
source /workspace/adamic-tools/env.sh
node stage3/interface-downcasts/lane2/originalB2/prepare.cjs /workspace/lane2-b-original-pin /workspace/lane2-b-declarations2 > /tmp/lane2-b2-prepare.log 2>&1
ADAMIC_ARRAYB2_ORIGINAL_DECLS=/workspace/lane2-b-declarations2 ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewArraysB2Original$|^TestCheckedViewArraysB2Mutants$' -count=1 -parallel 4 -v -timeout 20m > /tmp/lane2-b2-final.log 2>&1
ADAMIC_ARRAYB2_ORIGINAL_DECLS=/workspace/lane2-b-declarations2 ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewArraysB2Counts$' -count=1 -timeout 15m -args -update-counts > /tmp/lane2-b2-counts.log 2>&1
ADAMIC_ARRAYB2_ORIGINAL_DECLS=/workspace/lane2-b-declarations2 ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 20m -args -update-counts > /tmp/lane2-b2-global-counts.log 2>&1
go vet ./internal/oracle > /tmp/lane2-b2-vet.log 2>&1
git diff --check > /tmp/lane2-b2-diff-check.log 2>&1
```

Setup from the first delivery is reused: done 45.313s, nproc 5, cpu.max 400000 100000, Go 1.27.1, clang 20.1.8 and Node 24.19.0. Complete timing lines remain in originalB/REPORT.md and evidence/setup.log.gz. Evidence is retained as compressed logs; declarations remain external with repeatable preparation and hashes.
