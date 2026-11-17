# Readonly flow slots, adaptation 74

Adamic, task #5pstmvw. This one changes JavaScript: one statement pair in
`src/compiler/binder.ts` becomes one statement, and the object it makes is the
same. Adamic ruled that acceptable on #5pstmvw, on the precedent of road C at
sys.ts:514 (#1vkqj9w): a JavaScript-changing adaptation stands when it's
value-identical, counted, witnessed on Node, and a mutant brings the refusals
back.

## The edit

`adapt.cjs` makes two edits, each found through the stock AST and refused
unless it finds exactly upstream's declaration or statements. A second run
changes nothing.

**types.ts.** `readonly ` goes before `node` and `antecedent` on
`FlowNodeBase` and on every arm of the `FlowNode` union: FlowUnreachable,
FlowStart, FlowLabel, FlowAssignment, FlowCall, FlowCondition,
FlowSwitchClause, FlowArrayMutation and FlowReduceLabel. FlowLabel's
`antecedent` is the one exception and stays writable, because a label writes it
as it gathers its antecedents.

tsc reads a union's property as read-only when any arm's is (checker.ts:15626).
So a flow node seen as `FlowNode`, or as a union of some of its arms, becomes a
view that can't write one arm's `node` into another arm that reads it narrower.
For example, it can't write an `Expression` into a FlowArrayMutation, which
reads `CallExpression | BinaryExpression`.

**binder.ts.** Every flow node gets its node from `createFlowNode` at
construction except one: `bindContainer` makes a FlowStart with no node, then
assigns the function on the next statement through `currentFlow`, which is
typed `FlowNode`. Upstream's pair:

```ts
currentFlow = createFlowNode(FlowFlags.Start, /*node*/ undefined, /*antecedent*/ undefined);
if (containerFlags & (ContainerFlags.IsFunctionExpression | ContainerFlags.IsObjectLiteralOrClassExpressionMethodOrAccessor)) {
    currentFlow.node = node as FunctionExpression | ArrowFunction | MethodDeclaration | GetAccessorDeclaration | SetAccessorDeclaration;
}
```

becomes one statement, where the same condition chooses the node
`createFlowNode` is given:

```ts
currentFlow = createFlowNode(FlowFlags.Start, /*node*/ containerFlags & (ContainerFlags.IsFunctionExpression | ContainerFlags.IsObjectLiteralOrClassExpressionMethodOrAccessor) ? node as FunctionExpression | ArrowFunction | MethodDeclaration | GetAccessorDeclaration | SetAccessorDeclaration : undefined, /*antecedent*/ undefined);
```

`createFlowNode` (binder.ts:495) builds `{ flags, id: 0, node, antecedent }`
and hands it to `Debug.attachFlowNodeDebugInfo` (debug.ts:552). That call
reads none of the node's fields. When debug info is on, it sets a shared
prototype, or defines lazy getters that read only when a debugger asks. Nothing
reads `node` between the FlowStart's creation and the old
assignment. Once the pair has run, the object, its keys, their order and their
values are the same.

Both slots are `@internal`, so the published `typescript.d.ts` doesn't change
and there's no API sanction.

The directory sorts after 72-readonly-diagnostic-slots and before
75-optional-widening. Every later adaptation, and every consumer, reads apply's
tree with the slots readonly.

## Certificate

As with 72, it's certified by the compiled program's gate. Adamic compiles
`src/compiler` and judges every view at its relation site, so a write through a
wider view inside the compiled program is refused there, not trusted. The count
of writes below is checker-resolved over all of upstream's `src`, not textual.

## Measured, from J-e5bf815

The measure ran with adaptation 74 present (after) and absent (before), on
azure-proof-5, using the script on memory/scratch-5pstmvw-2.

| Measurement | Before | After |
|---|---:|---:|
| `node` written after construction, any flow node, all of src | 1 (binder.ts:1016, through `FlowNode`) | 0 |
| `antecedent` written, all of src | 8, all through `FlowLabel` | 8, all through `FlowLabel` |
| `Object.assign` onto a flow node | 0 | 0 |
| Stock tsc diagnostics over 605 roots | 7,057 | 7,057 |
| TS2540 new in after | | 0 |
| The meter's latent total | 3,941 | 3,933 |
| The flow family | 42 | 34, all one counted fact (below) |
| Lane, whole, on the after tree | | 398 units, 106,366 pass, the plan's 1 fail |

`stage3-lane-byte-comparisons` passed with `typescript.d.ts` byte-identical.
`shard-plan.json`'s runtime hash holds unchanged.

**The mutant** makes FlowAssignment's `node` writable again and nothing else.
That brings the family back to 42, node's refusals among them.

## Counted: FlowUnreachable fits FlowLabel (34 sites)

All 42 `node` refusals are gone after the edit. Eight sites clear outright. The
other 34 are the same sites (binder.ts 19, checker.ts 15), now refused for one
fact under four spellings, for example "a value of type FlowNode seen as
FlowNode | undefined, whose readonly field antecedent becomes writable".

The fact: FlowUnreachable is `{ node: undefined, antecedent: undefined }`, and
with both readonly it still fits FlowLabel `{ node: undefined, antecedent:
FlowNode[] | undefined }`, whose `antecedent` stays writable. Adamic judges a
value seen through a union as each arm it fits (pinned in
`internal/lower/flow_union_readonly_test.go`), and that's sound by Adamic's own
checks. In `.ts`, a cast to FlowLabel is checked by structure, and the check
can't tell an unreachable node from a label: both carry `flags: FlowFlags`.

tsc tells them apart by the Label bit in `flags`, which no type declares. A
literal discriminant would be a new lie, since `setFlowNodeReferenced` ORs
Referenced and Shared into every arm's flags. So no adaptation of types fixes
it. Adamic ruled to land 74 with the 34 counted. Whether the refusal at the
widening adds anything beyond the checked cast to an arm is a separate
question, filed as its own task for Compiler.

## Premises a reader can re-check

Readonly is a checker fact. Two roads could write a slot without the checker
seeing an assignment to it, and both are closed today by a stop in front of
them, not by this adaptation:

- `Object.assign`, which has a NotYet in front of it. Lifting it must judge
  readonly on the target's properties (#w9ysnpk).
- A generic key write `t[k] = v` with `K extends keyof T`, which is NotYet.
  Whoever lifts it judges readonly per property of K's constraint, under the
  ruling on writes through a key that isn't a literal (docs/rulings.md).

If either stop is lifted without that judgment, this certificate no longer
holds for the flow slots.

## Re-running

`node adapt.cjs <tree>` uses stock typescript@6.0.3 through NODE_PATH or
`CENSUS_TYPESCRIPT`, as apply runs every adaptation. The measure compares each
tree's diagnostics and refusal reasons with the tree's root read as `<tree>`,
since a message naming a declaration prints its path.
