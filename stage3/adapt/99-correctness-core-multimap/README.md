# MultiMap stop report

Base: origin/main f0271c6a7. TypeScript source: 050880ce59e30b356b686bd3144efe24f875ebc8.

Stopped before implementation under the requested prior-owner overlap rule.

## Blocking pin

`95-checker-axolotl-h/sites.json`, `literals[1]`, pins the whole `src/compiler/program.ts` owner `createProgram` with `ownerBeforeSha256` 505bb77211eb4818ad15ed0701812c6c51eb06059becdd73923a5f3e73b69152. Its adapt.cjs checks the whole owner hash at lines 78-79 and 92. The required rewrites in nested `findSourceFileWorker` (`redirectTargetsMap.add`) and `addFileIncludeReason` (`programDiagnostics.getFileReasons().add`) change that pinned owner. This needs an amendment decision.

Additional census hits: `30-indexed-reads/sites.json` records `arrayToMultiMap` at line 882; `32-indexed-reads-program` site manifests record `createProgram`. These are site inventories, not asserted here to be whole-owner hash pins. `95-checker-axolotl-b` mentions `createTypeChecker` as a caller inventory, not a whole-owner pin. `99-correctness-axolotl-j` pins the sibling `invokeCallbacks`, not the MultiMap call owner `createDirectoryWatcher`.

## Preliminary semantic census

Stock TypeScript 6.0.3 language service findReferences over 261 compiler, services, server and namespace files found 24 add references, 5 remove references and 34 createMultiMap references. Member totals include declarations and createMultiMap property writes. This is preliminary only: the full baseline apply was still running at adapter 20 when the blocking overlap was established and the process was terminated. It is not the required completed adapted-tree census. Flow and enumeration analysis was not completed; zero enumerations or value reads is not claimed. The JSON retains individual references and enclosing owner chains.

| File | Owner chain | add refs | remove refs |
| --- | --- | ---: | ---: |
| src/compiler/checker.ts | createTypeChecker / createNodeBuilder / mapToTypeNodes | 1 | 0 |
| src/compiler/core.ts | add | 1 | 0 |
| src/compiler/core.ts | arrayToMultiMap | 1 | 0 |
| src/compiler/core.ts | createMultiMap | 1 | 1 |
| src/compiler/core.ts | remove | 0 | 1 |
| src/compiler/program.ts | createProgram / addFileIncludeReason | 1 | 0 |
| src/compiler/program.ts | createProgram / findSourceFileWorker | 1 | 0 |
| src/compiler/sys.ts | createDirectoryWatcherSupportingRecursive / createDirectoryWatcher | 1 | 0 |
| src/compiler/sys.ts | createDirectoryWatcherSupportingRecursive / createDirectoryWatcher / <anonymous> | 0 | 1 |
| src/compiler/sys.ts | createUseFsEventsOnParentDirectoryWatchFile / nonPollingWatchFile | 1 | 0 |
| src/compiler/sys.ts | createUseFsEventsOnParentDirectoryWatchFile / nonPollingWatchFile / <anonymous> | 0 | 1 |
| src/compiler/utilities.ts | createSymlinkCache / <anonymous> | 1 | 0 |
| src/server/editorServices.ts | ProjectService / deleteScriptInfo | 0 | 1 |
| src/server/editorServices.ts | ProjectService / getSymlinkedProjects / combineProjects | 2 | 0 |
| src/server/scriptInfo.ts | ScriptInfo / ensureRealPath | 1 | 0 |
| src/services/codeFixProvider.ts | registerCodeFix | 1 | 0 |
| src/services/codefixes/convertToAsyncFunction.ts | renameCollidingVarNames / visit | 3 | 0 |
| src/services/codefixes/convertToEsModule.ts | collectFreeIdentifiers / <anonymous> | 1 | 0 |
| src/services/codefixes/importFixes.ts | createExistingImportMap | 2 | 0 |
| src/services/codefixes/importFixes.ts | getExportInfos / addSymbol | 1 | 0 |
| src/services/codefixes/inferFromUsage.ts | inferTypeFromReferences / combineAnonymousTypes | 1 | 0 |
| src/services/exportInfoMap.ts | createCacheableExportInfoMap / <anonymous> | 1 | 0 |
| src/services/services.ts | SourceFileObject / computeNamedDeclarations / addDeclaration | 1 | 0 |
| src/services/textChanges.ts | ChangeTracker / insertStatementsInNewFile | 1 | 0 |

## Proof status

Full applies: baseline interrupted at adapter 20; adapted not run. Paired replay: not run. Adamic types: not run. JavaScript/declarations comparison: not run. Core site census and mutant: not run. Lane runtime count and witnesses: not run. No adaptation or ledger claim was written. Setup and apply process groups were terminated on the stop condition.
