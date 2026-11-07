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
| src/compiler/sys.ts | createDirectoryWatcherSupportingRecursive / createDirectoryWatcher / &lt;anonymous&gt; | 0 | 1 | 0 |
| src/compiler/sys.ts | createUseFsEventsOnParentDirectoryWatchFile | 0 | 0 | 1 |
| src/compiler/sys.ts | createUseFsEventsOnParentDirectoryWatchFile / nonPollingWatchFile | 1 | 0 | 0 |
| src/compiler/sys.ts | createUseFsEventsOnParentDirectoryWatchFile / nonPollingWatchFile / &lt;anonymous&gt; | 0 | 1 | 0 |
| src/compiler/utilities.ts | createSymlinkCache / &lt;anonymous&gt; | 1 | 0 | 1 |
| src/harness/fourslashImpl.ts | TestState / rangesByText | 1 | 0 | 1 |
| src/harness/watchUtils.ts | createWatchUtils / initializeWatches | 0 | 0 | 1 |
| src/harness/watchUtils.ts | createWatchUtils / initializeWatches / add | 2 | 0 | 1 |
| src/harness/watchUtils.ts | createWatchUtils / initializeWatches / remove | 0 | 2 | 0 |
| src/server/editorServices.ts | ProjectService / &lt;anonymous&gt; | 0 | 0 | 1 |
| src/server/editorServices.ts | ProjectService / deleteScriptInfo | 0 | 1 | 0 |
| src/server/editorServices.ts | ProjectService / getSymlinkedProjects / combineProjects | 2 | 0 | 1 |
| src/server/scriptInfo.ts | ScriptInfo / ensureRealPath | 1 | 0 | 0 |
| src/services/codeFixProvider.ts |  | 0 | 0 | 1 |
| src/services/codeFixProvider.ts | registerCodeFix | 1 | 0 | 0 |
| src/services/codefixes/convertToAsyncFunction.ts | renameCollidingVarNames | 0 | 0 | 1 |
| src/services/codefixes/convertToAsyncFunction.ts | renameCollidingVarNames / visit | 3 | 0 | 0 |
| src/services/codefixes/convertToEsModule.ts | collectFreeIdentifiers | 0 | 0 | 1 |
| src/services/codefixes/convertToEsModule.ts | collectFreeIdentifiers / &lt;anonymous&gt; | 1 | 0 | 0 |
| src/services/codefixes/importFixes.ts | createExistingImportMap | 2 | 0 | 2 |
| src/services/codefixes/importFixes.ts | getExportInfos | 0 | 0 | 1 |
| src/services/codefixes/importFixes.ts | getExportInfos / addSymbol | 1 | 0 | 0 |
| src/services/codefixes/inferFromUsage.ts | inferTypeFromReferences / combineAnonymousTypes | 1 | 0 | 1 |
| src/services/exportInfoMap.ts | createCacheableExportInfoMap | 0 | 0 | 1 |
| src/services/exportInfoMap.ts | createCacheableExportInfoMap / &lt;anonymous&gt; | 1 | 0 | 0 |
| src/services/services.ts | SourceFileObject / computeNamedDeclarations | 0 | 0 | 1 |
| src/services/services.ts | SourceFileObject / computeNamedDeclarations / addDeclaration | 1 | 0 | 0 |
| src/services/textChanges.ts | ChangeTracker / insertStatementsInNewFile | 1 | 0 | 1 |
| src/testRunner/unittests/tsserver/cachingFileSystemInformation.ts | &lt;anonymous&gt; / createLoggerTrackingHostCalls / setCallsTrackingWithFiveArgFn | 0 | 0 | 1 |
| src/testRunner/unittests/tsserver/cachingFileSystemInformation.ts | &lt;anonymous&gt; / createLoggerTrackingHostCalls / setCallsTrackingWithFiveArgFn / &lt;anonymous&gt; | 1 | 0 | 0 |
| src/testRunner/unittests/tsserver/cachingFileSystemInformation.ts | &lt;anonymous&gt; / createLoggerTrackingHostCalls / setCallsTrackingWithSingleArgFn | 0 | 0 | 1 |
| src/testRunner/unittests/tsserver/cachingFileSystemInformation.ts | &lt;anonymous&gt; / createLoggerTrackingHostCalls / setCallsTrackingWithSingleArgFn / &lt;anonymous&gt; | 1 | 0 | 0 |

## Flow review

Map and ReadonlyMap destinations include RedirectTargetsMap, FreeIdentifiers, named declarations, export info, inferred anonymous properties, project symlinks, rangesByText and harness diffMap. The two unknown-context observations are the descriptor value in Object.defineProperty(cache, "__cache", { value: exportInfo }) under Debug.isDebugging. This stores the Map for debugging and performs no own-key enumeration. Harness serializeMultiMap uses Map.keys() and Map.forEach(), which enumerate Map entries. Original member assignments are confined to createMultiMap.

The static census examines for-in, object spread, object rest and own-key APIs on MultiMap-typed operands. The lane measurement additionally registers actual factory Maps and watches emitted for-in/spread operations and Object/Reflect enumeration calls, including paths through the wider Map/ReadonlyMap types. A zero dynamic enumeration result is asserted only when the completed count has passed.

## Owner overlap audit

The only whole-owner hash conflict is 95h createProgram, amended as authorized. Adapter 30 records arrayToMultiMap as an indexed-read site; adapter 32 records createProgram indexed-read sites. These are not whole-owner hash gates. 95b records createTypeChecker as a caller; 99x delegates createProgram to 95h. 99j pins the sibling invokeCallbacks in the recursive watcher, not createDirectoryWatcher. No other edited owner is pinned by an earlier adapter. owner-audit.json preserves the broad name search hits; the earlier complete ancestor audit is documented in the stop history.

## Completed type and flow references

The language service also reports 36 references to the MultiMap type declaration. The checker follows base classes and interfaces when identifying MultiMap expressions. All references and their locations are retained in census.json. The final-tree language-service census confirms 27 multiMapAddValue calls, 5 multiMapRemoveValue calls and 24 createMultiMap calls, with zero legacy helper identifiers.

Flow observations include overlapping contextual and argument checks; these are 70 observations rather than 70 unique transfers. Counts by file and lexical owner follow.

| File | Owner | Flow observations |
| --- | --- | ---: |
| src/compiler/checker.ts | createBasicNodeBuilderModuleSpecifierResolutionHost/&lt;anonymous&gt; | 1 |
| src/compiler/core.ts | arrayToMultiMap | 1 |
| src/compiler/core.ts | createMultiMap | 1 |
| src/compiler/program.ts | createProgram | 1 |
| src/compiler/program.ts | createProgram/&lt;anonymous&gt; | 1 |
| src/compiler/program.ts | createProgram/getEmitHost | 1 |
| src/compiler/program.ts | createProgram/tryReuseStructureFromOldProgram | 2 |
| src/compiler/programDiagnostics.ts | createProgramDiagnostics/getFileReasons | 1 |
| src/compiler/programDiagnostics.ts | createProgramDiagnostics/reuseStateFromOldProgram | 2 |
| src/compiler/utilities.ts | createSymlinkCache/&lt;anonymous&gt; | 2 |
| src/harness/fourslashImpl.ts | TestState/rangesByText | 4 |
| src/harness/watchUtils.ts | createWatchUtils/initializeWatches/add | 1 |
| src/harness/watchUtils.ts | createWatchUtils/initializeWatches/serialize | 2 |
| src/harness/watchUtils.ts | serializeMultiMap | 4 |
| src/server/editorServices.ts | ProjectService/&lt;anonymous&gt; | 2 |
| src/server/editorServices.ts | ProjectService/getSymlinkedProjects | 1 |
| src/server/editorServices.ts | ProjectService/getSymlinkedProjects/combineProjects | 4 |
| src/server/session.ts | Session/getProjects | 3 |
| src/services/codefixes/convertToAsyncFunction.ts | renameCollidingVarNames/visit | 4 |
| src/services/codefixes/convertToEsModule.ts | collectFreeIdentifiers | 1 |
| src/services/codefixes/importFixes.ts | createExistingImportMap | 2 |
| src/services/codefixes/importFixes.ts | getExportInfos | 1 |
| src/services/codefixes/inferFromUsage.ts | inferTypeFromReferences/combineAnonymousTypes | 2 |
| src/services/exportInfoMap.ts | createCacheableExportInfoMap | 2 |
| src/services/exportInfoMap.ts | createCacheableExportInfoMap/&lt;anonymous&gt; | 2 |
| src/services/findAllReferences.ts | getReferencedSymbolsForNode | 4 |
| src/services/findAllReferences.ts | getReferencesForFileName | 2 |
| src/services/services.ts | SourceFileObject/computeNamedDeclarations | 1 |
| src/services/textChanges.ts | ChangeTracker/insertStatementsInNewFile | 2 |
| src/services/utilities.ts | createModuleSpecifierResolutionHost | 2 |
| src/services/utilities.ts | createModuleSpecifierResolutionHost/&lt;anonymous&gt; | 1 |
| src/testRunner/unittests/tsserver/cachingFileSystemInformation.ts | &lt;anonymous&gt;/createLoggerTrackingHostCalls | 10 |

## Object-rest coverage

The seven source object-rest patterns (13 emitted copies) were checked separately against the same stock checker and all 70 MultiMap flow observations. Their operand types are CodeFixAction, NameAndKind, UserPreferences, AssignProjectResult, { metadata?: unknown }, RenameLocation and RenameTextSpan. None is Map-shaped or any/unknown, and no MultiMap flows to these operands. They introduce zero MultiMap enumerations. rest-types.json and the emitted-pattern audit are retained in proof-evidence.tar.gz.
