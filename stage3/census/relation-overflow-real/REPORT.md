# Relation overflow on the real tsc entry

Task ny0c3gb. Canonical entry-order checkpoint. Unit continuation and independent census end-state results are pending.

Main is c3b7baf0b3438bf5ac7aa9fcf6eb2ba385e2f9bd. TypeScript source is pinned at 050880ce59e30b356b686bd3144efe24f875ebc8. Setup, adapted tree, main adamic, and corrected measurement helpers built successfully.

After checking and after production lowering, RelationOverflow is empty, RelationOverflowMarked is false, all six caches carry zero overflow flags, and no TS2859 is held. The repeated measurements agree exactly with the initial samples. Sizes are in real-checked.json and real-lowered.json.

Production lowering stops in the refusal scan at src/compiler/core.ts:2599:15, the ambient process host binding, before declareModule and moduleStatements. It returns no IR.

The current continuation and independent census use canonical LatentEntryReach from stage3/meter/entry_overlay.py: tsc.ts stays first, then dependencies follow in checker order. An initial attempt used replay ordering and was terminated; it is archived under invalid-entry-order-* and excluded from the unit comparison. The corrected unit walks are still active. No final unit counts or flagged-pair claims are made in this checkpoint.
