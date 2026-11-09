# Relation overflow on the real tsc entry

Task ny0c3gb. Intermediate checkpoint after real checking and production lowering. Unit continuation and the independent census are pending.

Main: c3b7baf0b3438bf5ac7aa9fcf6eb2ba385e2f9bd. Pinned TypeScript: 050880ce59e30b356b686bd3144efe24f875ebc8. Required setup, adapted tree, main adamic, and both helpers built successfully.

After checking: RelationOverflow is empty, RelationOverflowMarked is false, all six relation caches carry zero overflow flags, and TS2859 is absent. Exact sizes are in real-checked.json.

After production lowering: the same readers remain clear, all six caches carry zero overflow flags, and TS2859 is absent. Exact sizes are in real-lowered.json. Lowering returns no IR and stops at src/compiler/core.ts:2599:15 on the process host binding, pending runtime/platform clearance. The same checker is now running the census's fresh lowering per top-level unit.

The final report will replace this intermediate checkpoint.
