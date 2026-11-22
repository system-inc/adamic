# Getter-backed operator calls through readonly views

Adamic’s Nov 22, 2026 ruling on #bbssdcy covers these thirteen direct calls. The
adapter inserts a local readonly callable view and changes only the call’s
receiver. The indexed access type retains the original call signatures,
including both assignment overloads. Rulings 113 and 119 govern getter and
method behavior. Ruling 397 governs the public API. The views use annotations,
so the assertion rule in row 14 is outside this edit.

Locations below are in TypeScript 6.0.3 (050880ce) after the existing adapters on
Adamic 0679623c2aa6. The table uses paths relative to `src/compiler/`.

| Site | Receiver declaration or route | Method and getter |
|---|---|---|
| `factory/emitHelpers.ts:315:25` | `factory/emitHelpers.ts:149`, `const factory = context.factory` | A |
| `factory/utilities.ts:415:13` | `factory/utilities.ts:412`, `factory: NodeFactory` forwarded by the object-literal dispatcher at 475 | A |
| `factory/utilities.ts:428:13` | `factory/utilities.ts:425`, `factory: NodeFactory` forwarded by the same dispatcher | A |
| `factory/utilities.ts:441:13` | `factory/utilities.ts:438`, `factory: NodeFactory` forwarded by the same dispatcher | A |
| `factory/utilities.ts:528:18` | `factory/utilities.ts:518`, `factory: NodeFactory` from class-fields and decorators transformation contexts | A |
| `factory/utilities.ts:537:21` | Same parameter and unit as the call at 528. The local view is inside the `if (resultVariable)` branch. | A |
| `factory/utilities.ts:1661:17` | `factory/utilities.ts:1650`, `factory: NodeFactory` forwarded by `transformers/classFields.ts:955` | A |
| `transformers/classFields.ts:3320:12` | `transformers/classFields.ts:3319`, `factory: NodeFactory` from the context binding at 357 | A |
| `transformers/destructuring.ts:591:56` | `transformers/destructuring.ts:67`, `context: TransformationContext`, then `types.ts:9599`, `readonly factory: NodeFactory` | A |
| `transformers/es2017.ts:1105:25` | `transformers/es2017.ts:1058`, `factory: NodeFactory` forwarded by the ES2017 and ES2018 transforms | A |
| `transformers/namedEvaluation.ts:95:24` | `transformers/namedEvaluation.ts:70`, `const { factory } = context` | A |
| `visitorPublic.ts:457:29` | `visitorPublic.ts:446`, `const { factory } = context` | E |
| `visitorPublic.ts:485:37` | `visitorPublic.ts:475`, `const factory = context.factory` | A |

In the compiler’s emit path, `emitter.ts:847,915` passes the exported factory
to `transformNodes`, which puts it on the transformation context in
`transformer.ts:248,272`. The export at `factory/nodeFactory.ts:7410` calls
`createNodeFactory`. That function creates the following getters on its
`factory: NodeFactory` literal. The generator transform passes its context
factory into the object-literal dispatcher at `transformers/generators.ts:1179`.

**A** combines `types.ts:9359–9360` with `factory/nodeFactory.ts:1027–1029`.

```ts
createAssignment(left: ObjectLiteralExpression | ArrayLiteralExpression, right: Expression): DestructuringAssignment;
createAssignment(left: Expression, right: Expression): AssignmentExpression<EqualsToken>;

get createAssignment() {
    return getBinaryCreateFunction(SyntaxKind.EqualsToken) as NodeFactory["createAssignment"];
},
```

**E** combines `types.ts:9366` with `factory/nodeFactory.ts:1045–1047`.

```ts
createStrictEquality(left: Expression, right: Expression): BinaryExpression;

get createStrictEquality() {
    return getBinaryCreateFunction(SyntaxKind.EqualsEqualsEqualsToken);
},
```

The getter producer at `factory/nodeFactory.ts:502` returns an arrow.
The member lookup stays at the original call, with the same runtime receiver
and argument evaluation. At the destructuring site the annotated local is
initialized after `hoistVariableDeclaration` and immediately before the outer
call. Its earlier read crosses only the lookup of `emitExpression`, an own data
property initialized with a local function in both `FlattenContext` literals
at `transformers/destructuring.ts:124–137,254–267`.

`sites.json` pins the exact enclosing statements before and after each change,
normalizing CRLF only for comparison. The adapter finds the statement and call
in the AST, checks all pins before writing any file, and preserves existing
line endings. It accepts a complete replay with zero writes and rejects mixed
original and adapted snapshots.


## Exposed stops

Family 19 removes thirteen getter-backed operator calls across two first-error
checkpoints. The pristine entry census records twelve. Clearing the first
assignment in `expandPreOrPostfixIncrementOrDecrementExpression` revealed
the thirteenth at `factory/utilities.ts:537:21`, which the Nov 22 standing
ruling includes in this adapter.

| Family | Count | Exposed site and unit | Census reason |
|---|---:|---|---|
| Branded checked-predicate array target | 1 | `factory/emitHelpers.ts:469:21`, unit `createEmitHelperFactory` at `factory/emitHelpers.ts:148:1` | a branded supertype inside a checked predicate's array target |
| Family 19, getter-backed method call | 1 | `factory/utilities.ts:541:18`, unit `expandPreOrPostfixIncrementOrDecrementExpression` at `factory/utilities.ts:517:1` | a method sharing its name with a getter; rename the getter or use a callable arrow field in the interface and implementation |

The earlier Family 19 stop at `factory/emitHelpers.ts:315:25` prevented
`createEmitHelperFactory` from lowering. Clearing it exposes this next
NotYet in the same unit. The Nov 21 meter already counted the exact reason
at three sites. The family table attributes the exposure to the branded
checked-predicate array target family.


The same unit that stopped at `factory/utilities.ts:528:18` in the pristine
pass and `:537:21` in round 2 now reaches `:541:18`. Those earlier stops
prevented it from lowering. The next stop is still NotYet, and the Nov 21
meter already counted the exact Family 19 reason at ten sites. This
additional exposure stays outside the thirteen edited calls.
