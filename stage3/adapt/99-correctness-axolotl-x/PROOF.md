# Current continuation: baseline equality is the only replay condition

The user superseded both earlier replay conditions. A shared refusal is a
baseline outcome regardless of file or owner names. 40-explicit-any's builder.ts
brand refusal is a recorded baseline, not a check failure. The paired replay
has resumed after its eight matching outcomes and is comparing every remaining
adapter; only a difference between the trees stops it. Earlier stopping notes
below are historical.

Check 2 has now passed: adamic types received all 79 sorted compiler roots
from /workspace/bind-after in one invocation, exited 0 in 3.650 seconds and
reported zero diagnostics. The complete type listing is types-after.log.gz;
arguments, root count, timing and exit are in types-after.json. All 46 sites
remain rewritten, zero are left as written, and no cast was added.

# 99x proof: revised check 1 stops at the affected-file guard

The user replaced the original zero-refusal replay rule with a paired replay
comparison. The new records are identical through 40-explicit-any, but both
refusals name builder.ts, one of the six files covered by the revised rule's
affected-file condition. That condition fails. The run stopped there; checks
2 through 6 remain unrun. No new 99x regression is established by this result.
The adapter is implemented, but the requested verification remains incomplete.

## Results

| Check | Result |
| --- | --- |
| 1 baseline full apply, --build-product | Pass, exit 0 |
| 1 after full apply, --build-product | Pass, exit 0; 46 rewrites, zero left, six files |
| 99x owner contract | Pass: six refused malformed states; foreign edit accepted |
| 99x replay on actual after tree | Zero files and edits; bytes and nanosecond mtimes preserved across 79 roots |
| CRLF contract | Pass |
| 1 original ordered replay | Superseded by the user |
| 1 revised paired replay comparison | Eight adapters compared, identical records, empty diff |
| 1 revised refusal file/owner guard | Failed at 40-explicit-any: both refusals name builder.ts |
| 2 Adamic types, all 79 roots | Not run: stopped at check 1 |
| 3 nine emitted JavaScript files | Not run: stopped at check 1; no emitted diff hunks |
| 4 upstream oracle and shards | Not run: stopped at check 1 |
| 5 restored-site negative witness | Fixture and patch prepared; detection not run |
| 6 latent and ordinary per-file census | Not run: stopped at check 1 |

## Historical stop under the original rule

The first three adapters, 00-setup, 10-type-imports and
20-optional-declarations, exited 0 with no byte or mtime changes. The next
adapter, 30-indexed-reads, exited 1 with:

```text
declined site changed: core.ts:992:21
```

Its sites.json row names addRange, expression from[i], occurrence 2 of 2,
action decline. On the completed after tree the expression is at core.ts:1002:21
and its parent is from[i]!. The stop is in addRange, which is outside this
adaptation's 15 owners. It is a declined-site guard, not an overlapping owner
pin refusal. No other adaptation directory was edited.

core.ts is byte-identical between the successful main baseline and the after
tree. Its SHA-256 is
3dde48b17a7ea7c9475f28b0dfa1b8c3a407f6b4966c4f6f6c4e6769e4920f3e.
That hash is measurement evidence only, never an adaptation precondition.
replay-stop-detail.json records the complete row, actual AST parent and owner.
ordered-replay/results.json records exits and changed-file lists; per-adapter
logs and the driver log are retained alongside it. The failure made zero writes.

## Commands and provenance

The base is unmodified main 3e395f9e5a7038d859ec8930be72cde68d8be3ff. An isolated worktree at
/workspace/bind-main provided the baseline adaptation set, excluding this
branch's initial evidence-only directory. The branch was never based on 99w.
99w's adapt.cjs, README.md and PROOF.md were read at
06dad75d0bb482b47046370a60341c054899afed. 99k's adapter and documentation were
read from main. No Apple compiler overlay was used.

```sh
source /workspace/adamic-tools/env.sh
bash cloud/setup.sh > /workspace/scratch/bind-setup.log 2>&1
npm ci --prefix stage3/api > /workspace/scratch/bind-npm-api.log 2>&1
# Setup does not supply the upstream source checkout; clone the authorized pin.
git clone --filter=blob:none --no-checkout https://github.com/microsoft/TypeScript.git /workspace/typescript-source
git -C /workspace/typescript-source checkout 050880ce59e30b356b686bd3144efe24f875ebc8
# Run from /workspace/bind-main for the baseline:
ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source stage3/apply.sh --build-product /workspace/bind-before
# Run from this branch for the after tree:
ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source stage3/apply.sh --build-product /workspace/bind-after
NODE_PATH=stage3/api/node_modules node stage3/adapt/99-correctness-axolotl-x/generate.cjs /workspace/bind-before /workspace/typescript-source
NODE_PATH=stage3/api/node_modules node stage3/adapt/99-correctness-axolotl-x/contract.cjs /workspace/bind-before
python3 stage3/adapt/99-correctness-axolotl-x/replay.py /workspace/adamic /workspace/bind-after /workspace/scratch/bind-ordered-replay
```

Pin generation ran after baseline apply and before after apply. The command
list groups apply commands together for readability. All check outputs went
to files. Full apply logs are baseline-apply.log.gz and after-apply.log.gz.
Setup, npm, pin generation, pristine symbol scan, contract and actual-tree
self replay each have their own compressed logs. The self replay checks all
79 compiler source files' SHA-256 and nanosecond mtime before and after calling
adapt on the actual after tree. An unmodified main Adamic and ordinary census
tool were built successfully, but no type or lowering measurement ran.

The initial missing-source stop remains in check1-before.log and
prerequisites.log as historical evidence, superseded by this continuation's
successful full applies. Completed evidence commits change product keys
without changing the frozen transformation. All commits touch only this
directory, and all pushes use typescript/maybebind-adapt without force.

## CW051 before and after

The six unit identifiers below come from typescript/core-walls-sort at
ad29102e9c145c5b480629227574d2ffa235a4b1. That reference records their shared
stop as core.ts:1523:12, a call through ?. (an optional call). Those historical
observations are not new baseline measurements. The container-fit selector
reference was inspected at 654092b95416580c3e5be1f968fb8b14755b4481.

| Recorded unit | New before run | New after run |
| --- | --- | --- |
| src/compiler/checker.ts:54277:1 | Not run | Not run |
| src/compiler/watch.ts:673:1 | Not run | Not run |
| src/compiler/watchPublic.ts:123:1 | Not run | Not run |
| src/compiler/executeCommandLine.ts:812:1 | Not run | Not run |
| src/compiler/executeCommandLine.ts:1011:1 | Not run | Not run |
| src/compiler/executeCommandLine.ts:1034:1 | Not run | Not run |

The ordinary 79-root census also did not run; no changed-first-stop result
is claimed. The negative fixture restores one builder.ts call and its import,
but its JavaScript comparison did not run. No Go file was added. All processes
started for this task have exited; no background work remains.

## Revised check 1: paired replay

The requested replacement condition compares adapter outcomes between main
and main plus 99x, excluding 99x itself, and requires that no refusal name any
of 99x's six source files or its named owners. paired_replay.py implements
both conditions and stops at their first failure.

Independent copies preserve the successful full-apply trees for future work:

```sh
cp -a --reflink=auto /workspace/bind-before /workspace/bind-replay-before
cp -a --reflink=auto /workspace/bind-after /workspace/bind-replay-after
python3 stage3/adapt/99-correctness-axolotl-x/paired_replay.py /workspace/adamic /workspace/bind-replay-before /workspace/bind-replay-after /workspace/scratch/bind-paired-replay > /workspace/scratch/bind-paired-replay.log 2>&1
```

For each adapter, snapshots record the actual files written, including writes
that preserve bytes. Text edits are recorded as affected files and added/removed
line counts; complete patches are retained if any edit occurs. Neither tree
had any writes or edits during these eight adapters. A nonzero exit records
the complete refusal output, with only the tree root normalized to <tree>.
Raw per-adapter logs are retained for both sides. Failures do not stop traversal
unless records differ or the refusal names a covered file or owner.

| Adapter | Before | After | Equal |
| --- | --- | --- | --- |
| 00-setup | Exit 0, zero writes | Exit 0, zero writes | Yes |
| 10-type-imports | Exit 0, zero writes | Exit 0, zero writes | Yes |
| 20-optional-declarations | Exit 0, zero writes | Exit 0, zero writes | Yes |
| 30-indexed-reads | Refuses core.ts:992:21 | Same refusal | Yes |
| 31-indexed-reads-checker | Exit 0, zero writes | Exit 0, zero writes | Yes |
| 32-indexed-reads-program | Refuses appendSourceMap runtime drift | Same refusal | Yes |
| 33-indexed-reads-emit | Refuses sourcemap.ts:215 site drift | Same refusal | Yes |
| 40-explicit-any | Refuses missing builder.ts brands | Same refusal | Yes |

The two records are paired-replay/before.json and paired-replay/after.json.
Their exact diff, paired-replay/records.diff, is empty (zero bytes).
paired-replay/status.json records the failed affected-file condition.
The full shared refusal at 40-explicit-any includes:

```text
Error: missing reviewed brands: src/compiler/builder.ts
```

40's classes.cjs matches brands by key and absolute line. The relevant owners
are IncrementalBuildInfoFileId and IncrementalBuildInfoFileIdListId, neither
of which is one of 99x's named function owners. The brand declarations are
already typed undefined in both trees. Its reviewed lines are 1074 and 1076;
main's actual lines are 1098 and 1100, and after's are 1097 and 1099. The one-line
after shift is the removed maybeBind import. Both trees miss the original
line pins, so this is a shared baseline failure. The rule's file-name condition
still fails because it names builder.ts. No other adaptation was changed.
paired-replay-stop-detail.json retains keys, owners and actual/pinned lines.

The source comparison in source-delta.json confirms that the original complete
src trees differ in exactly the intended six files, with core.ts unchanged.
The revised run did not reach 99x or later adapters. 99x's direct no-write
replay and contracts passed in the earlier completed piece. No types, emitted
JavaScript, oracle, negative detection or census measurements were added in
this continuation. The six CW051 rows above remain unmeasured. All processes
started for this continuation have exited.

## Check 3: all nine ordinary emitted JavaScript products

Both ordinary `npm run build:compiler` builds exit 0 using the same upstream
lockfile install. Seven top-level JavaScript products are byte-identical.
_tsc.js and typescript.js each contain exactly the 46 rewritten compiler
sites. Existing services calls outside src/compiler are preserved.

Upstream targets older JavaScript, so its normal esbuild output expands each
new optional binding into a conditional plus a generated temporary. _tsc.js
also eliminates the unused core maybeBind helper and consequently stops
renaming binder's separate local helper to maybeBind2. These are automatic
emitter consequences of the 46 rewrites and import cleanup. core.ts and
binder.ts source remain byte-identical to main.

compare.cjs parses the complete products with stock TypeScript 6.0.3. It matches
only the 46 inventoried owner/member pairs, verifies receiver and temporary
binding identities, and reconstructs the original call form for comparison.
Only compiler temporary names, parentheses around the reconstructed call,
the unused _tsc helper and binder helper's generated name are normalized.
The remaining complete ASTs must be identical. Both changed bundles pass,
with exactly 46 expected and 46 rewritten sites each; no other AST difference
remains. All nine results and byte hashes are in javascript/comparison.json.
Complete build and comparison logs are retained in javascript/.

All raw emitted diff hunks follow, including generated declarations and names.

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

## Check 5: one restored-site negative control

In a private copy, the first builder.ts call in
createBuilderProgram.emitBuildInfo was restored to maybeBind(host,
host.writeFile), with its required import. The prepared patch exactly matches
the actual source delta. The ordinary compiler build exits 0.

The same nine-product comparison exits 1 and reports exactly one unique site:
createBuilderProgram.emitBuildInfo, receiver host, member writeFile. Each of
the two changed bundles has 45 expected bound forms and exactly that one
remaining original form; the other seven products remain byte-identical.
The normalized comparison diff has one restored expression in each bundle.
This verifies that check 3 detects a missed adaptation site. Full restoration,
build and comparison logs, the exact report and the two negative diffs are
retained with the restored-site fixtures.

## Check 6 latent component: all six CW051 units, both trees

All 12 bounded pieces exit 0. The largest piece takes 393.4 seconds, below
25 minutes. The scheduling selector is copied from the requested container-fit
reference at 654092b95416580c3e5be1f968fb8b14755b4481. The overlay uses unmodified
main's latent measurement harness. Its only extra hooks select scheduled files
and top-level units; registration and dependency bodies remain unchanged.
All runs report checker_rejected false. Each retained piece contains exactly
one requested attempted unit and at most one first-error lowering finding.
File-wide refusal_scan findings are retained but are not unit lowering stops.

The current main baseline differs from the historical core-walls snapshot:
core's helper is now at line 1533, and the checker-host unit stops first on a
method-presence boundary. The measured table below reports current main.
The after locations include the removed import's one-line shift.

| Unit | Before first lowering stop | After first lowering stop |
| --- | --- | --- |
| createBasicNodeBuilderModuleSpecifierResolutionHost | NotYet: src/compiler/checker.ts:54279:37 | NotYet: src/compiler/checker.ts:54278:37 |
| createWatchHost | NotYet: src/compiler/core.ts:1533:12 | NotYet: src/compiler/watch.ts:676:20 |
| createIncrementalCompilerHost | NotYet: src/compiler/core.ts:1533:12 | NotYet: src/compiler/watchPublic.ts:124:23 |
| performBuild | NotYet: src/compiler/core.ts:1533:12 | NotYet: src/compiler/watch.ts:855:19 |
| createWatchOfConfigFile | NotYet: src/compiler/core.ts:1533:12 | NotYet: src/compiler/watch.ts:855:19 |
| createWatchOfFilesAndCompilerOptions | NotYet: src/compiler/core.ts:1533:12 | NotYet: src/compiler/watch.ts:855:19 |

The checker-host stop on both trees is a presence test of a method whose
presence the receiver's construction does not prove. The other five units
stop on a call through ?. before and after: before in core's helper, after
at the direct optional-binding expression. This measures moved stops and
does not claim native lowering completes. latent/comparison.json retains all
exact reasons, reference identifiers and timings. All per-piece records,
unit plans, logs and timers are retained under latent/before and latent/after.

## Oracle preparation piece

The standard prepare_shards.py producer completed in 804.7 seconds under a
1440-second outer limit. Apply, locked install, ordinary build and harness
phases all exit 0. Its source tree is byte-identical across all 746 src files
to the checked after tree. The frozen worktree at 21ad49c10 has the exact same
adapt.cjs, engine.cjs, owners.json and inventory.json bytes. Freezing its inputs
keeps later evidence commits from invalidating the prepared artifact hash.
The whole lane and shards run its unchanged main lane/oracle code.

The public and internal declaration outputs are byte-identical to baseline;
their hashes are retained in javascript/declarations.json. Oracle producer
logs, phase exits and source-identity evidence are under oracle/.

## Check 6 ordinary component: all 79 compiler roots

Both ordinary per-file runs complete successfully. All 79 TypeScript entry
roots are present on each side, and every normalized first-stop record is
identical. There are zero changed first stops. Seventy-three roots stop at
core.ts:810:89 with Refused: a cast the runtime can't check. The other six
records are unchanged as well; their exact kinds, locations and reasons are
in census/comparison.json and the two complete summaries.

The ordinary tool also emits rows for non-source JSON files and a combined
program row. Complete streams are retained; the 79-root comparison selects
only single-root .ts records. Nothing is suppressed or treated as accepted.
Census logs and all raw records are retained under census/. This completes
both ordinary and six-unit latent components of check 6.

## Check 4 whole-lane piece

The unchanged stage3/lane/run.sh --artifact command exits 0 with PASS in
365.7 seconds, below the 25-minute piece limit. It reports exactly 106,366
passing tests, one sanctioned public-API failure and zero pending tests.
All runners, eight workers, original test timeouts and sanctioned API lines
are unchanged. The oracle uses the frozen verified after-tree artifact.
The replay was temporarily paused for the whole test's CPU budget and has
now resumed. Whole task observations, verdict, execution metadata, API diff
and full logs are retained under oracle/whole/. The 398 standard shards
are the remaining part of check 4.
