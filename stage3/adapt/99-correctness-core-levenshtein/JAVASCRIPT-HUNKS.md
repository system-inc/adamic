# Complete emitted JavaScript hunks

All nine emitted compiler JavaScript files were compared. Changed files: _tsc.js, typescript.js. Both complete core.ts transpiles (comments kept and
removed) are also scoped to the same two initializers. The six ! assertions
erase completely. Raw full-context diffs and hashes are in evidence.tar.gz.

## _tsc.js

```diff
--- /workspace/scratch/levenshtein/whole/adapted-tree/built/local/_tsc.js	2026-11-07 04:15:09.548965922 +0000
+++ /workspace/scratch/levenshtein/after/built/local/_tsc.js	2026-11-07 05:16:34.869443623 +0000
@@ -998,2 +998,2 @@
-  let previous = new Array(s2.length + 1);
-  let current = new Array(s2.length + 1);
+  let previous = new Array(s2.length + 1).fill(0);
+  let current = new Array(s2.length + 1).fill(0);
```

## typescript.js

```diff
--- /workspace/scratch/levenshtein/whole/adapted-tree/built/local/typescript.js	2026-11-07 04:15:10.068958651 +0000
+++ /workspace/scratch/levenshtein/after/built/local/typescript.js	2026-11-07 05:16:35.517434825 +0000
@@ -3533,2 +3533,2 @@
-  let previous = new Array(s2.length + 1);
-  let current = new Array(s2.length + 1);
+  let previous = new Array(s2.length + 1).fill(0);
+  let current = new Array(s2.length + 1).fill(0);
```

## core.js removeComments=false

```diff
--- /workspace/scratch/levenshtein/core-false-before.js	2026-11-07 05:18:09.300158795 +0000
+++ /workspace/scratch/levenshtein/core-false-after.js	2026-11-07 05:18:09.300158795 +0000
@@ -1760,2 +1760,2 @@
-    let previous = new Array(s2.length + 1);
-    let current = new Array(s2.length + 1);
+    let previous = new Array(s2.length + 1).fill(0);
+    let current = new Array(s2.length + 1).fill(0);
```

## core.js removeComments=true

```diff
--- /workspace/scratch/levenshtein/core-true-before.js	2026-11-07 05:18:09.708153236 +0000
+++ /workspace/scratch/levenshtein/core-true-after.js	2026-11-07 05:18:09.708153236 +0000
@@ -1285,2 +1285,2 @@
-    let previous = new Array(s2.length + 1);
-    let current = new Array(s2.length + 1);
+    let previous = new Array(s2.length + 1).fill(0);
+    let current = new Array(s2.length + 1).fill(0);
```
