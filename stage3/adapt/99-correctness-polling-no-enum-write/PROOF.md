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

## Completed acceptance, November 13, 2026

Both full applies, both stock 6.0.3 checks and both ordinary adamic types checks
on 79 compiler roots exit 0. The stock checks use Adamic's existing pinned Node
library and host-axiom input, with no relaxed options or filtered diagnostics.
Owner contract and replay exit 0, with zero replay writes and preserved mtime.
Both compiler/harness builds exit 0; emitted typescript.d.ts and the API reference
baseline are byte-identical. Only the pinned sys.ts owner changes among 747
source files and 81,500 tracked upstream files, including a post-lane comparison.

The unchanged count census exits 0 with zero enum writes and all 36 class (a)
readers retained. Restoring the original owner makes that same check exit 1 on
the original sys.ts:176 assignment. The four actual-source witnesses exit 0:
custom intervals schedule 17/41/89 before and defaults 250/500/2000 after; default
callers throw before and succeed after. No-variable outputs are byte-identical.
Published bundle and shipped _tsc.js controls, and 64 plain-table cases, exit 0.

All 398 unchanged lane shards exit 0: 106,366 passing, one expected API failure,
zero skips. The longest shard is 57.494 seconds, under 30 minutes. Final
artifact/actual-file audit and raw-evidence archive both exit 0. The evidence
branch retains every shard's raw observations, test output, API diff and report,
plus all child exit codes and the unfiled, tested upstream issue draft. Initial
setup failures are explicitly retained in ATTEMPTS.md, separate from acceptance.
