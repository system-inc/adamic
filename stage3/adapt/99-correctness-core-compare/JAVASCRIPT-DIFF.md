# Every emitted JavaScript hunk

All nine emitted files were compared. Each nonempty hunk is below; the AST comparison also proves equality outside the complete compareComparableValues owner. Both standalone core.ts comment modes are included.

## _tsc.js.diff

```diff
--- /workspace/scratch/core-compare/before/built/local/_tsc.js	2026-11-07 05:02:08.237492877 +0000
+++ /workspace/scratch/core-compare/final-after/built/local/_tsc.js	2026-11-07 05:17:23.028455594 +0000
@@ -931,7 +931,7 @@
   return equateValues(a, b);
 }
 function compareComparableValues(a, b) {
-  return a === b ? 0 /* EqualTo */ : a === void 0 ? -1 /* LessThan */ : b === void 0 ? 1 /* GreaterThan */ : a < b ? -1 /* LessThan */ : 1 /* GreaterThan */;
+  return a === b ? 0 /* EqualTo */ : a === void 0 ? -1 /* LessThan */ : b === void 0 ? 1 /* GreaterThan */ : typeof a === "number" && typeof b === "number" ? a < b ? -1 /* LessThan */ : 1 /* GreaterThan */ : typeof a === "string" && typeof b === "string" ? a < b ? -1 /* LessThan */ : 1 /* GreaterThan */ : 1 /* GreaterThan */;
 }
 function compareValues(a, b) {
   return compareComparableValues(a, b);
```

## core-false.diff

```diff
--- before/core.js
+++ after/core.js
@@ -1552,8 +1552,8 @@
     return a === b ? Comparison.EqualTo :
         a === undefined ? Comparison.LessThan :
             b === undefined ? Comparison.GreaterThan :
-                a < b ? Comparison.LessThan :
-                    Comparison.GreaterThan;
+                typeof a === "number" && typeof b === "number" ? (a < b ? Comparison.LessThan : Comparison.GreaterThan) :
+                    typeof a === "string" && typeof b === "string" ? (a < b ? Comparison.LessThan : Comparison.GreaterThan) : Comparison.GreaterThan;
 }
 /**
  * Compare two numeric values for their order relative to each other.
```

## core-true.diff

```diff
--- before/core.js
+++ after/core.js
@@ -1170,8 +1170,8 @@
     return a === b ? Comparison.EqualTo :
         a === undefined ? Comparison.LessThan :
             b === undefined ? Comparison.GreaterThan :
-                a < b ? Comparison.LessThan :
-                    Comparison.GreaterThan;
+                typeof a === "number" && typeof b === "number" ? (a < b ? Comparison.LessThan : Comparison.GreaterThan) :
+                    typeof a === "string" && typeof b === "string" ? (a < b ? Comparison.LessThan : Comparison.GreaterThan) : Comparison.GreaterThan;
 }
 export function compareValues(a, b) {
     return compareComparableValues(a, b);
```

## typescript.js.diff

```diff
--- /workspace/scratch/core-compare/before/built/local/typescript.js	2026-11-07 05:02:09.117477869 +0000
+++ /workspace/scratch/core-compare/final-after/built/local/typescript.js	2026-11-07 05:17:23.940443241 +0000
@@ -3431,7 +3431,7 @@
   return equateValues(a, b);
 }
 function compareComparableValues(a, b) {
-  return a === b ? 0 /* EqualTo */ : a === void 0 ? -1 /* LessThan */ : b === void 0 ? 1 /* GreaterThan */ : a < b ? -1 /* LessThan */ : 1 /* GreaterThan */;
+  return a === b ? 0 /* EqualTo */ : a === void 0 ? -1 /* LessThan */ : b === void 0 ? 1 /* GreaterThan */ : typeof a === "number" && typeof b === "number" ? a < b ? -1 /* LessThan */ : 1 /* GreaterThan */ : typeof a === "string" && typeof b === "string" ? a < b ? -1 /* LessThan */ : 1 /* GreaterThan */ : 1 /* GreaterThan */;
 }
 function compareValues(a, b) {
   return compareComparableValues(a, b);
```
