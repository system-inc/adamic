# forEachAncestorDirectory: scratch shape measurements

**A is the smallest measured shape that removes the cast, preserves emitted JavaScript bytes, and needs no callback-body changes.** All four candidates have **zero stock TypeScript diagnostics on the 79 compiler roots**, and zero diagnostics in the separate extended caller check. This is a report only: no adaptation or compiler implementation is proposed for admission by this commit.

Main tip tested: **`2f32716054f245622c2ac3c8f692d00fa51e66e9`**. It has no `stage3/adapt` differences from the question's `3007955b`. Upstream is TypeScript 6.0.3, `050880ce59e30b356b686bd3144efe24f875ebc8`. A complete `stage3/apply.sh` run produced the baseline, child exit **0**, 360.15 seconds. It reproduces the original refusal at **path.ts:1105:28**, `a cast the runtime can't check`.

## Results

| Candidate | Stock diagnostics, 79 roots | Extended caller diagnostics | Emitted JavaScript | Full-mode census at implementation |
|---|---:|---:|---|---|
| Baseline | 0 | 0 | reference | cast Refused at 1105:28; generic implementation deferred |
| A: remove Path overload; string overload + string implementation | 0 | 0 | all 89 files byte-identical | no target findings; implementation at 1097 deferred |
| B: retain overloads; string implementation with `any` callback input | 0 | 0 | all 89 files byte-identical | no target findings; implementation at 1098 deferred |
| C: string form plus separate Path loop | 0 | 0 | 3 files change; exact hunks below | no findings in either loop; implementations at 1097 and 1114 deferred |
| D: single string declaration/implementation, no overloads | 0 | 0 | path.js comment changes only; exact hunk below | no target findings; implementation at 1094 deferred |

**Census limit:** every target generic declaration/body has status `deferred_generic_instantiation`, `instantiated: false`. These runs establish absence of recorded target findings after removing the cast, **not successful instantiated lowering**. In particular B's `any` must not be represented as an Adamic-approved callback relation. Its implementation body was deferred, not admitted. No method-bivariance spelling was used anywhere.

All five census headers have `latent_mode: full`, empty checker diagnostics, and `checker_rejected: false`. Each census child exited **0**, with `LATENT_FULL=1`, exact absolute `LATENT_ONLY_FILE` selecting that tree's `src/compiler/path.ts`, and `LATENT_ASSERT_NO_OUTPUT=1`. Each loaded the complete compiler directory; only path.ts was selected for measurement. The recorder's historical fixed label “measured on a checker-rejected program” does not supersede its actual header. [census-summary.json](census-summary.json) retains function ranges, exact findings and unit ledgers; the five `*-census.jsonl` files retain the entire raw streams.

## Exact shapes

A retains the existing string overload and changes the implementation to:

```ts
export function forEachAncestorDirectory<T>(directory: string, callback: (directory: string) => T | undefined): T | undefined
```

It removes only the Path overload signature, changes `P` to string in the implementation signature, and removes `as P`. Both original callback bodies remain unchanged. Keeping the string overload also keeps the original emitted implementation comment. See [A-source.diff](A-source.diff).

B keeps both original overloads and uses this cast-free implementation signature:

```ts
export function forEachAncestorDirectory<T>(directory: string, callback: (directory: any) => T | undefined): T | undefined
```

The loop is unchanged except for removing `as P`. The explicit `any` is the measured stock compatibility escape, not a claim of a checked relation. A plain string callback with both overloads is the known TS2394 failure from the prior evidence. B avoids TS2394 and preserves both caller-visible overloads, but its census deferral leaves the implementation relation unresolved. See [B-source.diff](B-source.diff).

C uses A's string form, then duplicates the original loop under:

```ts
export function forEachAncestorDirectoryPath<T>(directory: Path, callback: (directory: Path) => T | undefined): T | undefined
```

Neither loop has a cast. `getDirectoryPath` already declares `getDirectoryPath(path: Path): Path` at path.ts:278 and `getDirectoryPath(path: string): string` at path.ts:310. Thus the new loop's `getDirectoryPath(directory)` preserves Path through the existing overload; no new brand assertion is needed. The two Path callers and their imports change to the new function. See [C-source.diff](C-source.diff).

D removes both overload declarations and uses A's string implementation as the sole declaration. It is simpler in type syntax; retaining the existing leading documentation changes emitted comment bytes. Its executable statements and caller bodies match A. See [D-source.diff](D-source.diff).

## The two Path callers

Locations below are baseline adapted-tree locations and also remain the call locations in C. Resolved signatures, full expressions and owner chains are in [stock.json](stock.json).

1. **src/compiler/watchUtilities.ts:296**, nested `clearFirstAncestorEntry`, within `createCachedDirectoryStructureHost`:

   ```ts
   ancestor => cachedReadDirectoryResult.delete(ensureTrailingDirectorySeparator(ancestor)) ? true : undefined
   ```

   The cache is `Map<string, MutableFileSystemEntries | false>` (line 123). `ensureTrailingDirectorySeparator` accepts string, and `delete` accepts its string key. No operation in the callback needs the Path brand. Baseline resolution has a Path callback returning `true | undefined`; A/D resolve a string callback with the same result. C resolves the new Path function.

2. **src/typingsInstallerCore/typingsInstaller.ts:258**, `TypingsInstaller.installPackage`:

   ```ts
   directory => {
       if (this.installTypingHost.fileExists(combinePaths(directory, "package.json"))) {
           return directory;
       }
   }
   ```

   `combinePaths` accepts string; `fileExists` consumes the combined string. The callback returns its input. The call result is combined with `projectRootPath` using `||`, then passed to `installWorker(-1, [packageName], cwd, ...)` at line 264. `installWorker`'s cwd parameter is string (declaration at line 552). A/D change the callback/result from Path to string without an error; C preserves Path inference through the new function. No downstream Path consumer appears in this body.

There are five source calls: baseline/B retain string 3, Path 2; A/D resolve all five through string signatures; C has three old string-function calls and two new Path-function calls. The primary compiler project has **79 roots**. Because typingsInstaller is outside those roots, the additional check uses those roots plus its source entry (80 explicit roots, composite disabled for that extended check), retaining the same strict compiler options. Both checks have zero diagnostics for every candidate.

## Caller-owner pins

`grep` equivalent `rg` searches of `stage3/adapt` found **no direct text/hash/name pin on either changed caller owner** (`clearFirstAncestorEntry` / `createCachedDirectoryStructureHost`, or `TypingsInstaller.installPackage`). See [owner-pins-grep.txt](owner-pins-grep.txt), [caller-owner-pins-grep.txt](caller-owner-pins-grep.txt), and [owner-pin-matches.json](owner-pin-matches.json).

Adaptation 32 includes watchUtilities in its file list and has indexed-read site records there. Its additional runtime owner guard validates **getWatchFactory**, not createCachedDirectoryStructureHost: [watch-protocol.cjs.txt](watch-protocol.cjs.txt) and [watch-protocol.json](watch-protocol.json) retain the exact guard and pinned owner. File-name hits and copied import lists are not pins on the changed callback bodies. This audit does not claim a new adaptation would pass all adapter replay/drift contracts; no adapter was modified or replayed on a proposed patch.

## Public API

For **every candidate**, `forEachAncestorDirectory` is absent from both npm TypeScript **6.0.3's `lib/typescript.d.ts`** and **`stage3/lane/sanctioned-api.json`**. C's new `forEachAncestorDirectoryPath` name is also absent from both. Both original overloads are internal. [public-api.json](public-api.json) records the searches and [provenance.json](provenance.json) pins both inventory hashes. C nevertheless changes the internal JavaScript export surface: the path module exports another function, exposed through the internal ts namespace's existing re-export.

## Emitted JavaScript: exact hunks

Emission here means **stock TypeScript `Program.emit`**, using the compiler project's options with emit enabled, declaration emit retained as required by its composite setting, source maps disabled, and scratch output directories. The extended caller program's emission includes typingsInstaller and its dependency closure: **89 JavaScript files** per tree. All 89 byte lengths and SHA-256s are retained per candidate in [stock.json](stock.json). These are per-module JavaScript outputs, not upstream esbuild bundles. Public declaration/build bundle byte identity is not asserted by this measurement.

A and B have no JavaScript hunks. [comparison.json](comparison.json) records byte identity. C has precisely these hunks ([C-javascript.diff](C-javascript.diff)):

```diff
--- baseline/compiler/path.js
+++ C/compiler/path.js
@@ -42,6 +42,7 @@
 exports.getRelativePathFromFile = getRelativePathFromFile;
 exports.getRelativePathToDirectoryOrUrl = getRelativePathToDirectoryOrUrl;
 exports.forEachAncestorDirectory = forEachAncestorDirectory;
+exports.forEachAncestorDirectoryPath = forEachAncestorDirectoryPath;
 exports.isNodeModulesDirectory = isNodeModulesDirectory;
 const ts_js_1 = require("./_namespaces/ts.js");
 /**
@@ -813,6 +814,20 @@
     }
 }
 /** @internal */
+function forEachAncestorDirectoryPath(directory, callback) {
+    while (true) {
+        const result = callback(directory);
+        if (result !== undefined) {
+            return result;
+        }
+        const parentPath = getDirectoryPath(directory);
+        if (parentPath === directory) {
+            return undefined;
+        }
+        directory = parentPath;
+    }
+}
+/** @internal */
 function isNodeModulesDirectory(dirPath) {
     return (0, ts_js_1.endsWith)(dirPath, "/node_modules");
 }
--- baseline/compiler/watchUtilities.js
+++ C/compiler/watchUtilities.js
@@ -171,7 +171,7 @@
         return host.realpath ? host.realpath(s) : s;
     }
     function clearFirstAncestorEntry(fileOrDirectoryPath) {
-        (0, ts_js_1.forEachAncestorDirectory)((0, ts_js_1.getDirectoryPath)(fileOrDirectoryPath), ancestor => cachedReadDirectoryResult.delete((0, ts_js_1.ensureTrailingDirectorySeparator)(ancestor)) ? true : undefined);
+        (0, ts_js_1.forEachAncestorDirectoryPath)((0, ts_js_1.getDirectoryPath)(fileOrDirectoryPath), ancestor => cachedReadDirectoryResult.delete((0, ts_js_1.ensureTrailingDirectorySeparator)(ancestor)) ? true : undefined);
     }
     function addOrDeleteFileOrDirectory(fileOrDirectory, fileOrDirectoryPath) {
         const existingResult = getCachedFileSystemEntries(fileOrDirectoryPath);
--- baseline/typingsInstallerCore/typingsInstaller.js
+++ C/typingsInstallerCore/typingsInstaller.js
@@ -161,7 +161,7 @@
             this.sendResponse(response);
             return;
         }
-        const cwd = (0, ts_js_1.forEachAncestorDirectory)((0, ts_js_1.getDirectoryPath)(fileName), directory => {
+        const cwd = (0, ts_js_1.forEachAncestorDirectoryPath)((0, ts_js_1.getDirectoryPath)(fileName), directory => {
             if (this.installTypingHost.fileExists((0, ts_js_1.combinePaths)(directory, "package.json"))) {
                 return directory;
             }
```

D has precisely this comment hunk ([D-javascript.diff](D-javascript.diff)):

```diff
--- baseline/compiler/path.js
+++ D/compiler/path.js
@@ -798,7 +798,12 @@
     }
     return getPathFromPathComponents(pathComponents);
 }
-/** @internal */
+//// Path Traversal
+/**
+ * Calls `callback` on `directory` and every ancestor directory it has, returning the first defined result.
+ *
+ * @internal
+ */
 function forEachAncestorDirectory(directory, callback) {
     while (true) {
         const result = callback(directory);
```

## Reproduction and process evidence

[measure.cjs](measure.cjs) creates each candidate independently from the apply baseline, checks both programs, records source call resolutions and emits JavaScript. [compare.py](compare.py) generates exact source and JavaScript diffs. Each `*-path.ts.txt` preserves the complete measured path source. [build-overlay.py](build-overlay.py) generates the census overlay with physical Go evidence names ending `.go.txt`; no compiler source changes are needed. [run-candidates.py](run-candidates.py) supplies the required census environment. [summarize-census.py](summarize-census.py) extracts the target owners while preserving deferred-unit status.

Apply, dependency preparation, overlay/build, stock-driver and census commands used [run.py](run.py), which records its child's own exit code, bounds each child to 29 minutes, and reports termination separately. Completed commands stayed below 30 minutes; no timeout fired. Setup and measurement statuses and logs are included. Bounded census orchestration likewise reports its wrapper children's own exits. Main/upstream pins, every adaptation input hash, baseline path hash, overlay hashes and census binary hash are in [provenance.json](provenance.json). No binary is committed.

Two setup/exploratory results are explicitly superseded: `overlay-build` exited 1 before the pinned submodules were ready; `overlay-build-final` and `census-build` exited 0 afterward. The first `stock` emission driver incorrectly disabled declaration emission on a composite project and recorded TS6304 (one diagnostic per tree); `stock-final` retained declaration emission and supersedes it. Its child exited 0; the final [stock.json](stock.json) contains the corrected zero-diagnostic results. Those historical logs are preserved rather than represented as successful checks.

No candidate was applied to `stage3/adapt`. A is the best measured next shape; instantiated Adamic acceptance remains a separate decision gate because this requested single-file census defers the generic implementations.
