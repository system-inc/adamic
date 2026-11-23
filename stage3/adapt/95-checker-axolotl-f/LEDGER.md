## 3. Excess JavaScript type argument crashes getTypeArgumentConstraint

Status: fixed in our fork by JavaScript-changing adaptation 95f under the
November 1 Correctness ruling; draft issue, not filed.
Upstream: TypeScript 6.0.3, commit 050880ce59e30b356b686bd3144efe24f875ebc8.
Location: src/compiler/checker.ts:42849, getTypeArgumentConstraint calls
getConstraintOfTypeParameter with typeParameters[typeArgumentPosition] even
when the argument position exceeds the generic arity.

Node counterexample: a one-parameter JavaScript generic C and JSDoc
`@type {C<string, number, boolean>}` or `@extends {C<string, number, boolean>}`.
Querying the second or third argument through TypeChecker.getTypeArgumentConstraint
throws TypeError reading resolvedBaseConstraint. JavaScript's class-reference
resolution deliberately continues after invalid arity, so the error-type early
return used for TypeScript does not establish parameter presence here.

Our patch: stage3/adapt/95-checker-axolotl-f/adapt.cjs inserts only:

```diff
 const relevantTypeParameter = typeParameters[typeArgumentPosition];
+if (relevantTypeParameter === undefined) return undefined;
 const constraint = getConstraintOfTypeParameter(relevantTypeParameter);
```

The API already returns Type | undefined; internal and service callers already
handle no constraint. This is a crash fix, not a required-value assertion.
Official typescript-go 89d5d5b2849a0db0957065889ca58536fa6d2e4a and TypeScript
main 6161772c1370bdd4a82f808e6823bd755eb1870d both return nil past arity in
the equivalent Go checker function. Evidence is on branch
typescript/axolotl-checker-f-evidence under stage3/evidence/95-checker-axolotl-f/.

### Draft GitHub issue

Title: getTypeArgumentConstraint crashes on excess JSDoc type arguments in TypeScript 6.0.3

Version: TypeScript 6.0.3 on Node 24.19.0. Current Go main already bounds-checks
this position; this draft documents the 6.0.3 defect and fork repair.

Save the following as repro.cjs beside an installed typescript@6.0.3 and run
`node repro.cjs`:

```js
const ts = require('typescript');
const file = '/repro.js';
const text = '/** @template T */ class C {}\n/** @type {C<string, number, boolean>} */ let value;';
const host = ts.createCompilerHost({allowJs: true, checkJs: true, noEmit: true});
const original = host.getSourceFile;
host.getSourceFile = (name, ...args) => name === file
    ? ts.createSourceFile(name, text, ts.ScriptTarget.Latest, true)
    : original.call(host, name, ...args);
const program = ts.createProgram([file], {allowJs: true, checkJs: true, noEmit: true}, host);
const checker = program.getTypeChecker();
const source = program.getSourceFile(file);
const argument = source.statements[1].jsDoc[0].tags[0].typeExpression.type.typeArguments[1];
console.log(checker.getTypeArgumentConstraint(argument));
```

Expected: undefined (the excess argument has no corresponding parameter and
therefore no constraint). Actual: TypeError: Cannot read properties of undefined
(reading 'resolvedBaseConstraint'). Suggested change: handle the absent local
before calling getConstraintOfTypeParameter, as the Go checker already does.

This entry and the independent G entry both append to 95a's ledger end. Resolve
the append conflict when merging the second branch; preserve both issue drafts.
