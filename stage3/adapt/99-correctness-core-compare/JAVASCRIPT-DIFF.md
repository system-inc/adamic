# Every emitted JavaScript hunk

Amended owner against original main. All nine emitted files were checked; every nonempty JavaScript hunk follows, including both standalone core comment modes.

## _tsc.js.diff

```diff
--- /workspace/scratch/core-compare-amend/before/built/local/_tsc.js	2026-11-07 05:50:40.892468205 +0000
+++ /workspace/scratch/core-compare-amend/after/built/local/_tsc.js	2026-11-07 05:51:37.687697724 +0000
@@ -931,7 +931,7 @@
   return equateValues(a, b);
 }
 function compareComparableValues(a, b) {
-  return a === b ? 0 /* EqualTo */ : a === void 0 ? -1 /* LessThan */ : b === void 0 ? 1 /* GreaterThan */ : a < b ? -1 /* LessThan */ : 1 /* GreaterThan */;
+  return a === b ? 0 /* EqualTo */ : a === void 0 ? -1 /* LessThan */ : b === void 0 ? 1 /* GreaterThan */ : typeof a === "number" && typeof b === "number" ? a < b ? -1 /* LessThan */ : 1 /* GreaterThan */ : typeof a === "string" && typeof b === "string" ? a < b ? -1 /* LessThan */ : 1 /* GreaterThan */ : Debug.fail("compareComparableValues: a string compared with a number");
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
+                    typeof a === "string" && typeof b === "string" ? (a < b ? Comparison.LessThan : Comparison.GreaterThan) : Debug.fail("compareComparableValues: a string compared with a number");
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
+                    typeof a === "string" && typeof b === "string" ? (a < b ? Comparison.LessThan : Comparison.GreaterThan) : Debug.fail("compareComparableValues: a string compared with a number");
 }
 export function compareValues(a, b) {
     return compareComparableValues(a, b);
```

## typescript.js.diff

```diff
--- /workspace/scratch/core-compare-amend/before/built/local/typescript.js	2026-11-07 05:50:41.248463373 +0000
+++ /workspace/scratch/core-compare-amend/after/built/local/typescript.js	2026-11-07 05:51:37.935694359 +0000
@@ -3431,7 +3431,7 @@
   return equateValues(a, b);
 }
 function compareComparableValues(a, b) {
-  return a === b ? 0 /* EqualTo */ : a === void 0 ? -1 /* LessThan */ : b === void 0 ? 1 /* GreaterThan */ : a < b ? -1 /* LessThan */ : 1 /* GreaterThan */;
+  return a === b ? 0 /* EqualTo */ : a === void 0 ? -1 /* LessThan */ : b === void 0 ? 1 /* GreaterThan */ : typeof a === "number" && typeof b === "number" ? a < b ? -1 /* LessThan */ : 1 /* GreaterThan */ : typeof a === "string" && typeof b === "string" ? a < b ? -1 /* LessThan */ : 1 /* GreaterThan */ : Debug.fail("compareComparableValues: a string compared with a number");
 }
 function compareValues(a, b) {
   return compareComparableValues(a, b);
```
