# Factory update view, adaptation 99-memory-factory-update-view

Adamic, task #38kdd4d, under the ruling Adamic delegated on #k8ed1f2 and
99-readonly-family3-views applies: a view narrowed where tsc never writes.
This is type-only: the emitted JavaScript does not change. The one declaration
it touches is local to `createNodeFactory`, which is `@internal`, so the
published declaration files do not change, this directory carries no
`api-sanctions.json`, and `stage3/lane/sanctioned-api.json` is not
regenerated.

## The edit

One parameter type in `src/compiler/factory/nodeFactory.ts`:

```ts
function update<T extends Node>(updated: T, original: T): T {   // was updated: Mutable<T>
```

`adapt.cjs` parses nodeFactory.ts with stock TypeScript 6.0.3 and finds the
parameter by its function's name and its own, never by line, offset or
whole-line text. Before any write it refuses unless:

- there is exactly one function named `update` in the file, a statement of
  `createNodeFactory`'s own body, with no modifiers, `<T extends Node>`, two
  plain parameters `updated` and `original: T`, and the return type `T`;
- its body is upstream's, statement by statement: the `if` that calls
  `setOriginal(updated, original)` and `setTextRange(updated, original)`, and
  `return updated;`;
- `setOriginal` is `createNodeFactory`'s
  `const setOriginal = flags & NodeFactoryFlags.NoOriginalNode ? identity : setOriginalNode`;
- `setOriginalNode` assigns `node.original` and `node.emitNode` and nothing
  else, with no increment or delete;
- every use of the name `update` in `createNodeFactory` is a call of it, and
  none stands outside it;
- `updated`'s type is `Mutable<T>`, upstream's, or `T`, this adaptation's.

The edit replaces that one type node's span. A parameter already `T` keeps
its bytes, so a second run changes nothing. It prints one line: edited or
already adapted, where, and how many calls it counted.

No adaptation before it edits `update`, `setOriginal` or `setOriginalNode`, so
it expects upstream's text. 91-patcher-discard pins `createNodeFactory`'s
whole text, and it sorts first, so its pin reads upstream's `update`. The
eight that sort after it (99-method-value-properties through
99-wasm-crypto-import) neither read nor pin `update`'s declaration;
99-truthful-generated-name-emit-node edits two other reads in nodeFactory.ts
by their own text.

## The premise

The `Mutable<T>` view lifts `readonly` from every field of the node. What
`update` actually writes through it, by function, with upstream 6.0.3's lines:

- `update` itself (nodeFactory.ts:7179 to 7185) writes no field. It compares,
  calls the two helpers below, and returns the node.
- `setOriginal` (nodeFactory.ts:493) is `identity`, which writes nothing, or
  `setOriginalNode` (nodeFactory.ts:7420), which writes `node.original`
  (7422) and `node.emitNode` (7425). `mergeEmitNode` (7431) writes into the
  `EmitNode` object it is handed or makes, never into the node. Node declares
  both `original?: Node` and `emitNode?: EmitNode` writable (types.ts:949 and
  950, each `@internal`).
- `setTextRange` (factory/utilitiesPublic.ts:10) calls `setTextRangePosEnd`
  (utilities.ts:10664), which calls `setTextRangePos` and `setTextRangeEnd`,
  writing `pos` (10645) and `end` (10655). Upstream's Node extends
  `ReadonlyTextRange` (types.ts:942); 80-writable-node-range makes it extend
  `TextRange`, whose `pos` and `end` are writable (types.ts:28 to 31), and
  85-writable-range-setters constrains the setters to `TextRange`.

So every field `update` writes is already writable on Node, and `T extends
Node` carries all four. Node's readonly `kind` and `flags` (types.ts:943 and
944) are never written, so the view that made them writable claimed more than
tsc does. With `updated: T` there is nothing for the view to stand on.

`update` keeps `T` as its return type, so every caller sees the same result
type. Eight of its calls pass a value that is already a `Mutable<X>`, the
finishUpdate helpers' own `updated` (nodeFactory.ts:1225, 1761, 1942, 1977,
4645, 4994, 5953, 5989): a `Mutable<X>` is assignable to `X`, and for the
generic one, finishUpdateBaseSignatureDeclaration, tsc relates a homomorphic
`Mutable<T>` to `T`, so by reading those calls still type-check and infer the
same `T`. That is read, not run; verdict (1) is what settles it.

## Sites

Adamic refuses each `update(createX(...), node)` call as "a value of type X
seen as Mutable<X>, whose readonly field flags becomes writable" (or kind).
#38kdd4d's W2 rows (branch `memory/scratch-w2-scout`,
`scratch-w2/w2-rows-85.tsv`) list 44 at `update` calls, each the fresh node
argument, in the adapted tree's numbering of nodeFactory.ts:

1640:22 createTypeParameterDeclaration; 2669:22 createImportTypeNode; 2709:22
createTypeOperatorNode; 2754:22 createMappedTypeNode; 3629:22
createYieldExpression; 3871:22 createBlock; 3920:22 createExpressionStatement;
4091:22 createContinueStatement; 4110:22 createBreakStatement; 4131:22
createReturnStatement; 4174:22 createSwitchStatement; 4617:22 createCaseBlock;
4722:22 createImportDeclaration; 4753:22 createImportClause; 4771:22
createAssertClause; 4788:22 createAssertEntry; 4804:22
createImportTypeAssertionContainer; 4824:22 createImportAttributes; 4841:22
createImportAttribute; 4857:22 createNamespaceImport; 4874:22
createNamespaceExport; 4890:22 createNamedImports; 5102:22
createJSDocPrePostfixUnaryTypeWorker; 5112:22 createJSDocUnaryTypeWorker;
5135:22 createJSDocFunctionType; 5151:22 createJSDocTypeLiteral; 5165:22
createJSDocTypeExpression; 5187:22 createJSDocSignature; 5227:22
createJSDocTemplateTag; 5249:22 createJSDocTypedefTag; 5271:22
createJSDocParameterTag; 5293:22 createJSDocPropertyTag; 5315:22
createJSDocCallbackTag; 5331:22 createJSDocOverloadTag; 5347:22
createJSDocAugmentsTag; 5370:22 createJSDocSeeTag; 5384:22
createJSDocNameReference; 5417:22 createJSDocLink; 5432:22 createJSDocLinkCode;
5447:22 createJSDocLinkPlain; 5456:22 createJSDocImplementsTag; 5524:22
createJSDocUnknownTag; 5563:22 createJSDocImportTag; 5577:22 createJSDocText.

Two of them, 5102 and 5112, are generic: the census prints the source as `T`
and the field as `T["flags"]`.

## Left out: the finishUpdate helpers

Thirteen finishUpdate helpers in nodeFactory.ts (1220, 1756, 1937, 1972,
2026, 2090, 2150, 2354, 4372, 4640, 4987, 5945, 5981) still take
`updated: Mutable<X>`, and this adaptation leaves every one as it is. Each
writes a field through that view before it calls `update`, and most of those
fields are readonly on the node's own interface: PropertySignature's
`initializer`, MethodDeclaration's `exclamationToken`,
ConstructorDeclaration's `typeParameters` and `type`, the `modifiers` of
ClassStaticBlockDeclaration and FunctionTypeNode, and the rest. `T` there would
not type-check. finishUpdateBaseSignatureDeclaration writes only
`typeArguments`, writable and `@internal` on SignatureDeclarationBase, but it
is generic over that interface's subtypes. Whether those copies of children
"used only for error reporting" or "for quick info" deserve a writable view is
a separate ruling. The W2 rows hold 7 of
its sites, each the fresh node handed to a finishUpdate helper: 1752:45
(finishUpdatePropertySignature), 1968:55
(finishUpdateClassStaticBlockDeclaration), 2350:44
(finishUpdateFunctionTypeNode), and 2188:52, 2221:52, 2254:52 and 2416:52
(finishUpdateBaseSignatureDeclaration). They stay refused.

The W2 rows' other nodeFactory.ts sites are casts and writes this adaptation
doesn't reach either: 2646:14, 5079:22, 6120:22, 6395:10, 6396:10 and 7396:6.

## Public consequence

None. `update` is a function declared in `createNodeFactory`'s body, and
`createNodeFactory` is `@internal`, so no line of typescript.d.ts moves. A
parameter's type erases, so no line of the emitted JavaScript moves either.

## What certifies it

`scratch-upd/measure.sh` on branch `memory/factory-update-view`, run as a
sealed Loom job at that branch's measure commit, prints a PASS or FAIL line per
verdict:

- (0) sources: the adapted tree with this adaptation differs from main's
  adapted tree (the same inputs without this directory) in nodeFactory.ts
  alone, and there in `update`'s one line, `Mutable<T>` to `T`; a second run of
  `adapt.cjs` over the adapted file changes nothing;
- (1) stock tsc: adaptation 70's program over all of src, no diagnostic new
  and none gone, by code, file, line and message;
- (2) the JavaScript: every src file transpiled by stock TypeScript on both
  trees, byte-identical;
- (3) the census: the sites seen as `Mutable<...>` at `update(...)` calls go
  from the census's count (44 by the W2 rows) to 0, the finishUpdate helpers'
  sites stay, and no site anywhere is new;
- (4) the mutant: the adapted tree with `Mutable<T>` restored is main's tree
  byte for byte, and its census brings back exactly the sites (3) saw go.

Until that job's verdicts are in, this directory has not been applied: the
claims above are read from upstream 6.0.3's source.

## Re-running

`node adapt.cjs <tree>` uses stock typescript@6.0.3 through NODE_PATH or
`CENSUS_TYPESCRIPT`, as apply runs every adaptation.
