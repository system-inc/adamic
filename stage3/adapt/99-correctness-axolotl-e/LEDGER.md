## 5. Empty repeated extends clause crashes declaration emit

Status: fixed in our fork by adaptation 99e; draft issue, not filed.
Upstream: TypeScript 6.0.3, commit 050880ce59e30b356b686bd3144efe24f875ebc8.
Location: src/compiler/transformers/declarations.ts:1678, the heritage callback
inside transformTopLevelDeclaration's expression-base rewrite.

Node counterexample input: `export class C extends f() extends {}`.
A further reduced standalone witness is `class C extends 0 extends {}`; it
also crashes both CLI declaration emit and transpileDeclaration, with TS2507
for the numeric base and TS1172 for the repeated clause.
The parser reports no diagnostics and creates heritage type lengths `[1, 0]`.
Semantic checking reports TS1172, `'extends' clause already seen.`, for the
second clause. `tsc --declaration`, `ts.transpileDeclaration`, and program.emit
throw `TypeError: Cannot read properties of undefined (reading 'kind')` in
isVariableDeclaration, called by createGetSymbolAccessibilityDiagnosticForNode.
The outer getEffectiveBaseTypeNode guard proves only the first clause's type.

Our patch: stage3/adapt/99-correctness-axolotl-e/adapt.cjs:

```diff
- const heritageClauses = factory.createNodeArray(map(input.heritageClauses, clause => {
+ const heritageClauses = factory.createNodeArray(mapDefined(input.heritageClauses, clause => {
      if (clause.token === SyntaxKind.ExtendsKeyword) {
+         const firstType = clause.types[0];
+         if (!firstType) return undefined;
          const oldDiag = getSymbolAccessibilityDiagnostic;
-         getSymbolAccessibilityDiagnostic = createGetSymbolAccessibilityDiagnosticForNode(clause.types[0]);
+         getSymbolAccessibilityDiagnostic = createGetSymbolAccessibilityDiagnosticForNode(firstType);
```

The empty clause is omitted, matching the ordinary heritage transformation's
empty-clause filtering. Nonempty clauses retain their existing treatment.
The declaration becomes `declare const C_base: any;` followed by
`export declare class C extends C_base {}` and `export {};`.
Node witnesses, full stacks, control equality, diagnostic counts and baseline
results are recorded in the adaptation's README and proofs.

The pinned native compiler at d92d9bfee114c80be2c375d72edae966176e3a4f
and TypeScript main at 6161772c1370bdd4a82f808e6823bd755eb1870d use a
first-effective-base rewrite and elide empty clauses. The legacy release-6.0
branch still has the unsafe JavaScript transform; no JavaScript fix commit
is claimed. Both native and fixed JavaScript emit the same declaration for
the original f() witness. On the reduced numeric witness, native emits a
literal base type 0 while TypeScript 6 emits number, matching its unchanged
single-clause control. This patch does not alter type inference.

### Draft GitHub issue

Title: Repeated empty extends clause crashes declaration emit in TypeScript 6.0.3

Version: TypeScript 6.0.3, reproduced on Node 24.19.0.

Minimal reproduction, in a fresh directory without tsconfig.json:

```sh
npm install --no-audit --no-fund typescript@6.0.3
printf '%s\n' 'export class C extends f() extends {}' > input.ts
node node_modules/typescript/lib/tsc.js --declaration input.ts
```

API reproduction:

```js
const ts = require("typescript");
ts.transpileDeclaration("export class C extends f() extends {}", { fileName: "input.ts" });
```

Expected: report applicable diagnostics and finish declaration emit without an
exception. The empty repeated clause should be omitted, consistently with the
ordinary heritage transformer and the native compiler.

Actual: both commands throw a TypeError reading kind on undefined, from
isVariableDeclaration through createGetSymbolAccessibilityDiagnosticForNode,
map's heritage callback, and transformTopLevelDeclaration. The source parses
without errors; semantic checking reports TS1172 for the second extends clause
and TS2304 for the undeclared f. Declaration emit crashes even though
noEmitOnError is false, its default.

Suggested upstream change for the legacy implementation: use mapDefined for
the expression-base heritage mapping, return undefined for an empty extends
clause, and pass the checked first type to the diagnostic helper. Native main
already handles this case; this draft concerns the TypeScript 6 JavaScript
compiler. No claim is made that erroneous input must yield a valid program.
