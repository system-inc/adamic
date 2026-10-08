Prove a fresh automatic array's slot storage from all checker observations and intrinsic writes.
Commits: codex/generic-function-value, 2026-10-08 UTC; host integration stays on our scratch branch.
Checks: positive Node/native/JavaScript oracle 0.460s; negative Node and pinned NotYet; full lower 16.171s; vet clean.
Mutant: keep only the final observation instead of requiring all observations to agree; TestEvolvingArrayRejectsChangingStorage fails with got nil.
Uncovered: non-const automatic arrays, erased escapes, differing observed element contracts, general any, and whole-repository/count regeneration gates.

The initializer must be a fresh empty literal with no annotation. Every typed read supplies exactly the same concrete checker element type; every push value must be assignable with invariant writable slots and matching ownership. Erased any[] is allowed only at the intrinsic write receiver or length observation. The helper chooses allocation and write storage without changing checker types, signatures, or contextual propagation. Explicit any arrays are unaffected.

collect<T> at number, string, and an object type agrees with Node in both backends. The storage-changing source prints 1 and 2 on Node but retains the pinned unsupported array-of-any diagnostic. /tmp/evolving-array-root-{oracle,controls,negative-node,mutant,counts,lower,vet}.log records evidence. Counts alloc/free 12/12, retain/release 14/22, peak 8, regions 0.

The pinned host fixture advances past result.push(v) at 1100:17 to reading console at 1296:5. The earlier array-union report described 1100:17 as addRange incorrectly; the exact checker trace shows result.push(v), with any[] at that receiver and NonNullable<T>[] at typed reads and return. Full fixture 25 does not yet pass. No additional nullable kind has been reached.
