# Emitted JavaScript review hunk

Whitespace-normalized for display. Exact bundle and source hunks are retained
in elementat-step-b.tar.gz, with hashes and scoped comparisons.

```diff
--- /workspace/scratch/elementat-b/whole-control/adapted-tree/built/local/typescript.js	2026-11-06 22:39:12.557628752 +0000
+++ /workspace/scratch/elementat-b/step-b/after/built/local/typescript.js	2026-11-06 23:02:13.067136158 +0000
@@ -2936,7 +2936,7 @@
   }
   return true;
 }
-var elementAt = !!Array.prototype.at ? (array, offset) => array == null ? void 0 : array.at(offset) : (array, offset) => {
+var elementAt = (array, offset) => {
   if (array !== void 0) {
     offset = toOffset(array, offset);
     if (offset < array.length) {
```
