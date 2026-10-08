Certified eight of the final ten original array-field pairs / sixteen static reads toward roadmap step 09.
Commits: previous delivery 4bf6da09a3c8a35c99ae76b26bb892879de0ee90; the final delivery SHA accompanies its push.
Checks: 68 sources, 56 runtime controls and nine mutants pass in 33.815s; 56 scoped count rows pass in 55.436s; global counts remain red; vet passes.
Mutants: eight reached numeric member replacements and one missing-field admission are caught in sanitized native, release native and JavaScript.
Not covered: optional SourceFile.imports conversion, optional EmitNode.helpers.name admission, retained handoffs and production runtime reachability.

This final delivery contains ten pairs because the eligible tail ends at the worker's three-read rank. The two preceding deliveries contain twenty pairs each. Certification stops before CircularBuildOrder.buildOrder, the first three-read row encountered from the bottom. The worker's newest report at 7cd0281388208bd439b5a20d4f4294a55a4ac91e was read. Every selected pair across all four deliveries is disjoint from every published original-pair config on that tip. No other lane or worker branch was merged. No production code changed.

| Original pair | Type ID | Static reads | Result |
| --- | ---: | ---: | --- |
| `PrivateIdentifierPropertyDeclaration.modifiers` | 9209 | 2 | certified |
| `UnionOrIntersectionType.resolvedProperties` | 8655 | 2 | certified |
| `UnionType.arrayFallbackSignatures` | 8654 | 2 | certified |
| `InterfaceType.outerTypeParameters` | 8600 | 2 | certified |
| `SourceFile | undefined.imports` | 7378 | 2 | code needed |
| `SignatureDeclaration.typeArguments` | 7331 | 2 | certified |
| `SignatureDeclaration.jsDoc` | 7331 | 2 | certified |
| `JSDocFunctionType.typeParameters` | 7327 | 2 | certified |
| `SourceFile.additionalSyntacticDiagnostics` | 6998 | 2 | certified |
| `EmitNode | undefined.helpers` | 6995 | 2 | code needed |

Scheduling correction: PrivateIdentifierPropertyDeclaration and UnionOrIntersectionType are named interfaces in the original declarations. The first tail scan incorrectly treated their names as union-target handoffs. They are now included in reverse ranked order, with complete original declarations and independent mutants. The previous delivery report's remaining-tail count is corrected from eight to ten. Other explicit union-target, mixed-member and own-array-field handoffs retain their owners; progress.json records every disposition. JSDocArray.jsDocCache remains a separate own-array-field production joint, as recorded in RANKED6_ARRAYS_REPORT.md.

Preparation preserves all 79 complete declarations from pristine TypeScript 050880ce59e30b356b686bd3144efe24f875ebc8, verifies every original span and declared array-field type, and checks complete receiver and selected element field sets against lowered metadata. Fixtures include valid reads, unread malformed arrays, later unread malformed elements, pinned stopping reads, source aliases, optional fields and optional receivers. No declaration was reduced and no checker option changed.

All 68 sources run on Node with pinned stdout, empty stderr and exit zero. The 56 executable controls meet their normal or exact stopping result in ASan/UBSan native, release native and JavaScript. All 47 finishing controls and all nine finishing mutants pass native leak checks. Twelve demanded sources pin compile refusals; the two unread malformed-array controls for those pairs still run successfully. Unsupported demanded reads receive no pair or mutant credit.

Observed code requirements:

- SourceFile | undefined.imports: five demanded controls refuse `Adamic 0.1 refuses checked view read of field imports with unsupported representation conversion contract; prove or implement the representation conversion contract before reading this field`. The original array member is readonly StringLiteralLike[]. Checked conversion for this optional receiver's array of original union elements is needed. Its lazy array control passes.
- EmitNode | undefined.helpers: seven demanded controls refuse `Adamic 0.1 refuses checked view read of field name with unsupported never contract; prove or implement the never contract before reading this field`. The original EmitHelper name member is string. Its member demand needs investigation and implementation without collapsing the original type to never. Its lazy array control passes. This observation does not establish the underlying cause.

mutants.json records all nine independent executions. Eight replace one reached numeric member read and its conversion with a string read, printing bad. One admits a missing required nullable additionalSyntacticDiagnostics field and prints absent. Every mutant changes exactly one operation, executes with the expected Node stdout and empty stderr, passes sanitizers and leaks, and disagrees with the unchanged exit-70 stopping oracle. Clang failures and sanitizer failures are not credited.

The final scoped count update records exactly 56 new own rows. count-rows.json explains every addition; the twelve compile refusals have no runtime rows. Removing the own prefix reproduces the previous delivery's table byte for byte. The initial eight-pair measurement was rerun after adding the two corrected interfaces; the final complete measurement is authoritative. The required global refresh failed in 50.298s in existing graph-region invalid frees, process.exit, fs and other baseline fixtures. It is not a successful global refresh. No whole package or full gate ran.

Across this continuation, three pushes process 20, 20 and 10 pairs. Forty-eight pairs / seventy-four static reads are newly certified, with two code frontiers retained. There are 338 Node source controls, 326 executable controls and 53 independent mutants; all focused checks pass. All 326 new runtime count rows are isolated to originalB2, originalB3 and originalB4. This worker's four deliveries total 67 certified pairs / 93 reads, with three code frontiers. The worker reports 202 / 2957, giving a disjoint combined scheduling ledger of 269 / 3050 and 65 / 139 remaining out of 334 / 3189. These are static candidate obligations. Runtime reachability remains unmeasured; consumer, intrinsic and own-array-field totals receive no new credit.

```sh
source /workspace/adamic-tools/env.sh
node stage3/interface-downcasts/lane2/originalB4/prepare.cjs /workspace/lane2-b-original-pin /workspace/lane2-b-declarations4 > /tmp/lane2-b4-prepare.log 2>&1
ADAMIC_ARRAYB4_ORIGINAL_DECLS=/workspace/lane2-b-declarations4 ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewArraysB4Original$|^TestCheckedViewArraysB4Mutants$' -count=1 -parallel 4 -v -timeout 20m > /tmp/lane2-b4-complete.log 2>&1
ADAMIC_ARRAYB4_ORIGINAL_DECLS=/workspace/lane2-b-declarations4 ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewArraysB4Counts$' -count=1 -timeout 15m -args -update-counts > /tmp/lane2-b4-counts-complete.log 2>&1
ADAMIC_ARRAYB4_ORIGINAL_DECLS=/workspace/lane2-b-declarations4 ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 20m -args -update-counts > /tmp/lane2-b4-global-counts.log 2>&1
go vet ./internal/oracle > /tmp/lane2-b4-vet.log 2>&1
git diff --check > /tmp/lane2-b4-diff-check.log 2>&1
```

Setup is reused from the first delivery: done 45.313s, nproc 5, cpu.max 400000 100000, Go 1.27.1, clang 20.1.8, Node 24.19.0. Full timing lines remain in originalB/REPORT.md and evidence/setup.log.gz. All logs are retained compressed here, together with original declaration hashes and the latest worker report. The branch contains only this worker's commits after requested base 70522aa1d. Only codex/views-arrays-b is pushed.
