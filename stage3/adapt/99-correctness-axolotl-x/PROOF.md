# 99x proof: 28 rewrites, 18 earlier-owner handoffs

The current ownership revision passes checks 1 through 6. All current receipts
are under owned/. Receipts outside owned/ describe earlier 46-site experiments
and superseded stopping rules; they do not certify this scope.

javascriptChanging true. No earlier adaptation directory is edited. The exact
46-site inventory and every handoff reason are in README.md and inventory.json.
Six sites belong to 95-checker-axolotl-h: two checker-host calls and four calls
under program.ts createProgram/getEmitHost. Twelve belong to the complete
watch-host runtime pins in 32-indexed-reads-program. Their adapter directories
remain untouched. The runtime-input audit and ownership.json record these pins.
The current adapter edits ten owners in five files and removes three imports.
core.ts, binder.ts and checker.ts source are byte-identical to the baseline.

## Results

| Check | Current result |
| --- | --- |
| 1 full applies with --build-product | Before and after exit 0 |
| 1 paired replay | All 80 baseline outcomes identical; no differences |
| Direct replay | Zero writes and edits; bytes and nanosecond mtimes preserved |
| Owner contracts | Six malformed states refuse before writes; CRLF and unrelated edits pass |
| 2 types | All 79 roots together, exit 0 and zero diagnostics |
| 3 emitted JavaScript | Nine products pass; seven byte-identical, two with exactly 28 rewritten sites each |
| 4 upstream oracle | Whole lane and all 398 standard shards pass; their exact observation union and API bytes agree |
| 5 negative control | Exactly one restored source site detected, in both compiler bundles |
| 6 ordinary census | All 79 first stops unchanged; 73 stop at core.ts:810:89 |
| 6 CW051 | All six units measured on both trees; table below |

## Provenance and full applies

Base main: 3e395f9e5a7038d859ec8930be72cde68d8be3ff.
Upstream TypeScript 6.0.3: 050880ce59e30b356b686bd3144efe24f875ebc8.
99w's complete adapt.cjs, README.md and PROOF.md were read at
06dad75d0bb482b47046370a60341c054899afed. 99k was read on main. The branch
remains based on main, never 99w. No Apple compiler overlay was used.

cloud/setup.sh and npm ci --prefix stage3/api completed after sourcing
/workspace/adamic-tools/env.sh. Setup did not supply the upstream source,
so the authorized pinned checkout was cloned to /workspace/typescript-source.
Setup and package logs are setup.log.gz and npm-api.log.gz. Both complete
applications used --build-product:

```sh
# Isolated unmodified main worktree /workspace/bind-main:
ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source stage3/apply.sh --build-product /workspace/bind-before
# This branch with the current adapter:
ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source stage3/apply.sh --build-product /workspace/bind-owned-after
NODE_PATH=stage3/api/node_modules node stage3/adapt/99-correctness-axolotl-x/contract.cjs /workspace/bind-before
```

baseline-apply.log.gz records the unchanged before build. owned/after-apply.log.gz
records the current after build. owned/source-delta.json confirms the five
changed files. owned/contract.log.gz records direct replay, byte/mtime preservation,
atomic refusals and CRLF checks. owners.json contains exact before/after texts
and SHA-256 values for each edited owner, never whole-file hash gates.

## Check 1: all adapter outcomes

paired_replay.py runs the real Python order over independent copies of both
complete adapted trees. owned/paired-replay/before.json and after.json each
contain 81 records: 80 baseline adapters and the separately handled 99x entry.
Every baseline outcome matches: exit status, files written, exact edits and
exact refusal text. Only tree root paths are normalized. records.diff is empty.
No early stopping is used. 99x's after replay exits 0, writes zero files and
makes zero edits; its 18 handed-off calls remain inventoried.

Shared refusals are baseline outcomes. In particular 40-explicit-any reports
`Error: missing reviewed brands: src/compiler/builder.ts` on both trees.
30-indexed-reads reports `declined site changed: core.ts:992:21` on both.
The complete exact text and all other outcomes are retained in the records
and separate compressed adapter logs, including 95h's shared program owner
refusal. None is treated as a stop when identical on both trees.

A passive filesystem journal records sync and async writes, including unchanged
byte writes and inherited child-process writes. Whole-tree byte/mtime manifests
before and after verify zero unjournaled changes. owned/journal/ retains the
control and runtime write audit; owned/paired-replay/ retains every journal,
manifest and coverage.json. This changes observation cost, not adapter behavior.

## Check 2: 79-root types

The unchanged main Adamic binary receives all 79 sorted compiler TypeScript
roots in one invocation. It exits 0 with zero diagnostics. owned/types.json
records every argument and timing; owned/types.log.gz is the complete output.
No type-driven fallback or cast is needed for the 28 remaining rewrites.

## Check 3: emitted products

Both ordinary npm run build:compiler builds exit 0 with the same locked
upstream dependencies. owned/javascript/ contains every raw diff, build log,
comparison report and declaration hash. Both public and internal declaration
files are byte-identical. Seven normal JavaScript products have no changes.
_tsc.js and typescript.js each contain the same 28 expected source-site changes.

The normal emitter lowers each optional binding into a nullish conditional
and a generated temporary. compare.cjs uses stock TypeScript 6.0.3 to match
only inventoried owner/member pairs with receiver/temporary symbol identity.
It reconstructs those expressions and normalizes generated temporary names,
their declarations and resulting parentheses. The remaining complete ASTs
must agree. All 18 handed-off core helper calls remain, so the helper stays
in both bundles and binder's local helper name stays unchanged. There is no
helper deletion or binder rename exemption. The complete raw hunks appear below.

## Check 4: unchanged whole lane and all shards

The frozen oracle worktree commit is a94047e854fd6699fc156a53cb98de97ca5c4e01. Its five functional adapter
inputs match this branch byte for byte. All 746 source files in the prepared
artifact match /workspace/bind-owned-after. owned/oracle/source-identity.json
records those assertions. Evidence commits do not alter frozen oracle inputs.

The unchanged standard prepare_shards.py producer completes all apply, locked
install, build and harness phases successfully. Its elapsed time is recorded
in owned/oracle/prepare-metrics.json. The unchanged lane command is:

```sh
stage3/lane/run.sh /workspace/scratch/bind-owned-lane-whole --artifact ARTIFACT
```

The whole lane passes with 106,366 passing, one sanctioned public API failure
and zero pending. All runners, eight workers, the normal 40-second timeout and
main's sanctioned API lines are unchanged. Its complete reports, task
observations, API diff and logs are under owned/oracle/whole/.

All 398 standard manifest shards pass in ten sequential pieces of at most
40 shards each. Each piece runs two standard shards concurrently, preserves
their normal 30-second limit, and is committed and plainly pushed immediately
after completion. Preparation, whole lane and each piece have a 1440-second
outer limit, below 25 minutes. Exact timings are in the piece receipts and this table.

| Piece | Seconds |
| --- | ---: |
| Preparation | 436.374 |
| Whole lane | 380.284 |
| Shard piece 0 | 94.139 |
| Shard piece 1 | 92.229 |
| Shard piece 2 | 91.580 |
| Shard piece 3 | 96.292 |
| Shard piece 4 | 93.772 |
| Shard piece 5 | 91.528 |
| Shard piece 6 | 96.295 |
| Shard piece 7 | 94.365 |
| Shard piece 8 | 100.884 |
| Shard piece 9 | 85.320 |

Shards use main's supported STAGE3_ADAPTED_TREE product path with symlink-only
views of the same artifact and its exact locked dependency installation.
No expected count, runner, API line, timeout or test is changed. Standard
shard.py performs its usual identity, input hash, count, timeout and API checks.
owned/oracle/shards/ retains all unit reports, observations and logs;
owned/oracle/pieces/ retains all ten result lists, driver logs and timings.
Successful redundant scratch views are deleted after their receipts are saved.

check_oracle_union.py independently verifies all 398 standard verdicts,
manifest coverage, runtime and product identity, every test identity/status
multiset against the whole lane, the summed 106,366/1/0 counts, and the exact
sanctioned failure title and paths. Concatenated shard API diff bytes equal
the whole lane's bytes. Main's unchanged check_results also passes.
owned/oracle/union.json is the complete union receipt.

## Check 5: restored-site negative control

The negative fixture restores the builder.ts host.writeFile call in
createBuilderProgram.emitBuildInfo and its required import. The fixture patch
matches the actual source delta. The normal build exits 0, then the identical
JavaScript comparison exits 1 and reports exactly that one unique original
site in both bundles. Each bundle has 27 bound sites and one original call;
the other seven products stay byte-identical. Its dedicated fixture, patch,
build log and comparison receipt are retained under their negative-control
file names, including owned/mutant-comparison.json. No extra change is hidden
by the normalizer.

## Check 6: ordinary and latent census

owned/census/ contains all before/after records, logs, summaries and the
comparison. The unchanged /workspace/bind-before baseline measurement is
reused; the current after measurement was rerun. All 79 single-root TypeScript
first-stop records agree exactly. There are zero changes. Seventy-three stop
at core.ts:810:89, Refused: a cast the runtime cannot check. The other six
records are unchanged too. Extra JSON and whole-program rows are retained
but excluded from the 79-root comparison.

The six CW051 unit identities come from core-walls-sort at
ad29102e9c145c5b480629227574d2ffa235a4b1. The scheduling selector comes from
container-fit-census at 654092b95416580c3e5be1f968fb8b14755b4481. The census-only
overlay uses main's latent harness. Its two hooks select files and top-level
units without changing registration, dependency bodies or first-error behavior.
Full and speculative modes are disabled. Every added Go source ends in .go.txt.

All 12 bounded pieces exit 0, select exactly one requested unit and report
checker_rejected false. File-wide refusal scans remain separate from selected
unit lowering stops. The maximum piece wall time is 212.227 seconds, below
25 minutes. owned/latent/ retains selections, raw streams, logs, timers and
comparison.json. These are the actual current-main locations, all NotYet:

| Unit | Before | After |
| --- | --- | --- |
| createBasicNodeBuilderModuleSpecifierResolutionHost | src/compiler/checker.ts:54279:37 | src/compiler/checker.ts:54279:37 |
| createWatchHost | src/compiler/core.ts:1533:12 | src/compiler/watch.ts:677:20 |
| createIncrementalCompilerHost | src/compiler/core.ts:1533:12 | src/compiler/watchPublic.ts:124:23 |
| performBuild | src/compiler/core.ts:1533:12 | src/compiler/core.ts:1533:12 |
| createWatchOfConfigFile | src/compiler/core.ts:1533:12 | src/compiler/core.ts:1533:12 |
| createWatchOfFilesAndCompilerOptions | src/compiler/core.ts:1533:12 | src/compiler/core.ts:1533:12 |

The checker unit remains at its method-presence boundary because checker.ts
is unchanged. createWatchHost and createIncrementalCompilerHost move from
core's optional call to direct optional calls, the form Apple's head lowers.
The three command-line units stay at core's optional call because the relevant
watch host owners belong to 32 and remain untouched. No native-lowering
completion is claimed. Historical core-walls line 1523 is current line 1533.

## All emitted JavaScript diff hunks

Seven files have no hunks. These are the complete raw hunks for the other two.

### _tsc.js

```diff
--- before/_tsc.js
+++ after/_tsc.js
@@ -122090,6 +122090,7 @@
   };
 }
 function createCompilerHostWorker(options, setParentNodes, system = sys) {
+  var _a;
   const existingDirectories = /* @__PURE__ */ new Map();
   const getCanonicalFileName = createGetCanonicalFileName(system.useCaseSensitiveFileNames);
   function directoryExists(directoryPath) {
@@ -122129,7 +122130,7 @@
     realpath,
     readDirectory: (path, extensions, include, exclude, depth) => system.readDirectory(path, extensions, include, exclude, depth),
     createDirectory: (d) => system.createDirectory(d),
-    createHash: maybeBind(system, system.createHash)
+    createHash: (_a = system.createHash) == null ? void 0 : _a.bind(system)
   };
   return compilerHost;
 }
@@ -125776,6 +125777,7 @@
   return filter(diagnostic, (d) => !d.skippedOn || !option[d.skippedOn]);
 }
 function parseConfigHostFromCompilerHostLike(host, directoryStructureHost = host) {
+  var _a, _b, _c;
   return {
     fileExists: (f) => directoryStructureHost.fileExists(f),
     readDirectory(root, extensions, excludes, includes, depth) {
@@ -125783,9 +125785,9 @@
       return directoryStructureHost.readDirectory(root, extensions, excludes, includes, depth);
     },
     readFile: (f) => directoryStructureHost.readFile(f),
-    directoryExists: maybeBind(directoryStructureHost, directoryStructureHost.directoryExists),
-    getDirectories: maybeBind(directoryStructureHost, directoryStructureHost.getDirectories),
-    realpath: maybeBind(directoryStructureHost, directoryStructureHost.realpath),
+    directoryExists: (_a = directoryStructureHost.directoryExists) == null ? void 0 : _a.bind(directoryStructureHost),
+    getDirectories: (_b = directoryStructureHost.getDirectories) == null ? void 0 : _b.bind(directoryStructureHost),
+    realpath: (_c = directoryStructureHost.realpath) == null ? void 0 : _c.bind(directoryStructureHost),
     useCaseSensitiveFileNames: host.useCaseSensitiveFileNames(),
     getCurrentDirectory: () => host.getCurrentDirectory(),
     onUnRecoverableConfigFileDiagnostic: host.onUnRecoverableConfigFileDiagnostic || returnUndefined,
@@ -127473,10 +127475,11 @@
   }
   return builderProgram;
   function emitBuildInfo(writeFile2, cancellationToken) {
+    var _a;
     Debug.assert(isBuilderProgramStateWithDefinedProgram(state));
     if (getBuildInfoEmitPending(state)) {
       const result = state.program.emitBuildInfo(
-        writeFile2 || maybeBind(host, host.writeFile),
+        writeFile2 || ((_a = host.writeFile) == null ? void 0 : _a.bind(host)),
         cancellationToken
       );
       state.buildInfoEmitPending = false;
@@ -127485,7 +127488,7 @@
     return emitSkippedWithNoDiagnostics;
   }
   function emitNextAffectedFileOrDtsErrors(writeFile2, cancellationToken, emitOnlyDtsFiles, customTransformers, isForDtsErrors) {
-    var _a, _b, _c, _d;
+    var _a, _b, _c, _d, _e;
     Debug.assert(isBuilderProgramStateWithDefinedProgram(state));
     let affected = getNextAffectedFile(state, cancellationToken, host);
     const programEmitKind = getBuilderFileEmit(state.compilerOptions);
@@ -127542,7 +127545,7 @@
         if (isForDtsErrors || !getBuildInfoEmitPending(state)) return void 0;
         const affected2 = state.program;
         const result2 = affected2.emitBuildInfo(
-          writeFile2 || maybeBind(host, host.writeFile),
+          writeFile2 || ((_b = host.writeFile) == null ? void 0 : _b.bind(host)),
           cancellationToken
         );
         state.buildInfoEmitPending = false;
@@ -127574,12 +127577,12 @@
       state.seenAffectedFiles.add(affectedSourceFile.resolvedPath);
       if (state.affectedFilesIndex !== void 0) state.affectedFilesIndex++;
       state.buildInfoEmitPending = true;
-      const existing = ((_b = state.seenEmittedFiles) == null ? void 0 : _b.get(affectedSourceFile.resolvedPath)) || 0 /* None */;
+      const existing = ((_c = state.seenEmittedFiles) == null ? void 0 : _c.get(affectedSourceFile.resolvedPath)) || 0 /* None */;
       (state.seenEmittedFiles ?? (state.seenEmittedFiles = /* @__PURE__ */ new Map())).set(affectedSourceFile.resolvedPath, emitKind | existing);
-      const existingPending = ((_c = state.affectedFilesPendingEmit) == null ? void 0 : _c.get(affectedSourceFile.resolvedPath)) || programEmitKind;
+      const existingPending = ((_d = state.affectedFilesPendingEmit) == null ? void 0 : _d.get(affectedSourceFile.resolvedPath)) || programEmitKind;
       const pendingKind = getPendingEmitKind(existingPending, emitKind | existing);
       if (pendingKind) (state.affectedFilesPendingEmit ?? (state.affectedFilesPendingEmit = /* @__PURE__ */ new Map())).set(affectedSourceFile.resolvedPath, pendingKind);
-      else (_d = state.affectedFilesPendingEmit) == null ? void 0 : _d.delete(affectedSourceFile.resolvedPath);
+      else (_e = state.affectedFilesPendingEmit) == null ? void 0 : _e.delete(affectedSourceFile.resolvedPath);
       if (result.diagnostics.length) (state.emitDiagnosticsPerFile ?? (state.emitDiagnosticsPerFile = /* @__PURE__ */ new Map())).set(affectedSourceFile.resolvedPath, result.diagnostics);
     } else {
       state.changedFilesSet.clear();
@@ -127611,10 +127614,11 @@
     );
   }
   function getWriteFileCallback(writeFile2, customTransformers) {
+    var _a;
     Debug.assert(isBuilderProgramStateWithDefinedProgram(state));
-    if (!getEmitDeclarations(state.compilerOptions)) return writeFile2 || maybeBind(host, host.writeFile);
+    if (!getEmitDeclarations(state.compilerOptions)) return writeFile2 || ((_a = host.writeFile) == null ? void 0 : _a.bind(host));
     return (fileName, text, writeByteOrderMark, onError, sourceFiles, data) => {
-      var _a, _b, _c;
+      var _a2, _b, _c;
       if (isDeclarationFileName(fileName)) {
         if (!state.compilerOptions.outFile) {
           Debug.assert((sourceFiles == null ? void 0 : sourceFiles.length) === 1);
@@ -127630,7 +127634,7 @@
                 host,
                 data
               );
-              if (!((_a = data == null ? void 0 : data.diagnostics) == null ? void 0 : _a.length)) emitSignature = signature;
+              if (!((_a2 = data == null ? void 0 : data.diagnostics) == null ? void 0 : _a2.length)) emitSignature = signature;
               if (signature !== file.version) {
                 if (host.storeSignatureInfo) (state.signatureInfo ?? (state.signatureInfo = /* @__PURE__ */ new Map())).set(file.resolvedPath, 1 /* StoredSignatureAtEmit */);
                 if (state.affectedFiles) {
@@ -129615,13 +129619,14 @@
 var noopFileWatcher = { close: noop };
 var returnNoopFileWatcher = () => noopFileWatcher;
 function createWatchHost(system = sys, reportWatchStatus2) {
+  var _a, _b, _c, _d;
   const onWatchStatusChange = reportWatchStatus2 || createWatchStatusReporter(system);
   return {
     onWatchStatusChange,
-    watchFile: maybeBind(system, system.watchFile) || returnNoopFileWatcher,
-    watchDirectory: maybeBind(system, system.watchDirectory) || returnNoopFileWatcher,
-    setTimeout: maybeBind(system, system.setTimeout) || noop,
-    clearTimeout: maybeBind(system, system.clearTimeout) || noop,
+    watchFile: ((_a = system.watchFile) == null ? void 0 : _a.bind(system)) || returnNoopFileWatcher,
+    watchDirectory: ((_b = system.watchDirectory) == null ? void 0 : _b.bind(system)) || returnNoopFileWatcher,
+    setTimeout: ((_c = system.setTimeout) == null ? void 0 : _c.bind(system)) || noop,
+    clearTimeout: ((_d = system.clearTimeout) == null ? void 0 : _d.bind(system)) || noop,
     preferNonRecursiveWatch: system.preferNonRecursiveWatch
   };
 }
@@ -129846,13 +129851,14 @@
   return createBuilderProgramUsingIncrementalBuildInfo(buildInfo, buildInfoPath, host);
 }
 function createIncrementalCompilerHost(options, system = sys) {
+  var _a;
   const host = createCompilerHostWorker(
     options,
     /*setParentNodes*/
     void 0,
     system
   );
-  host.createHash = maybeBind(system, system.createHash);
+  host.createHash = (_a = system.createHash) == null ? void 0 : _a.bind(system);
   host.storeSignatureInfo = system.storeSignatureInfo;
   setGetSourceFileAsHashVersioned(host);
   changeCompilerHostLikeToUseCache(host, (fileName) => toPath(fileName, host.getCurrentDirectory(), host.getCanonicalFileName));
@@ -129872,6 +129878,7 @@
   return createProgram2(rootNames, options, host, oldProgram, configFileParsingDiagnostics, projectReferences);
 }
 function createWatchProgram(host) {
+  var _a, _b, _c, _d, _e, _f, _g, _h;
   let builderProgram;
   let updateLevel;
   let missingFilesMap;
@@ -129929,7 +129936,7 @@
   compilerHost.onReleaseParsedCommandLine = onReleaseParsedCommandLine;
   compilerHost.toPath = toPath3;
   compilerHost.getCompilationSettings = () => compilerOptions;
-  compilerHost.useSourceOfProjectReferenceRedirect = maybeBind(host, host.useSourceOfProjectReferenceRedirect);
+  compilerHost.useSourceOfProjectReferenceRedirect = (_a = host.useSourceOfProjectReferenceRedirect) == null ? void 0 : _a.bind(host);
   compilerHost.preferNonRecursiveWatch = host.preferNonRecursiveWatch;
   compilerHost.watchDirectoryOfFailedLookupLocation = (dir, cb, flags) => watchDirectory(dir, cb, flags, watchOptions, WatchType.FailedLookupLocations);
   compilerHost.watchAffectingFileLocation = (file, cb) => watchFile3(file, cb, 2e3 /* High */, watchOptions, WatchType.AffectingFileLocation);
@@ -129948,21 +129955,21 @@
     /*logChangesWhenResolvingModule*/
     false
   );
-  compilerHost.resolveModuleNameLiterals = maybeBind(host, host.resolveModuleNameLiterals);
-  compilerHost.resolveModuleNames = maybeBind(host, host.resolveModuleNames);
+  compilerHost.resolveModuleNameLiterals = (_b = host.resolveModuleNameLiterals) == null ? void 0 : _b.bind(host);
+  compilerHost.resolveModuleNames = (_c = host.resolveModuleNames) == null ? void 0 : _c.bind(host);
   if (!compilerHost.resolveModuleNameLiterals && !compilerHost.resolveModuleNames) {
     compilerHost.resolveModuleNameLiterals = resolutionCache.resolveModuleNameLiterals.bind(resolutionCache);
   }
-  compilerHost.resolveTypeReferenceDirectiveReferences = maybeBind(host, host.resolveTypeReferenceDirectiveReferences);
-  compilerHost.resolveTypeReferenceDirectives = maybeBind(host, host.resolveTypeReferenceDirectives);
+  compilerHost.resolveTypeReferenceDirectiveReferences = (_d = host.resolveTypeReferenceDirectiveReferences) == null ? void 0 : _d.bind(host);
+  compilerHost.resolveTypeReferenceDirectives = (_e = host.resolveTypeReferenceDirectives) == null ? void 0 : _e.bind(host);
   if (!compilerHost.resolveTypeReferenceDirectiveReferences && !compilerHost.resolveTypeReferenceDirectives) {
     compilerHost.resolveTypeReferenceDirectiveReferences = resolutionCache.resolveTypeReferenceDirectiveReferences.bind(resolutionCache);
   }
   compilerHost.resolveLibrary = !host.resolveLibrary ? resolutionCache.resolveLibrary.bind(resolutionCache) : host.resolveLibrary.bind(host);
-  compilerHost.getModuleResolutionCache = host.resolveModuleNameLiterals || host.resolveModuleNames ? maybeBind(host, host.getModuleResolutionCache) : (() => resolutionCache.getModuleResolutionCache());
+  compilerHost.getModuleResolutionCache = host.resolveModuleNameLiterals || host.resolveModuleNames ? (_f = host.getModuleResolutionCache) == null ? void 0 : _f.bind(host) : (() => resolutionCache.getModuleResolutionCache());
   const userProvidedResolution = !!host.resolveModuleNameLiterals || !!host.resolveTypeReferenceDirectiveReferences || !!host.resolveModuleNames || !!host.resolveTypeReferenceDirectives;
-  const customHasInvalidatedResolutions = userProvidedResolution ? maybeBind(host, host.hasInvalidatedResolutions) || returnTrue : returnFalse;
-  const customHasInvalidLibResolutions = host.resolveLibrary ? maybeBind(host, host.hasInvalidatedLibResolutions) || returnTrue : returnFalse;
+  const customHasInvalidatedResolutions = userProvidedResolution ? ((_g = host.hasInvalidatedResolutions) == null ? void 0 : _g.bind(host)) || returnTrue : returnFalse;
+  const customHasInvalidLibResolutions = host.resolveLibrary ? ((_h = host.hasInvalidatedLibResolutions) == null ? void 0 : _h.bind(host)) || returnTrue : returnFalse;
   builderProgram = readBuilderProgram(compilerOptions, compilerHost);
   synchronizeProgram();
   return configFileName ? { getCurrentProgram: getCurrentBuilderProgram, getProgram: updateProgram, close, getResolutionCache } : { getCurrentProgram: getCurrentBuilderProgram, getProgram: updateProgram, updateRootFileNames, close, getResolutionCache };
@@ -129995,8 +130002,8 @@
     }
     if (parsedConfigs) {
       clearMap(parsedConfigs, (config) => {
-        var _a;
-        (_a = config.watcher) == null ? void 0 : _a.close();
+        var _a2;
+        (_a2 = config.watcher) == null ? void 0 : _a2.close();
         config.watcher = void 0;
         if (config.watchedDirectories) clearMap(config.watchedDirectories, closeFileWatcherOf);
         config.watchedDirectories = void 0;
@@ -130360,13 +130367,13 @@
     return parsedCommandLine;
   }
   function onReleaseParsedCommandLine(fileName) {
-    var _a;
+    var _a2;
     const path = toPath3(fileName);
     const config = parsedConfigs == null ? void 0 : parsedConfigs.get(path);
     if (!config) return;
     parsedConfigs.delete(path);
     if (config.watchedDirectories) clearMap(config.watchedDirectories, closeFileWatcherOf);
-    (_a = config.watcher) == null ? void 0 : _a.close();
+    (_a2 = config.watcher) == null ? void 0 : _a2.close();
     clearSharedExtendedConfigFileWatcher(path, sharedExtendedConfigFileWatchers);
   }
   function watchFilePath(path, file, callback, pollingInterval, options, watchType) {
@@ -130453,10 +130460,10 @@
       (extendedConfigFileName, extendedConfigFilePath) => watchFile3(
         extendedConfigFileName,
         (_fileName, eventKind) => {
-          var _a;
+          var _a2;
           updateCachedSystemWithFile(extendedConfigFileName, extendedConfigFilePath, eventKind);
           if (extendedConfigCache) cleanExtendedConfigCache(extendedConfigCache, extendedConfigFilePath, toPath3);
-          const projects = (_a = sharedExtendedConfigFileWatchers.get(extendedConfigFilePath)) == null ? void 0 : _a.projects;
+          const projects = (_a2 = sharedExtendedConfigFileWatchers.get(extendedConfigFilePath)) == null ? void 0 : _a2.projects;
           if (!(projects == null ? void 0 : projects.size)) return;
           projects.forEach((projectPath) => {
             if (configFileName && toPath3(configFileName) === projectPath) {
@@ -130477,7 +130484,7 @@
     );
   }
   function watchReferencedProject(configFileName2, configPath, commandLine) {
-    var _a, _b, _c, _d;
+    var _a2, _b2, _c2, _d2;
     commandLine.watcher || (commandLine.watcher = watchFile3(
       configFileName2,
       (_fileName, eventKind) => {
@@ -130488,14 +130495,14 @@
         scheduleProgramUpdate();
       },
       2e3 /* High */,
-      ((_a = commandLine.parsedCommandLine) == null ? void 0 : _a.watchOptions) || watchOptions,
+      ((_a2 = commandLine.parsedCommandLine) == null ? void 0 : _a2.watchOptions) || watchOptions,
       WatchType.ConfigFileOfReferencedProject
     ));
     updateWatchingWildcardDirectories(
       commandLine.watchedDirectories || (commandLine.watchedDirectories = /* @__PURE__ */ new Map()),
-      (_b = commandLine.parsedCommandLine) == null ? void 0 : _b.wildcardDirectories,
+      (_b2 = commandLine.parsedCommandLine) == null ? void 0 : _b2.wildcardDirectories,
       (directory, flags) => {
-        var _a2;
+        var _a3;
         return watchDirectory(
           directory,
           (fileOrDirectory) => {
@@ -130524,15 +130531,15 @@
             }
           },
           flags,
-          ((_a2 = commandLine.parsedCommandLine) == null ? void 0 : _a2.watchOptions) || watchOptions,
+          ((_a3 = commandLine.parsedCommandLine) == null ? void 0 : _a3.watchOptions) || watchOptions,
           WatchType.WildcardDirectoryOfReferencedProject
         );
       }
     );
     updateExtendedConfigFilesWatches(
       configPath,
-      (_c = commandLine.parsedCommandLine) == null ? void 0 : _c.options,
-      ((_d = commandLine.parsedCommandLine) == null ? void 0 : _d.watchOptions) || watchOptions,
+      (_c2 = commandLine.parsedCommandLine) == null ? void 0 : _c2.options,
+      ((_d2 = commandLine.parsedCommandLine) == null ? void 0 : _d2.watchOptions) || watchOptions,
       WatchType.ExtendedConfigOfReferencedProject
     );
   }
@@ -130577,13 +130584,14 @@
   };
 }
 function createSolutionBuilderHostBase(system, createProgram2, reportDiagnostic, reportSolutionBuilderStatus) {
+  var _a;
   const host = createProgramHost(system, createProgram2);
   host.getModifiedTime = system.getModifiedTime ? (path) => system.getModifiedTime(path) : returnUndefined;
   host.setModifiedTime = system.setModifiedTime ? (path, date) => system.setModifiedTime(path, date) : noop;
   host.deleteFile = system.deleteFile ? (path) => system.deleteFile(path) : noop;
   host.reportDiagnostic = reportDiagnostic || createDiagnosticReporter(system);
   host.reportSolutionBuilderStatus = reportSolutionBuilderStatus || createBuilderStatusReporter(system);
-  host.now = maybeBind(system, system.now);
+  host.now = (_a = system.now) == null ? void 0 : _a.bind(system);
   return host;
 }
 function createSolutionBuilderHost(system = sys, createProgram2, reportDiagnostic, reportSolutionBuilderStatus, reportErrorSummary2) {
@@ -130625,18 +130633,19 @@
   );
 }
 function createSolutionBuilderState(watch2, hostOrHostWithWatch, rootNames, options, baseWatchOptions) {
+  var _a, _b, _c, _d, _e, _f, _g;
   const host = hostOrHostWithWatch;
   const hostWithWatch = hostOrHostWithWatch;
   const baseCompilerOptions = getCompilerOptionsOfBuildOptions(options);
   const compilerHost = createCompilerHostFromProgramHost(host, () => state.projectCompilerOptions);
   setGetSourceFileAsHashVersioned(compilerHost);
   compilerHost.getParsedCommandLine = (fileName) => parseConfigFile(state, fileName, toResolvedConfigFilePath(state, fileName));
-  compilerHost.resolveModuleNameLiterals = maybeBind(host, host.resolveModuleNameLiterals);
-  compilerHost.resolveTypeReferenceDirectiveReferences = maybeBind(host, host.resolveTypeReferenceDirectiveReferences);
-  compilerHost.resolveLibrary = maybeBind(host, host.resolveLibrary);
-  compilerHost.resolveModuleNames = maybeBind(host, host.resolveModuleNames);
-  compilerHost.resolveTypeReferenceDirectives = maybeBind(host, host.resolveTypeReferenceDirectives);
-  compilerHost.getModuleResolutionCache = maybeBind(host, host.getModuleResolutionCache);
+  compilerHost.resolveModuleNameLiterals = (_a = host.resolveModuleNameLiterals) == null ? void 0 : _a.bind(host);
+  compilerHost.resolveTypeReferenceDirectiveReferences = (_b = host.resolveTypeReferenceDirectiveReferences) == null ? void 0 : _b.bind(host);
+  compilerHost.resolveLibrary = (_c = host.resolveLibrary) == null ? void 0 : _c.bind(host);
+  compilerHost.resolveModuleNames = (_d = host.resolveModuleNames) == null ? void 0 : _d.bind(host);
+  compilerHost.resolveTypeReferenceDirectives = (_e = host.resolveTypeReferenceDirectives) == null ? void 0 : _e.bind(host);
+  compilerHost.getModuleResolutionCache = (_f = host.getModuleResolutionCache) == null ? void 0 : _f.bind(host);
   let moduleResolutionCache, typeReferenceDirectiveResolutionCache;
   if (!compilerHost.resolveModuleNameLiterals && !compilerHost.resolveModuleNames) {
     moduleResolutionCache = createModuleResolutionCache(compilerHost.getCurrentDirectory(), compilerHost.getCanonicalFileName);
@@ -130701,7 +130710,7 @@
     host,
     hostWithWatch,
     parseConfigFileHost: parseConfigHostFromCompilerHostLike(host),
-    write: maybeBind(host, host.trace),
+    write: (_g = host.trace) == null ? void 0 : _g.bind(host),
     // State of solution
     options,
     baseCompilerOptions,
```

### typescript.js

```diff
--- before/typescript.js
+++ after/typescript.js
@@ -126922,6 +126922,7 @@
   };
 }
 function createCompilerHostWorker(options, setParentNodes, system = sys) {
+  var _a;
   const existingDirectories = /* @__PURE__ */ new Map();
   const getCanonicalFileName = createGetCanonicalFileName(system.useCaseSensitiveFileNames);
   function directoryExists(directoryPath) {
@@ -126961,7 +126962,7 @@
     realpath,
     readDirectory: (path, extensions, include, exclude, depth) => system.readDirectory(path, extensions, include, exclude, depth),
     createDirectory: (d) => system.createDirectory(d),
-    createHash: maybeBind(system, system.createHash)
+    createHash: (_a = system.createHash) == null ? void 0 : _a.bind(system)
   };
   return compilerHost;
 }
@@ -130642,6 +130643,7 @@
   return filter(diagnostic, (d) => !d.skippedOn || !option[d.skippedOn]);
 }
 function parseConfigHostFromCompilerHostLike(host, directoryStructureHost = host) {
+  var _a, _b, _c;
   return {
     fileExists: (f) => directoryStructureHost.fileExists(f),
     readDirectory(root, extensions, excludes, includes, depth) {
@@ -130649,9 +130651,9 @@
       return directoryStructureHost.readDirectory(root, extensions, excludes, includes, depth);
     },
     readFile: (f) => directoryStructureHost.readFile(f),
-    directoryExists: maybeBind(directoryStructureHost, directoryStructureHost.directoryExists),
-    getDirectories: maybeBind(directoryStructureHost, directoryStructureHost.getDirectories),
-    realpath: maybeBind(directoryStructureHost, directoryStructureHost.realpath),
+    directoryExists: (_a = directoryStructureHost.directoryExists) == null ? void 0 : _a.bind(directoryStructureHost),
+    getDirectories: (_b = directoryStructureHost.getDirectories) == null ? void 0 : _b.bind(directoryStructureHost),
+    realpath: (_c = directoryStructureHost.realpath) == null ? void 0 : _c.bind(directoryStructureHost),
     useCaseSensitiveFileNames: host.useCaseSensitiveFileNames(),
     getCurrentDirectory: () => host.getCurrentDirectory(),
     onUnRecoverableConfigFileDiagnostic: host.onUnRecoverableConfigFileDiagnostic || returnUndefined,
@@ -132373,10 +132375,11 @@
   }
   return builderProgram;
   function emitBuildInfo(writeFile2, cancellationToken) {
+    var _a;
     Debug.assert(isBuilderProgramStateWithDefinedProgram(state));
     if (getBuildInfoEmitPending(state)) {
       const result = state.program.emitBuildInfo(
-        writeFile2 || maybeBind(host, host.writeFile),
+        writeFile2 || ((_a = host.writeFile) == null ? void 0 : _a.bind(host)),
         cancellationToken
       );
       state.buildInfoEmitPending = false;
@@ -132385,7 +132388,7 @@
     return emitSkippedWithNoDiagnostics;
   }
   function emitNextAffectedFileOrDtsErrors(writeFile2, cancellationToken, emitOnlyDtsFiles, customTransformers, isForDtsErrors) {
-    var _a, _b, _c, _d;
+    var _a, _b, _c, _d, _e;
     Debug.assert(isBuilderProgramStateWithDefinedProgram(state));
     let affected = getNextAffectedFile(state, cancellationToken, host);
     const programEmitKind = getBuilderFileEmit(state.compilerOptions);
@@ -132442,7 +132445,7 @@
         if (isForDtsErrors || !getBuildInfoEmitPending(state)) return void 0;
         const affected2 = state.program;
         const result2 = affected2.emitBuildInfo(
-          writeFile2 || maybeBind(host, host.writeFile),
+          writeFile2 || ((_b = host.writeFile) == null ? void 0 : _b.bind(host)),
           cancellationToken
         );
         state.buildInfoEmitPending = false;
@@ -132474,12 +132477,12 @@
       state.seenAffectedFiles.add(affectedSourceFile.resolvedPath);
       if (state.affectedFilesIndex !== void 0) state.affectedFilesIndex++;
       state.buildInfoEmitPending = true;
-      const existing = ((_b = state.seenEmittedFiles) == null ? void 0 : _b.get(affectedSourceFile.resolvedPath)) || 0 /* None */;
+      const existing = ((_c = state.seenEmittedFiles) == null ? void 0 : _c.get(affectedSourceFile.resolvedPath)) || 0 /* None */;
       (state.seenEmittedFiles ?? (state.seenEmittedFiles = /* @__PURE__ */ new Map())).set(affectedSourceFile.resolvedPath, emitKind | existing);
-      const existingPending = ((_c = state.affectedFilesPendingEmit) == null ? void 0 : _c.get(affectedSourceFile.resolvedPath)) || programEmitKind;
+      const existingPending = ((_d = state.affectedFilesPendingEmit) == null ? void 0 : _d.get(affectedSourceFile.resolvedPath)) || programEmitKind;
       const pendingKind = getPendingEmitKind(existingPending, emitKind | existing);
       if (pendingKind) (state.affectedFilesPendingEmit ?? (state.affectedFilesPendingEmit = /* @__PURE__ */ new Map())).set(affectedSourceFile.resolvedPath, pendingKind);
-      else (_d = state.affectedFilesPendingEmit) == null ? void 0 : _d.delete(affectedSourceFile.resolvedPath);
+      else (_e = state.affectedFilesPendingEmit) == null ? void 0 : _e.delete(affectedSourceFile.resolvedPath);
       if (result.diagnostics.length) (state.emitDiagnosticsPerFile ?? (state.emitDiagnosticsPerFile = /* @__PURE__ */ new Map())).set(affectedSourceFile.resolvedPath, result.diagnostics);
     } else {
       state.changedFilesSet.clear();
@@ -132511,10 +132514,11 @@
     );
   }
   function getWriteFileCallback(writeFile2, customTransformers) {
+    var _a;
     Debug.assert(isBuilderProgramStateWithDefinedProgram(state));
-    if (!getEmitDeclarations(state.compilerOptions)) return writeFile2 || maybeBind(host, host.writeFile);
+    if (!getEmitDeclarations(state.compilerOptions)) return writeFile2 || ((_a = host.writeFile) == null ? void 0 : _a.bind(host));
     return (fileName, text, writeByteOrderMark, onError, sourceFiles, data) => {
-      var _a, _b, _c;
+      var _a2, _b, _c;
       if (isDeclarationFileName(fileName)) {
         if (!state.compilerOptions.outFile) {
           Debug.assert((sourceFiles == null ? void 0 : sourceFiles.length) === 1);
@@ -132530,7 +132534,7 @@
                 host,
                 data
               );
-              if (!((_a = data == null ? void 0 : data.diagnostics) == null ? void 0 : _a.length)) emitSignature = signature;
+              if (!((_a2 = data == null ? void 0 : data.diagnostics) == null ? void 0 : _a2.length)) emitSignature = signature;
               if (signature !== file.version) {
                 if (host.storeSignatureInfo) (state.signatureInfo ?? (state.signatureInfo = /* @__PURE__ */ new Map())).set(file.resolvedPath, 1 /* StoredSignatureAtEmit */);
                 if (state.affectedFiles) {
@@ -134542,13 +134546,14 @@
 var noopFileWatcher = { close: noop };
 var returnNoopFileWatcher = () => noopFileWatcher;
 function createWatchHost(system = sys, reportWatchStatus2) {
+  var _a, _b, _c, _d;
   const onWatchStatusChange = reportWatchStatus2 || createWatchStatusReporter(system);
   return {
     onWatchStatusChange,
-    watchFile: maybeBind(system, system.watchFile) || returnNoopFileWatcher,
-    watchDirectory: maybeBind(system, system.watchDirectory) || returnNoopFileWatcher,
-    setTimeout: maybeBind(system, system.setTimeout) || noop,
-    clearTimeout: maybeBind(system, system.clearTimeout) || noop,
+    watchFile: ((_a = system.watchFile) == null ? void 0 : _a.bind(system)) || returnNoopFileWatcher,
+    watchDirectory: ((_b = system.watchDirectory) == null ? void 0 : _b.bind(system)) || returnNoopFileWatcher,
+    setTimeout: ((_c = system.setTimeout) == null ? void 0 : _c.bind(system)) || noop,
+    clearTimeout: ((_d = system.clearTimeout) == null ? void 0 : _d.bind(system)) || noop,
     preferNonRecursiveWatch: system.preferNonRecursiveWatch
   };
 }
@@ -134773,13 +134778,14 @@
   return createBuilderProgramUsingIncrementalBuildInfo(buildInfo, buildInfoPath, host);
 }
 function createIncrementalCompilerHost(options, system = sys) {
+  var _a;
   const host = createCompilerHostWorker(
     options,
     /*setParentNodes*/
     void 0,
     system
   );
-  host.createHash = maybeBind(system, system.createHash);
+  host.createHash = (_a = system.createHash) == null ? void 0 : _a.bind(system);
   host.storeSignatureInfo = system.storeSignatureInfo;
   setGetSourceFileAsHashVersioned(host);
   changeCompilerHostLikeToUseCache(host, (fileName) => toPath(fileName, host.getCurrentDirectory(), host.getCanonicalFileName));
@@ -134824,6 +134830,7 @@
   }
 }
 function createWatchProgram(host) {
+  var _a, _b, _c, _d, _e, _f, _g, _h;
   let builderProgram;
   let updateLevel;
   let missingFilesMap;
@@ -134881,7 +134888,7 @@
   compilerHost.onReleaseParsedCommandLine = onReleaseParsedCommandLine;
   compilerHost.toPath = toPath3;
   compilerHost.getCompilationSettings = () => compilerOptions;
-  compilerHost.useSourceOfProjectReferenceRedirect = maybeBind(host, host.useSourceOfProjectReferenceRedirect);
+  compilerHost.useSourceOfProjectReferenceRedirect = (_a = host.useSourceOfProjectReferenceRedirect) == null ? void 0 : _a.bind(host);
   compilerHost.preferNonRecursiveWatch = host.preferNonRecursiveWatch;
   compilerHost.watchDirectoryOfFailedLookupLocation = (dir, cb, flags) => watchDirectory(dir, cb, flags, watchOptions, WatchType.FailedLookupLocations);
   compilerHost.watchAffectingFileLocation = (file, cb) => watchFile3(file, cb, 2e3 /* High */, watchOptions, WatchType.AffectingFileLocation);
@@ -134900,21 +134907,21 @@
     /*logChangesWhenResolvingModule*/
     false
   );
-  compilerHost.resolveModuleNameLiterals = maybeBind(host, host.resolveModuleNameLiterals);
-  compilerHost.resolveModuleNames = maybeBind(host, host.resolveModuleNames);
+  compilerHost.resolveModuleNameLiterals = (_b = host.resolveModuleNameLiterals) == null ? void 0 : _b.bind(host);
+  compilerHost.resolveModuleNames = (_c = host.resolveModuleNames) == null ? void 0 : _c.bind(host);
   if (!compilerHost.resolveModuleNameLiterals && !compilerHost.resolveModuleNames) {
     compilerHost.resolveModuleNameLiterals = resolutionCache.resolveModuleNameLiterals.bind(resolutionCache);
   }
-  compilerHost.resolveTypeReferenceDirectiveReferences = maybeBind(host, host.resolveTypeReferenceDirectiveReferences);
-  compilerHost.resolveTypeReferenceDirectives = maybeBind(host, host.resolveTypeReferenceDirectives);
+  compilerHost.resolveTypeReferenceDirectiveReferences = (_d = host.resolveTypeReferenceDirectiveReferences) == null ? void 0 : _d.bind(host);
+  compilerHost.resolveTypeReferenceDirectives = (_e = host.resolveTypeReferenceDirectives) == null ? void 0 : _e.bind(host);
   if (!compilerHost.resolveTypeReferenceDirectiveReferences && !compilerHost.resolveTypeReferenceDirectives) {
     compilerHost.resolveTypeReferenceDirectiveReferences = resolutionCache.resolveTypeReferenceDirectiveReferences.bind(resolutionCache);
   }
   compilerHost.resolveLibrary = !host.resolveLibrary ? resolutionCache.resolveLibrary.bind(resolutionCache) : host.resolveLibrary.bind(host);
-  compilerHost.getModuleResolutionCache = host.resolveModuleNameLiterals || host.resolveModuleNames ? maybeBind(host, host.getModuleResolutionCache) : (() => resolutionCache.getModuleResolutionCache());
+  compilerHost.getModuleResolutionCache = host.resolveModuleNameLiterals || host.resolveModuleNames ? (_f = host.getModuleResolutionCache) == null ? void 0 : _f.bind(host) : (() => resolutionCache.getModuleResolutionCache());
   const userProvidedResolution = !!host.resolveModuleNameLiterals || !!host.resolveTypeReferenceDirectiveReferences || !!host.resolveModuleNames || !!host.resolveTypeReferenceDirectives;
-  const customHasInvalidatedResolutions = userProvidedResolution ? maybeBind(host, host.hasInvalidatedResolutions) || returnTrue : returnFalse;
-  const customHasInvalidLibResolutions = host.resolveLibrary ? maybeBind(host, host.hasInvalidatedLibResolutions) || returnTrue : returnFalse;
+  const customHasInvalidatedResolutions = userProvidedResolution ? ((_g = host.hasInvalidatedResolutions) == null ? void 0 : _g.bind(host)) || returnTrue : returnFalse;
+  const customHasInvalidLibResolutions = host.resolveLibrary ? ((_h = host.hasInvalidatedLibResolutions) == null ? void 0 : _h.bind(host)) || returnTrue : returnFalse;
   builderProgram = readBuilderProgram(compilerOptions, compilerHost);
   synchronizeProgram();
   return configFileName ? { getCurrentProgram: getCurrentBuilderProgram, getProgram: updateProgram, close, getResolutionCache } : { getCurrentProgram: getCurrentBuilderProgram, getProgram: updateProgram, updateRootFileNames, close, getResolutionCache };
@@ -134947,8 +134954,8 @@
     }
     if (parsedConfigs) {
       clearMap(parsedConfigs, (config) => {
-        var _a;
-        (_a = config.watcher) == null ? void 0 : _a.close();
+        var _a2;
+        (_a2 = config.watcher) == null ? void 0 : _a2.close();
         config.watcher = void 0;
         if (config.watchedDirectories) clearMap(config.watchedDirectories, closeFileWatcherOf);
         config.watchedDirectories = void 0;
@@ -135312,13 +135319,13 @@
     return parsedCommandLine;
   }
   function onReleaseParsedCommandLine(fileName) {
-    var _a;
+    var _a2;
     const path = toPath3(fileName);
     const config = parsedConfigs == null ? void 0 : parsedConfigs.get(path);
     if (!config) return;
     parsedConfigs.delete(path);
     if (config.watchedDirectories) clearMap(config.watchedDirectories, closeFileWatcherOf);
-    (_a = config.watcher) == null ? void 0 : _a.close();
+    (_a2 = config.watcher) == null ? void 0 : _a2.close();
     clearSharedExtendedConfigFileWatcher(path, sharedExtendedConfigFileWatchers);
   }
   function watchFilePath(path, file, callback, pollingInterval, options, watchType) {
@@ -135405,10 +135412,10 @@
       (extendedConfigFileName, extendedConfigFilePath) => watchFile3(
         extendedConfigFileName,
         (_fileName, eventKind) => {
-          var _a;
+          var _a2;
           updateCachedSystemWithFile(extendedConfigFileName, extendedConfigFilePath, eventKind);
           if (extendedConfigCache) cleanExtendedConfigCache(extendedConfigCache, extendedConfigFilePath, toPath3);
-          const projects = (_a = sharedExtendedConfigFileWatchers.get(extendedConfigFilePath)) == null ? void 0 : _a.projects;
+          const projects = (_a2 = sharedExtendedConfigFileWatchers.get(extendedConfigFilePath)) == null ? void 0 : _a2.projects;
           if (!(projects == null ? void 0 : projects.size)) return;
           projects.forEach((projectPath) => {
             if (configFileName && toPath3(configFileName) === projectPath) {
@@ -135429,7 +135436,7 @@
     );
   }
   function watchReferencedProject(configFileName2, configPath, commandLine) {
-    var _a, _b, _c, _d;
+    var _a2, _b2, _c2, _d2;
     commandLine.watcher || (commandLine.watcher = watchFile3(
       configFileName2,
       (_fileName, eventKind) => {
@@ -135440,14 +135447,14 @@
         scheduleProgramUpdate();
       },
       2e3 /* High */,
-      ((_a = commandLine.parsedCommandLine) == null ? void 0 : _a.watchOptions) || watchOptions,
+      ((_a2 = commandLine.parsedCommandLine) == null ? void 0 : _a2.watchOptions) || watchOptions,
       WatchType.ConfigFileOfReferencedProject
     ));
     updateWatchingWildcardDirectories(
       commandLine.watchedDirectories || (commandLine.watchedDirectories = /* @__PURE__ */ new Map()),
-      (_b = commandLine.parsedCommandLine) == null ? void 0 : _b.wildcardDirectories,
+      (_b2 = commandLine.parsedCommandLine) == null ? void 0 : _b2.wildcardDirectories,
       (directory, flags) => {
-        var _a2;
+        var _a3;
         return watchDirectory(
           directory,
           (fileOrDirectory) => {
@@ -135476,15 +135483,15 @@
             }
           },
           flags,
-          ((_a2 = commandLine.parsedCommandLine) == null ? void 0 : _a2.watchOptions) || watchOptions,
+          ((_a3 = commandLine.parsedCommandLine) == null ? void 0 : _a3.watchOptions) || watchOptions,
           WatchType.WildcardDirectoryOfReferencedProject
         );
       }
     );
     updateExtendedConfigFilesWatches(
       configPath,
-      (_c = commandLine.parsedCommandLine) == null ? void 0 : _c.options,
-      ((_d = commandLine.parsedCommandLine) == null ? void 0 : _d.watchOptions) || watchOptions,
+      (_c2 = commandLine.parsedCommandLine) == null ? void 0 : _c2.options,
+      ((_d2 = commandLine.parsedCommandLine) == null ? void 0 : _d2.watchOptions) || watchOptions,
       WatchType.ExtendedConfigOfReferencedProject
     );
   }
@@ -135550,13 +135557,14 @@
   };
 }
 function createSolutionBuilderHostBase(system, createProgram2, reportDiagnostic, reportSolutionBuilderStatus) {
+  var _a;
   const host = createProgramHost(system, createProgram2);
   host.getModifiedTime = system.getModifiedTime ? (path) => system.getModifiedTime(path) : returnUndefined;
   host.setModifiedTime = system.setModifiedTime ? (path, date) => system.setModifiedTime(path, date) : noop;
   host.deleteFile = system.deleteFile ? (path) => system.deleteFile(path) : noop;
   host.reportDiagnostic = reportDiagnostic || createDiagnosticReporter(system);
   host.reportSolutionBuilderStatus = reportSolutionBuilderStatus || createBuilderStatusReporter(system);
-  host.now = maybeBind(system, system.now);
+  host.now = (_a = system.now) == null ? void 0 : _a.bind(system);
   return host;
 }
 function createSolutionBuilderHost(system = sys, createProgram2, reportDiagnostic, reportSolutionBuilderStatus, reportErrorSummary2) {
@@ -135598,18 +135606,19 @@
   );
 }
 function createSolutionBuilderState(watch2, hostOrHostWithWatch, rootNames, options, baseWatchOptions) {
+  var _a, _b, _c, _d, _e, _f, _g;
   const host = hostOrHostWithWatch;
   const hostWithWatch = hostOrHostWithWatch;
   const baseCompilerOptions = getCompilerOptionsOfBuildOptions(options);
   const compilerHost = createCompilerHostFromProgramHost(host, () => state.projectCompilerOptions);
   setGetSourceFileAsHashVersioned(compilerHost);
   compilerHost.getParsedCommandLine = (fileName) => parseConfigFile(state, fileName, toResolvedConfigFilePath(state, fileName));
-  compilerHost.resolveModuleNameLiterals = maybeBind(host, host.resolveModuleNameLiterals);
-  compilerHost.resolveTypeReferenceDirectiveReferences = maybeBind(host, host.resolveTypeReferenceDirectiveReferences);
-  compilerHost.resolveLibrary = maybeBind(host, host.resolveLibrary);
-  compilerHost.resolveModuleNames = maybeBind(host, host.resolveModuleNames);
-  compilerHost.resolveTypeReferenceDirectives = maybeBind(host, host.resolveTypeReferenceDirectives);
-  compilerHost.getModuleResolutionCache = maybeBind(host, host.getModuleResolutionCache);
+  compilerHost.resolveModuleNameLiterals = (_a = host.resolveModuleNameLiterals) == null ? void 0 : _a.bind(host);
+  compilerHost.resolveTypeReferenceDirectiveReferences = (_b = host.resolveTypeReferenceDirectiveReferences) == null ? void 0 : _b.bind(host);
+  compilerHost.resolveLibrary = (_c = host.resolveLibrary) == null ? void 0 : _c.bind(host);
+  compilerHost.resolveModuleNames = (_d = host.resolveModuleNames) == null ? void 0 : _d.bind(host);
+  compilerHost.resolveTypeReferenceDirectives = (_e = host.resolveTypeReferenceDirectives) == null ? void 0 : _e.bind(host);
+  compilerHost.getModuleResolutionCache = (_f = host.getModuleResolutionCache) == null ? void 0 : _f.bind(host);
   let moduleResolutionCache, typeReferenceDirectiveResolutionCache;
   if (!compilerHost.resolveModuleNameLiterals && !compilerHost.resolveModuleNames) {
     moduleResolutionCache = createModuleResolutionCache(compilerHost.getCurrentDirectory(), compilerHost.getCanonicalFileName);
@@ -135674,7 +135683,7 @@
     host,
     hostWithWatch,
     parseConfigFileHost: parseConfigHostFromCompilerHostLike(host),
-    write: maybeBind(host, host.trace),
+    write: (_g = host.trace) == null ? void 0 : _g.bind(host),
     // State of solution
     options,
     baseCompilerOptions,
```

## Packed evidence

To keep the change inside Loom's file bound, four folders of raw per-item results are packed, each as `<folder>.tar.gz` with `<folder>.sha256` listing every packed file's SHA-256 as it was committed: `oracle/shards`, `owned/oracle/shards`, `paired-replay` and `owned/paired-replay`. Links above into those folders resolve inside the archives. Summaries, union records and whole-run logs stay unpacked.
