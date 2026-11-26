# visitNode's implementation takes the undefined it already checks

Type-only, task #ytagqs5. Adamic ruled it to land on Nov 26, 2026.

`src/compiler/visitorPublic.ts` declares visitNode as two overloads and one
implementation. Both overloads take `node: TIn` with `TIn extends Node | undefined`, so
a caller may pass undefined, and the implementation's body handles it first:

```ts
export function visitNode(
    node: Node,
    visitor: Visitor,
    test?: (node: Node) => boolean,
    lift?: (node: readonly Node[]) => Node,
): Node | undefined {
    if (node === undefined) {
        // If the input type is undefined, then the output type can be undefined.
        return node;
    }
```

The declared parameter says `Node`, which the body's own first test contradicts. Adamic's
overload census holds an implementation to every overload it serves, so overload 1's
`node: TIn` is refused against the implementation's `node: Node`: 17 tsc units stop on
"overload 1 of visitNode parameter node cannot be served by implementation parameter
node" (J-yjptyr4, main's first stops at ef741d4f).

## The edit

One exact text, counted once, inside the one top-level visitNode with a body, right
after its two overloads (adapted visitorPublic.ts:148):

```ts
export function visitNode(
    node: Node | undefined,
    visitor: Visitor,
```

Nothing else moves. The body already reads `node` as possibly undefined, returns it
when it is, and from there on `node` narrows to `Node` as before, so every statement
after the first sees the same type it saw.

- **The JavaScript is the same.** A parameter's type annotation emits nothing: the file
  transpiles to the same bytes at ES2020 and ESNext, with and without comments.
- **No declaration moves.** The implementation of an overloaded function is never
  emitted as a declaration: tsc's declaration transformer elides it (`// Elide
  implementation signatures from overload sets`, `resolver.isImplementationOfOverload`,
  transformers/declarations.ts:1036 and :1405 at 050880ce), so the bundled
  `typescript.d.ts` carries visitNode's two overloads alone. The file's declaration emit
  is the same bytes with and without `stripInternal`.
- **Stock tsc agrees.** Widening an implementation's parameter keeps both overloads
  compatible with it, and stock 6.0.3 reports the same diagnostics over src/compiler.

## The count

J-rbv9ac2 is main's first-stop census (ef741d4f, entry pass) over a tree with exactly
this edit. All 17 units moved off the parameter stop to the next one, overload 1's
result (the `return node` at visitorPublic.ts:156 against `TOut | (TIn & undefined) |
(TVisited & undefined)`), and no other unit's first stop changed: 390 of 4,593 units
stop on both trees. That next stop is #ytagqs5's own, and this adaptation doesn't touch
it.

## How it is held

- `adapt.cjs` finds the owner through the parsed source: the one top-level visitNode
  with a body, exported, standing right after exactly two overloads. It refuses before
  any write unless the owner's text is the pinned before or after (SHA-256 and exact
  text, on LF), and unless the edit's text is in it once and its replacement not at all,
  or the other way round. It holds the planned file to the before file's JavaScript and
  declarations, byte for byte, keeps CRLF, and makes no write on replay.
- `contract.cjs TREE` checks LF and CRLF, a zero-edit replay that keeps bytes and mtime,
  and refusal, with the file untouched, of a renamed owner, a duplicated one, a third
  overload, and four drifts of one token each (the undefined test, the final return, the
  parameter, the return type).
- `check.cjs BEFORE AFTER` is run on two complete applies, before without this
  directory. It checks (a) that only visitorPublic.ts differs under src, and it is the
  before file with only the owner replaced; (b) that stock 6.0.3 reports the same
  diagnostics over src/compiler, and the implementation's node parameter is `Node`
  before and `Node | undefined` after; (c) that the file's JavaScript is byte-identical
  at ES2020 and ESNext, with and without comments; and (d) that its declaration emit is
  byte-identical, with and without `stripInternal`. Each tree's `node_modules` is a link
  to `stage3/typescript/node_modules`, for `@types/node`.

No later adaptation in apply order pins or edits visitNode's implementation: the ones
after this one that edit visitorPublic.ts own visitArrayWorker and visitNodes
(99-apple-visit-nodes-guard), one factory call each in
addDefaultValueAssignmentForBindingPattern and addDefaultValueAssignmentForInitializer
(99-getter-callable-views), visitFunctionBody and visitIterationBody
(99-seats-build22-52-nonempty-checks), and addDefaultValueAssignmentsIfNeeded
(99-seats-build22-62-node-array-range).
