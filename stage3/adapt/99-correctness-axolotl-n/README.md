# 99n: hasNode observes a readonly FlowNode

H10 / CW157 is kept. The other 25 sites are dropped with independent checker and lowering evidence. The branch is based on main `ab8a85bebf2a14a06fa8c3c4ca8cc93201a2309f`; the upstream source is TypeScript 6.0.3 at `050880ce59e30b356b686bd3144efe24f875ebc8`.

The only shipped source edit is:

```ts
function hasNode(f: Readonly<FlowNode>): f is Readonly<Extract<FlowNode, { node?: Node; }>> {
    return !!(f.flags & hasNodeFlags);
}
```

It reads flags and its caller only reads node. The original CW157 writable union view is gone, with zero new findings anywhere in debug.ts. hasNode has no recorded lowering stop. renderFlowNode retains the existing checked-predicate depth bound for FlowSwitchClause at debug.ts:937:9.

| Hand | Sites | Decision | Evidence |
| --- | --- | --- | --- |
| H06 | CW093, CW095, CW096, CW097 | Dropped | New weak/strong retention NotYet in assertNotNode |
| H07 | CW102 through CW136, even IDs | Dropped | Required enum-record NotYet transition absent; two new cache-storage scan findings |
| H08 | CW155 | Dropped | The assumed literal declaration is absent; introducing one changes JavaScript |
| H09 | CW156 | Dropped | New Array-as-value lowering and string initializer scan finding |
| H10 | CW157 | Kept | One finding removed; zero new findings or regressions |
| H11 | CW158 | Dropped | New Array-as-value lowering in renderGraph and inherited fill storage |

Every individual ID and reason is in [sites.json](sites.json). H07's exact counts, the raw census, the real pair result, and all 105 declaration next stops are in [CENSUS.md](CENSUS.md). Passing proof details are in [PROOF.md](PROOF.md).

## Adapter contract

The pinned stock parser finds the unique `Debug.formatControlFlowGraph.hasNode` implementation by named lexical ancestry. Its complete exact before/after owner text is checked by SHA-256, and the exact old header is replaced only within that owner. A changed, moved, duplicate, missing, or partial owner is refused before writing. The planned file must parse and every owner must have its complete after state. Both whole-file transpiles, with comments kept and removed, must be byte-identical. The source is reread before the single write. A second run makes zero edits and writes. Foreign text elsewhere in debug.ts is accepted and preserved.

`contract.cjs` measures first apply, replay with unchanged bytes/mtime, foreign-edit acceptance, changed-before-owner refusal, partial-signature refusal and changed-after-owner refusal. The three refusal fixtures intercept writes and record zero. `proposed-pins.json` is evidence for rejected trials; the adapter reads only `pins.json`.

Main contains no 99m directory. The real `adaptation_inputs` function sorts 99n after 99j and before 99-wasm on this base. A scratch fixture containing 99m uses that same function and places 99n immediately after it. No other adaptation is edited; see `sort.json`.
