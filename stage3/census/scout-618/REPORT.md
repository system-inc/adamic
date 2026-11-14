# Scout: sys.ts:618:33

Recommend **(b): retain the string-returning canonicalizer and mint `Path` directly
at its four Path-consuming calls**. Stock TypeScript 6.0.3: **0 diagnostics / 79
roots (0 new)**. Full single-file CommonJSBundle sys.ts census: **41 NotYet,
18 Refused, 55 Boundary**, versus **41/19/55** at base. **Only the 618:33
unchecked-function-cast finding disappears; no new or changed finding. Emitted
sys.js is byte-identical.** This is a scout recommendation, not an adaptation.

## Results and limits

| Shape | Stock diagnostics / roots | Unique NotYet | Refused | Boundary | Semantic finding delta | sys.js |
| --- | --- | ---: | ---: | ---: | --- | --- |
| Base | 0 / 79 | 41 | 19 | 55 | — | reference |
| (a) cached canonicalizer + typed toPath wrapper | 0 / 79 | 41 | 18 | 55 | only 618:33 gone | changed: import, wrapper, normalization |
| (b) string function + four direct result mints | 0 / 79 | 41 | 18 | 55 | only 618:33 gone | byte-identical |
| (c) cached canonicalizer + typed direct-mint wrapper | 0 / 79 | 41 | 18 | 55 | only 618:33 gone | changed: wrapper |

All counts above deduplicate `(kind, where, reason, text)` across attempts and
combine refusal/contract scans with actual lowering. Boundary is separate from
NotYet/Refused. All four runs finish with a relation-report trailer, no overflow,
no errors/panics, and zero checker diagnostics. The tool's fixed label says
“measured on a checker-rejected program”; the actual headers say
`checker_rejected:false`. These are full-mode censuses, not scan-only runs.

**Coverage limit:** even after removing the cast, the owner's attempt stops in
its prologue at `sys.ts:594:5`, Refused “a method in object destructuring”, with a
Boundary at `590:1`. Therefore these results do not establish end-to-end lowering
of either new wrapper or of the owner. The full census exercises eligible nested
units and detects no newly exposed lowering stop. No claim that all 73 formerly
blocked roots now compile is made; their new ordinary first stops were not measured.

## Precedent

At the exact base, `stage3/adapt/40-explicit-any/BRANDS-VOID.md` is absent.
Its successor `BRANDS-UNDEFINED.md` explicitly supersedes the old void ruling:
phantom members retain names and required/optional flags but become undefined.
The class rules contain rule 147 for `types.ts`'s `__pathBrand`, plus rule 103 for
`PathPathComponents`. No Path-mint helper is introduced by adaptation 40.

Rule 103's final form is:

```ts
export interface PathPathComponents<T extends Path = Path> extends Array<T> {
    __pathComponensBrand: undefined;
}
```

The README records all 13 uses as bare uses of the default, zero casts, and a
separate phantom-array road. It is not the rule that admits a primitive Path mint.

In the final adapted path.ts, `toPath` is at 764:1; its cast is at 768:

```ts
export function toPath(fileName: string, basePath: string | undefined,
    getCanonicalFileName: (path: string) => string): Path {
    const nonCanonicalizedPath = isRootedDiskPath(fileName)
        ? normalizePath(fileName)
        : getNormalizedAbsolutePath(fileName, basePath);
    return getCanonicalFileName(nonCanonicalizedPath) as Path;
}
```

Its **full single-file census** completes with **7 unique NotYet, 3 Refused,
10 Boundary** (raw attempts: 13/3/16), zero errors/panics. `toPath` is an attempted
unit with no finding: its own `as Path` mint actually lowers. `castProof` first
calls `phantomCast`; `phantomAssignable` compares the primitive bases admitted by
`phantomBase` and `phantomField`. The undefined, string-keyed `__pathBrand` has no
runtime property or primitive-prototype name conflict. See
`internal/lower/{cast_proof,phantom_brands,view_unions_primitive_brands}.go` and
`docs/rulings.md:29–31,35–37`. This is distinct from `brandMintOf`'s interface
extension / array mint road. An unchecked function cast does not get this primitive
proof merely because its return type is branded.

Thus invoking a genuinely Path-returning helper removes the function cast honestly.
But using `toPath` here adds normalization, not just a brand. A focused stock-6.0.3
API check with the case-sensitive canonicalizer gives:

| Input | Existing canonicalizer | toPath(input, undefined, canonicalizer) |
| --- | --- | --- |
| `/a/../b` | `/a/../b` | `/b` |
| `a/./b` | `a/./b` | `a/b` |
| `C:\a\b` | `C:\a\b` | `C:/a/b` |

This is a helper-behavior witness, not a claim those strings reach this watcher.
The watcher inventory supplies no universal proof that normalization is redundant.
Prefer (b); (c) uses the same accepted direct primitive mint but adds a closure.

## Ownership

Read `stage3/adapt/99-correctness-timer-closures/README.md` and `owners.json`.
The latter pins `createDirectoryWatcherSupportingRecursive` whole, before/after,
with SHA-256. The freshly adapted sys.ts contains the exact after owner once.
Any eventual implementation must amend that owner in rule 99. No owner, adapter,
compiler, sanction or test expectation was changed by this scout.

## Reader inventory (whole rule-99 after owner)

There are five call expressions plus one function-value use:

- `createDirectoryWatcher`: `dirPath` keys cache.get/set/delete,
  callbackCache.add/remove, callback dispatch, and child-watch updates. Captured
  callbacks and close retain this key; updates use it in the pending map too.
- `toPathInLink`: passes the function to getRelativePathFromDirectory as its
  string canonicalizer; the returned Path is consumed as a string comparison
  key, and combinePaths produces a string filename.
- Both recursive `invokeCallbacks` branches: canonicalize `link` for dirPath;
  this keys cache/links and participates in callback-root equality, prefix,
  separator, and length tests. One branch carries an InvokeMap, the other a
  single filename. Callback payloads remain filenames, not branded keys.
- `removeChildWatches`: canonicalized childWatcher.dirName keys cache.get before
  recursively removing its children.
- `isInPath`: canonicalized path is used only for string.includes(searchPath)
  after the direct test fails and case-sensitive mode is excluded.

No consumer observes the phantom brand. The Path-consuming boundaries are
map/multimap keys and the dirPath parameters. The function-value consumer and
includes consumer need only strings.


## Scratch candidate source edits

All edits below are relative to the fully adapted base. (b) leaves the function
string-returning at the callback and includes consumers, minting only at the four
calls whose consumers require Path. Its direct casts use the same accepted
primitive mint as path.ts, preserving evaluation count and runtime operations.
(a) additionally imports the existing toPath helper; (c) is the typed wrapper
using the house direct-result mint.

### (a)

```diff
--- base.sys.ts
+++ a.sys.ts
@@ -37 +37 @@
-    normalizeSlashes,
+    normalizeSlashes, toPath,
@@ -618 +618,2 @@
-    const toCanonicalFilePath = createGetCanonicalFileName(useCaseSensitiveFileNames) as (fileName: string) => Path;
+    const getCanonicalFileName = createGetCanonicalFileName(useCaseSensitiveFileNames);
+    const toCanonicalFilePath = (fileName: string): Path => toPath(fileName, undefined, getCanonicalFileName);
```

### (b)

```diff
--- base.sys.ts
+++ b.sys.ts
@@ -618 +618 @@
-    const toCanonicalFilePath = createGetCanonicalFileName(useCaseSensitiveFileNames) as (fileName: string) => Path;
+    const toCanonicalFilePath = createGetCanonicalFileName(useCaseSensitiveFileNames);
@@ -634 +634 @@
-        const dirPath = toCanonicalFilePath(dirName);
+        const dirPath = (toCanonicalFilePath(dirName) as Path);
@@ -733 +733 @@
-                invokeCallbacks(link, toCanonicalFilePath(link), invokeMap, fileNames?.map(toPathInLink));
+                invokeCallbacks(link, (toCanonicalFilePath(link) as Path), invokeMap, fileNames?.map(toPathInLink));
@@ -736 +736 @@
-                invokeCallbacks(link, toCanonicalFilePath(link), toPathInLink(fileName!));
+                invokeCallbacks(link, (toCanonicalFilePath(link) as Path), toPathInLink(fileName!));
@@ -818 +818 @@
-            removeChildWatches(cache.get(toCanonicalFilePath(childWatcher.dirName)));
+            removeChildWatches(cache.get((toCanonicalFilePath(childWatcher.dirName) as Path)));
```

### (c)

```diff
--- base.sys.ts
+++ c.sys.ts
@@ -618 +618,2 @@
-    const toCanonicalFilePath = createGetCanonicalFileName(useCaseSensitiveFileNames) as (fileName: string) => Path;
+    const getCanonicalFileName = createGetCanonicalFileName(useCaseSensitiveFileNames);
+    const toCanonicalFilePath = (fileName: string): Path => getCanonicalFileName(fileName) as Path;
```

## Complete census delta

For each of (a), (b), (c), the only gone finding is:

- **Refused**, `src/compiler/sys.ts:618:33`, phase `refusal_scan`, attempting owner
  `src/compiler/sys.ts:590:1`, reason **“a cast the runtime can't check”**.
  Full diagnostic: `src/compiler/sys.ts:618:33: Adamic 0.1 refuses a cast the
  runtime can't check; use a proven upcast, cast a discriminated object union with
  unique literal or enum tags to members or a sub-union, or downcast along nominal
  class ancestry (adamic/no-unchecked-cast)`.

New findings: **none**. Changed kind, reason, text after coordinate normalization,
or phase: **none**. (b) has no coordinate shifts. (a) and (c) insert one line;
unchanged findings after line 618 move down one line. The following is the complete
coordinate-only ledger, shared by (a) and (c); each row identifies one unique
finding. `text` is the baseline diagnostic, without the scratch path prefix.

| Kind | Baseline where | (a)/(c) where | Reason | Baseline text |
| --- | --- | --- | --- | --- |
| Boundary | src/compiler/sys.ts:1026:9 | src/compiler/sys.ts:1027:9 | failed statement; state rolled back; continued at next statement or declaration | src/compiler/sys.ts:1026:54: stage 0 can't lower a boolean \| undefined variable a function value captures yet |
| Boundary | src/compiler/sys.ts:1048:21 | src/compiler/sys.ts:1049:21 | failed statement; state rolled back; continued at next statement or declaration | src/compiler/sys.ts:1048:140: stage 0 can't lower a boolean \| undefined variable a function value captures yet |
| Boundary | src/compiler/sys.ts:1052:17 | src/compiler/sys.ts:1053:17 | failed statement; state rolled back; continued at next statement or declaration | src/compiler/debug.ts:251:33: stage 0 can't lower a value of type never yet |
| Boundary | src/compiler/sys.ts:1057:9 | src/compiler/sys.ts:1058:9 | failed statement; state rolled back; continued at next statement or declaration | src/compiler/sys.ts:1057:84: stage 0 can't lower reading getModifiedTime yet |
| Boundary | src/compiler/sys.ts:1061:9 | src/compiler/sys.ts:1062:9 | failed statement; state rolled back; continued at next statement or declaration | src/compiler/sys.ts:1061:90: stage 0 can't lower reading getModifiedTime yet |
| Boundary | src/compiler/sys.ts:1096:9 | src/compiler/sys.ts:1097:9 | failed statement; state rolled back; continued at next statement or declaration | src/compiler/sys.ts:1096:40: stage 0 can't lower optional record property access yet |
| Boundary | src/compiler/sys.ts:1097:9 | src/compiler/sys.ts:1098:9 | failed statement; state rolled back; continued at next statement or declaration | src/compiler/sys.ts:1099:30: stage 0 can't lower reading defaultFallbackPolling yet |
| Boundary | src/compiler/sys.ts:1118:13 | src/compiler/sys.ts:1119:13 | failed statement; state rolled back; continued at next statement or declaration | src/compiler/sys.ts:1121:17: stage 0 can't lower reading fileSystemEntryExists yet |
| Boundary | src/compiler/sys.ts:1168:17 | src/compiler/sys.ts:1169:17 | failed statement; state rolled back; continued at next statement or declaration | src/compiler/debug.ts:251:33: stage 0 can't lower a value of type never yet |
| Boundary | src/compiler/sys.ts:1182:17 | src/compiler/sys.ts:1183:17 | failed statement; state rolled back; continued at next statement or declaration | src/compiler/sys.ts:1182:48: stage 0 can't lower optional record property access yet |
| Boundary | src/compiler/sys.ts:1251:17 | src/compiler/sys.ts:1252:17 | failed statement; state rolled back; continued at next statement or declaration | src/compiler/sys.ts:1251:77: stage 0 can't lower function identity comparison in a program with specialized generic function values yet |
| Boundary | src/compiler/sys.ts:1267:17 | src/compiler/sys.ts:1268:17 | failed statement; state rolled back; continued at next statement or declaration | src/compiler/sys.ts:1267:89: stage 0 can't lower a boolean \| undefined variable a function value captures yet |
| Boundary | src/compiler/sys.ts:1275:17 | src/compiler/sys.ts:1276:17 | failed statement; state rolled back; continued at next statement or declaration | src/compiler/sys.ts:1275:17: stage 0 can't lower reading presentWatcher yet |
| Boundary | src/compiler/sys.ts:1279:17 | src/compiler/sys.ts:1280:17 | failed statement; state rolled back; continued at next statement or declaration | src/compiler/sys.ts:1279:24: stage 0 can't lower reading presentWatcher yet |
| Boundary | src/compiler/sys.ts:1386:5 | src/compiler/sys.ts:1387:5 | failed statement; state rolled back; continued at next statement or declaration | src/compiler/sys.ts:1391:49: Adamic 0.1 refuses inherited library member call read as an own field; prototype members are not stored in an object's shape; call the method on its receiver, or wrap that call in an arrow (unbound-method) |
| Boundary | src/compiler/sys.ts:1517:9 | src/compiler/sys.ts:1518:9 | failed statement; state rolled back; continued at next statement or declaration | src/compiler/sys.ts:1588:191: stage 0 can't lower process.recordreplay as a value (use its supported process operation directly) yet |
| Boundary | src/compiler/sys.ts:1540:17 | src/compiler/sys.ts:1541:17 | failed statement; state rolled back; continued at next statement or declaration | src/compiler/sys.ts:1517:15: stage 0 can't lower a function value that captures the variable its own initializer declares yet |
| Boundary | src/compiler/sys.ts:1540:65 | src/compiler/sys.ts:1541:65 | failed statement; state rolled back; continued at next statement or declaration | src/compiler/sys.ts:1517:15: stage 0 can't lower a function value that captures the variable its own initializer declares yet |
| Boundary | src/compiler/sys.ts:1543:21 | src/compiler/sys.ts:1544:21 | failed statement; state rolled back; continued at next statement or declaration | src/compiler/sys.ts:1517:15: stage 0 can't lower a function value that captures the variable its own initializer declares yet |
| Boundary | src/compiler/sys.ts:1544:25 | src/compiler/sys.ts:1545:25 | failed statement; state rolled back; continued at next statement or declaration | src/compiler/sys.ts:1517:15: stage 0 can't lower a function value that captures the variable its own initializer declares yet |
| Boundary | src/compiler/sys.ts:1547:25 | src/compiler/sys.ts:1548:25 | failed statement; state rolled back; continued at next statement or declaration | src/compiler/sys.ts:1517:15: stage 0 can't lower a function value that captures the variable its own initializer declares yet |
| Boundary | src/compiler/sys.ts:1547:56 | src/compiler/sys.ts:1548:56 | failed statement; state rolled back; continued at next statement or declaration | src/compiler/sys.ts:1517:15: stage 0 can't lower a function value that captures the variable its own initializer declares yet |
| Boundary | src/compiler/sys.ts:1549:29 | src/compiler/sys.ts:1550:29 | failed statement; state rolled back; continued at next statement or declaration | src/compiler/sys.ts:1517:15: stage 0 can't lower a function value that captures the variable its own initializer declares yet |
| Boundary | src/compiler/sys.ts:1816:17 | src/compiler/sys.ts:1817:17 | failed statement; state rolled back; continued at next statement or declaration | src/compiler/sys.ts:1816:24: stage 0 can't lower reading byteOrderMarkIndicator yet |
| Boundary | src/compiler/sys.ts:1867:17 | src/compiler/sys.ts:1868:17 | failed statement; state rolled back; continued at next statement or declaration | src/compiler/sys.ts:1867:17: Adamic 0.1 refuses sort without a comparator; pass one: the default compares numbers as strings, so [10, 9, 1].sort() is [1, 10, 9] |
| Boundary | src/compiler/sys.ts:1868:17 | src/compiler/sys.ts:1869:17 | failed statement; state rolled back; continued at next statement or declaration | src/compiler/sys.ts:1868:17: Adamic 0.1 refuses sort without a comparator; pass one: the default compares numbers as strings, so [10, 9, 1].sort() is [1, 10, 9] |
| Boundary | src/compiler/sys.ts:1965:1 | src/compiler/sys.ts:1966:1 | failed statement; state rolled back; continued at next statement or declaration | src/compiler/sys.ts:1965:12: stage 0 can't lower a literal method through a view that erases its receiver yet |
| Boundary | src/compiler/sys.ts:646:25 | src/compiler/sys.ts:647:25 | failed statement; state rolled back; continued at next statement or declaration | src/compiler/sys.ts:646:29: stage 0 can't lower optional record property access yet |
| Boundary | src/compiler/sys.ts:764:9 | src/compiler/sys.ts:765:9 | failed statement; state rolled back; continued at next statement or declaration | src/compiler/sys.ts:764:13: stage 0 can't lower reading timerToUpdateChildWatches yet |
| Boundary | src/compiler/sys.ts:768:9 | src/compiler/sys.ts:769:9 | failed statement; state rolled back; continued at next statement or declaration | src/compiler/sys.ts:768:9: stage 0 can't lower assigning to an Identifier yet |
| Boundary | src/compiler/sys.ts:777:9 | src/compiler/sys.ts:778:9 | failed statement; state rolled back; continued at next statement or declaration | src/compiler/sys.ts:777:9: stage 0 can't lower assigning to an Identifier yet |
| Boundary | src/compiler/sys.ts:782:9 | src/compiler/sys.ts:783:9 | failed statement; state rolled back; continued at next statement or declaration | src/compiler/sys.ts:782:17: stage 0 can't lower reading timerToUpdateChildWatches yet |
| Boundary | src/compiler/sys.ts:809:9 | src/compiler/sys.ts:810:9 | failed statement; state rolled back; continued at next statement or declaration | src/compiler/sys.ts:809:117: stage 0 can't lower reading timerToUpdateChildWatches yet |
| Boundary | src/compiler/sys.ts:829:5 | src/compiler/sys.ts:830:5 | function signature or prologue failed; continued at next declaration | src/compiler/sys.ts:834:13: Adamic 0.1 refuses a value of type any: any has no proven type; type it unknown and narrow it where it's read, or give it its real type |
| Boundary | src/compiler/sys.ts:937:5 | src/compiler/sys.ts:938:5 | failed statement; state rolled back; continued at next statement or declaration | src/compiler/sys.ts:937:13: stage 0 can't lower optional record property access yet |
| Boundary | src/compiler/sys.ts:993:1 | src/compiler/sys.ts:994:1 | function signature or prologue failed; continued at next declaration | src/compiler/sys.ts:1003:5: Adamic 0.1 refuses a method in object destructuring; call it on its receiver or wrap that call in an arrow; destructuring would lose this |
| NotYet | src/compiler/sys.ts:1026:54 | src/compiler/sys.ts:1027:54 | a boolean \| undefined variable a function value captures | src/compiler/sys.ts:1026:54: stage 0 can't lower a boolean \| undefined variable a function value captures yet |
| NotYet | src/compiler/sys.ts:1048:140 | src/compiler/sys.ts:1049:140 | a boolean \| undefined variable a function value captures | src/compiler/sys.ts:1048:140: stage 0 can't lower a boolean \| undefined variable a function value captures yet |
| NotYet | src/compiler/sys.ts:1057:84 | src/compiler/sys.ts:1058:84 | reading getModifiedTime | src/compiler/sys.ts:1057:84: stage 0 can't lower reading getModifiedTime yet |
| NotYet | src/compiler/sys.ts:1061:90 | src/compiler/sys.ts:1062:90 | reading getModifiedTime | src/compiler/sys.ts:1061:90: stage 0 can't lower reading getModifiedTime yet |
| NotYet | src/compiler/sys.ts:1096:40 | src/compiler/sys.ts:1097:40 | optional record property access | src/compiler/sys.ts:1096:40: stage 0 can't lower optional record property access yet |
| NotYet | src/compiler/sys.ts:1099:30 | src/compiler/sys.ts:1100:30 | reading defaultFallbackPolling | src/compiler/sys.ts:1099:30: stage 0 can't lower reading defaultFallbackPolling yet |
| NotYet | src/compiler/sys.ts:1121:17 | src/compiler/sys.ts:1122:17 | reading fileSystemEntryExists | src/compiler/sys.ts:1121:17: stage 0 can't lower reading fileSystemEntryExists yet |
| NotYet | src/compiler/sys.ts:1182:48 | src/compiler/sys.ts:1183:48 | optional record property access | src/compiler/sys.ts:1182:48: stage 0 can't lower optional record property access yet |
| NotYet | src/compiler/sys.ts:1251:77 | src/compiler/sys.ts:1252:77 | function identity comparison in a program with specialized generic function values | src/compiler/sys.ts:1251:77: stage 0 can't lower function identity comparison in a program with specialized generic function values yet |
| NotYet | src/compiler/sys.ts:1267:89 | src/compiler/sys.ts:1268:89 | a boolean \| undefined variable a function value captures | src/compiler/sys.ts:1267:89: stage 0 can't lower a boolean \| undefined variable a function value captures yet |
| NotYet | src/compiler/sys.ts:1275:17 | src/compiler/sys.ts:1276:17 | reading presentWatcher | src/compiler/sys.ts:1275:17: stage 0 can't lower reading presentWatcher yet |
| NotYet | src/compiler/sys.ts:1279:24 | src/compiler/sys.ts:1280:24 | reading presentWatcher | src/compiler/sys.ts:1279:24: stage 0 can't lower reading presentWatcher yet |
| NotYet | src/compiler/sys.ts:1500:57 | src/compiler/sys.ts:1501:57 | node:timers.clearTimeout value needs exactly (timeoutId: unknown) => void, not (timeout: unknown) => void with no slot | src/compiler/sys.ts:1500:57: stage 0 can't lower node:timers.clearTimeout value needs exactly (timeoutId: unknown) => void, not (timeout: unknown) => void with no slot yet |
| NotYet | src/compiler/sys.ts:1517:15 | src/compiler/sys.ts:1518:15 | a function value that captures the variable its own initializer declares | src/compiler/sys.ts:1517:15: stage 0 can't lower a function value that captures the variable its own initializer declares yet |
| NotYet | src/compiler/sys.ts:1588:191 | src/compiler/sys.ts:1589:191 | process.recordreplay as a value (use its supported process operation directly) | src/compiler/sys.ts:1588:191: stage 0 can't lower process.recordreplay as a value (use its supported process operation directly) yet |
| NotYet | src/compiler/sys.ts:1588:192 | src/compiler/sys.ts:1589:192 | a namespace object used as a value; no runtime container is emitted, so identity, receiver behavior, live export aliases and staged properties are not represented; use qualified members or named module imports | src/compiler/sys.ts:1588:192: stage 0 can't lower a namespace object used as a value; no runtime container is emitted, so identity, receiver behavior, live export aliases and staged properties are not represented; use qualified members or named module imports yet |
| NotYet | src/compiler/sys.ts:1598:57 | src/compiler/sys.ts:1599:57 | node:timers.clearTimeout value needs exactly (timeoutId: unknown) => void, not (timeout: unknown) => void with no slot | src/compiler/sys.ts:1598:57: stage 0 can't lower node:timers.clearTimeout value needs exactly (timeoutId: unknown) => void, not (timeout: unknown) => void with no slot yet |
| NotYet | src/compiler/sys.ts:1816:24 | src/compiler/sys.ts:1817:24 | reading byteOrderMarkIndicator | src/compiler/sys.ts:1816:24: stage 0 can't lower reading byteOrderMarkIndicator yet |
| NotYet | src/compiler/sys.ts:1965:12 | src/compiler/sys.ts:1966:12 | a literal method through a view that erases its receiver | src/compiler/sys.ts:1965:12: stage 0 can't lower a literal method through a view that erases its receiver yet |
| NotYet | src/compiler/sys.ts:646:29 | src/compiler/sys.ts:647:29 | optional record property access | src/compiler/sys.ts:646:29: stage 0 can't lower optional record property access yet |
| NotYet | src/compiler/sys.ts:764:13 | src/compiler/sys.ts:765:13 | reading timerToUpdateChildWatches | src/compiler/sys.ts:764:13: stage 0 can't lower reading timerToUpdateChildWatches yet |
| NotYet | src/compiler/sys.ts:768:9 | src/compiler/sys.ts:769:9 | assigning to an Identifier | src/compiler/sys.ts:768:9: stage 0 can't lower assigning to an Identifier yet |
| NotYet | src/compiler/sys.ts:777:9 | src/compiler/sys.ts:778:9 | assigning to an Identifier | src/compiler/sys.ts:777:9: stage 0 can't lower assigning to an Identifier yet |
| NotYet | src/compiler/sys.ts:782:17 | src/compiler/sys.ts:783:17 | reading timerToUpdateChildWatches | src/compiler/sys.ts:782:17: stage 0 can't lower reading timerToUpdateChildWatches yet |
| NotYet | src/compiler/sys.ts:809:117 | src/compiler/sys.ts:810:117 | reading timerToUpdateChildWatches | src/compiler/sys.ts:809:117: stage 0 can't lower reading timerToUpdateChildWatches yet |
| NotYet | src/compiler/sys.ts:937:13 | src/compiler/sys.ts:938:13 | optional record property access | src/compiler/sys.ts:937:13: stage 0 can't lower optional record property access yet |
| Refused | src/compiler/sys.ts:1003:5 | src/compiler/sys.ts:1004:5 | a method in object destructuring | src/compiler/sys.ts:1003:5: Adamic 0.1 refuses a method in object destructuring; call it on its receiver or wrap that call in an arrow; destructuring would lose this |
| Refused | src/compiler/sys.ts:1385:31 | src/compiler/sys.ts:1386:31 | a method read as a value (writeFile would lose its object, and this with it) | src/compiler/sys.ts:1385:31: Adamic 0.1 refuses a method read as a value (writeFile would lose its object, and this with it); call it in an arrow that keeps the object: (path, data, writeByteOrderMark) => sys.writeFile(path, data, writeByteOrderMark) (unbound-method) |
| Refused | src/compiler/sys.ts:1391:49 | src/compiler/sys.ts:1392:49 | inherited library member call read as an own field | src/compiler/sys.ts:1391:49: Adamic 0.1 refuses inherited library member call read as an own field; prototype members are not stored in an object's shape; call the method on its receiver, or wrap that call in an arrow (unbound-method) |
| Refused | src/compiler/sys.ts:1532:13 | src/compiler/sys.ts:1533:13 | a function taking PollingInterval seen as one taking number \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | src/compiler/sys.ts:1532:13: Adamic 0.1 refuses a function taking PollingInterval seen as one taking number \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take; write the method as a property holding a function (handle: (animal: Animal) => void), which tsc checks one way, or take the wider type in the method (method-signature-style) |
| Refused | src/compiler/sys.ts:1533:13 | src/compiler/sys.ts:1534:13 | a function taking boolean seen as one taking boolean \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | src/compiler/sys.ts:1533:13: Adamic 0.1 refuses a function taking boolean seen as one taking boolean \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take; write the method as a property holding a function (handle: (animal: Animal) => void), which tsc checks one way, or take the wider type in the method (method-signature-style) |
| Refused | src/compiler/sys.ts:1610:22 | src/compiler/sys.ts:1611:22 | a value of type (baseDir: string, moduleName: string) => { module: any; modulePath: string; error: undefined; } \| { module: undefined; modulePath: undefined; error: unknown; } seen as (baseDir: string, moduleName: string) => ModuleImportResult, which can write string \| undefined where string is read | src/compiler/sys.ts:1610:22: Adamic 0.1 refuses a value of type (baseDir: string, moduleName: string) => { module: any; modulePath: string; error: undefined; } \| { module: undefined; modulePath: undefined; error: unknown; } seen as (baseDir: string, moduleName: string) => ModuleImportResult, which can write string \| undefined where string is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable) |
| Refused | src/compiler/sys.ts:1612:77 | src/compiler/sys.ts:1613:77 | a value of type System seen as ModuleResolutionHost, whose method directoryExists becomes a writable field: a write of ((directoryName: string) => boolean) \| undefined would replace a method of type (path: string) => boolean | src/compiler/sys.ts:1612:77: Adamic 0.1 refuses a value of type System seen as ModuleResolutionHost, whose method directoryExists becomes a writable field: a write of ((directoryName: string) => boolean) \| undefined would replace a method of type (path: string) => boolean; make directoryExists readonly in the type it's seen as (readonly directoryExists: ...), or declare it there as a method, which isn't written (adamic/invariant-mutable) |
| Refused | src/compiler/sys.ts:1867:17 | src/compiler/sys.ts:1868:17 | sort without a comparator | src/compiler/sys.ts:1867:17: Adamic 0.1 refuses sort without a comparator; pass one: the default compares numbers as strings, so [10, 9, 1].sort() is [1, 10, 9] |
| Refused | src/compiler/sys.ts:1868:17 | src/compiler/sys.ts:1869:17 | sort without a comparator | src/compiler/sys.ts:1868:17: Adamic 0.1 refuses sort without a comparator; pass one: the default compares numbers as strings, so [10, 9, 1].sort() is [1, 10, 9] |
| Refused | src/compiler/sys.ts:1965:12 | src/compiler/sys.ts:1966:12 | a method read as a value (getEnvironmentVariable would lose its object, and this with it) | src/compiler/sys.ts:1965:12: Adamic 0.1 refuses a method read as a value (getEnvironmentVariable would lose its object, and this with it); call it in an arrow that keeps the object: (name) => sys.getEnvironmentVariable(name) (unbound-method) |
| Refused | src/compiler/sys.ts:790:87 | src/compiler/sys.ts:791:87 | a value of type Map<Path, string[]> seen as InvokeMap, which can write true \| string[] where string[] is read | src/compiler/sys.ts:790:87: Adamic 0.1 refuses a value of type Map<Path, string[]> seen as InvokeMap, which can write true \| string[] where string[] is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable) |
| Refused | src/compiler/sys.ts:834:13 | src/compiler/sys.ts:835:13 | a value of type any: any has no proven type | src/compiler/sys.ts:834:13: Adamic 0.1 refuses a value of type any: any has no proven type; type it unknown and narrow it where it's read, or give it its real type |
| Refused | src/compiler/sys.ts:877:34 | src/compiler/sys.ts:878:34 | a value of type never[] seen as ChildDirectoryWatcher[], which can write ChildDirectoryWatcher where never is read | src/compiler/sys.ts:877:34: Adamic 0.1 refuses a value of type never[] seen as ChildDirectoryWatcher[], which can write ChildDirectoryWatcher where never is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable) |

## Emitted JavaScript diffs

Stock 6.0.3 `transpileModule`, ES2024 / ESNext, same options and full sys.ts input
for every case; comments retained. This is a per-file emission comparison, not a
bundle or source-map comparison. The following diffs are complete, with zero context lines.

### (a)

```diff
--- base.sys.js
+++ a.sys.js
@@ -1 +1 @@
-import { multiMapAddValue, multiMapRemoveValue, AssertionLevel, closeFileWatcher, closeFileWatcherOf, combinePaths, Comparison, contains, containsPath, createGetCanonicalFileName, createMultiMap, Debug, directorySeparator, emptyArray, emptyFileSystemEntries, endsWith, enumerateInsertsAndDeletes, getDirectoryPath, getFallbackOptions, getNormalizedAbsolutePath, getRelativePathFromDirectory, getRelativePathToDirectoryOrUrl, getRootLength, getStringComparer, isArray, isNodeLikeSystem, isString, mapDefined, matchesExclude, matchFiles, memoize, noop, normalizePath, normalizeSlashes, orderedRemoveItem, PollingWatchKind, resolveJSModule, some, startsWith, timestamp, unorderedRemoveItem, WatchDirectoryKind, WatchFileKind, writeFileEnsuringDirectories, } from "./_namespaces/ts.js";
+import { multiMapAddValue, multiMapRemoveValue, AssertionLevel, closeFileWatcher, closeFileWatcherOf, combinePaths, Comparison, contains, containsPath, createGetCanonicalFileName, createMultiMap, Debug, directorySeparator, emptyArray, emptyFileSystemEntries, endsWith, enumerateInsertsAndDeletes, getDirectoryPath, getFallbackOptions, getNormalizedAbsolutePath, getRelativePathFromDirectory, getRelativePathToDirectoryOrUrl, getRootLength, getStringComparer, isArray, isNodeLikeSystem, isString, mapDefined, matchesExclude, matchFiles, memoize, noop, normalizePath, normalizeSlashes, toPath, orderedRemoveItem, PollingWatchKind, resolveJSModule, some, startsWith, timestamp, unorderedRemoveItem, WatchDirectoryKind, WatchFileKind, writeFileEnsuringDirectories, } from "./_namespaces/ts.js";
@@ -421 +421,2 @@
-    const toCanonicalFilePath = createGetCanonicalFileName(useCaseSensitiveFileNames);
+    const getCanonicalFileName = createGetCanonicalFileName(useCaseSensitiveFileNames);
+    const toCanonicalFilePath = (fileName) => toPath(fileName, undefined, getCanonicalFileName);
```

### (b)

Empty diff; byte-identical.

### (c)

```diff
--- base.sys.js
+++ c.sys.js
@@ -421 +421,2 @@
-    const toCanonicalFilePath = createGetCanonicalFileName(useCaseSensitiveFileNames);
+    const getCanonicalFileName = createGetCanonicalFileName(useCaseSensitiveFileNames);
+    const toCanonicalFilePath = (fileName) => getCanonicalFileName(fileName);
```

## Timer at line 615

`let timerToUpdateChildWatches: any;` remains in the exact rule-99 after owner.
It is **not settled by the timer-host closure work**. Adaptation 40's
TIMER-SLOT.md covers the solution builder's `timerToBuildInvalidatedProject`, not
this local. Baseline and all candidates retain the following nested-unit lowering
observations (baseline coordinates; a/c add one):

| NotYet where | Reason | Failed-statement Boundary |
| --- | --- | --- |
| 764:13 | reading timerToUpdateChildWatches | 764:9 |
| 782:17 | reading timerToUpdateChildWatches | 782:9 |
| 809:117 | reading timerToUpdateChildWatches | 809:9 |

These are real further lowering stops. The full run does not diagnose a new
Refused at declaration 615: the enclosing owner's prologue fails before lowering
that local. Consequently it would be inaccurate to claim this experiment proves
that simply changing this local's type resolves these reads, or that the ordinary
compiler reaches them immediately after removing 618. No timer edit was tried.

## Reproduction and provenance

Base Adamic: `e71fab21338c77ed365f56647da147f149cf8114`.
Fresh `git worktree add --detach` under `/tmp/scout-618/adamic`, then new
`typescript/scout-618` branch for this report only; no force push.
Ran `git submodule update --init --recursive --depth 1`, cloned TypeScript with
`--depth 1 --branch v6.0.3`, confirmed
`050880ce59e30b356b686bd3144efe24f875ebc8`, exported
`ADAMIC_TYPESCRIPT_SOURCE=/tmp/scout-618/TypeScript`, and ran `npm ci` in stage3/api.
Full `python3 stage3/apply.py /tmp/scout-618/adapted` exited 0.

Go is `/workspace/adamic-tools/go/bin/go`, version go1.27.1; `/usr/bin/go` is an
unrelated command. GOTMPDIR and TMPDIR were `/tmp/scout-618/tmp`, GOCACHE
`/tmp/scout-618/gocache`, GOPATH `/tmp/scout-618/gopath`. Build:

```sh
python3 stage3/census/latent/make_overlay.py /tmp/scout-618/adamic /tmp/scout-618/overlay
go build -overlay /tmp/scout-618/overlay/overlay.json -o /tmp/scout-618/latent ./stage3/census/latent/tool
LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 \
LATENT_ONLY_FILE=/tmp/scout-618/adapted/src/compiler/sys.ts \
/tmp/scout-618/latent /tmp/scout-618/adapted/src/compiler OUTPUT.jsonl
# Repeat with LATENT_ONLY_FILE=.../path.ts for the mint precedent.
```

Module format defaults to CommonJSBundle and every header confirms it. No scan
substitute, speculative mode, Adamic compiler source edit or usable output IR.
The measurement overlay is scratch-only. Each sys candidate starts from identical
base bytes; the runner restores baseline sys.ts after (c). The path file never changes.

Stock oracle: `stage3/api/node_modules/typescript` 6.0.3, all recursively discovered
79 `.ts` compiler roots, strict, noUncheckedIndexedAccess,
exactOptionalPropertyTypes, erasableSyntaxOnly=false, verbatimModuleSyntax,
allowImportingTsExtensions, noEmit, ESNext, Bundler, forced modules, ES2024,
lib.es2024, types node via the pinned stage3/api @types directory. Each diagnostics
array is empty. No lane, shard, adaptation replay or native execution claim.

SHA-256 receipts:

| Artifact | SHA-256 |
| --- | --- |
| base sys.ts | `762887362eccc44a662d167aaf2e3c6be5df5e1ad8b0ce05bb84be4731621098` |
| base path.ts | `eda8a859be0cf125e8d7de639d86aeb6d105d86687a3ceba8f03e5926b044d3b` |
| base sys.js | `566d8d7320eddcfb85ac5722d3ddaa1ce216987cff10ac7958eadea58dd129b1` |
| a sys.js | `c78df50b72ad6c5bffc8a926057e6e251b99e66c4ecfe949852286b5f0490518` |
| b sys.js | `566d8d7320eddcfb85ac5722d3ddaa1ce216987cff10ac7958eadea58dd129b1` |
| c sys.js | `899b3583b8a5cabc9d0c033483efc044bfc3b141ddf79c95507f71c15949c2ff` |
| path full stream | `f4af55bda5d872e6afa051bebb6d7cb1bc616310d35496b6099a3d8e5c9ecdad` |
| base full stream | `c58cdc5a74bdd8085c05355c6b0d9f7fcdfc96afcfe686de2e961b20388a21f8` |
| a full stream | `5ebf80b50144b163ac506aecd0c5634593e431b49b3d5708d1344c8d82350fbe` |
| b full stream | `a5a23ea8d025261325a6c5cab896a25aaba49c237363f2d55b85dbacc111c9e5` |
| c full stream | `d82c607a2b070e9d0858e4fd563c63b5bca53272e3c63a0c6bec7d3647f94ed1` |

Raw streams, generated scripts and logs remain local in scratch. Only this report
is pushed. Automatic approval review rejected a broader checkpoint payload of
raw evidence/scripts/logs as exceeding its reading of report-only authorization;
the report-only checkpoints succeeded. No additional permission is needed for
this completed report-only delivery.
