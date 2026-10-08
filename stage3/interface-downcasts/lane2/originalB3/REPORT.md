Certified 20 original array-field pairs / 38 static reads toward roadmap step 09.
Commits: previous delivery 4be16240660318473e294465cbd170fefd58aa92; this delivery SHA accompanies its push.
Checks: 126 controls and 22 mutants pass in 82.036s; 126 scoped count rows pass in 95.458s; global counts retain existing failures; vet passes.
Mutants: each reached bad element or member and each missing required nullable field is caught by its pinned stopping oracle in all three modes.
Not covered: callable element replacement, delegated union and mixed-member work, the remaining two-read tail and production runtime reachability.

This is the third 20-pair delivery. It remains on the requested lane 2 base with only this worker's commits. No production code changed. All selected original pairs were checked against the worker's published original config files through 7cd0281388208bd439b5a20d4f4294a55a4ac91e: no overlaps. The newest RANKED29 report was read; the worker remains in the three-read rank. Its private files-array certification is distinct from this worker's directories and branded sorted-array fields. No worker branch was merged.

| Original pair | Type ID | Static reads |
| --- | ---: | ---: |
| `{ kind: TypeMapKind.Deferred; sources: readonly Type[]; targets: (() => Type)[]; }.targets` | 6066 | 1 |
| `{ kind: TypeMapKind.Deferred; sources: readonly Type[]; targets: (() => Type)[]; }.sources` | 6066 | 1 |
| `NonIncrementalBuildInfo.root` | 97938 | 2 |
| `BuilderProgramState.affectedFiles` | 97245 | 2 |
| `SortedAndCanonicalizedMutableFileSystemEntries.sortedAndCanonicalizedFiles` | 93573 | 2 |
| `SortedAndCanonicalizedMutableFileSystemEntries.sortedAndCanonicalizedDirectories` | 93573 | 2 |
| `UnscopedEmitHelper.dependencies` | 11590 | 2 |
| `WatchOptions | undefined.excludeFiles` | 11219 | 2 |
| `WatchOptions | undefined.excludeDirectories` | 11219 | 2 |
| `DiagnosticWithDetachedLocation.relatedInformation` | 11106 | 2 |
| `InterfaceTypeWithDeclaredMembers.declaredIndexInfos` | 10788 | 2 |
| `InterfaceTypeWithDeclaredMembers.declaredConstructSignatures` | 10788 | 2 |
| `InterfaceTypeWithDeclaredMembers.declaredCallSignatures` | 10788 | 2 |
| `CompilerOptions.types` | 9858 | 2 |
| `CompilerOptions.moduleSuffixes` | 9858 | 2 |
| `CompilerOptions.customConditions` | 9858 | 2 |
| `ConfigFileSpecs | undefined.validatedIncludeSpecs` | 9854 | 2 |
| `ConfigFileSpecs.validatedExcludeSpecs` | 9853 | 2 |
| `TsConfigSourceFile.statements` | 9852 | 2 |
| `DiagnosticWithLocation.relatedInformation` | 9760 | 2 |

Preparation keeps all 79 complete declaration files at original TypeScript pin 050880ce59e30b356b686bd3144efe24f875ebc8, checks every original span and full declared member type, and verifies full receiver and selected element field sets against the lowered descriptors. DeferredMapper uses Extract<TypeMapper, {kind: TypeMapKind.Deferred}> over the original union, preserving all three fields and complete Type return descriptors. Both Canonicalized branded string arrays retain their original declarations; reads and aliases work on this base and the numeric negative pins Canonicalized by name. No proof is erased or declaration reduced.

Every pair retains valid, unread malformed array, later unread malformed element, stopping and source-alias controls. Optional arrays and receivers retain their absent and undefined controls. Required nullable config fields distinguish a missing field from initialized undefined. IndexInfo reads its original isReadonly boolean. EmitHelper dependencies keep both complete original helper arms; controls provide the original required scoped/text/name members and read the original name. Literal scoped:false is preserved with as const so the base proof is true. No checker option changed.

All 126 sources run on Node. All 126 compiled controls match their normal or exact stopping result in ASan/UBSan native, release native and JavaScript. All 103 finishing controls and all 22 finishing mutants pass native leak checks. The initial generator used the wrong shared field for optional WatchOptions and did not preserve the helper's false literal; those fixture errors were corrected. Both initial logs remain recorded. Final controls and mutants pass together.

The deferred target alias mutates the returned source object's flags; the source callable itself stays in the slot. A separate complete-original witness replaces that callable slot. Node prints 9, while native modes stop with `element read failed: <array write> expected function, found uncertified source element contract`. JavaScript stops with `element read failed: <array write> expected uncertified storage, found uncertified source element contract`. Both exact messages are pinned independently. This write needs producer certification for callable elements; it receives no write or mutant credit. Read certification does not claim this replacement is implemented.

mutants.json records all independent executions. Ten string-consumer mutations admit numeric elements through string | number. Ten reached member mutations replace a numeric, boolean or string read and its conversion with the observed alternative scalar representation. Two required-field mutations admit absent arrays. There are 22 executions total: ten consumer, ten member and two presence mutations. Each changes exactly one reached operation, finishes with Node output and empty stderr, passes sanitizers and leaks, and disagrees with the unchanged exit-70 oracle. Invalid C and sanitizer failures are not credited.

Exactly 126 count rows are added, explained in count-rows.json. Removing the own prefix reproduces the previous delivery's table byte for byte. The required global refresh failed in 49.306s in existing graph-region invalid frees, process.exit, fs and other baseline fixtures. The scoped refresh succeeded; no pre-existing row changed. No whole package or full gate ran.

The worker now reports 202 pairs / 2957 static reads. This worker has 59 / 77 across three disjoint deliveries, making the combined scheduling ledger 261 / 3034, with 73 / 155 remaining of 334 / 3189. Runtime reachability remains unmeasured; consumer, intrinsic and own-array-field totals receive no credit.

```sh
source /workspace/adamic-tools/env.sh
node stage3/interface-downcasts/lane2/originalB3/prepare.cjs /workspace/lane2-b-original-pin /workspace/lane2-b-declarations3 > /tmp/lane2-b3-prepare.log 2>&1
ADAMIC_ARRAYB3_ORIGINAL_DECLS=/workspace/lane2-b-declarations3 ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewArraysB3Original$|^TestCheckedViewArraysB3Mutants$' -count=1 -parallel 4 -v -timeout 20m > /tmp/lane2-b3-complete.log 2>&1
ADAMIC_ARRAYB3_ORIGINAL_DECLS=/workspace/lane2-b-declarations3 ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewArraysB3Counts$' -count=1 -timeout 15m -args -update-counts > /tmp/lane2-b3-counts.log 2>&1
ADAMIC_ARRAYB3_ORIGINAL_DECLS=/workspace/lane2-b-declarations3 ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 20m -args -update-counts > /tmp/lane2-b3-global-counts.log 2>&1
go vet ./internal/oracle > /tmp/lane2-b3-vet.log 2>&1
git diff --check > /tmp/lane2-b3-diff-check.log 2>&1
```

Setup is reused from the first delivery: done 45.313s, nproc 5, cpu.max 400000 100000, Go 1.27.1, clang 20.1.8 and Node 24.19.0. Full timing lines remain in originalB/REPORT.md and evidence/setup.log.gz. Evidence logs and declaration manifest hashes are retained here. The final tail contains eight eligible pairs before CircularBuildOrder.buildOrder enters the worker's three-read rank.
