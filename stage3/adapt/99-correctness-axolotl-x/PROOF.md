# 99x proof: stopped at the first unequal replay outcome

The final check 1 condition compares replay outcomes only: files written,
text edits and exact refusal text must match between main and main plus 99x,
apart from 99x itself. Shared outcomes are recorded baselines regardless of
which file or owner they name. No file-name or owner-name guard remains.

The first 44 adapter pairs match. Pair 45, 95-checker-axolotl-h, differs.
The replay stopped there and no later adapter or shard piece was started.
No other adaptation directory was edited. The branch is not ready to land.

## Results

| Check | Result |
| --- | --- |
| 1 full baseline and after applies, --build-product | Both exit 0 |
| 1 paired replay | First difference at 95-checker-axolotl-h; 44 earlier outcomes match |
| 99x owner contract | Six malformed/partial states refused; unrelated edit accepted |
| 99x direct replay | Zero writes and edits; all 79 roots preserve bytes and nanosecond mtimes |
| CRLF contract | Pass |
| 2 adamic types, all 79 roots together | Exit 0, zero diagnostics |
| 3 all nine emitted JavaScript products | Pass; seven byte-identical, two contain exactly 46 rewritten sites |
| 4 unchanged whole lane | PASS: 106,366 passing, one sanctioned API failure, zero pending |
| 4 standard shards | First 40 of 398 pass; remaining 358 not run after check 1 differs |
| 5 restored-site negative control | Build exits 0; comparison exits 1 and reports exactly one missed site |
| 6 ordinary census | All 79 first-stop records unchanged; 73 stop at core.ts:810:89 |
| 6 CW051 latent census | All six units measured before and after, in 12 bounded pieces |

Rewritten: 46. Left as written: zero. Six unused import specifiers are removed.
No cast is added. core.ts and binder.ts source remain byte-identical to main.

## Exact check 1 difference

Both sides exit 1 with zero files written and zero edits. Their complete
refusal text differs because 95h reaches different writer-owner pins:

| Tree | File | Refused owner |
| --- | --- | --- |
| Before, main | src/compiler/program.ts | createProgram |
| After, main plus 99x | src/compiler/checker.ts | createBasicNodeBuilderModuleSpecifierResolutionHost |

95h asserts `writer owner shape` at adapt.cjs:35. On main, the checker-host
writer matches its pinned SHA-256 9171aad5d35e92052c682093a2f2b7fb0ad49a63787488f78d0aa74161d9b9b1.
After 99x, that same writer hashes to
48bed054dbe79b4c7ed5441561ca9eda360dd0892c7d3414c6bce35a1409f2a3.
That function is one of 99x's 15 owners and contains its two checker rewrites.

Main gets past that writer, then hits an existing createProgram mismatch:
actual 729a032f2ab8b7149432d2725cdc921422210dd6850b6cf450999fae56863854,
expected 505bb77211eb4818ad15ed0701812c6c51eb06059becdd73923a5f3e73b69152.
This is not the earlier shared baseline outcome. The first refusal changes,
and the exact-text records differ under the user's final condition.

The complete paired records are paired-replay/before.json and after.json.
Their nonempty records.diff contains the exact differing assertion payloads.
Every executed adapter has separate compressed logs for both trees.
final-replay-stop.json maps both refusal hashes to their precise owners;
replay-checker-owner-hashes.json independently shows the before checker writer
matches 95h's pin and the after writer does not. 40's builder.ts brand refusal,
30's addRange refusal and all other matching failures remain baseline records.
Historical stop files from superseded rules are retained as audit history.

## Provenance and reproduction

Base: unmodified main 3e395f9e5a7038d859ec8930be72cde68d8be3ff.
Upstream: TypeScript 6.0.3, 050880ce59e30b356b686bd3144efe24f875ebc8.
99w's adapt.cjs, README.md and PROOF.md were read from
06dad75d0bb482b47046370a60341c054899afed; 99k's adapter and documentation were
read from main. This branch was never based on 99w. No Apple compiler overlay
or compiler implementation change was used.

After sourcing /workspace/adamic-tools/env.sh, cloud/setup.sh completed and
npm ci installed stage3/api's pinned packages. Setup does not supply the
upstream source checkout, so the authorized Microsoft TypeScript pin was
cloned into /workspace/typescript-source. Both full applies used --build-product:

```sh
# From an isolated main worktree at /workspace/bind-main:
ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source stage3/apply.sh --build-product /workspace/bind-before
# From this branch, after generating the owner pins:
ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source stage3/apply.sh --build-product /workspace/bind-after
NODE_PATH=stage3/api/node_modules node stage3/adapt/99-correctness-axolotl-x/contract.cjs /workspace/bind-before
```

Full logs are baseline-apply.log.gz, after-apply.log.gz, setup.log.gz,
npm-api.log.gz, generate.log.gz and contract.log.gz. self-replay.log.gz records
all 79 byte/mtime checks on the actual completed after tree. owners.json pins
15 exact before/after function texts and their SHA-256 values; no whole-file
hash gates adaptation. source-delta.json confirms exactly six changed source
files. Evidence additions change product keys without changing the frozen
transformation. All commits touch only this directory; every push is plain
to typescript/maybebind-adapt.

The revised paired run uses independent copies to preserve both measured trees:

```sh
python3 stage3/adapt/99-correctness-axolotl-x/paired_replay.py /workspace/adamic /workspace/bind-replay-before /workspace/bind-replay-after /workspace/scratch/bind-paired-replay
```

It resumed after eight already-matching records with --resume when the user
removed the old file-name guard. It records actual writes, including writes
with unchanged bytes, and exact edits/refusals. Only root paths are normalized.
The replay was paused during the whole lane's CPU-sensitive test run, then
resumed and reached the first actual outcome difference reported above.

## Check 2

The unmodified main Adamic binary receives all 79 sorted src/compiler .ts
paths in one invocation. It exits 0 in 3.650 seconds with zero diagnostics.
The full type listing is types-after.log.gz; exact arguments and timing are
in types-after.json. No site required fallback or a cast.

## Check 3 and negative control

Both ordinary `npm run build:compiler` builds exit 0 using the same upstream
lockfile install. Public and internal declaration files are byte-identical;
javascript/declarations.json records their hashes. Seven of the nine normal
JavaScript files are byte-identical. _tsc.js and typescript.js each contain
exactly 46 corresponding compiler-site rewrites. Existing services calls
outside src/compiler remain unchanged.

Upstream's normal esbuild target expands the new optional bindings into
conditionals and generated temporaries. _tsc.js eliminates the now-unused
core helper, which also removes its naming collision with binder's local
helper. These are automatic emitter consequences of the 46 source rewrites
and import cleanup, and are visible in the raw hunks below.

compare.cjs parses the complete products with stock TypeScript 6.0.3. It matches
only the inventoried owner/member pairs and verifies receiver/temporary symbol
identity. It reconstructs those 46 expressions, normalizes generated temporary
names and their bookkeeping declarations, parentheses around the reconstructed
call, the unused _tsc helper, and binder's generated helper name. The remaining
complete ASTs must be identical. Both changed files pass with 46 expected
and 46 bound sites each; no other AST difference remains. Complete hashes,
comparison results, ordinary build logs and raw diffs are under javascript/.

The negative fixture restores one builder.ts call in
createBuilderProgram.emitBuildInfo, with its required import. Its prepared
patch exactly matches the actual source delta. The normal build exits 0.
The same emitted comparison exits 1 and reports exactly one unique original
site, host.writeFile in emitBuildInfo. Each changed bundle has 45 bound forms
and that one restored original form; seven files remain byte-identical.
The exact report, full logs and normalized negative diffs are retained with
the restored-site fixtures.

## Check 4, completed pieces and stop

A frozen worktree at 21ad49c10 keeps oracle input hashing stable as evidence
commits are published. Its adapt.cjs, engine.cjs, owners.json and inventory.json
bytes match this branch, and all 746 src files in its artifact match the
checked after tree. oracle/source-identity.json records these assertions.
The unchanged standard prepare_shards.py producer exits 0 in 804.7 seconds;
apply, locked install, build and harness phases all exit 0.

The unchanged whole lane command is:

```sh
stage3/lane/run.sh /workspace/scratch/bind-lane-whole --artifact ARTIFACT
```

It exits 0 with PASS in 365.7 seconds. All runners, eight workers, original
40-second test timeouts and sanctioned API lines are unchanged. The exact
result is 106,366 passing, one sanctioned public-API failure and zero pending.
The complete task observations, execution metadata, report, verdict, raw API
diff and logs are retained under oracle/whole/.

The first bounded shard piece executes 40 unchanged standard manifest units
and all 40 pass in 98.4 seconds. Shards use main's supported
STAGE3_ADAPTED_TREE product path with a symlink-only view of that same artifact
and its exact locked dependency installation. Standard shard.py performs its
normal identity, count, timeout and API checks. No test or expected result is
changed. Each unit's reports, observations and full logs are retained under
oracle/shards/; piece metrics and driver output are under oracle/pieces/0/.
Successful redundant scratch views are discarded; all observations and logs
are retained. The first piece includes the sanctioned API unit.

The check 1 difference was observed after this completed piece. The remaining
358 shards and the complete shard-to-whole union comparison were not run.
Check 4 is therefore incomplete, despite the passing whole lane and first
40 shards. Producer, whole and shard pieces each have a 1440-second outer
bound and each completed in less than 25 minutes.

## Check 6 ordinary census

Both ordinary runs complete. All 79 single-root TypeScript records are
present, and every normalized first-stop record is identical. There are zero
changed first stops. Seventy-three stop at core.ts:810:89 with Refused: a cast
the runtime can't check. The other six first stops are unchanged as well.
Their exact kinds, locations and reasons are in census/comparison.json and
the complete before/after summaries. Raw streams and logs are compressed
under census/. Non-source JSON rows and the combined-program row emitted by
the tool are retained but excluded from the 79-root comparison.

## CW051 before and after

The six unit identities come from core-walls-sort at
ad29102e9c145c5b480629227574d2ffa235a4b1. The unit selector comes from
container-fit-census at 654092b95416580c3e5be1f968fb8b14755b4481.
The census-only overlay uses main's existing latent harness; its two added
hooks select scheduled files and top-level units without changing registration,
dependency bodies or first-error behavior. Full/speculative modes are disabled.
All 12 bounded pieces exit 0, each contains exactly one requested attempted
unit, and every run reports checker_rejected false. File-wide refusal_scan
findings are retained separately from the selected unit's lowering stop.
The maximum piece wall time is 393.4 seconds, below 25 minutes.

Historical core-walls locations differ from current main. The helper is now
at line 1533, and the checker-host unit stops earlier on a method-presence
boundary. These are the current measured stops, all NotYet:

| Unit | Before | After |
| --- | --- | --- |
| createBasicNodeBuilderModuleSpecifierResolutionHost | src/compiler/checker.ts:54279:37 | src/compiler/checker.ts:54278:37 |
| createWatchHost | src/compiler/core.ts:1533:12 | src/compiler/watch.ts:676:20 |
| createIncrementalCompilerHost | src/compiler/core.ts:1533:12 | src/compiler/watchPublic.ts:124:23 |
| performBuild | src/compiler/core.ts:1533:12 | src/compiler/watch.ts:855:19 |
| createWatchOfConfigFile | src/compiler/core.ts:1533:12 | src/compiler/watch.ts:855:19 |
| createWatchOfFilesAndCompilerOptions | src/compiler/core.ts:1533:12 | src/compiler/watch.ts:855:19 |

The checker-host reason on both sides is a presence test of a method whose
presence the receiver's construction does not prove. All other rows report
a call through ?. (an optional call), moving from core's helper to the direct
optional-binding expression. No native-lowering completion is claimed.
latent/comparison.json retains exact reasons, original unit identifiers and
timings. Every piece's selection, raw records, log and timers is retained
under latent/before and latent/after. build_latent.py reproduces the existing
measurement overlay plus the scheduling-only selector; all added Go sources
are named .go.txt. No background task or suspended replay remains.

## All emitted JavaScript diff hunks

Seven files have no hunks. The complete raw differences in the other two are:

### _tsc.js

```diff
--- before/_tsc.js
+++ after/_tsc.js
@@ -793,9 +793,6 @@
       first2[id] = second[id];
     }
   }
-}
-function maybeBind(obj, fn) {
-  return fn == null ? void 0 : fn.bind(obj);
 }
 function createMultiMap() {
   const map2 = /* @__PURE__ */ new Map();
@@ -43662,7 +43659,7 @@
     }
     function onLeft(left, state, node) {
       if (!state.skip) {
-        const maybeBound = maybeBind2(left);
+        const maybeBound = maybeBind(left);
         if (node.operatorToken.kind === 28 /* CommaToken */) {
           maybeBindExpressionFlowIfCall(left);
         }
@@ -43676,7 +43673,7 @@
     }
     function onRight(right, state, node) {
       if (!state.skip) {
-        const maybeBound = maybeBind2(right);
+        const maybeBound = maybeBind(right);
         if (node.operatorToken.kind === 28 /* CommaToken */) {
           maybeBindExpressionFlowIfCall(right);
         }
@@ -43707,7 +43704,7 @@
       state.skip = false;
       state.stackIndex--;
     }
-    function maybeBind2(node) {
+    function maybeBind(node) {
       if (node && isBinaryExpression(node) && !isDestructuringAssignment(node)) {
         return node;
       }
@@ -91038,13 +91035,14 @@
   return !!(s.flags & 2 /* HasLiteralTypes */);
 }
 function createBasicNodeBuilderModuleSpecifierResolutionHost(host) {
+  var _a, _b;
   return {
     getCommonSourceDirectory: !!host.getCommonSourceDirectory ? () => host.getCommonSourceDirectory() : () => "",
     getCurrentDirectory: () => host.getCurrentDirectory(),
-    getSymlinkCache: maybeBind(host, host.getSymlinkCache),
+    getSymlinkCache: (_a = host.getSymlinkCache) == null ? void 0 : _a.bind(host),
     getPackageJsonInfoCache: () => {
-      var _a;
-      return (_a = host.getPackageJsonInfoCache) == null ? void 0 : _a.call(host);
+      var _a2;
+      return (_a2 = host.getPackageJsonInfoCache) == null ? void 0 : _a2.call(host);
     },
     useCaseSensitiveFileNames: () => host.useCaseSensitiveFileNames(),
     redirectTargetsMap: host.redirectTargetsMap,
@@ -91055,7 +91053,7 @@
     readFile: host.readFile ? ((fileName) => host.readFile(fileName)) : void 0,
     getDefaultResolutionModeForFile: (file) => host.getDefaultResolutionModeForFile(file),
     getModeForResolutionAtIndex: (file, index) => host.getModeForResolutionAtIndex(file, index),
-    getGlobalTypingsCacheLocation: maybeBind(host, host.getGlobalTypingsCacheLocation)
+    getGlobalTypingsCacheLocation: (_b = host.getGlobalTypingsCacheLocation) == null ? void 0 : _b.bind(host)
   };
 }
 var SymbolTrackerImpl = class _SymbolTrackerImpl {
@@ -122090,6 +122088,7 @@
   };
 }
 function createCompilerHostWorker(options, setParentNodes, system = sys) {
+  var _a;
   const existingDirectories = /* @__PURE__ */ new Map();
   const getCanonicalFileName = createGetCanonicalFileName(system.useCaseSensitiveFileNames);
   function directoryExists(directoryPath) {
@@ -122129,7 +122128,7 @@
     realpath,
     readDirectory: (path, extensions, include, exclude, depth) => system.readDirectory(path, extensions, include, exclude, depth),
     createDirectory: (d) => system.createDirectory(d),
-    createHash: maybeBind(system, system.createHash)
+    createHash: (_a = system.createHash) == null ? void 0 : _a.bind(system)
   };
   return compilerHost;
 }
@@ -122734,7 +122733,7 @@
   };
 }
 function createProgram(_rootNamesOrOptions, _options, _host, _oldProgram, _configFileParsingDiagnostics) {
-  var _a, _b, _c, _d, _e, _f, _g, _h, _i, _j, _k, _l, _m, _n, _o, _p;
+  var _a, _b, _c, _d, _e, _f, _g, _h, _i, _j, _k, _l, _m, _n, _o, _p, _q;
   let _createProgramOptions = isArray(_rootNamesOrOptions) ? createCreateProgramOptions(_rootNamesOrOptions, _options, _host, _oldProgram, _configFileParsingDiagnostics) : _rootNamesOrOptions;
   const { rootNames, options, configFileParsingDiagnostics, projectReferences, typeScriptVersion: typeScriptVersion2, host: createProgramOptionsHost } = _createProgramOptions;
   let { oldProgram } = _createProgramOptions;
@@ -123108,7 +123107,7 @@
     getFileIncludeReasons: () => programDiagnostics.getFileReasons(),
     structureIsReused,
     writeFile: writeFile2,
-    getGlobalTypingsCacheLocation: maybeBind(host, host.getGlobalTypingsCacheLocation)
+    getGlobalTypingsCacheLocation: (_p = host.getGlobalTypingsCacheLocation) == null ? void 0 : _p.bind(host)
   };
   onProgramCreateComplete();
   if (!skipVerifyCompilerOptions) {
@@ -123116,7 +123115,7 @@
   }
   mark("afterProgram");
   measure("Program", "beforeProgram", "afterProgram");
-  (_p = tracing) == null ? void 0 : _p.pop();
+  (_q = tracing) == null ? void 0 : _q.pop();
   return program;
   function getResolvedModule(file, moduleName, mode) {
     var _a2;
@@ -123591,6 +123590,7 @@
     return 2 /* Completely */;
   }
   function getEmitHost(writeFileCallback) {
+    var _a2, _b2, _c2;
     return {
       getCanonicalFileName,
       getCommonSourceDirectory: program.getCommonSourceDirectory,
@@ -123616,18 +123616,18 @@
         if (missingFileNames.has(path)) return false;
         return host.fileExists(f);
       },
-      realpath: maybeBind(host, host.realpath),
+      realpath: (_a2 = host.realpath) == null ? void 0 : _a2.bind(host),
       useCaseSensitiveFileNames: () => host.useCaseSensitiveFileNames(),
       getBuildInfo: () => {
-        var _a2;
-        return (_a2 = program.getBuildInfo) == null ? void 0 : _a2.call(program);
+        var _a3;
+        return (_a3 = program.getBuildInfo) == null ? void 0 : _a3.call(program);
       },
       getSourceFileFromReference: (file, ref) => program.getSourceFileFromReference(file, ref),
       redirectTargetsMap,
       getFileIncludeReasons: program.getFileIncludeReasons,
-      createHash: maybeBind(host, host.createHash),
+      createHash: (_b2 = host.createHash) == null ? void 0 : _b2.bind(host),
       getModuleResolutionCache: () => program.getModuleResolutionCache(),
-      trace: maybeBind(host, host.trace),
+      trace: (_c2 = host.trace) == null ? void 0 : _c2.bind(host),
       getGlobalTypingsCacheLocation: program.getGlobalTypingsCacheLocation
     };
   }
@@ -125776,6 +125776,7 @@
   return filter(diagnostic, (d) => !d.skippedOn || !option[d.skippedOn]);
 }
 function parseConfigHostFromCompilerHostLike(host, directoryStructureHost = host) {
+  var _a, _b, _c;
   return {
     fileExists: (f) => directoryStructureHost.fileExists(f),
     readDirectory(root, extensions, excludes, includes, depth) {
@@ -125783,9 +125784,9 @@
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
@@ -127473,10 +127474,11 @@
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
@@ -127485,7 +127487,7 @@
     return emitSkippedWithNoDiagnostics;
   }
   function emitNextAffectedFileOrDtsErrors(writeFile2, cancellationToken, emitOnlyDtsFiles, customTransformers, isForDtsErrors) {
-    var _a, _b, _c, _d;
+    var _a, _b, _c, _d, _e;
     Debug.assert(isBuilderProgramStateWithDefinedProgram(state));
     let affected = getNextAffectedFile(state, cancellationToken, host);
     const programEmitKind = getBuilderFileEmit(state.compilerOptions);
@@ -127542,7 +127544,7 @@
         if (isForDtsErrors || !getBuildInfoEmitPending(state)) return void 0;
         const affected2 = state.program;
         const result2 = affected2.emitBuildInfo(
-          writeFile2 || maybeBind(host, host.writeFile),
+          writeFile2 || ((_b = host.writeFile) == null ? void 0 : _b.bind(host)),
           cancellationToken
         );
         state.buildInfoEmitPending = false;
@@ -127574,12 +127576,12 @@
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
@@ -127611,10 +127613,11 @@
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
@@ -127630,7 +127633,7 @@
                 host,
                 data
               );
-              if (!((_a = data == null ? void 0 : data.diagnostics) == null ? void 0 : _a.length)) emitSignature = signature;
+              if (!((_a2 = data == null ? void 0 : data.diagnostics) == null ? void 0 : _a2.length)) emitSignature = signature;
               if (signature !== file.version) {
                 if (host.storeSignatureInfo) (state.signatureInfo ?? (state.signatureInfo = /* @__PURE__ */ new Map())).set(file.resolvedPath, 1 /* StoredSignatureAtEmit */);
                 if (state.affectedFiles) {
@@ -129615,13 +129618,14 @@
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
@@ -129656,6 +129660,7 @@
   return result;
 }
 function createCompilerHostFromProgramHost(host, getCompilerOptions, directoryStructureHost = host) {
+  var _a, _b, _c, _d, _e, _f, _g, _h;
   const useCaseSensitiveFileNames2 = host.useCaseSensitiveFileNames();
   const compilerHost = {
     getSourceFile: createGetSourceFile(
@@ -129663,7 +129668,7 @@
       /*setParentNodes*/
       void 0
     ),
-    getDefaultLibLocation: maybeBind(host, host.getDefaultLibLocation),
+    getDefaultLibLocation: (_a = host.getDefaultLibLocation) == null ? void 0 : _a.bind(host),
     getDefaultLibFileName: (options) => host.getDefaultLibFileName(options),
     writeFile: createWriteFileMeasuringIO(
       (path, data, writeByteOrderMark) => host.writeFile(path, data, writeByteOrderMark),
@@ -129676,13 +129681,13 @@
     getNewLine: () => getNewLineCharacter(getCompilerOptions()),
     fileExists: (f) => host.fileExists(f),
     readFile: (f) => host.readFile(f),
-    trace: maybeBind(host, host.trace),
-    directoryExists: maybeBind(directoryStructureHost, directoryStructureHost.directoryExists),
-    getDirectories: maybeBind(directoryStructureHost, directoryStructureHost.getDirectories),
-    realpath: maybeBind(host, host.realpath),
-    getEnvironmentVariable: maybeBind(host, host.getEnvironmentVariable) || (() => ""),
-    createHash: maybeBind(host, host.createHash),
-    readDirectory: maybeBind(host, host.readDirectory),
+    trace: (_b = host.trace) == null ? void 0 : _b.bind(host),
+    directoryExists: (_c = directoryStructureHost.directoryExists) == null ? void 0 : _c.bind(directoryStructureHost),
+    getDirectories: (_d = directoryStructureHost.getDirectories) == null ? void 0 : _d.bind(directoryStructureHost),
+    realpath: (_e = host.realpath) == null ? void 0 : _e.bind(host),
+    getEnvironmentVariable: ((_f = host.getEnvironmentVariable) == null ? void 0 : _f.bind(host)) || (() => ""),
+    createHash: (_g = host.createHash) == null ? void 0 : _g.bind(host),
+    readDirectory: (_h = host.readDirectory) == null ? void 0 : _h.bind(host),
     storeSignatureInfo: host.storeSignatureInfo,
     jsDocParsingMode: host.jsDocParsingMode
   };
@@ -129732,6 +129737,7 @@
   };
 }
 function createProgramHost(system, createProgram2) {
+  var _a, _b, _c, _d;
   const getDefaultLibLocation = memoize(() => getDirectoryPath(normalizePath(system.getExecutingFilePath())));
   return {
     useCaseSensitiveFileNames: () => system.useCaseSensitiveFileNames,
@@ -129744,15 +129750,15 @@
     directoryExists: (path) => system.directoryExists(path),
     getDirectories: (path) => system.getDirectories(path),
     readDirectory: (path, extensions, exclude, include, depth) => system.readDirectory(path, extensions, exclude, include, depth),
-    realpath: maybeBind(system, system.realpath),
-    getEnvironmentVariable: maybeBind(system, system.getEnvironmentVariable),
+    realpath: (_a = system.realpath) == null ? void 0 : _a.bind(system),
+    getEnvironmentVariable: (_b = system.getEnvironmentVariable) == null ? void 0 : _b.bind(system),
     trace: (s) => system.write(s + system.newLine),
     createDirectory: (path) => system.createDirectory(path),
     writeFile: (path, data, writeByteOrderMark) => system.writeFile(path, data, writeByteOrderMark),
-    createHash: maybeBind(system, system.createHash),
+    createHash: (_c = system.createHash) == null ? void 0 : _c.bind(system),
     createProgram: createProgram2 || createEmitAndSemanticDiagnosticsBuilderProgram,
     storeSignatureInfo: system.storeSignatureInfo,
-    now: maybeBind(system, system.now)
+    now: (_d = system.now) == null ? void 0 : _d.bind(system)
   };
 }
 function createWatchCompilerHost(system = sys, createProgram2, reportDiagnostic, reportWatchStatus2) {
@@ -129846,13 +129852,14 @@
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
@@ -129872,6 +129879,7 @@
   return createProgram2(rootNames, options, host, oldProgram, configFileParsingDiagnostics, projectReferences);
 }
 function createWatchProgram(host) {
+  var _a, _b, _c, _d, _e, _f, _g, _h;
   let builderProgram;
   let updateLevel;
   let missingFilesMap;
@@ -129929,7 +129937,7 @@
   compilerHost.onReleaseParsedCommandLine = onReleaseParsedCommandLine;
   compilerHost.toPath = toPath3;
   compilerHost.getCompilationSettings = () => compilerOptions;
-  compilerHost.useSourceOfProjectReferenceRedirect = maybeBind(host, host.useSourceOfProjectReferenceRedirect);
+  compilerHost.useSourceOfProjectReferenceRedirect = (_a = host.useSourceOfProjectReferenceRedirect) == null ? void 0 : _a.bind(host);
   compilerHost.preferNonRecursiveWatch = host.preferNonRecursiveWatch;
   compilerHost.watchDirectoryOfFailedLookupLocation = (dir, cb, flags) => watchDirectory(dir, cb, flags, watchOptions, WatchType.FailedLookupLocations);
   compilerHost.watchAffectingFileLocation = (file, cb) => watchFile3(file, cb, 2e3 /* High */, watchOptions, WatchType.AffectingFileLocation);
@@ -129948,21 +129956,21 @@
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
@@ -129995,8 +130003,8 @@
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
@@ -130360,13 +130368,13 @@
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
@@ -130453,10 +130461,10 @@
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
@@ -130477,7 +130485,7 @@
     );
   }
   function watchReferencedProject(configFileName2, configPath, commandLine) {
-    var _a, _b, _c, _d;
+    var _a2, _b2, _c2, _d2;
     commandLine.watcher || (commandLine.watcher = watchFile3(
       configFileName2,
       (_fileName, eventKind) => {
@@ -130488,14 +130496,14 @@
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
@@ -130524,15 +130532,15 @@
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
@@ -130577,13 +130585,14 @@
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
@@ -130625,18 +130634,19 @@
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
@@ -130701,7 +130711,7 @@
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
@@ -95665,13 +95665,14 @@
   return !!(s.flags & 2 /* HasLiteralTypes */);
 }
 function createBasicNodeBuilderModuleSpecifierResolutionHost(host) {
+  var _a, _b;
   return {
     getCommonSourceDirectory: !!host.getCommonSourceDirectory ? () => host.getCommonSourceDirectory() : () => "",
     getCurrentDirectory: () => host.getCurrentDirectory(),
-    getSymlinkCache: maybeBind(host, host.getSymlinkCache),
+    getSymlinkCache: (_a = host.getSymlinkCache) == null ? void 0 : _a.bind(host),
     getPackageJsonInfoCache: () => {
-      var _a;
-      return (_a = host.getPackageJsonInfoCache) == null ? void 0 : _a.call(host);
+      var _a2;
+      return (_a2 = host.getPackageJsonInfoCache) == null ? void 0 : _a2.call(host);
     },
     useCaseSensitiveFileNames: () => host.useCaseSensitiveFileNames(),
     redirectTargetsMap: host.redirectTargetsMap,
@@ -95682,7 +95683,7 @@
     readFile: host.readFile ? ((fileName) => host.readFile(fileName)) : void 0,
     getDefaultResolutionModeForFile: (file) => host.getDefaultResolutionModeForFile(file),
     getModeForResolutionAtIndex: (file, index) => host.getModeForResolutionAtIndex(file, index),
-    getGlobalTypingsCacheLocation: maybeBind(host, host.getGlobalTypingsCacheLocation)
+    getGlobalTypingsCacheLocation: (_b = host.getGlobalTypingsCacheLocation) == null ? void 0 : _b.bind(host)
   };
 }
 var SymbolTrackerImpl = class _SymbolTrackerImpl {
@@ -126922,6 +126923,7 @@
   };
 }
 function createCompilerHostWorker(options, setParentNodes, system = sys) {
+  var _a;
   const existingDirectories = /* @__PURE__ */ new Map();
   const getCanonicalFileName = createGetCanonicalFileName(system.useCaseSensitiveFileNames);
   function directoryExists(directoryPath) {
@@ -126961,7 +126963,7 @@
     realpath,
     readDirectory: (path, extensions, include, exclude, depth) => system.readDirectory(path, extensions, include, exclude, depth),
     createDirectory: (d) => system.createDirectory(d),
-    createHash: maybeBind(system, system.createHash)
+    createHash: (_a = system.createHash) == null ? void 0 : _a.bind(system)
   };
   return compilerHost;
 }
@@ -127600,7 +127602,7 @@
   };
 }
 function createProgram(_rootNamesOrOptions, _options, _host, _oldProgram, _configFileParsingDiagnostics) {
-  var _a, _b, _c, _d, _e, _f, _g, _h, _i, _j, _k, _l, _m, _n, _o, _p;
+  var _a, _b, _c, _d, _e, _f, _g, _h, _i, _j, _k, _l, _m, _n, _o, _p, _q;
   let _createProgramOptions = isArray(_rootNamesOrOptions) ? createCreateProgramOptions(_rootNamesOrOptions, _options, _host, _oldProgram, _configFileParsingDiagnostics) : _rootNamesOrOptions;
   const { rootNames, options, configFileParsingDiagnostics, projectReferences, typeScriptVersion: typeScriptVersion3, host: createProgramOptionsHost } = _createProgramOptions;
   let { oldProgram } = _createProgramOptions;
@@ -127974,7 +127976,7 @@
     getFileIncludeReasons: () => programDiagnostics.getFileReasons(),
     structureIsReused,
     writeFile: writeFile2,
-    getGlobalTypingsCacheLocation: maybeBind(host, host.getGlobalTypingsCacheLocation)
+    getGlobalTypingsCacheLocation: (_p = host.getGlobalTypingsCacheLocation) == null ? void 0 : _p.bind(host)
   };
   onProgramCreateComplete();
   if (!skipVerifyCompilerOptions) {
@@ -127982,7 +127984,7 @@
   }
   mark("afterProgram");
   measure("Program", "beforeProgram", "afterProgram");
-  (_p = tracing) == null ? void 0 : _p.pop();
+  (_q = tracing) == null ? void 0 : _q.pop();
   return program;
   function getResolvedModule(file, moduleName, mode) {
     var _a2;
@@ -128457,6 +128459,7 @@
     return 2 /* Completely */;
   }
   function getEmitHost(writeFileCallback) {
+    var _a2, _b2, _c2;
     return {
       getCanonicalFileName,
       getCommonSourceDirectory: program.getCommonSourceDirectory,
@@ -128482,18 +128485,18 @@
         if (missingFileNames.has(path)) return false;
         return host.fileExists(f);
       },
-      realpath: maybeBind(host, host.realpath),
+      realpath: (_a2 = host.realpath) == null ? void 0 : _a2.bind(host),
       useCaseSensitiveFileNames: () => host.useCaseSensitiveFileNames(),
       getBuildInfo: () => {
-        var _a2;
-        return (_a2 = program.getBuildInfo) == null ? void 0 : _a2.call(program);
+        var _a3;
+        return (_a3 = program.getBuildInfo) == null ? void 0 : _a3.call(program);
       },
       getSourceFileFromReference: (file, ref) => program.getSourceFileFromReference(file, ref),
       redirectTargetsMap,
       getFileIncludeReasons: program.getFileIncludeReasons,
-      createHash: maybeBind(host, host.createHash),
+      createHash: (_b2 = host.createHash) == null ? void 0 : _b2.bind(host),
       getModuleResolutionCache: () => program.getModuleResolutionCache(),
-      trace: maybeBind(host, host.trace),
+      trace: (_c2 = host.trace) == null ? void 0 : _c2.bind(host),
       getGlobalTypingsCacheLocation: program.getGlobalTypingsCacheLocation
     };
   }
@@ -130642,6 +130645,7 @@
   return filter(diagnostic, (d) => !d.skippedOn || !option[d.skippedOn]);
 }
 function parseConfigHostFromCompilerHostLike(host, directoryStructureHost = host) {
+  var _a, _b, _c;
   return {
     fileExists: (f) => directoryStructureHost.fileExists(f),
     readDirectory(root, extensions, excludes, includes, depth) {
@@ -130649,9 +130653,9 @@
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
@@ -132373,10 +132377,11 @@
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
@@ -132385,7 +132390,7 @@
     return emitSkippedWithNoDiagnostics;
   }
   function emitNextAffectedFileOrDtsErrors(writeFile2, cancellationToken, emitOnlyDtsFiles, customTransformers, isForDtsErrors) {
-    var _a, _b, _c, _d;
+    var _a, _b, _c, _d, _e;
     Debug.assert(isBuilderProgramStateWithDefinedProgram(state));
     let affected = getNextAffectedFile(state, cancellationToken, host);
     const programEmitKind = getBuilderFileEmit(state.compilerOptions);
@@ -132442,7 +132447,7 @@
         if (isForDtsErrors || !getBuildInfoEmitPending(state)) return void 0;
         const affected2 = state.program;
         const result2 = affected2.emitBuildInfo(
-          writeFile2 || maybeBind(host, host.writeFile),
+          writeFile2 || ((_b = host.writeFile) == null ? void 0 : _b.bind(host)),
           cancellationToken
         );
         state.buildInfoEmitPending = false;
@@ -132474,12 +132479,12 @@
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
@@ -132511,10 +132516,11 @@
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
@@ -132530,7 +132536,7 @@
                 host,
                 data
               );
-              if (!((_a = data == null ? void 0 : data.diagnostics) == null ? void 0 : _a.length)) emitSignature = signature;
+              if (!((_a2 = data == null ? void 0 : data.diagnostics) == null ? void 0 : _a2.length)) emitSignature = signature;
               if (signature !== file.version) {
                 if (host.storeSignatureInfo) (state.signatureInfo ?? (state.signatureInfo = /* @__PURE__ */ new Map())).set(file.resolvedPath, 1 /* StoredSignatureAtEmit */);
                 if (state.affectedFiles) {
@@ -134542,13 +134548,14 @@
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
@@ -134583,6 +134590,7 @@
   return result;
 }
 function createCompilerHostFromProgramHost(host, getCompilerOptions, directoryStructureHost = host) {
+  var _a, _b, _c, _d, _e, _f, _g, _h;
   const useCaseSensitiveFileNames2 = host.useCaseSensitiveFileNames();
   const compilerHost = {
     getSourceFile: createGetSourceFile(
@@ -134590,7 +134598,7 @@
       /*setParentNodes*/
       void 0
     ),
-    getDefaultLibLocation: maybeBind(host, host.getDefaultLibLocation),
+    getDefaultLibLocation: (_a = host.getDefaultLibLocation) == null ? void 0 : _a.bind(host),
     getDefaultLibFileName: (options) => host.getDefaultLibFileName(options),
     writeFile: createWriteFileMeasuringIO(
       (path, data, writeByteOrderMark) => host.writeFile(path, data, writeByteOrderMark),
@@ -134603,13 +134611,13 @@
     getNewLine: () => getNewLineCharacter(getCompilerOptions()),
     fileExists: (f) => host.fileExists(f),
     readFile: (f) => host.readFile(f),
-    trace: maybeBind(host, host.trace),
-    directoryExists: maybeBind(directoryStructureHost, directoryStructureHost.directoryExists),
-    getDirectories: maybeBind(directoryStructureHost, directoryStructureHost.getDirectories),
-    realpath: maybeBind(host, host.realpath),
-    getEnvironmentVariable: maybeBind(host, host.getEnvironmentVariable) || (() => ""),
-    createHash: maybeBind(host, host.createHash),
-    readDirectory: maybeBind(host, host.readDirectory),
+    trace: (_b = host.trace) == null ? void 0 : _b.bind(host),
+    directoryExists: (_c = directoryStructureHost.directoryExists) == null ? void 0 : _c.bind(directoryStructureHost),
+    getDirectories: (_d = directoryStructureHost.getDirectories) == null ? void 0 : _d.bind(directoryStructureHost),
+    realpath: (_e = host.realpath) == null ? void 0 : _e.bind(host),
+    getEnvironmentVariable: ((_f = host.getEnvironmentVariable) == null ? void 0 : _f.bind(host)) || (() => ""),
+    createHash: (_g = host.createHash) == null ? void 0 : _g.bind(host),
+    readDirectory: (_h = host.readDirectory) == null ? void 0 : _h.bind(host),
     storeSignatureInfo: host.storeSignatureInfo,
     jsDocParsingMode: host.jsDocParsingMode
   };
@@ -134659,6 +134667,7 @@
   };
 }
 function createProgramHost(system, createProgram2) {
+  var _a, _b, _c, _d;
   const getDefaultLibLocation = memoize(() => getDirectoryPath(normalizePath(system.getExecutingFilePath())));
   return {
     useCaseSensitiveFileNames: () => system.useCaseSensitiveFileNames,
@@ -134671,15 +134680,15 @@
     directoryExists: (path) => system.directoryExists(path),
     getDirectories: (path) => system.getDirectories(path),
     readDirectory: (path, extensions, exclude, include, depth) => system.readDirectory(path, extensions, exclude, include, depth),
-    realpath: maybeBind(system, system.realpath),
-    getEnvironmentVariable: maybeBind(system, system.getEnvironmentVariable),
+    realpath: (_a = system.realpath) == null ? void 0 : _a.bind(system),
+    getEnvironmentVariable: (_b = system.getEnvironmentVariable) == null ? void 0 : _b.bind(system),
     trace: (s) => system.write(s + system.newLine),
     createDirectory: (path) => system.createDirectory(path),
     writeFile: (path, data, writeByteOrderMark) => system.writeFile(path, data, writeByteOrderMark),
-    createHash: maybeBind(system, system.createHash),
+    createHash: (_c = system.createHash) == null ? void 0 : _c.bind(system),
     createProgram: createProgram2 || createEmitAndSemanticDiagnosticsBuilderProgram,
     storeSignatureInfo: system.storeSignatureInfo,
-    now: maybeBind(system, system.now)
+    now: (_d = system.now) == null ? void 0 : _d.bind(system)
   };
 }
 function createWatchCompilerHost(system = sys, createProgram2, reportDiagnostic, reportWatchStatus2) {
@@ -134773,13 +134782,14 @@
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
@@ -134824,6 +134834,7 @@
   }
 }
 function createWatchProgram(host) {
+  var _a, _b, _c, _d, _e, _f, _g, _h;
   let builderProgram;
   let updateLevel;
   let missingFilesMap;
@@ -134881,7 +134892,7 @@
   compilerHost.onReleaseParsedCommandLine = onReleaseParsedCommandLine;
   compilerHost.toPath = toPath3;
   compilerHost.getCompilationSettings = () => compilerOptions;
-  compilerHost.useSourceOfProjectReferenceRedirect = maybeBind(host, host.useSourceOfProjectReferenceRedirect);
+  compilerHost.useSourceOfProjectReferenceRedirect = (_a = host.useSourceOfProjectReferenceRedirect) == null ? void 0 : _a.bind(host);
   compilerHost.preferNonRecursiveWatch = host.preferNonRecursiveWatch;
   compilerHost.watchDirectoryOfFailedLookupLocation = (dir, cb, flags) => watchDirectory(dir, cb, flags, watchOptions, WatchType.FailedLookupLocations);
   compilerHost.watchAffectingFileLocation = (file, cb) => watchFile3(file, cb, 2e3 /* High */, watchOptions, WatchType.AffectingFileLocation);
@@ -134900,21 +134911,21 @@
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
@@ -134947,8 +134958,8 @@
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
@@ -135312,13 +135323,13 @@
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
@@ -135405,10 +135416,10 @@
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
@@ -135429,7 +135440,7 @@
     );
   }
   function watchReferencedProject(configFileName2, configPath, commandLine) {
-    var _a, _b, _c, _d;
+    var _a2, _b2, _c2, _d2;
     commandLine.watcher || (commandLine.watcher = watchFile3(
       configFileName2,
       (_fileName, eventKind) => {
@@ -135440,14 +135451,14 @@
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
@@ -135476,15 +135487,15 @@
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
@@ -135550,13 +135561,14 @@
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
@@ -135598,18 +135610,19 @@
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
@@ -135674,7 +135687,7 @@
     host,
     hostWithWatch,
     parseConfigFileHost: parseConfigHostFromCompilerHostLike(host),
-    write: maybeBind(host, host.trace),
+    write: (_g = host.trace) == null ? void 0 : _g.bind(host),
     // State of solution
     options,
     baseCompilerOptions,
```
