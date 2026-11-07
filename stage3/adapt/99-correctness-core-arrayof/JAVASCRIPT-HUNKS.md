# Complete JavaScript and source hunks

Blank context-line marker spaces are omitted in this display. The raw .diff
files in proof.tar.gz retain the exact unified diffs.

All nine emitted compiler JavaScript files were checked. Only _tsc.js and
typescript.js differ, inside arrayOf. Both complete core.ts emits, with
comments kept and removed, also differ only in that owner.

## source.diff

```diff
--- /workspace/scratch/arrayof/before/src/compiler/core.ts	2026-11-07 03:46:17.348880597 +0000
+++ /workspace/scratch/arrayof/after/src/compiler/core.ts	2026-11-07 04:37:43.357547256 +0000
@@ -1332,9 +1332,9 @@

 /** @internal */
 export function arrayOf<T>(count: number, f: (index: number) => T): T[] {
-    const result = new Array(count);
+    const result: T[] = [];
     for (let i = 0; i < count; i++) {
-        result[i] = f(i);
+        result.push(f(i));
     }
     return result;
 }
```

## core-false.diff

```diff
--- /workspace/scratch/arrayof/core-false-before.js	2026-11-07 04:41:05.738829692 +0000
+++ /workspace/scratch/arrayof/core-false-after.js	2026-11-07 04:41:05.738829692 +0000
@@ -1019,9 +1019,9 @@
 }
 /** @internal */
 export function arrayOf(count, f) {
-    const result = new Array(count);
+    const result = [];
     for (let i = 0; i < count; i++) {
-        result[i] = f(i);
+        result.push(f(i));
     }
     return result;
 }
```

## core-true.diff

```diff
--- /workspace/scratch/arrayof/core-true-before.js	2026-11-07 04:41:05.954826793 +0000
+++ /workspace/scratch/arrayof/core-true-after.js	2026-11-07 04:41:05.954826793 +0000
@@ -788,9 +788,9 @@
     return values;
 }
 export function arrayOf(count, f) {
-    const result = new Array(count);
+    const result = [];
     for (let i = 0; i < count; i++) {
-        result[i] = f(i);
+        result.push(f(i));
     }
     return result;
 }
```

## _tsc.js.diff

```diff
--- /workspace/scratch/arrayof/before/built/local/_tsc.js	2026-11-07 03:47:08.700184281 +0000
+++ /workspace/scratch/arrayof/after/built/local/_tsc.js	2026-11-07 04:40:38.831190818 +0000
@@ -701,9 +701,9 @@
   return values;
 }
 function arrayOf(count, f) {
-  const result = new Array(count);
+  const result = [];
   for (let i = 0; i < count; i++) {
-    result[i] = f(i);
+    result.push(f(i));
   }
   return result;
 }
```

## typescript.js.diff

```diff
--- /workspace/scratch/arrayof/before/built/local/typescript.js	2026-11-07 03:47:09.764169839 +0000
+++ /workspace/scratch/arrayof/after/built/local/typescript.js	2026-11-07 04:40:39.159186415 +0000
@@ -3070,9 +3070,9 @@
   return values;
 }
 function arrayOf(count, f) {
-  const result = new Array(count);
+  const result = [];
   for (let i = 0; i < count; i++) {
-    result[i] = f(i);
+    result.push(f(i));
   }
   return result;
 }
```
