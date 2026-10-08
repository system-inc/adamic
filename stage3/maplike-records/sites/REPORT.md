Built: Step 22 (#p9v82wa), source reductions for 10 of 56 MapLike and Record operation sites, in static-use order.
Commits: continued from e379d2321c081e0e3b02a0dad326eb1a7b3d9a8a on codex/maplike-records; delivery tips are reported after each batch push.
Commands and outputs: 2 sites compile and match Node in native and JavaScript backends with ASan, UBSan and Linux leaks; 4 stop in the checker, 3 are Refused, and 1 are NotYet. Counts are regenerated for each added compiling reduction.
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
| 1 | 36 | core.ts:1267:12 | Checker | stage3/maplike-records/sites/fixtures/hasProperty.a:10:44: error TS2554: Expected 1 arguments, but got 2. |
| 2 | 13 | core.ts:1289:5 | Compiles |  |
| 3 | 13 | core.ts:1290:13 | Compiles |  |
| 4 | 7 | moduleNameResolver.ts:2820:43 | Refused | stage3/maplike-records/sites/fixtures/loadModuleFromTargetExportOrImport.a:7:28: Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast) |
| 5 | 7 | utilities.ts:8605:43 | NotYet | stage3/maplike-records/sites/fixtures/getLocaleSpecificMessage.a:6:12: stage 0 can't lower a logical record operand converted to string without a proven result representation yet |
| 6 | 5 | moduleNameResolver.ts:2351:54 | Refused | stage3/maplike-records/sites/fixtures/loadEntrypointsFromTargetExports.a:5:35: Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast) |
| 7 | 5 | moduleSpecifiers.ts:1111:35 | Refused | stage3/maplike-records/sites/fixtures/tryGetModuleNameFromExportsOrImports.a:6:35: Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast) |
| 8 | 4 | core.ts:1314:5 | Checker | stage3/maplike-records/sites/fixtures/getOwnValues.a:7:63: error TS2345: Argument of type 'T &#124; undefined' is not assignable to parameter of type 'T'. |
| 9 | 4 | core.ts:1315:13 | Checker | stage3/maplike-records/sites/fixtures/getOwnValues.a:7:63: error TS2345: Argument of type 'T &#124; undefined' is not assignable to parameter of type 'T'. |
| 10 | 4 | core.ts:1316:25 | Checker | stage3/maplike-records/sites/fixtures/getOwnValues.a:7:63: error TS2345: Argument of type 'T &#124; undefined' is not assignable to parameter of type 'T'. |

The concrete Node observations and each reduction adaptation are in results.json.
