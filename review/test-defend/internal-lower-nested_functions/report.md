# Nested-function defense

Main: d29d80ceb5d5d42d9b0ffb7b528a57a2272c76f4. Both rows are defended in the completed whole-package matrices. Each mutant failed exactly one top-level test, passed 273, and skipped TestOptionalWideningCensus and TestOriginalCycleLedger. Results JSON files list every observed passing row. Uniqueness is limited by these two skips.

Code under test: Adamic Go lowering, including capture ownership and cycle analysis. Oracle: the cycle family expects Refused and adamic/cycle-capable diagnostics written in the tests; the captured-parameter row compares execution with Node and independently asserts a self-written IR ownership invariant.

Family members: TestNestedFunctionCycleIsRefused, TestNestedEnvironmentCycleIncludesDisjointSlots, TestNestedCallbackCycleIsRefused. D2 only fails the disjoint-slot member, so it uniquely catches this family. Named nested functions sharing a frame must retain all its cells even when they directly capture disjoint slots. The literal-method subsumer does not exercise this frame-layout propagation. Coverage comparison shows 221 exclusive covered blocks for the family versus the subsumer.

D1 preserves capture cells but marks a reference parameter borrowed. The ownership row checks this flag directly. Its subsumer captures numbers and checks allocation structure. They share ownership code but differ semantically in parameter representation and assertions. Coverage comparison shows nine exclusive covered blocks for the ownership row.

Both standalone diffs were compiled by go vet ./internal/lower/ and exercised in the full package. Source was restored after each run. No tests were changed. D1 wall time: 47.561 seconds; D2: 67.218 seconds. Clean baseline binary time: 46.873 seconds. Toolchain warm, setup skipped; npm ci completed before baseline. nproc: 5. Four individual coverage runs and their profiles are retained.

Friction and limitations: /tmp is only 8.8 GB total, so the requested 15 GB free threshold cannot be met there. Earlier /tmp/regexp-deletion scratch was removed; afterward /tmp had 5.2 GB free and /workspace 14 GB free. They are separate filesystems. No disk-full failure occurred. The audit used an older main and current ownership/allocation tests additionally compare behavior with Node. The supplied audit snippets omit commands described as below; the fetched artifacts supplied them. Go coverage measures lowering, not the native execution itself. Neither target's name promises an assertion absent from its body. No further mutant attempts were necessary after unique catches.
