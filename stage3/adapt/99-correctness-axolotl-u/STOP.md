# 99u stop report and third-form measurement

Task: #b5w3ycg. This directory is a stop artifact, not a working adaptation.
`adapt.cjs` refuses immediately, before reading or writing source, and names
STOP.md. Adding this directory deliberately stops a full apply. Do not treat its
presence as evidence of an applied source change.

Base: `5ff432f962b1c7c9145533946c46a35450ebfacc`.
Evidence branch: `typescript/debug-type-claims-stop`, one commit on that base.
99t reference: `typescript/debug-visible-checks` at
`9b3410c37fd40d75161d628ac75d90740abd4d4e`. Its README.md,
references.json and callers.json were read before the initial measurements.

## Measurement scope

The original baseline is the complete main adaptation tree at the base above,
using TypeScript 6.0.3 source `050880ce59e30b356b686bd3144efe24f875ebc8`.
The full baseline apply exited 0 during the original stop investigation.
This follow-up ran no build, full apply, oracle, Adamic types, compiler census,
public declaration generation, shard generation or lane tests.

Stock TypeScript 6.0.3 checks the same sorted 79 compiler roots with strict,
exactOptionalPropertyTypes, noUncheckedIndexedAccess, verbatimModuleSyntax,
allowImportingTsExtensions and noEmit true; erasableSyntaxOnly false;
module ESNext; moduleDetection Force; moduleResolution Bundler; target ES2024;
lib lib.es2024.d.ts; types empty. noImplicitReturns and
noFallthroughCasesInSwitch are omitted, as in Adamic's measurement profile.
`diagnostics.json` retains the exact options and the original two edit forms'
diagnostics without rewriting them.

The original delete and honest-local forms used in-memory CompilerHost text.
The third form was physically written to `/workspace/scratch/99u-third/src`,
copied from the complete baseline source, and checked with noEmit. It is not
consumed by adapt.cjs. The scratch measurement script is
`/workspace/scratch/99u-third-measure.cjs`; its output is
`/workspace/scratch/99u-third-measure.log`. Complete third-form owners, owner
hashes, file diffs and all program diagnostics are embedded in diagnostics.json.
The unchanged baseline has 59 other standalone stock-program diagnostics in
this scratch setup. The third form has those 59 plus the five pins below,
64 total. This does not claim a zero-diagnostic whole-program baseline or an
Adamic checker result. All five original owners have zero diagnostics.

## Per-owner counts

Counts below are diagnostics inside each named owner, not lowering census rows.
`getDiagnostic` is nested in `createGetIsolatedDeclarationErrors`.

| Owner | Original | Delete only | Honest local | Existing guards, third form |
| --- | ---: | ---: | ---: | ---: |
| isDeclarationWithTypeParameters | 0 | 1 | 1 | 1 |
| isDeclarationWithTypeParameterChildren | 0 | 1 | 1 | 1 |
| isPrimitiveLiteralValue | 0 | 6 | 6 | 1 |
| hasInferredType | 0 | 1 | 1 | 1 |
| getDiagnostic | 0 | 10 | 10 | 1 |

The honest local is Node, or Expression for isPrimitiveLiteralValue, matching
the parameter's declared type. Node and Expression are interfaces, not closed
discriminated unions; switching on kind does not recover their concrete AST
variant fields.

## Third form

Delete all five Debug.type calls. In the PrefixUnaryExpression case, use the
existing isPrefixUnaryExpression guard through Debug.assert before reading
operator or operand. Keep the original isNumericLiteral/isBigIntLiteral tests.
In getDiagnostic, use Debug.assert with the existing tsc guard for each kind,
or the disjunction of existing guards for grouped cases, before each helper
argument or narrowed field read. Add the needed existing-guard imports.
The default's pre-existing `node as Expression` cast is removed and replaced by
Debug.assert(isExpression(node)) before the fallback helper call. This broad
expression guard is separate from the case guards and from the false X claim.
It is included only in this measurement; behavior on default inputs has not
been approved or measured. No new guard function or cast is introduced. Each
of the five measured owners contains zero cast syntax nodes after this form.

All existing assertType pins are intentionally left in place. The third form
removes the field/helper diagnostics; its only remaining diagnostics are these
five pins. It is still not accepted source and is not shipped by the adapter.

## Remaining diagnostics, verbatim

Locations are in the physically written third-form files. Each owner has
exactly the single diagnostic printed below. In particular, the first pin
claims DeclarationWithTypeParameterChildren; it is not a never pin.

### isDeclarationWithTypeParameters

`src/compiler/utilities.ts:2239`, TS2345:

```text
Argument of type 'Node' is not assignable to parameter of type 'DeclarationWithTypeParameterChildren'.
  Type 'Node' is missing the following properties from type 'JSDocTemplateTag': constraint, typeParameters, tagName
```

### isDeclarationWithTypeParameterChildren

`src/compiler/utilities.ts:2268`, TS2345:

```text
Argument of type 'Node' is not assignable to parameter of type 'never'.
```

### isPrimitiveLiteralValue

`src/compiler/utilities.ts:12051`, TS2345:

```text
Argument of type 'Expression' is not assignable to parameter of type 'never'.
```

### hasInferredType

`src/compiler/utilities.ts:12082`, TS2345:

```text
Argument of type 'Node' is not assignable to parameter of type 'never'.
```

### getDiagnostic

`src/compiler/transformers/declarations/diagnostics.ts:720`, TS2345:

```text
Argument of type 'Node' is not assignable to parameter of type 'never'.
```

## Empty pins and real oracle inputs

The baseline definition is `src/compiler/core.ts:2502`:

```ts
export function assertType<T>(_: T): void {}
```

It is an empty function, with no assertion signature. It checks nothing at
runtime. Its generic parameter merely requires assignability when checking
the call. These pins are reached on legitimate inputs for which their claims
are false, so making them runtime assertions would change valid behavior.
They are listed separately from the case guards and remain unchanged.

| Owner | Remaining pin | Historical evidence that the claim is reached falsely |
| --- | --- | --- |
| isDeclarationWithTypeParameters | assertType<DeclarationWithTypeParameterChildren>(node) | Yes. All 4,230 outside-X calls take default, rather than one of the three JSDoc true cases, and also lie outside the child union. |
| isDeclarationWithTypeParameterChildren | assertType<never>(node) | Yes. Its kind cases are exactly X's kind union; 4,230 outside-X calls take default. |
| isPrimitiveLiteralValue | assertType<never>(node) | Yes, by the historical title count and source route described below. The aggregate 1,788 outside-X calls are not a default-branch count. |
| hasInferredType | assertType<never>(node) | Yes. Its cases are exactly X's kind union; 19,422 outside-X calls take default. |
| getDiagnostic | assertType<never>(node) | Yes. Its cases are exactly X's eighteen-kind union; 26 outside-X calls take default after the early returns. |

These counts are inherited from 99t's counter-results.json, measured on
`a7b6da7f622bd768a99a443706490e80667dd547`. They are not a fresh oracle run on
5ff432f9 or the third form. The historical runs/outside-X counts are retained
in diagnostics.json with their provenance. No measurement is invented for a
pin that 99t did not instrument separately.

PrimitiveLiteral includes operator/operand restrictions as well as kind.
An outside-X prefix unary expression can return false inside its case without
ever reaching default, so 1,788 alone cannot establish a default count.
There is also source-backed default reach evidence: 99t records nine false
PrimitiveLiteral evaluations in the real oracle title
`transpile tests tests/cases/transpile/declarationComputedPropertyNames.ts transpile test declarationComputedPropertyNames has expected declaration output`.
The fixture's object D contains nine nonliteral computed names (identifiers,
property accesses, parentheses and a conditional), plus the two literal names
1 and "2". None of its computed-name expressions is PrefixUnaryExpression.
`createSyntacticTypeNodeBuilder.canGetTypeFromObjectLiteral` in
src/compiler/expressionToTypeNode.ts:1045 calls isPrimitiveLiteralValue on each
computed-name expression. Those nonliteral kinds reach the default pin. This
is a source-path inference paired with a historical title counter, not a newly
observed branch counter or a claim that all 1,788 failures reached default.

## Every line changed by the third form

These diffs are against the fully applied baseline. Blank context lines are
shown empty for Markdown; diagnostics.json retains exact unified diff text.
They include
all five owners and every import change. They are measurements only; adapt.cjs
does not contain or apply them. The three owners without field/helper errors
change only by dropping their Debug.type line and keep their one failing pin.

```diff
--- main/src/compiler/utilities.ts
+++ measured-third/src/compiler/utilities.ts
@@ -2230,7 +2230,6 @@

 /** @internal */
 export function isDeclarationWithTypeParameters(node: Node): node is DeclarationWithTypeParameters {
-    Debug.type<DeclarationWithTypeParameters>(node);
     switch (node.kind) {
         case SyntaxKind.JSDocCallbackTag:
         case SyntaxKind.JSDocTypedefTag:
@@ -2244,7 +2243,6 @@

 /** @internal */
 export function isDeclarationWithTypeParameterChildren(node: Node): node is DeclarationWithTypeParameterChildren {
-    Debug.type<DeclarationWithTypeParameterChildren>(node);
     switch (node.kind) {
         case SyntaxKind.CallSignature:
         case SyntaxKind.ConstructSignature:
@@ -12031,7 +12029,6 @@

 /** @internal */
 export function isPrimitiveLiteralValue(node: Expression, includeBigInt = true): node is PrimitiveLiteral {
-    Debug.type<PrimitiveLiteral>(node);
     switch (node.kind) {
         case SyntaxKind.TrueKeyword:
         case SyntaxKind.FalseKeyword:
@@ -12042,6 +12039,7 @@
         case SyntaxKind.BigIntLiteral:
             return includeBigInt;
         case SyntaxKind.PrefixUnaryExpression:
+            Debug.assert(isPrefixUnaryExpression(node));
             if (node.operator === SyntaxKind.MinusToken) {
                 return isNumericLiteral(node.operand) || (includeBigInt && isBigIntLiteral(node.operand));
             }
@@ -12065,7 +12063,6 @@

 /** @internal */
 export function hasInferredType(node: Node): node is HasInferredType {
-    Debug.type<HasInferredType>(node);
     switch (node.kind) {
         case SyntaxKind.Parameter:
         case SyntaxKind.PropertySignature:
```

```diff
--- main/src/compiler/transformers/declarations/diagnostics.ts
+++ measured-third/src/compiler/transformers/declarations/diagnostics.ts
@@ -34,6 +34,16 @@
     hasSyntacticModifier,
     type ImportEqualsDeclaration,
     type IndexSignatureDeclaration,
+    isComputedPropertyName,
+    isShorthandPropertyAssignment,
+    isSpreadAssignment,
+    isArrayLiteralExpression,
+    isSpreadElement,
+    isFunctionExpression,
+    isArrowFunction,
+    isPropertyAssignment,
+    isClassExpression,
+    isExpression,
     isAsExpression,
     isBinaryExpression,
     isBindingElement,
@@ -669,38 +679,47 @@
         if ((isPartOfTypeNode(node) || isTypeQueryNode(node.parent)) && (isEntityName(node) || isEntityNameExpression(node))) {
             return createEntityInTypeNodeError(node);
         }
-        Debug.type<WithIsolatedDeclarationDiagnostic>(node);
         switch (node.kind) {
             case SyntaxKind.GetAccessor:
             case SyntaxKind.SetAccessor:
+                Debug.assert(isGetAccessor(node) || isSetAccessor(node));
                 return createAccessorTypeError(node);
             case SyntaxKind.ComputedPropertyName:
             case SyntaxKind.ShorthandPropertyAssignment:
             case SyntaxKind.SpreadAssignment:
+                Debug.assert(isComputedPropertyName(node) || isShorthandPropertyAssignment(node) || isSpreadAssignment(node));
                 return createObjectLiteralError(node);
             case SyntaxKind.ArrayLiteralExpression:
             case SyntaxKind.SpreadElement:
+                Debug.assert(isArrayLiteralExpression(node) || isSpreadElement(node));
                 return createArrayLiteralError(node);
             case SyntaxKind.MethodDeclaration:
             case SyntaxKind.ConstructSignature:
             case SyntaxKind.FunctionExpression:
             case SyntaxKind.ArrowFunction:
             case SyntaxKind.FunctionDeclaration:
+                Debug.assert(isMethodDeclaration(node) || isConstructSignatureDeclaration(node) || isFunctionExpression(node) || isArrowFunction(node) || isFunctionDeclaration(node));
                 return createReturnTypeError(node);
             case SyntaxKind.BindingElement:
+                Debug.assert(isBindingElement(node));
                 return createBindingElementError(node);
             case SyntaxKind.PropertyDeclaration:
             case SyntaxKind.VariableDeclaration:
+                Debug.assert(isPropertyDeclaration(node) || isVariableDeclaration(node));
                 return createVariableOrPropertyError(node);
             case SyntaxKind.Parameter:
+                Debug.assert(isParameter(node));
                 return createParameterError(node);
             case SyntaxKind.PropertyAssignment:
+                Debug.assert(isPropertyAssignment(node));
                 return createExpressionError(node.initializer);
             case SyntaxKind.ClassExpression:
+                Debug.assert(isClassExpression(node));
                 return createClassExpressionError(node);
             default:
                 assertType<never>(node);
-                return createExpressionError(node as Expression);
+                Debug.assert(isExpression(node));
+                return createExpressionError(node);
         }
     }

```

## Original stop report

The following is the original report at the time of the clean stop. Statements
about no directory or commit describe that earlier state. The new third-form
measurement above follows the user's additional instructions; it does not
reclassify the stopped work as an accepted adaptation.

```text
Adaptation 99u stop report

Base: origin/main 5ff432f962b1c7c9145533946c46a35450ebfacc.
Reference 99t: typescript/debug-visible-checks 9b3410c37fd40d75161d628ac75d90740abd4d4e.
Read its README.md, references.json and callers.json before probing.

Full stage3/apply.sh on main exited 0. Log: /tmp/99u-main-apply.log.
Stock TypeScript 6.0.3 checked the complete sorted 79 compiler roots.
The probe substitutes source text only through an in-memory CompilerHost.
No adapted source or adaptation was written by the probe.

Owner                                      Before  Delete  Honest local
isDeclarationWithTypeParameters             0       1       1
isDeclarationWithTypeParameterChildren      0       1       1
isPrimitiveLiteralValue                     0       6       6
hasInferredType                             0       1       1
createGetIsolatedDeclarationErrors.getDiagnostic 0  10      10

The honest local is Node, or Expression for isPrimitiveLiteralValue, matching
its parameter's declared type. The switch and its node uses refer to that local.
Node is an interface with kind: SyntaxKind; Expression extends Node. Neither is
a discriminated union, so the switch cannot recover concrete AST variants.

isDeclarationWithTypeParameters still passes Node to an existing empty
assertType<DeclarationWithTypeParameterChildren>. Three utilities owners and
getDiagnostic pass their broad argument to assertType<never> in default.
isPrimitiveLiteralValue additionally cannot read operator or operand from
Expression. getDiagnostic cannot pass Node to eight concrete helper argument
types or read initializer from Node. TS2339 and TS2345 diagnostics are retained
in 99u-main-owner-probe.json with exact messages and line numbers.

These are owner diagnostics, not compiler census results. The standalone stock
program has 59 other diagnostics before and 78 after in this scratch setup;
this does not claim a zero-diagnostic whole-program baseline or an Adamic types
result. The five baseline owners themselves have zero diagnostics.

Stop: meeting type checking needs changes beyond deleting Debug.type or
switching on an honestly typed broad local. No cast, replacement check of X,
or further owner rewrite was attempted. The existing assertType calls were
not deleted. No new directory, commit or remote branch was created. The local
requested branch points at the base, and the repository is clean.

Mutant lowering control: not run, no accepted deletion.
Oracle identity check: not run, no accepted adapted candidate.
Public d.ts, Adamic types, compiler census, replay and lane proofs for a new
adaptation: not run. This is a stop report, not a completed adaptation.
```

## Refusal verification

The adapter is an unconditional top-level throw. It imports no modules, reads
no source and writes no files. A scratch invocation with fs mutation APIs
replaced by throwing traps exits 1, names STOP.md, and reaches none of the
traps. The executable adapter contains no third-form edits. No build was run.
