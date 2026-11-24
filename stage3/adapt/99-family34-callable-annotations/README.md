# Give three utilities callbacks their receiving return types

Compiler's Nov 22 ruling on #bbssdcy covers utilities.ts:902:29, :2325:38
and :2332:38. The diagnostic callback uses the existing
`import("./types.js").RepopulateDiagnosticChainInfo` alias. The two
`findAncestor` predicates declare `boolean | "quit"` explicitly.
The adapter checks each enclosing AST expression and exact callback text.

Compiler permits these two emitted JavaScript changes. Stock TypeScript
6.0.3 emits the literal-result callback byte-identically. The two
predicate lines in `compiler/utilities.js` are:

```diff
@@ emitted utilities.js:1555 @@
-    return findAncestor(node.parent, n => !!(getContainerFlags(n) & 1 /* ContainerFlags.IsContainer */));
+    return findAncestor(node.parent, (n) => !!(getContainerFlags(n) & 1 /* ContainerFlags.IsContainer */));
@@ emitted utilities.js:1561 @@
-    return findAncestor(node.parent, current => isBlockScope(current, current.parent));
+    return findAncestor(node.parent, (current) => isBlockScope(current, current.parent));
```

The adapter compares the complete stock JavaScript against those exact
replacements and compares the public declaration output byte for byte.
Rulings 347 and 363 to 365 explain the storage comparison. Row 392 says
“NotYet means intended support.” Row 402 gives the meaning-preserving
meter adaptation constraint.
