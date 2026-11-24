# Method parameter adaptations

Family 7, #bbssdcy. Recut at `5ff5724930bb75f01952aec3c06fdc9e8d74385a`. Original site coordinates and ruling references use Adamic pin `0679623c2aa66cce5144ba0f894f020e3c5c6d99`. The pins refer to the fully composed TypeScript 6.0.3 tree. Each change matches an AST owner and exact source text.

This adapter changes implementation types and erased private overload declarations. It makes no edit to a public declaration. The public optionality, plain-list ranges, and write callback stay as declared. In particular, neither the public `WriteFileCallback` nor `SolutionBuilderHostBase.writeFile` is widened in this head.

Rulings are line numbers in `docs/rulings.md` at the Adamic pin. Row 102 is the parameter relation, row 96 covers shorthand, row 397 governs public changes, and row 398 governs internal adaptations.

## Candidate edits

- `src/compiler/builder.ts:392:26` Keep the key type shared by both collections (102, 127, 398).
- `src/compiler/builder.ts:392:100` Keep the key type shared by both collections (102, 127, 398).
- `src/compiler/checker.ts:1631:9` Widen backing annotations where omission already has a runtime path (96, 102, 397, 398).
- `src/compiler/checker.ts:1810:9` Widen backing annotations where omission already has a runtime path (96, 102, 397, 398).
- `src/compiler/checker.ts:6457:13` Keep the existing missing-symbol fallback (102, 398).
- `src/compiler/checker.ts:51469:13` Widen backing annotations where omission already has a runtime path (96, 102, 397, 398).
- `src/compiler/checker.ts:51470:13` Widen backing annotations where omission already has a runtime path (96, 102, 397, 398).
- `src/compiler/emitter.ts:1294:9` Expose the existing broad implementation and erase the two narrower private overloads (102, 131, 398).
- `src/compiler/emitter.ts:1295:9` Widen backing annotations where omission already has a runtime path (96, 102, 397, 398).
- `src/compiler/factory/emitHelpers.ts:171:9` Widen backing annotations where omission already has a runtime path (96, 102, 397, 398).
- `src/compiler/factory/nodeFactory.ts:641:36` Match the input of the forwarded factory call (102).
- `src/compiler/factory/nodeFactory.ts:678:9` Match the input of the underlying constructor (96, 102).
- `src/compiler/factory/nodeFactory.ts:689:9` Widen backing annotations where omission already has a runtime path (96, 102, 397, 398).
- `src/compiler/factory/nodeFactory.ts:843:9` Widen backing annotations where omission already has a runtime path (96, 102, 397, 398).
- `src/compiler/factory/nodeFactory.ts:1155:9` Permit the absence that the visitor path already handles (102, 131).
- `src/compiler/factory/nodeFactory.ts:1157:9` Permit the absence that the visitor path already handles (102, 131).
- `src/compiler/factory/nodeFactory.ts:1161:9` Widen backing annotations where omission already has a runtime path (96, 102, 397, 398).
- `src/compiler/factory/nodeFactory.ts:1162:9` Widen backing annotations where omission already has a runtime path (96, 102, 397, 398).
- `src/compiler/factory/parenthesizerRules.ts:63:9` Use the key domain already accepted by the parenthesizer and advertised by its slot (96, 102, 398).
- `src/compiler/factory/parenthesizerRules.ts:64:9` Use the key domain already accepted by the parenthesizer and advertised by its slot (96, 102, 398).

## Readonly-list edits held by Adamic

Adamic held these three edits under the readonly-view rule cited in her
ruling as `docs/rulings.md:95`.

- `src/compiler/factory/emitHelpers.ts:156:9` (`createDecorateHelper.decoratorExpressions`).
- `src/compiler/factory/emitHelpers.ts:163:9` (`createAssignHelper.attributesSegments`).
- `src/compiler/factory/parenthesizerRules.ts:91:9` (`parenthesizeTypeArguments.typeArguments`).

They may return for Adamic's review only with a trace showing that the
constructor writes metadata exclusively to a fresh copy at every call
that reaches it, leaving each caller's array untouched.

## Accepted stays (12)

- `src/compiler/checker.ts:1687:42` public kind contract (102, 397).
- `src/compiler/checker.ts:1842:9` narrow node/index storage contract (102, 397).
- `src/compiler/checker.ts:7354:140` overloaded visitor proof (102, 397).
- `src/compiler/expressionToTypeNode.ts:582:83` overloaded visitor proof (102, 397).
- `src/compiler/factory/nodeFactory.ts:750:9` narrow node/index storage contract (102, 397).
- `src/compiler/factory/nodeFactory.ts:751:9` narrow node/index storage contract (102, 397).
- `src/compiler/factory/nodeFactory.ts:794:9` narrow node/index storage contract (102, 397).
- `src/compiler/factory/nodeFactory.ts:795:9` narrow node/index storage contract (102, 397).
- `src/compiler/factory/nodeFactory.ts:850:9` narrow node/index storage contract (102, 397).
- `src/compiler/factory/nodeFactory.ts:851:9` narrow node/index storage contract (102, 397).
- `src/compiler/factory/nodeFactory.ts:894:9` required tag field (102, 397).
- `src/compiler/factory/parenthesizerRules.ts:75:9` plain-list source-range contract (102, 397).

## Held sites

- `src/compiler/checker.ts:28251:58` left — checked below, decision held for Adamic.
- `src/compiler/checker.ts:51439:13` stopped — widening the optional meaning flag would have to propagate through accessibility and symbol formatting.
- `src/compiler/transformers/destructuring.ts:185:34` left — checked below, decision held for Adamic.
- `src/compiler/tsbuildPublic.ts:791:22` stopped — optional BOM callback requires the maintainer’s write-callback review.
- `src/compiler/program.ts:1900:9` stopped. The required change lies inside the exact `createProgram` pin of the later `99-method-value-views` adapter. That pin is outside this change (102, 398).


## Checked held signatures

Run `node stage3/adapt/99-method-parameter-adaptations/signatures.cjs TREE trace`
on the composed tree. This uses the pinned stock TypeScript checker, resolves
each call and checks the selected overload, the reducer accumulator, both
predicate branches and their common signature.

- **Seeded reducer, checker.ts:28251:58.** The selected overload is
  `core.ts:1236`, `reduceLeft<T, U>(array, f, initial, start?, count?): U`,
  with `T = Type` and `U = number`. The arrow is
  `(facts: TypeFacts.None, t: Type) => number`. Its slot is
  `(memo: number, value: Type, i: number) => number`.
  Stock assignability succeeds both ways at the first parameter. Adamic's
  enum-domain check rejects feeding a number into the literal enum member.
  There is no method relation here. A scratch seed of `TypeFacts.None as number`
  records `facts: number` and fails the literal-accumulator pin.
- **Conditional predicate, transformers/destructuring.ts:185:34.** The
  call resolves to the generic predicate overload of `Debug.assertNode` in
  `debug.ts:281`. The true branch is `(node: Node) => node is Identifier`,
  the false branch is `(node: Node) => node is Expression`, and the common
  signature and test slot are `(node: Node) => node is Expression`.
  The Adamic refusal compares the ternary's condition,
  `createAssignmentCallback`, against that predicate slot. The condition's
  callback takes `Identifier`. `Node -> Identifier` fails and
  `Identifier -> Node` succeeds. The callable-context return in
  `internal/lower/invariance.go:2033` precedes its branch-membership check
  at line 2036. Both actual predicate branches accept Node. In a scratch
  control, narrowing the true branch declaration's input to Identifier
  changes its signature and the common signature and fails the branch pin.

These traces leave both decisions with Adamic.

## Landing-lane regression check

The landing lane's API and byte-comparison check runs
`signatures.cjs TREE printer` on every current landing. Historical evidence
produced before this adapter can be replayed with
`python3 stage3/lane/check.py RESULTS --pre-family7-replay`. The
standalone report records the selected replay mode.
It checks the two private backing signatures under `createPrinter`
independently of the adapter's edit list. Restoring exactly
`writeList.nodes: NodeArray<T>` fails the gate, which expects
`NodeArray<T> | undefined`. Stock TypeScript emit stays byte-identical
under that mutation.
