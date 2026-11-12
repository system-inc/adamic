# Debug assertion overloads — task 68x1t2y

Measured main `219dbb6e7e20ea99d3e5d4b61c0cb14102dc9205` and exactly the same adapted tree as `overload-results-main` (all 747 file hashes verified). The census uses the existing own-tree entry overlay and binary, unsplit with no shard selectors. This directory contains evidence and scratch instrumentation only.

## 1. Single-file full-mode census

`LATENT_FULL=1`, `LATENT_ONLY_FILE=/workspace/orm-adapted/src/compiler/transformers/destructuring.ts`, `LATENT_ASSERT_NO_OUTPUT=1`; entry root remains adapted `src/tsc/tsc.ts`. The stock run has zero checker diagnostics and 26 measured units. It records **8 overload-parameter observations**, at **3 unique declaration/reason sites**. Every matching observation follows; locations are in the frozen adapted tree.

| Unit | Unit location | Finding location | Exact reason |
| --- | --- | --- | --- |
| flattenDestructuringBinding | src/compiler/transformers/destructuring.ts:244:1 | src/compiler/debug.ts:281:82 | overload 1 of assertNode parameter test cannot be served by implementation parameter test |
| emitBindingOrAssignment | src/compiler/transformers/destructuring.ts:318:5 | src/compiler/debug.ts:281:82 | overload 1 of assertNode parameter test cannot be served by implementation parameter test |
| flattenBindingOrAssignmentElement | src/compiler/transformers/destructuring.ts:338:1 | src/compiler/visitorPublic.ts:123:5 | overload 1 of visitNode parameter node cannot be served by implementation parameter node |
| flattenObjectBindingOrAssignmentPattern | src/compiler/transformers/destructuring.ts:386:1 | src/compiler/visitorPublic.ts:123:5 | overload 1 of visitNode parameter node cannot be served by implementation parameter node |
| makeArrayBindingPattern | src/compiler/transformers/destructuring.ts:600:1 | src/compiler/debug.ts:259:98 | overload 4 of assertEachNode parameter test cannot be served by implementation parameter test |
| makeArrayAssignmentPattern | src/compiler/transformers/destructuring.ts:605:1 | src/compiler/debug.ts:259:98 | overload 4 of assertEachNode parameter test cannot be served by implementation parameter test |
| makeObjectBindingPattern | src/compiler/transformers/destructuring.ts:610:1 | src/compiler/debug.ts:259:98 | overload 4 of assertEachNode parameter test cannot be served by implementation parameter test |
| makeObjectAssignmentPattern | src/compiler/transformers/destructuring.ts:615:1 | src/compiler/debug.ts:259:98 | overload 4 of assertEachNode parameter test cannot be served by implementation parameter test |


Full mode includes the nested `emitBindingOrAssignment` unit at :318 as well as its enclosing `flattenDestructuringBinding` unit at :244. The two assertNode observations are the same :319 call reached in those two unit attempts. The other assertNode call at :185 is included in the checker evidence for completeness; the full-mode census does not report an overload-parameter finding for it.

## 2. Stock TypeScript 6.0.3 checker

The stock program uses the same strict options and adapted entry root, plus main’s exact bundled @types/node 25.3.3 declarations, as the census loader supplies. It has **zero diagnostics**. No adapted source is changed. `checker.json` retains every resolved declaration, instantiated signature, implementation and source hash. Overload indices are one-based among signature declarations, excluding the implementation.

The requested `assertNode` “around :244” belongs to `flattenDestructuringBinding` at :244: its nested `emitBindingOrAssignment` starts at :318 and calls the assertion at :319.

### assertEachNode at src/compiler/transformers/destructuring.ts:601:5

```ts
Debug.assertEachNode(elements, isArrayBindingElement)
```

| Fact | Checker evidence |
| --- | --- |
| Argument expression | elements |
| Argument declared type | BindingOrAssignmentElement[] |
| Argument flow type at call | BindingOrAssignmentElement[] |
| test expression | isArrayBindingElement |
| test type | (node: Node) =&gt; node is ArrayBindingElement |
| Predicate target | ArrayBindingElement |
| Resolved overload | 2 @ src/compiler/debug.ts:257:5 |
| Implementation test parameter | test: ((node: Node) =&gt; boolean) \| undefined |
| Implementation test parameter type | ((node: Node) =&gt; boolean) \| undefined |


Resolved overload declaration:

```ts
export function assertEachNode<T extends Node, U extends T>(nodes: readonly T[], test: (node: T) => node is U, message?: string, stackCrawlMark?: AnyFunction): asserts nodes is readonly U[];
```

Instantiated resolved signature:

```ts
(nodes: readonly BindingOrAssignmentElement[], test: (node: BindingOrAssignmentElement) => node is ArrayBindingElement, message?: string | undefined, stackCrawlMark?: AnyFunction | undefined): asserts nodes is readonly ArrayBindingElement[]
```

### assertNode at src/compiler/transformers/destructuring.ts:319:9

```ts
Debug.assertNode(target, isBindingName)
```

| Fact | Checker evidence |
| --- | --- |
| Argument expression | target |
| Argument declared type | BindingOrAssignmentElementTarget |
| Argument flow type at call | BindingOrAssignmentElementTarget |
| test expression | isBindingName |
| test type | (node: Node) =&gt; node is BindingName |
| Predicate target | BindingName |
| Resolved overload | 1 @ src/compiler/debug.ts:281:5 |
| Implementation test parameter | test: ((node: Node) =&gt; boolean) \| undefined |
| Implementation test parameter type | ((node: Node) =&gt; boolean) \| undefined |


Resolved overload declaration:

```ts
export function assertNode<T extends Node, U extends T>(node: T | undefined, test: (node: T) => node is U, message?: string, stackCrawlMark?: AnyFunction): asserts node is U;
```

Instantiated resolved signature:

```ts
(node: ObjectBindingPattern | ArrayBindingPattern | ObjectLiteralExpression | ArrayLiteralExpression | Identifier | ElementAccessExpression | PropertyAccessExpression | OmittedExpression | undefined, test: (node: ObjectBindingPattern | ArrayBindingPattern | ObjectLiteralExpression | ArrayLiteralExpression | Identifier | ElementAccessExpression | PropertyAccessExpression | OmittedExpression) => node is BindingName, message?: string | undefined, stackCrawlMark?: AnyFunction | undefined): asserts node is BindingName
```

The :185 call is `Debug.assertNode(target, createAssignmentCallback ? isIdentifier : isExpression)`: target is also declared `BindingOrAssignmentElementTarget`; its test expression has type `(node: Node) => node is Expression`, target `Expression`, and resolves assertNode overload 1.

The checker-resolved assertEachNode overload is **2**, whereas the refusal names **4**. Lowering’s `censusOverloads` walks **every** overload declaration, stopping at the first unserved declaration, rather than checking only the signature chosen for this call. The probe records overloads 1–3 passing the body-use exception with rigid T, then overload 4 failing with the concrete `BindingOrAssignmentElement` callback parameter.

## 3. Exact refusal branch and reachable any

**Both requested calls fail because `overloadCallbackHoldsAny` reaches its `depth >= 32` bound. No TypeFlagsAny branch triggers; the bound also stops caller-type recovery before any flow-type equality test.**

The unmodified and instrumented full-census streams are **exactly equal**, including the relation trailer. The scratch instrumentation preserves branch order, short-circuiting, seen-set use and return values. It adds complete depth paths and parameter/door logs. All 12 true walker events (two assertNode unit attempts × two guards, four assertEachNode unit attempts × two guards) are bound events at depth 32, with non-any flags. `depth-traces.json` retains all events; `probe.log` retains their order and callback gate outcomes.

The outer emitter in `census_small.go:228` refuses only if **both** `censusRelated(given, takes)` and `overloadCallbackUse(...)` are false. The direct relation is false for the narrow predicate callback against `((node: Node) => boolean) | undefined`. The fallback fails as follows:

- assertEachNode: `overloadCallbackElement` reaches its admitted-array guard (`element == nil || overloadCallbackHoldsAny(admitted, ...)`, original :329). The index element exists, but the admitted `readonly BindingOrAssignmentElement[] | undefined` hits the bound. Recovery returns the existing flow type **Node**, before checking `flow == concrete(own)` or its nonnullable variant. `overloadCallbackCall` then checks promised `BindingOrAssignmentElement`; its holds-any guard again hits the bound, and the relation operand is short-circuited.
- assertNode: `overloadCallbackArgument` reaches the admitted `BindingOrAssignmentElementTarget | undefined` guard (original :246), hits the bound and returns flow type **Node**, before the flow-type match. `overloadCallbackCall` then checks the promised nonnullable target union, hits the bound again, and short-circuits the relation.

Thus the failure is neither a flow-type mismatch nor a discovered non-library any. The trace’s `relatedIfEvaluated=false` at the final CALLGATE is an unevaluated sentinel when holdsAny=true; it is not evidence that a second relation ran.

### Reachable any distinction

With exactly the production library-member exemption and no depth cutoff, the stock type-graph traversal visits **1,896** types from the array argument, **1,890** from the assertNode target, and **1,889** from each test signature, finding **zero any types**. Predicate-target graphs are also checked and contain no any. The traversal follows unions/intersections, reference type arguments, properties, call/construct parameter and return types, and index value types; seen types break cycles. Generic type-parameter constraints are not followed, matching the production walker.

If library members are included, actual any witnesses do exist. These are excluded library method parameters, not the cause of the recorded refusal:

- assertEachNode argument: `root: BindingOrAssignmentElement[] → property every: { <S extends BindingOrAssignmentElement>(predicate: (value: BindingOrAssignmentElement, index: number, array: BindingOrAssignmentElement[]) => value is S, thisArg?: any): this is S[]; (predicate: (value: BindingOrAssignmentElement, index: number, array: BindingOrAssignmentElement[]) => unknown, thisArg?: any): boolean; } → signature.parameter thisArg: any`.

- assertEachNode test: `root: (node: Node) => node is ArrayBindingElement → signature.parameter node: Node → property emitNode: EmitNode | undefined → union/intersection[1]: EmitNode → property annotatedNodes: Node[] → property every: { <S extends Node>(predicate: (value: Node, index: number, array: Node[]) => value is S, thisArg?: any): this is S[]; (predicate: (value: Node, index: number, array: Node[]) => unknown, thisArg?: any): boolean; } → signature.parameter thisArg: any`.

- assertNode argument: `root: BindingOrAssignmentElementTarget → union/intersection[0]: ObjectBindingPattern → property elements: NodeArray<BindingElement> → property every: { <S extends BindingElement>(predicate: (value: BindingElement, index: number, array: readonly BindingElement[]) => value is S, thisArg?: any): this is readonly S[]; (predicate: (value: BindingElement, index: number, array: readonly BindingElement[]) => unknown, thisArg?: any): boolean; } → signature.parameter thisArg: any`.

- assertNode test: `root: (node: Node) => node is BindingName → signature.parameter node: Node → property emitNode: EmitNode | undefined → union/intersection[1]: EmitNode → property annotatedNodes: Node[] → property every: { <S extends Node>(predicate: (value: Node, index: number, array: Node[]) => value is S, thisArg?: any): this is S[]; (predicate: (value: Node, index: number, array: Node[]) => unknown, thisArg?: any): boolean; } → signature.parameter thisArg: any`.

The first unrestricted exploratory traversal tried enumerating the entire library graph and stopped at its 100,000-type guard (exit 1); it is retained as `checker-unrestricted-initial.*`. The final script exhaustively traverses the production-exempt graph and records just the first witness when library members are included, avoiding an unbounded expansion of generic library signatures.

### Full depth paths for the requested calls

Depth zero is the checked root. Each following row is exactly one recursive edge, including union selections and type arguments. The following four paths are the first matching events for the requested calls; all repetitions are retained in JSON.

#### assertEachNode caller-type recovery guard

`PARAMETER overload 4 of assertEachNode index 1 given (node: BindingOrAssignmentElement) => node is ArrayBindingElement takes ((node: Node) => boolean) | undefined`

| Depth | Exact edge and type |
| --- | --- |
| 0 | root: readonly BindingOrAssignmentElement[] \| undefined |
| 1 | union/intersection[1]: readonly BindingOrAssignmentElement[] |
| 2 | typeArgument[0]: BindingOrAssignmentElement |
| 3 | union/intersection[0]: ArrayLiteralExpression |
| 4 | property elements: NodeArray&lt;Expression&gt; |
| 5 | typeArgument[0]: Expression |
| 6 | property parent: Node |
| 7 | property emitNode: EmitNode \| undefined |
| 8 | union/intersection[1]: EmitNode |
| 9 | property externalHelpersModuleName: Identifier \| undefined |
| 10 | union/intersection[1]: Identifier |
| 11 | property jsDoc: JSDocArray \| undefined |
| 12 | union/intersection[1]: JSDocArray |
| 13 | property jsDocCache: readonly JSDocTag[] \| undefined |
| 14 | union/intersection[1]: readonly JSDocTag[] |
| 15 | typeArgument[0]: JSDocTag |
| 16 | property parent: JSDoc \| JSDocTypeLiteral |
| 17 | union/intersection[0]: JSDoc |
| 18 | property parent: HasJSDoc |
| 19 | union/intersection[0]: ArrowFunction |
| 20 | property body: ConciseBody |
| 21 | union/intersection[0]: Block |
| 22 | property nextContainer: HasLocals \| undefined |
| 23 | union/intersection[3]: CallSignatureDeclaration |
| 24 | property typeParameters: NodeArray&lt;TypeParameterDeclaration&gt; \| undefined |
| 25 | union/intersection[1]: NodeArray&lt;TypeParameterDeclaration&gt; |
| 26 | typeArgument[0]: TypeParameterDeclaration |
| 27 | property parent: InferTypeNode \| DeclarationWithTypeParameterChildren |
| 28 | union/intersection[2]: ClassDeclaration |
| 29 | property modifiers: NodeArray&lt;ModifierLike&gt; \| undefined |
| 30 | union/intersection[1]: NodeArray&lt;ModifierLike&gt; |
| 31 | typeArgument[0]: ModifierLike |
| 32 | union/intersection[5]: Decorator |

Bound hit: depth **32**, flags **1048576**, TypeFlagsAny bit **0**.

#### assertEachNode promised callback-parameter guard

`PARAMETER overload 4 of assertEachNode index 1 given (node: BindingOrAssignmentElement) => node is ArrayBindingElement takes ((node: Node) => boolean) | undefined`

| Depth | Exact edge and type |
| --- | --- |
| 0 | root: BindingOrAssignmentElement |
| 1 | union/intersection[0]: ArrayLiteralExpression |
| 2 | property elements: NodeArray&lt;Expression&gt; |
| 3 | typeArgument[0]: Expression |
| 4 | property parent: Node |
| 5 | property emitNode: EmitNode \| undefined |
| 6 | union/intersection[1]: EmitNode |
| 7 | property externalHelpersModuleName: Identifier \| undefined |
| 8 | union/intersection[1]: Identifier |
| 9 | property jsDoc: JSDocArray \| undefined |
| 10 | union/intersection[1]: JSDocArray |
| 11 | property jsDocCache: readonly JSDocTag[] \| undefined |
| 12 | union/intersection[1]: readonly JSDocTag[] |
| 13 | typeArgument[0]: JSDocTag |
| 14 | property parent: JSDoc \| JSDocTypeLiteral |
| 15 | union/intersection[0]: JSDoc |
| 16 | property parent: HasJSDoc |
| 17 | union/intersection[0]: ArrowFunction |
| 18 | property body: ConciseBody |
| 19 | union/intersection[0]: Block |
| 20 | property nextContainer: HasLocals \| undefined |
| 21 | union/intersection[3]: CallSignatureDeclaration |
| 22 | property typeParameters: NodeArray&lt;TypeParameterDeclaration&gt; \| undefined |
| 23 | union/intersection[1]: NodeArray&lt;TypeParameterDeclaration&gt; |
| 24 | typeArgument[0]: TypeParameterDeclaration |
| 25 | property parent: InferTypeNode \| DeclarationWithTypeParameterChildren |
| 26 | union/intersection[2]: ClassDeclaration |
| 27 | property modifiers: NodeArray&lt;ModifierLike&gt; \| undefined |
| 28 | union/intersection[1]: NodeArray&lt;ModifierLike&gt; |
| 29 | typeArgument[0]: ModifierLike |
| 30 | union/intersection[5]: Decorator |
| 31 | property parent: NamedDeclaration |
| 32 | property name: DeclarationName \| undefined |

Bound hit: depth **32**, flags **134217728**, TypeFlagsAny bit **0**.

#### assertNode caller-type recovery guard

`PARAMETER overload 1 of assertNode index 1 given (node: ArrayBindingPattern | ArrayLiteralExpression | ElementAccessExpression | Identifier | ObjectBindingPattern | ObjectLiteralExpression | OmittedExpression | PropertyAccessExpression) => node is BindingName takes ((node: Node) => boolean) | undefined`

| Depth | Exact edge and type |
| --- | --- |
| 0 | root: ArrayBindingPattern \| ArrayLiteralExpression \| ElementAccessExpression \| Identifier \| ... 4 more ... \| undefined |
| 1 | union/intersection[1]: ArrayBindingPattern |
| 2 | property parent: BindingElement \| ParameterDeclaration \| VariableDeclaration |
| 3 | union/intersection[0]: BindingElement |
| 4 | property parent: BindingPattern |
| 5 | union/intersection[1]: ObjectBindingPattern |
| 6 | property original: Node \| undefined |
| 7 | union/intersection[1]: Node |
| 8 | property emitNode: EmitNode \| undefined |
| 9 | union/intersection[1]: EmitNode |
| 10 | property externalHelpersModuleName: Identifier \| undefined |
| 11 | union/intersection[1]: Identifier |
| 12 | property jsDoc: JSDocArray \| undefined |
| 13 | union/intersection[1]: JSDocArray |
| 14 | property jsDocCache: readonly JSDocTag[] \| undefined |
| 15 | union/intersection[1]: readonly JSDocTag[] |
| 16 | typeArgument[0]: JSDocTag |
| 17 | property parent: JSDoc \| JSDocTypeLiteral |
| 18 | union/intersection[0]: JSDoc |
| 19 | property parent: HasJSDoc |
| 20 | union/intersection[0]: ArrowFunction |
| 21 | property body: ConciseBody |
| 22 | union/intersection[0]: Block |
| 23 | property nextContainer: HasLocals \| undefined |
| 24 | union/intersection[3]: CallSignatureDeclaration |
| 25 | property typeParameters: NodeArray&lt;TypeParameterDeclaration&gt; \| undefined |
| 26 | union/intersection[1]: NodeArray&lt;TypeParameterDeclaration&gt; |
| 27 | typeArgument[0]: TypeParameterDeclaration |
| 28 | property parent: InferTypeNode \| DeclarationWithTypeParameterChildren |
| 29 | union/intersection[2]: ClassDeclaration |
| 30 | property modifiers: NodeArray&lt;ModifierLike&gt; \| undefined |
| 31 | union/intersection[1]: NodeArray&lt;ModifierLike&gt; |
| 32 | typeArgument[0]: ModifierLike |

Bound hit: depth **32**, flags **134217728**, TypeFlagsAny bit **0**.

#### assertNode promised callback-parameter guard

`PARAMETER overload 1 of assertNode index 1 given (node: ArrayBindingPattern | ArrayLiteralExpression | ElementAccessExpression | Identifier | ObjectBindingPattern | ObjectLiteralExpression | OmittedExpression | PropertyAccessExpression) => node is BindingName takes ((node: Node) => boolean) | undefined`

| Depth | Exact edge and type |
| --- | --- |
| 0 | root: ArrayBindingPattern \| ArrayLiteralExpression \| ElementAccessExpression \| Identifier \| ObjectBindingPattern \| ObjectLiteralExpression \| OmittedExpression \| PropertyAccessExpression |
| 1 | union/intersection[0]: ArrayBindingPattern |
| 2 | property parent: BindingElement \| ParameterDeclaration \| VariableDeclaration |
| 3 | union/intersection[0]: BindingElement |
| 4 | property parent: BindingPattern |
| 5 | union/intersection[1]: ObjectBindingPattern |
| 6 | property original: Node \| undefined |
| 7 | union/intersection[1]: Node |
| 8 | property emitNode: EmitNode \| undefined |
| 9 | union/intersection[1]: EmitNode |
| 10 | property externalHelpersModuleName: Identifier \| undefined |
| 11 | union/intersection[1]: Identifier |
| 12 | property jsDoc: JSDocArray \| undefined |
| 13 | union/intersection[1]: JSDocArray |
| 14 | property jsDocCache: readonly JSDocTag[] \| undefined |
| 15 | union/intersection[1]: readonly JSDocTag[] |
| 16 | typeArgument[0]: JSDocTag |
| 17 | property parent: JSDoc \| JSDocTypeLiteral |
| 18 | union/intersection[0]: JSDoc |
| 19 | property parent: HasJSDoc |
| 20 | union/intersection[0]: ArrowFunction |
| 21 | property body: ConciseBody |
| 22 | union/intersection[0]: Block |
| 23 | property nextContainer: HasLocals \| undefined |
| 24 | union/intersection[3]: CallSignatureDeclaration |
| 25 | property typeParameters: NodeArray&lt;TypeParameterDeclaration&gt; \| undefined |
| 26 | union/intersection[1]: NodeArray&lt;TypeParameterDeclaration&gt; |
| 27 | typeArgument[0]: TypeParameterDeclaration |
| 28 | property parent: InferTypeNode \| DeclarationWithTypeParameterChildren |
| 29 | union/intersection[2]: ClassDeclaration |
| 30 | property modifiers: NodeArray&lt;ModifierLike&gt; \| undefined |
| 31 | union/intersection[1]: NodeArray&lt;ModifierLike&gt; |
| 32 | typeArgument[0]: ModifierLike |

Bound hit: depth **32**, flags **134217728**, TypeFlagsAny bit **0**.

## 4. Difference from the small shape

The callback bodies have the same relevant shape: an unchanged input parameter (or const for-of element), presence tests, and a direct `test(...)` call with that value. The real types add tsc’s large recursive AST graph. `BindingOrAssignmentElement` expands to declaration, binding, assignment and expression variants; their members lead through parent/emitNode/JSDoc/declaration graphs far enough for the depth-first guard to reach 32 before completing. The real assertNode target is a similarly connected union. Neither root itself is any.

Main’s in-tree small-shape tests use `interface N { readonly kind: number }` and `interface Id extends N { readonly kind: 1; readonly text: string }`. That graph has no tsc parent/emitNode/JSDoc links. Both `TestOverloadCallbackAssertNode` and `TestOverloadCallbackAssertEachNode` pass on this exact main (they lower and agree with Node). The repository’s assertEachNode twin also omits NodeArray-specific overloads and stackCrawlMark; those shape differences are documented in its source, but the observed blocker here is the depth-bound path. Optimizer’s separate small-program source was not supplied, so this comparison names the retained repository twin rather than claiming to inspect that unpublished program.

This is a diagnosis, not a proposed relaxation: production deliberately treats reaching the bound as unsafe. No optimizer code or adaptation was modified.

## Reproduction and checks

See `commands.md`, `measure.sh`, `checker.cjs` and `instrument.py`. All children run through `run.py` with a 1,740-second limit and a ten-second termination grace; each wrapper reports and propagates its own child’s status. All final census, probe, checker, build and small-shape test children exit zero. The exploratory graph-size refusal above is the only retained nonzero child. Go source evidence uses `.go.txt`. No compiler binary or scratch tree is committed.

| Child | Wall seconds | Own exit |
| --- | --- | --- |
| census | 36.928 | 0 |
| probe-build | 18.039 | 0 |
| probe | 37.91 | 0 |
| checker | 11.628 | 0 |
| small-shape-tests | 20.482 | 0 |
