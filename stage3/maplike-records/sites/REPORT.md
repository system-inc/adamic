Built: Step 22 (#p9v82wa), source reductions for 56 of 56 MapLike and Record operation sites, in static-use order.
Commits: continued from e379d2321c081e0e3b02a0dad326eb1a7b3d9a8a on codex/maplike-records; delivery tips are reported after each batch push.
Commands and outputs: 24 sites compile and match Node in native and JavaScript backends with ASan, UBSan and Linux leaks; 19 stop in the checker, 10 are Refused, and 3 are NotYet. Counts are regenerated for each added compiling reduction.
Mutants: removing an integer dictionary input is caught only by Node stdout in both backends, with valid builds, exit 0 and clean leaks; the earlier step 22 ordering and entries omission mutants remain available.
Not covered: complete tsc functions, host filesystem/package resolution, runtime frequency, or backends for reductions that stop before emission; the full gate was not run.

Source: microsoft/TypeScript 050880ce59e30b356b686bd3144efe24f875ebc8 (v6.0.3). The scout 6945a8a4 contributes 22 object-operation sites with MapLike/Record receivers, and the pinned typed string-index census contributes 34 read/write sites. Other receiver types, including CompilerOptions-only sites, are outside this list. No previously supplied acceptance fixture is modified. Sites in the same function may share a reduction, with every operation listed separately. The complete 56-site inventory and per-file source hashes are in inventory.json. The first batch did not add its 46 later sites to results.json until their batches were tested.

Order: count direct calls whose callee name matches the containing named function in all compiler sources, highest first, then file, line and column. This is a static name-based ranking, not a runtime profile or resolved call graph; same-name calls may inflate it. Anonymous enclosing callbacks fall back to their containing named function, and top-level sites rank last. rank.cjs reproduces it from the scout output and pinned typed census. The scout checker reports 22 diagnostics on this checkout with the supplied type roots; its operation count is an AST observation, not proof the entire tsc checkout passes Adamic's checker.

Each .a reduction uses the original indexed expression or object operation and describes removed dependencies in results.json. Generic helpers keep their generic signatures where possible; unknown casts and unchecked reads remain visible rather than being silently rewritten to compile. Reduced path/version/trace consumers are observed values, not complete host implementations. For unknown inputs, the source casts remain the first proof boundary. All source programs first run directly on Node with types stripped, including programs the checker rejects. After initial recording, their stdout, stderr and exit observations cannot be changed by the update flag. Only initial compiler outcomes or newly compiling reductions may be recorded.

Commands, with output written directly to evidence logs:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -count=1 -timeout 30m -v -run '^TestTSCRecordSites$/batch_XX' -args -update-tsc-record-sites
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -count=1 -timeout 30m -v -run '^TestTSCRecordSiteKeyMutant$'
go vet ./internal/oracle
go test ./internal/oracle -count=1 -timeout 30m -run '^TestCountsAreRecorded$' -args -update-counts
```

The evidence log for each batch names every outcome. Compiles means its Node stdout, stderr and exit agree with both backends and native sanitizers/leaks pass. Stops retain the compiler's complete returned diagnostic in results.json; the table shows its first line. The counts update is the only full fixture enumeration run. No compiler or runtime implementation file was changed for this extension.

| Rank | Static calls | Source | Result | First stop |
| --- | --- | --- | --- | --- |
| 1 | 36 | core.ts:1267:12 | Compiles |  |
| 2 | 13 | core.ts:1289:5 | Compiles |  |
| 3 | 13 | core.ts:1290:13 | Compiles |  |
| 4 | 7 | moduleNameResolver.ts:2820:43 | Refused | stage3/maplike-records/sites/fixtures/loadModuleFromTargetExportOrImport.a:7:28: Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast) |
| 5 | 7 | utilities.ts:8605:43 | NotYet | stage3/maplike-records/sites/fixtures/getLocaleSpecificMessage.a:6:12: stage 0 can't lower a logical record operand converted to string without a proven result representation yet |
| 6 | 5 | moduleNameResolver.ts:2351:54 | Refused | stage3/maplike-records/sites/fixtures/loadEntrypointsFromTargetExports.a:5:35: Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast) |
| 7 | 5 | moduleSpecifiers.ts:1111:35 | Refused | stage3/maplike-records/sites/fixtures/tryGetModuleNameFromExportsOrImports.a:6:35: Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast) |
| 8 | 4 | core.ts:1314:5 | Checker | stage3/maplike-records/sites/fixtures/getOwnValues.a:7:63: error TS2345: Argument of type 'T &#124; undefined' is not assignable to parameter of type 'T'. |
| 9 | 4 | core.ts:1315:13 | Checker | stage3/maplike-records/sites/fixtures/getOwnValues.a:7:63: error TS2345: Argument of type 'T &#124; undefined' is not assignable to parameter of type 'T'. |
| 10 | 4 | core.ts:1316:25 | Checker | stage3/maplike-records/sites/fixtures/getOwnValues.a:7:63: error TS2345: Argument of type 'T &#124; undefined' is not assignable to parameter of type 'T'. |
| 11 | 4 | moduleNameResolver.ts:2630:26 | NotYet | stage3/maplike-records/sites/fixtures/loadModuleFromExports.a:4:21: stage 0 can't lower a { '.': string; } seen as MapLike<unknown> (fixed objects and records have different storage; copy explicitly) yet |
| 12 | 3 | moduleNameResolver.ts:3177:34 | Compiles |  |
| 13 | 3 | watchUtilities.ts:522:21 | Compiles |  |
| 14 | 2 | commandLineParser.ts:2688:5 | Checker | stage3/maplike-records/sites/fixtures/convertToTSConfig.a:12:21: error TS2532: Object is possibly 'undefined'. |
| 15 | 2 | commandLineParser.ts:2690:29 | Checker | stage3/maplike-records/sites/fixtures/convertToTSConfig.a:12:21: error TS2532: Object is possibly 'undefined'. |
| 16 | 2 | commandLineParser.ts:2691:34 | Checker | stage3/maplike-records/sites/fixtures/convertToTSConfig.a:12:21: error TS2532: Object is possibly 'undefined'. |
| 17 | 2 | commandLineParser.ts:2693:17 | Checker | stage3/maplike-records/sites/fixtures/convertToTSConfig.a:12:21: error TS2532: Object is possibly 'undefined'. |
| 18 | 2 | commandLineParser.ts:2693:50 | Checker | stage3/maplike-records/sites/fixtures/convertToTSConfig.a:12:21: error TS2532: Object is possibly 'undefined'. |
| 19 | 2 | commandLineParser.ts:2707:25 | NotYet | stage3/maplike-records/sites/fixtures/optionDependsOnRecursive.a:13:16: stage 0 can't lower a call through ?. (an optional call) yet |
| 20 | 2 | moduleNameResolver.ts:464:5 | Compiles |  |
| 21 | 2 | moduleNameResolver.ts:474:43 | Compiles |  |
| 22 | 2 | moduleNameResolver.ts:2712:24 | Refused | stage3/maplike-records/sites/fixtures/loadModuleFromExportsOrImports.a:5:17: Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast) |
| 23 | 2 | moduleNameResolver.ts:2718:28 | Refused | stage3/maplike-records/sites/fixtures/loadModuleFromExportsOrImports.a:5:17: Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast) |
| 24 | 2 | moduleNameResolver.ts:2724:28 | Refused | stage3/maplike-records/sites/fixtures/loadModuleFromExportsOrImports.a:5:17: Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast) |
| 25 | 2 | moduleNameResolver.ts:2729:28 | Refused | stage3/maplike-records/sites/fixtures/loadModuleFromExportsOrImports.a:5:17: Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast) |
| 26 | 2 | moduleSpecifiers.ts:927:5 | Checker | stage3/maplike-records/sites/fixtures/tryGetModuleNameFromPaths.a:7:31: error TS2532: Object is possibly 'undefined'. |
| 27 | 2 | moduleSpecifiers.ts:928:35 | Checker | stage3/maplike-records/sites/fixtures/tryGetModuleNameFromPaths.a:7:31: error TS2532: Object is possibly 'undefined'. |
| 28 | 1 | commandLineParser.ts:3322:22 | Compiles |  |
| 29 | 1 | commandLineParser.ts:3323:77 | Compiles |  |
| 30 | 1 | commandLineParser.ts:3325:9 | Compiles |  |
| 31 | 1 | commandLineParser.ts:4142:68 | Compiles |  |
| 32 | 1 | commandLineParser.ts:4144:21 | Compiles |  |
| 33 | 1 | commandLineParser.ts:4154:9 | Compiles |  |
| 34 | 1 | commandLineParser.ts:4159:32 | Compiles |  |
| 35 | 1 | core.ts:1373:5 | Checker | stage3/maplike-records/sites/fixtures/equalOwnProperties.a:10:35: error TS2345: Argument of type 'T &#124; undefined' is not assignable to parameter of type 'T'. |
| 36 | 1 | core.ts:1374:13 | Checker | stage3/maplike-records/sites/fixtures/equalOwnProperties.a:10:35: error TS2345: Argument of type 'T &#124; undefined' is not assignable to parameter of type 'T'. |
| 37 | 1 | core.ts:1375:18 | Checker | stage3/maplike-records/sites/fixtures/equalOwnProperties.a:10:35: error TS2345: Argument of type 'T &#124; undefined' is not assignable to parameter of type 'T'. |
| 38 | 1 | core.ts:1376:35 | Checker | stage3/maplike-records/sites/fixtures/equalOwnProperties.a:10:35: error TS2345: Argument of type 'T &#124; undefined' is not assignable to parameter of type 'T'. |
| 39 | 1 | core.ts:1376:46 | Checker | stage3/maplike-records/sites/fixtures/equalOwnProperties.a:10:35: error TS2345: Argument of type 'T &#124; undefined' is not assignable to parameter of type 'T'. |
| 40 | 1 | core.ts:1380:5 | Checker | stage3/maplike-records/sites/fixtures/equalOwnProperties.a:10:35: error TS2345: Argument of type 'T &#124; undefined' is not assignable to parameter of type 'T'. |
| 41 | 1 | core.ts:1381:13 | Checker | stage3/maplike-records/sites/fixtures/equalOwnProperties.a:10:35: error TS2345: Argument of type 'T &#124; undefined' is not assignable to parameter of type 'T'. |
| 42 | 1 | core.ts:1382:18 | Checker | stage3/maplike-records/sites/fixtures/equalOwnProperties.a:10:35: error TS2345: Argument of type 'T &#124; undefined' is not assignable to parameter of type 'T'. |
| 43 | 1 | core.ts:1470:27 | Checker | stage3/maplike-records/sites/fixtures/groupBy.a:10:24: error TS2345: Argument of type 'T &#124; undefined' is not assignable to parameter of type 'T'. |
| 44 | 1 | moduleNameResolver.ts:432:9 | Compiles |  |
| 45 | 1 | moduleNameResolver.ts:2295:46 | Refused | stage3/maplike-records/sites/fixtures/loadEntrypointsFromExportMap.a:5:21: Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast) |
| 46 | 1 | moduleNameResolver.ts:2428:5 | Compiles |  |
| 47 | 1 | moduleSpecifiers.ts:1134:122 | Refused | stage3/maplike-records/sites/fixtures/tryGetModuleNameFromExports.a:5:33: Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast) |
| 48 | 1 | moduleSpecifiers.ts:1165:121 | Refused | stage3/maplike-records/sites/fixtures/tryGetModuleNameFromPackageJsonImports.a:4:29: Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast) |
| 49 | 1 | program.ts:4141:13 | Compiles |  |
| 50 | 1 | program.ts:4148:29 | Compiles |  |
| 51 | 1 | program.ts:4149:33 | Compiles |  |
| 52 | 1 | program.ts:4154:39 | Compiles |  |
| 53 | 0 | core.ts:1279:12 | Compiles |  |
| 54 | 0 | core.ts:1279:44 | Compiles |  |
| 55 | 0 | parser.ts:2356:38 | Compiles |  |
| 56 | 0 | scanner.ts:222:31 | Compiles |  |

The concrete Node observations and each reduction adaptation are in results.json.

Observation corrections: ranks 1, 20, 21, 28, 29 and 30 in batch 3, and 44, 49 and 50 in batch 5, now compile after prints were combined into single arguments and optional output reads were guarded. The source operations were retained and all recorded Node bytes stayed unchanged. Their prior checker stops were artifacts of the reductions. See the observer-corrections evidence logs.

Delivery was five batches of ten sites, then the remaining six sites. The earlier pushed tips were:

| Sites | Tip |
| --- | --- |
| 1–10 | c08788be6dee9ed5c7668449ee2ba3f1447138f7 |
| 11–20 | 71fee6ffb3c2bd38483fc17a200363a64916b2ba |
| 21–30 | 9caf2e93ef1119ce972deda31f6f79495ea64ff2 |
| 31–40 | d4733ffdbc1b1c3c4a4fbc2d468c3952d3b71fb2 |
| 41–50 | b0716ede79023958c5dc2d9a0b2bd0881544b881 |

The final tip for sites 51–56 is supplied with the delivery message. All 56 coordinates and source hashes reproduce with rank.cjs. Every Node observation from each previous pushed manifest remains byte for byte; evidence/inventory-check.log records the check.

Compiling reductions cover own membership and keys, guarded property reads, path-array lookup, copied substitutions, watch-directory entries and deletion, nested types-version records, version-key traversal, peer-dependency keys, path validation, and keyword keys/entries. The remaining first stops are generic reads passed to non-optional parameters, unchecked unknown-to-record casts, optional invocation, a logical record-to-string result, and a fixed-object-to-record view.
