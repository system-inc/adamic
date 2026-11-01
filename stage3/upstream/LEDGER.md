# TypeScript upstream bug ledger

Fork fixes stay separate from soundness adaptations. Kirk decides when to file
these drafted issues. Nothing in this ledger has been sent upstream.

## 1. Empty single-quoted triple-slash pragma argument throws

Status: fixed in our fork by adaptation 46; draft issue, not filed.
Upstream: TypeScript 6.0.3, commit 050880ce59e30b356b686bd3144efe24f875ebc8.
Location: src/compiler/parser.ts:10731 selects the argument using
`matchResult[2] || matchResult[3]`; parser.ts:10737 reads `value.length` and throws.
The quote-alternative matcher is constructed at parser.ts:10707.

Node counterexample input: `/// <reference path='' />`.
Before: `createSourceFile` throws
`TypeError: Cannot read properties of undefined (reading 'length')` in
extractPragmas. The successful match has group 2 equal to `""` and group 3
undefined, so truthy selection discards the participating empty capture.
After: parsing succeeds, referencedFiles has one entry whose fileName is `""`
and whose capture span has length 0. Nonempty single/double quotes and empty
double quotes retain their previous results.

Our patch: stage3/adapt/46-fix-pragma-empty-argument/adapt.cjs changes only the
operator in adaptation 45's checked boundary:

```diff
- const value = (matchResult[2] || matchResult[3])!;
+ const value = (matchResult[2] ?? matchResult[3])!;
```

Groups 2 and 3 are alternative quote branches and exactly one participates in
every successful match. Either participating capture may be empty. Nullish
selection keeps the empty string, and the entire result is therefore present.
Neither individual capture is asserted mandatory.

### Draft GitHub issue

Title: Empty single-quoted reference path crashes createSourceFile in TypeScript 6.0.3

Version: TypeScript 6.0.3, reproduced on Node 24.19.0.

Minimal reproduction:

```sh
mkdir tsc-empty-reference-repro
cd tsc-empty-reference-repro
npm install --no-audit --no-fund typescript@6.0.3
```

Save this as repro.cjs and run `node repro.cjs`:

```js
const ts = require("typescript");
const source = ts.createSourceFile("empty.ts", "/// <reference path='' />", ts.ScriptTarget.Latest);
console.log(JSON.stringify(source.referencedFiles.map(file => file.fileName)));
console.log(source.referencedFiles[0].end - source.referencedFiles[0].pos);
```

Expected: the parser completes and represents the empty value consistently
with a double-quoted empty reference path, printing:

```text
[""]
0
```

Actual: the parser throws
`TypeError: Cannot read properties of undefined (reading 'length')` in
extractPragmas. The equivalent `/// <reference path="" />` parses successfully.

Suggested upstream change: use `matchResult[2] ?? matchResult[3]` when choosing
between the single-quoted and double-quoted argument captures. The `!` in our
fork patch belongs to our sounder regex declarations; it is not needed in
upstream's standard string-valued capture declarations.

## 2. Truthiness treats a returned exit code as a watch handle

Status: soundness adaptation 79 under the Oct 19 language-lead ruling for
#se1zcgr; draft issue, not filed.
Upstream: TypeScript 6.0.3, commit 050880ce59e30b356b686bd3144efe24f875ebc8.
Locations: src/compiler/sys.ts System.exit; src/compiler/executeCommandLine.ts;
src/testRunner/unittests/helpers/tscWatch.ts:78 and
tscWatch/incremental.ts:77 (locations in the numbered adapted tree).

System.exit promises void, but test hosts override it with
`sys.exit = exitCode => sys.exitCode = exitCode`. executeCommandLine forwards
that value through its void arm. The harness treats the result as void or a
watch/solution object, and `if (closer) closer.close()` incorrectly excludes
all numbers using the declared void arm. A nonzero exit code is truthy.

Node 24.19.0 counterexample, using the real built TypeScript compiler:

```js
const ts = require("./built/local/typescript.js");
const sys = { ...ts.sys, write() {}, exit(_code) { return 1; } };
const closer = ts.executeCommandLine(sys, () => {}, ["--invalid-adamic-option"]);
console.log(closer); // 1
if (closer) closer.close(); // TypeError: closer.close is not a function
```

Observed: the invalid-option path calls exit with 1, executeCommandLine returns
1, and the harness's exact old close statement throws the TypeError above.
[The runnable proof](../adapt/79-exit-code/proofs/host-witness.cjs) extracts that
statement from the actual consumer, rather than maintaining a second copy.
This is a counterexample to the narrowing, not evidence that the unchanged
upstream suite exercises that invalid-option input in this particular consumer.

Our adaptation declares System.exit as number | undefined, gives the two
noop-as-exit test hosts dedicated stubs with that signature, and narrows the
two consumers using `typeof result === "object"`. Compiler JavaScript stays
byte-identical. The full unchanged upstream suite and its exact passing count
are the condition on accepting the harness changes; no test or baseline input
is added, deleted, filtered, or rewritten by adaptation 79.

## 3. Empty repeated extends clause crashes declaration emit

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
