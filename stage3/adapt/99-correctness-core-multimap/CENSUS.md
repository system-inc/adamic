# Checker reference and flow census

Stock TypeScript 6.0.3, over the completed main 698965fe6 adaptation stack, upstream 050880ce59e30b356b686bd3144efe24f875ebc8. The language service has 711 src roots, including compiler, services, server, harness and test runner; imported compiler/services/server source coverage is 262 files. References come from findReferences on the two interface members and the factory definition, not text matches.

27 add calls, 5 remove calls, 24 factory calls. The add reference set also includes one declaration and one factory write; remove likewise. Factory references include 16 import/definition bindings. There are zero member reads as values and zero checker-typed own-key enumerations. There are 70 contextual, argument, assignment and assertion flow observations, including duplicate observations of one flow through different checker interfaces. All raw locations, source and target types, owner chains and summary counts are in census.json.

| File | Owner chain | add | remove | createMultiMap |
| --- | --- | ---: | ---: | ---: |
| src/compiler/checker.ts | createTypeChecker / createNodeBuilder / mapToTypeNodes | 1 | 0 | 1 |
| src/compiler/core.ts | arrayToMultiMap | 1 | 0 | 1 |
| src/compiler/program.ts | createProgram | 0 | 0 | 1 |
| src/compiler/program.ts | createProgram / addFileIncludeReason | 1 | 0 | 0 |
| src/compiler/program.ts | createProgram / findSourceFileWorker | 1 | 0 | 0 |
| src/compiler/programDiagnostics.ts | createProgramDiagnostics | 0 | 0 | 1 |
| src/compiler/sys.ts | createDirectoryWatcherSupportingRecursive | 0 | 0 | 1 |
| src/compiler/sys.ts | createDirectoryWatcherSupportingRecursive / createDirectoryWatcher | 1 | 0 | 0 |
| src/compiler/sys.ts | createDirectoryWatcherSupportingRecursive / createDirectoryWatcher / <anonymous> | 0 | 1 | 0 |
| src/compiler/sys.ts | createUseFsEventsOnParentDirectoryWatchFile | 0 | 0 | 1 |
| src/compiler/sys.ts | createUseFsEventsOnParentDirectoryWatchFile / nonPollingWatchFile | 1 | 0 | 0 |
| src/compiler/sys.ts | createUseFsEventsOnParentDirectoryWatchFile / nonPollingWatchFile / <anonymous> | 0 | 1 | 0 |
| src/compiler/utilities.ts | createSymlinkCache / <anonymous> | 1 | 0 | 1 |
| src/harness/fourslashImpl.ts | TestState / rangesByText | 1 | 0 | 1 |
| src/harness/watchUtils.ts | createWatchUtils / initializeWatches | 0 | 0 | 1 |
| src/harness/watchUtils.ts | createWatchUtils / initializeWatches / add | 2 | 0 | 1 |
| src/harness/watchUtils.ts | createWatchUtils / initializeWatches / remove | 0 | 2 | 0 |
| src/server/editorServices.ts | ProjectService / <anonymous> | 0 | 0 | 1 |
| src/server/editorServices.ts | ProjectService / deleteScriptInfo | 0 | 1 | 0 |
| src/server/editorServices.ts | ProjectService / getSymlinkedProjects / combineProjects | 2 | 0 | 1 |
| src/server/scriptInfo.ts | ScriptInfo / ensureRealPath | 1 | 0 | 0 |
| src/services/codeFixProvider.ts |  | 0 | 0 | 1 |
| src/services/codeFixProvider.ts | registerCodeFix | 1 | 0 | 0 |
| src/services/codefixes/convertToAsyncFunction.ts | renameCollidingVarNames | 0 | 0 | 1 |
| src/services/codefixes/convertToAsyncFunction.ts | renameCollidingVarNames / visit | 3 | 0 | 0 |
| src/services/codefixes/convertToEsModule.ts | collectFreeIdentifiers | 0 | 0 | 1 |
| src/services/codefixes/convertToEsModule.ts | collectFreeIdentifiers / <anonymous> | 1 | 0 | 0 |
| src/services/codefixes/importFixes.ts | createExistingImportMap | 2 | 0 | 2 |
| src/services/codefixes/importFixes.ts | getExportInfos | 0 | 0 | 1 |
| src/services/codefixes/importFixes.ts | getExportInfos / addSymbol | 1 | 0 | 0 |
| src/services/codefixes/inferFromUsage.ts | inferTypeFromReferences / combineAnonymousTypes | 1 | 0 | 1 |
| src/services/exportInfoMap.ts | createCacheableExportInfoMap | 0 | 0 | 1 |
| src/services/exportInfoMap.ts | createCacheableExportInfoMap / <anonymous> | 1 | 0 | 0 |
| src/services/services.ts | SourceFileObject / computeNamedDeclarations | 0 | 0 | 1 |
| src/services/services.ts | SourceFileObject / computeNamedDeclarations / addDeclaration | 1 | 0 | 0 |
| src/services/textChanges.ts | ChangeTracker / insertStatementsInNewFile | 1 | 0 | 1 |
| src/testRunner/unittests/tsserver/cachingFileSystemInformation.ts | <anonymous> / createLoggerTrackingHostCalls / setCallsTrackingWithFiveArgFn | 0 | 0 | 1 |
| src/testRunner/unittests/tsserver/cachingFileSystemInformation.ts | <anonymous> / createLoggerTrackingHostCalls / setCallsTrackingWithFiveArgFn / <anonymous> | 1 | 0 | 0 |
| src/testRunner/unittests/tsserver/cachingFileSystemInformation.ts | <anonymous> / createLoggerTrackingHostCalls / setCallsTrackingWithSingleArgFn | 0 | 0 | 1 |
| src/testRunner/unittests/tsserver/cachingFileSystemInformation.ts | <anonymous> / createLoggerTrackingHostCalls / setCallsTrackingWithSingleArgFn / <anonymous> | 1 | 0 | 0 |

## Flow review

Map and ReadonlyMap destinations include RedirectTargetsMap, FreeIdentifiers, named declarations, export info, inferred anonymous properties, project symlinks, rangesByText and harness diffMap. The two unknown-context observations are the descriptor value in Object.defineProperty(cache, "__cache", { value: exportInfo }) under Debug.isDebugging. This stores the Map for debugging and performs no own-key enumeration. Harness serializeMultiMap uses Map.keys() and Map.forEach(), which enumerate Map entries. Original member assignments are confined to createMultiMap.

The static census examines for-in, object spread, object rest and own-key APIs on MultiMap-typed operands. The lane measurement additionally registers actual factory Maps and watches emitted for-in/spread operations and Object/Reflect enumeration calls, including paths through the wider Map/ReadonlyMap types. A zero dynamic enumeration result is asserted only when the completed count has passed.

## Owner overlap audit

The only whole-owner hash conflict is 95h createProgram, amended as authorized. Adapter 30 records arrayToMultiMap as an indexed-read site; adapter 32 records createProgram indexed-read sites. These are not whole-owner hash gates. 95b records createTypeChecker as a caller; 99x delegates createProgram to 95h. 99j pins the sibling invokeCallbacks in the recursive watcher, not createDirectoryWatcher. No other edited owner is pinned by an earlier adapter. owner-audit.json preserves the broad name search hits; the earlier complete ancestor audit is documented in the stop history.
