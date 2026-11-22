# The factory's memoized creators as plain properties

TypeScript v6.0.3, 050880ce59e30b356b686bd3144efe24f875ebc8. Task #s8mc8gs, Adamic's ruling of Nov 22 on that task: the
internal form. It changes the emitted JavaScript of `src/compiler/factory/nodeFactory.ts` alone. typescript.d.ts does not
move, so this directory carries no `api-sanctions.json`.

## Why

createNodeFactory builds the factory as one object literal, and upstream writes 70 of its members as getters that return
a memoized creator for one SyntaxKind:

```ts
get createAssignment() {
    return getBinaryCreateFunction(SyntaxKind.EqualsToken) as NodeFactory["createAssignment"];
},
```

The NodeFactory interface declares the same 70 as methods. Adamic refuses a method that shares its name with a getter
("a method sharing its name with a getter; rename the getter or use a callable arrow field in the interface and
implementation"), so every call through the factory stopped there: 112 latent sites in full mode (J-vt1zf5t at main
91fa7305), 12 counted the meter's way on 20261122T131040Z.tNkys1. createAssignment alone is about 60 of them; the binary
and unary operators, the JSDoc primary types behind parser.ts's finishNode calls and the JSDoc tags are the rest.

## The edit

Each getter becomes a property holding the expression it returned, evaluated once when the factory is built:

```ts
createAssignment: getBinaryCreateFunction(SyntaxKind.EqualsToken) as NodeFactory["createAssignment"],
```

The 70 are the 12 JSDoc type creators and updaters, the 26 JSDoc tag creators and updaters, the 24 binary operators, the
6 prefix and the 2 postfix unary operators. `parenthesizer` and `converters` stay getters: the interface declares them as
readonly properties, and their bodies build the factory's own rules lazily.

adapt.cjs refuses before any write unless:

- every one of the 70 getter bodies is exactly `return <helper>(SyntaxKind.<kind>)`, optionally cast to its own member's
  type `NodeFactory["<name>"]`, with the helper upstream calls for that member;
- each helper is a `const` of createNodeFactory's own body, declared before the literal, initialized with
  `memoizeOne(<an arrow returning an arrow>)` (all eleven are, at nodeFactory.ts:500 to 511 upstream);
- the 70 are upstream's, in upstream's order, and the tree holds either all getters or all adapted properties, never a
  mix.

A second run writes nothing. The tree keeps upstream's CRLF endings: each edit replaces one member with one line and
writes no line ending.

## Value identity, Adamic's four conditions

Measured in J-y1g0tqd (sealed, at 78654adc on main 776ae922), over two adapted trees built by apply, without and with
this directory.

1. **No side effects at construction.** The adapter holds each body to a single call of a `memoizeOne(arrow => arrow)`
   helper with a SyntaxKind argument, so evaluating it eagerly builds one closure and calls nothing else. Every helper
   is declared before the literal and each closure reaches only hoisted function declarations, so nothing is read in its
   temporal dead zone. Node loads stock typescript.js with the same edit made to its bundled createNodeFactory and builds
   `ts.factory` and a fresh `createNodeFactory(0, ...)` without error.
2. **Identity and descriptors go unread.** The checker's trace over all 713 files of src (the program stock tsc checks)
   finds 236 calls through the 70 members and 9 other references, all in type positions: `NodeFactory["createAssignment"]`
   (nodeFactory.ts), and the eight views 99-method-value-views writes: `(typeof factory)["createComma"]` in
   nodeFactory.ts and parser.ts's seven `(typeof factory)["createJSDoc<X>Tag"]`. This adaptation makes those views
   unnecessary and leaves them harmless, as Adamic ruled, so they stay. Nothing reads one of them as a value,
   compares it, reads its `name` or `length`, or assigns to it. Nothing hands the factory as a whole to `Object.keys`,
   `values`, `entries`, `getOwnPropertyNames`, `getOwnPropertyDescriptor(s)`, `defineProperty`, `assign`, `freeze` or
   `Reflect`, spreads it, or walks it with `for in`.
3. **Key order is unchanged.** In Node, `Object.keys` and `Object.getOwnPropertyNames` read the same 491 names in the same
   order on `ts.factory` and on a fresh factory, before and after. Each of the 70 goes from an accessor to a
   function-valued data property with the same enumerable and configurable flags.
4. **The memo still applies.** Each of the 70 builds the same node, printed the same, before and after (70 of 70), and
   each property reads the same function every time.

## What else was measured

| Measurement | Result | Job |
| --- | --- | --- |
| Stock tsc over all of src, as (file, code, message) | 2,719 before and after, 0 new, 0 gone | J-y1g0tqd |
| Emitted JavaScript, 604 files (109 .d.ts emit none) | nodeFactory.js alone differs: 70 getters out, 70 properties in | J-y1g0tqd |
| Trees after apply, `src/` | nodeFactory.ts alone differs | J-y1g0tqd |
| The adapter alone on the tree without it | the same nodeFactory.ts as apply's | J-y1g0tqd |
| Second run | writes nothing | J-y1g0tqd |
| A body that isn't its helper; a half-adapted tree | each refused, nothing written | J-y1g0tqd |
| The adapted tree product and the lane | green | J-3bs6fs3 |

## The census, before and after

Full-mode latent census of the whole adapted tree at 78654adc (main 776ae922), without this directory (J-gbmvbsz) and
with it (J-c6gfanz). Each run is six processes in parallel by `LATENT_ONLY_FILE` (checker.ts alone, the rest dealt by
size), since one process passed two hours (J-pnxbwv8). Sites are compared by (kind, where, reason), with nodeFactory.ts's
lines mapped back across the two lines each getter gives up.

| | Before | After |
| --- | ---: | ---: |
| Distinct sites | 6,408 | 6,241 |
| Refused / NotYet / panic / error | 1,970 / 4,435 / 2 / 1 | 1,970 / 4,268 / 2 / 1 |
| "a method sharing its name with a getter" | 112 | 1 |

**178 gone, 11 new.**

- **111 of the 112 getter-name sites go.** Each is a call through one of the 70 creators, and each now has no finding at
  its place. The one left, builder.ts `state.referencedMap?.size()`, is a ManyToManyMap's `size()` beside a Map's `size`
  getter: another cause, outside this family, there before and after.
- **65 more NotYet go with them**, all inside the same units: "reading <name>" of a local whose initializer held a
  getter-name stop (esnext.ts's `assignment`, ts.ts's `enumStatement` and `moduleStatement`, es2018.ts's `statements`
  and kin), and four "assigning to an Identifier" beside them.
- **Two Refused move within their member, not out of it.** `createAssignment`'s cast, "a cast the runtime can't check",
  goes from its getter's return line (1028:20) to its property's line (1027:27). `updateJSDocThisTag`'s "a function
  taking JSDocTypeExpression seen as one taking JSDocTypeExpression | undefined" moves column from 894:9 to 894:29.
  Each is the same fact at the same member. The second is #kr1v0tn's neighbor (the kept refusal of
  99-cli-jsdoc-type-like-tags).
- **9 NotYet are exposed, none created.** Each sits inside an expression whose first stop, before, was a getter-name
  stop or a read it caused, so the census never reached it:
  - eight "reading <name>" at parser.ts:9493, classFields.ts:2622, es2015.ts:3207, es2016.ts:109, es2018.ts:846, 1269
    and 1293, and esDecorators.ts:1954, each in the arguments of a creator call whose getter-name stop sat on its line
    or the one or two above it;
  - module.ts:540, "an array going where a plain object type is held": the `node.statements` location of a
    `setTextRange` in the UMD wrapper, whose first stop before was 517's "reading umdHeader", gone now. It's #hae423v's
    road, the NodeArray seen as a TextRange.

So the family goes from 112 to 0 of its own, and what it uncovers was there behind it.

## Re-running

`node adapt.cjs <tree>` uses stock typescript@6.0.3 through NODE_PATH or `CENSUS_TYPESCRIPT`, as apply runs every
adaptation.
