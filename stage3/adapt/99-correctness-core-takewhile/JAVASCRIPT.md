# Complete JavaScript hunks

Only takeWhile changes in both full-core transpilation modes and the two
changed compiler bundles. The other seven compiler JavaScript files are
byte-identical. comparison.json records every hash and scope assertion.

## core-false.diff

```diff
--- /workspace/scratch/takewhile/raw/core-false-before.js	2026-11-07 03:56:12.274968756 +0000
+++ /workspace/scratch/takewhile/raw/core-false-after.js	2026-11-07 03:56:12.274968756 +0000
@@ -2058,12 +2058,17 @@
 }
 export function takeWhile(array, predicate) {
     if (array !== undefined) {
+        const result = [];
         const len = array.length;
         let index = 0;
-        while (index < len && predicate(array[index])) {
+        while (index < len) {
+            const element = array[index];
+            if (!predicate(element))
+                break;
+            result.push(element);
             index++;
         }
-        return array.slice(0, index);
+        return result;
     }
 }
 /** @internal */
```

## core-true.diff

```diff
--- /workspace/scratch/takewhile/raw/core-true-before.js	2026-11-07 03:56:12.394971650 +0000
+++ /workspace/scratch/takewhile/raw/core-true-after.js	2026-11-07 03:56:12.394971650 +0000
@@ -1520,12 +1520,17 @@
 }
 export function takeWhile(array, predicate) {
     if (array !== undefined) {
+        const result = [];
         const len = array.length;
         let index = 0;
-        while (index < len && predicate(array[index])) {
+        while (index < len) {
+            const element = array[index];
+            if (!predicate(element))
+                break;
+            result.push(element);
             index++;
         }
-        return array.slice(0, index);
+        return result;
     }
 }
 export function skipWhile(array, predicate) {
```

## _tsc.js.diff

```diff
--- /workspace/scratch/takewhile/before/built/local/_tsc.js	2026-11-07 03:40:38.316453136 +0000
+++ /workspace/scratch/takewhile/after/built/local/_tsc.js	2026-11-07 03:55:51.914477589 +0000
@@ -1193,12 +1193,16 @@
 }
 function takeWhile(array, predicate) {
   if (array !== void 0) {
+    const result = [];
     const len = array.length;
     let index = 0;
-    while (index < len && predicate(array[index])) {
+    while (index < len) {
+      const element = array[index];
+      if (!predicate(element)) break;
+      result.push(element);
       index++;
     }
-    return array.slice(0, index);
+    return result;
   }
 }
 function skipWhile(array, predicate) {
```

## typescript.js.diff

```diff
--- /workspace/scratch/takewhile/before/built/local/typescript.js	2026-11-07 03:40:38.280452268 +0000
+++ /workspace/scratch/takewhile/after/built/local/typescript.js	2026-11-07 03:55:51.998479613 +0000
@@ -3766,12 +3766,16 @@
 }
 function takeWhile(array, predicate) {
   if (array !== void 0) {
+    const result = [];
     const len = array.length;
     let index = 0;
-    while (index < len && predicate(array[index])) {
+    while (index < len) {
+      const element = array[index];
+      if (!predicate(element)) break;
+      result.push(element);
       index++;
     }
-    return array.slice(0, index);
+    return result;
   }
 }
 function skipWhile(array, predicate) {
```
