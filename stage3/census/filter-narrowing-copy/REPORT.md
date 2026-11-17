# Narrowing filter copy stop report — #ev14ch9

Base: `49f1d643064a9231f7f343d8bbd02ba718866e6a`.
TypeScript: `050880ce59e30b356b686bd3144efe24f875ebc8` (v6.0.3).
Freeze evidence: `df886e406b521bacf030d7e8ca76d6e828fed249`.

## Gate result

The two out-of-scope readonly narrowing overloads both refuse in
`lowering.declareModule` for the same generic result mismatch:

- Original overload 3, core.ts:265:1: `readonly U[]` cannot be served by
  `readonly T[] | undefined`.
- Original overload 7, core.ts:273:1: `readonly U[] | undefined` cannot be
  served by `readonly T[] | undefined`.

Each control removes preceding signatures only in a scratch copy, retaining
blank lines so original source coordinates survive. The resulting ordinal in
the diagnostic is 1; the original ordinals are 3 and 7. The implementation
is the complete unchanged adapted filter body, including baseline indexed-read
non-null assertions. No landing filter edit is proposed. The original-owner
probe stops at core.ts:261:1 on overload 1, with or without callers: declaration
judgment does not depend on whether the signature has callers.

An initial isolated probe used upstream's unadapted indexed reads and failed
Load with three TS2345 diagnostics. Those are not lowering evidence. The final
controls use actual adapted indexed reads and load successfully. Earlier controls
removing only mutable narrowing signatures first encounter mutable boolean
overload 2; they do not establish anything about later readonly signatures.
The separately isolated original-3 and original-7 controls do.

The scratch helper probe is checker-clean and its guarded full latent run has
zero findings, with concrete instantiation witnesses for both plain helpers,
including the undefined-array call. This reduced probe alone does not establish
full core.ts coverage, general freshness certification, backend acceptance, or
any of the unrun lane/declaration/mutant checks.

The authorized change stops here under condition 6. It is not widened to readonly
narrowing or boolean overloads. No adapter, ledger acceptance line or landing
branch is created.

## Remaining proof obligations (not claimed)

A complete accepted copy requires the premise that no reader compares a
narrowing filter's result with its argument by identity, directly or through a
slot. The freeze run preserves identity and cannot witness that premise. Its
242,183 additional allocations (232,451 at checker.ts:16764) remain the measured
price from df886e40, not a new copying-lane measurement. Should #4yrezf5 is the
specified later erasure task. There is no successful-copy claim for the ledger.

The four sites were read, not run: checker.ts:9417 and :9419 create temporary
arrays whose filtered identifiers flow through flatMap/map into export specifiers;
transformers/ts.ts:1088 reads a skipWhile result and adds elements into a separate
modifiersArray; services/callHierarchy.ts:482 groups a temporary reference-result
array and maps entries to spans. The inspected paths do not compare argument/result
array identity or write through either handle after the filter call. This hand
review does not replace a copying lane or prove arbitrary transitive aliases.

Not run after the stop: copying lane (106,366/1), alias-restoring mutant, Node
runtime witnesses from a built adapted compiler, stock 79-root diagnostic delta,
`adamic types`, public declaration builds/comparison, adapter contract/replay,
restoring adapter mutant, and TestProduct_Stage3AdaptedTree. The reduced probe's
measurement success does not stand in for those checks.

## Checkpoint

Unchanged full adapter build still running. Complete-tree census/count logs will be appended.
