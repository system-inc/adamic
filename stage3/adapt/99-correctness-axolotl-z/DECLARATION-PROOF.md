# 99z declaration amendment proof, k220r1g

Built on origin/main `e8f3723bd41abcebedabfcc7546fdf104a9bbf08`. Compiler accepts this established-shape
amendment under Adamic's delegation, task b5w3ycg, Nov 7 03:20Z.
No compiler or lane-runner source is changed. The adapter remains
javascriptChanging true. The only upstream source change is elementAt's
const-arrow owner becoming the generic function declaration; the fallback
body and attached JSDoc are byte-identical.

## Completed checks

| Check | Result |
| --- | --- |
| Full baseline apply from isolated main checkout | Exit 0 |
| Full amended apply | Exit 0 |
| Owner contract | 39 checks pass; pristine and previous const accepted; declaration replay preserves bytes and mtime; half owners refuse without writes; LF/CRLF and JSDoc preserved |
| Paired complete-tree replay | 95 adapters; every other adapter has identical exit and normalized output; only 99z reports one edit before and zero after |
| Ordinary adamic types, all 79 sorted roots together | Before and after exit 0, zero diagnostics |
| Ordinary compiler builds | Both exit 0 |
| Source scope | All 746 files compared; only compiler/core.ts differs |
| All nine emitted compiler JavaScript files | Only _tsc.js and typescript.js differ; their bodies and every other executable statement match after removing only elementAt's owner |
| typescript.d.ts and tsserverlibrary.d.ts | Byte-identical |
| typescript.internal.d.ts | Exactly two hunks, shown below |
| Real amended whole lane | 398 units exactly once in 20 pieces; 106,366 passing, one sanctioned failure, zero skips; every unit report passes |
| Core.ts site census | Before, after and mutant exit 0 with zero checker diagnostics; F9 gone, no new Refused/NotYet sites, no elementAt or toOffset findings |
| Const-restoring mutant | Restores F9, core.ts:1078:5, NotYet: a function returning T \| undefined |
| Fresh Node P2 witnesses | All 30 rows and 30 expected-value/identity checks pass; extracted before and after owners agree byte for byte |

The complete before source is byte-identical to the source in accepted
measurement commit `5ab9992068bb004caacc8637b49a319533b5a002`.
Its accepted counts remain 11 call sites, 1,936 calls, 0 differences, and
0 each of early reads, new, prototype/name/length reads, value uses,
reassignments, and ts.elementAt reads. This run exercises the actual amended
compiler through the complete lane; it does not substitute a shadow helper.

All lane identity, baseline, API, failure and count guards remain enabled.
The largest piece is 125.630 seconds. Each piece checks frozen src and
built/local bytes, retains each child's own exit code and full unit logs,
and has its own SHA256SUMS. Independent reconciliation reopens every archive,
rebuilds raw test identities and counts, verifies unit input hashes and
baseline guards, and rejects duplicates or missing units.

Paired replay preserves 14 existing nonzero results on both sides.
These matched refusals are recorded explicitly, not claimed as successful
idempotent replays of those other adapters. Their full outputs and the
95-row comparison are retained in replay.json. 99z itself applies to the
old output and replays the declaration with no write.

The site census uses LATENT_FULL=1, LATENT_ONLY_FILE=core.ts, and
LATENT_ASSERT_NO_OUTPUT=1, with no speculative or replay flags. The guarded
measurement binary cannot return usable IR. Its legacy checker-rejected
label is retained, while actual checker_rejected metadata is false.
Both elementAt and toOffset are deferred generic-instantiation units, with
zero findings. No claim of a concrete native compiler build follows from
that deferral. Site comparison maps later line coordinates across the four
removed owner lines; kind, reason, phase and source locations otherwise
match exactly. The sole removed blocker is F9. Restoring the previous const
makes the mutant's core.ts byte-identical to the baseline and restores F9.

P2 retains interface Item, number and string arrays, offsets -1, 0, 1, 5,
-5, and undefined arrays. The fresh witness sources extract each actual
full-applied owner and toOffset, then reuse the prior P2 driver. Their
90 output lines include a label, value and identity/equality assertion per row.

## Every changed emitted JavaScript hunk

All nine output files are inventoried in comparison.log with both hashes.
_tsc.js and typescript.js each have two hunks. The other seven are identical:
_tsserver.js, _typingsInstaller.js, tsc.js, tsserver.js, tsserverlibrary.js,
typingsInstaller.js and watchGuard.js. Core.ts is also transpiled with
comments retained and removed; all hunks from both modes follow. The AST
comparison proves the elementAt body and the remaining executable program
are unchanged; only its binding form and placement differ.

### _tsc.js

```diff
--- before/_tsc.js
+++ after/_tsc.js
@@ -583,7 +583,7 @@
   }
   return true;
 }
-var elementAt = (array, offset) => {
+function elementAt(array, offset) {
   if (array !== void 0) {
     offset = toOffset(array, offset);
     if (offset < array.length) {
@@ -591,7 +591,7 @@
     }
   }
   return void 0;
-};
+}
 function firstOrUndefined(array) {
   return array === void 0 || array.length === 0 ? void 0 : array[0];
 }
```

### typescript.js

```diff
--- before/typescript.js
+++ after/typescript.js
@@ -2937,7 +2937,7 @@
   }
   return true;
 }
-var elementAt = (array, offset) => {
+function elementAt(array, offset) {
   if (array !== void 0) {
     offset = toOffset(array, offset);
     if (offset < array.length) {
@@ -2945,7 +2945,7 @@
     }
   }
   return void 0;
-};
+}
 function firstOrUndefined(array) {
   return array === void 0 || array.length === 0 ? void 0 : array[0];
 }
```

### core-js-false

```diff
--- before/core-js-false
+++ after/core-js-false
@@ -804,7 +804,7 @@
  *
  * @internal
  */
-export const elementAt = (array, offset) => {
+export function elementAt(array, offset) {
     if (array !== undefined) {
         offset = toOffset(array, offset);
         if (offset < array.length) {
@@ -812,7 +812,7 @@
         }
     }
     return undefined;
-};
+}
 /**
  * Returns the first element of an array if non-empty, `undefined` otherwise.
  *
```

### core-js-true

```diff
--- before/core-js-true
+++ after/core-js-true
@@ -649,7 +649,7 @@
     }
     return true;
 }
-export const elementAt = (array, offset) => {
+export function elementAt(array, offset) {
     if (array !== undefined) {
         offset = toOffset(array, offset);
         if (offset < array.length) {
@@ -657,7 +657,7 @@
         }
     }
     return undefined;
-};
+}
 export function firstOrUndefined(array) {
     return array === undefined || array.length === 0 ? undefined : array[0];
 }
```

## Both internal declaration hunks

```diff
--- before/typescript.internal.d.ts
+++ after/typescript.internal.d.ts
@@ -7243,6 +7243,13 @@
     /** @internal */
     function rangeEquals<T>(array1: readonly T[], array2: readonly T[], pos: number, end: number): boolean;
     /**
+     * Returns the element at a specific offset in an array if non-empty, `undefined` otherwise.
+     * A negative offset indicates the element should be retrieved from the end of the array.
+     *
+     * @internal
+     */
+    function elementAt<T>(array: readonly T[] | undefined, offset: number): T | undefined;
+    /**
      * Returns the first element of an array if non-empty, `undefined` otherwise.
      *
      * @internal
@@ -7755,13 +7762,6 @@
     const emptyArray: never[];
     /** @internal */
     const emptyMap: ReadonlyMap<never, never>;
-    /**
-     * Returns the element at a specific offset in an array if non-empty, `undefined` otherwise.
-     * A negative offset indicates the element should be retrieved from the end of the array.
-     *
-     * @internal
-     */
-    const elementAt: <T>(array: readonly T[] | undefined, offset: number) => T | undefined;
     /** @internal */
     type MultiMap<K, V> = Map<K, V[]>;
     /** @internal */
```

## Evidence and reproduction

[DECLARATION-PROOF.json](DECLARATION-PROOF.json) records statuses, compiler
hashes, comparisons, counts and archive identity.
[elementat-declaration-proof.tar.gz](elementat-declaration-proof.tar.gz)
contains the complete commands and child exit codes, source manifests,
source inputs, all JavaScript/declaration diffs, 79-root type logs, paired
replay outputs, census streams and timings, Node witnesses, lane-piece
archives and their independent reconciler. It includes the pinned meter
overlay generator and its physical .go.txt inputs. Internal SHA256SUMS
covers every member; the adjacent .sha256 file covers the outer archive.
The outer archive is 6337796 bytes, below 20 MB; each lane
piece is also below 20 MB. Bulky new proof data adds only four files beside
this directory's existing historical evidence.

The raw commands preserve an early amended-apply preparation run stopped
before finalizing the owner contract. It is excluded from the two accepted
full applies. All accepted transformation inputs are frozen and hashed.
The final source and built-product manifests are rechecked at closeout.
No native binary, compiler binary, node_modules directory or full built
TypeScript product is submitted. The scripts run from the archived scratch
paths after sourcing /workspace/adamic-tools/env.sh and setting
ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source at the pinned
050880ce59e30b356b686bd3144efe24f875ebc8 checkout. No upstream issue was sent.
