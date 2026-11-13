# Proof contract

Base: adamic main 28a8ac1fbe11d397d625d1dcdab52213b80d72a1.
Count: c7f2bc3d5d811387f053c2c7a392a424225702b5.
Upstream: TypeScript 6.0.3, 050880ce59e30b356b686bd3144efe24f875ebc8.

Evidence belongs only on typescript/polling-no-enum-write-evidence under
stage3/census/polling-no-enum-write. Its report records actual process exits,
including incomplete or failed checks, rather than treating this contract as
an assertion that a run passed.

Run full apply and contract.cjs on the completed tree; replay must report zero
files/edits and preserve bytes and mtime. Compare all source files against main:
only this whole sys.ts owner may change. Compare emitted typescript.d.ts and
the API reference baseline byte for byte. Stock 6.0.3 and ordinary adamic types
must each accept all 79 compiler roots with zero diagnostics. Run all 398
manifest shards in bounded processes and retain each exit and count.

Reuse census.cjs from the count to inspect the entire adapted source tree.
check-no-enum-write.cjs must report zero writes; restoring owners.json's before
function in a disposable copy must make the same check fail on the original
levels[level] write. No string-search substitute establishes this condition.

Run the count's node-reference.cjs source CommonJS mode before and after with
DynamicPriorityPolling and numeric interval variables 17/41/89. Before schedules
17/41/89 and default callers throw; after schedules 250/500/2000 and default
callers do not throw. Run both with all TSC_WATCH variables cleared and compare
their behavior. Retain published-package and shipped _tsc.js outputs separately;
for _tsc.js replace only the terminal executeCommandLine dispatch with exports
so its real initialized getNodeSystem can be exercised without running the CLI.

Native limitation: tsc's sys.ts does not lower yet, so a native run of this code
is not possible today. The Node source witness uses stock transpileModule and
is not native evidence.
