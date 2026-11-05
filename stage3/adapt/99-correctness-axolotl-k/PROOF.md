# 99k proof

All measured checks below pass on initial main 6761c7374 plus this adaptation.
The adapted upstream pin is 050880ce59e30b356b686bd3144efe24f875ebc8.

| Check | Result |
| --- | --- |
| Count before creating the adapter | 6 identity-return calls, 0 non-full; 7 wider calls, 0 non-full |
| Stock TypeScript 6.0.3 compiler project | 0 diagnostics |
| Adamic types, complete sorted 79-root input | 0 diagnostics before, 0 after; both commands exit 0 |
| Node witnesses using actual extracted functions and tsc's real repopulation helpers | 11 valid before/after rows identical; 4 invalid identity-return cases checked |
| Mutant restoring the unchecked chain cast | Red, exit 1; all 4 invalid cases escape the guard |
| Owner contract | 13 no-write refusals; unrelated edit accepted; second run preserves bytes and mtimes with zero writes |
| Full stage3/lane/run.sh | PASS; apply exit 0; 106,366 passes, one sanctioned API failure, 0 pending |
| Complete pass/fail identity lists | Byte-identical sorted lists, including multiplicities; 106,367 outcomes |
| Both public declaration files | Byte-identical to main |
| core.ts emitted JavaScript | Byte-identical, with comments kept and removed |
| generate_shards.py --check | Exit 0 |
| Go stage3/lane fixtures | Exit 0 using 99j's fixture command |
| Per-file meter census | All 79 roots measured; all 73 core.ts:356 movers now stop at core.ts:381:26 |

The Go fixture command is `go test -count=1 -skip 'TestShard_|TestProduct_' ./stage3/lane/`,
the same boundary as 99j's retained proof. It runs the lane fixtures; the complete
upstream suite is independently covered by the unfiltered, eight-worker oracle.
No oracle runner, timeout, test expectation, or worker count was changed.

The initial scratch S2a build hit a Git VCS-status error because its dependency
was symlinked. Replacing that task-created symlink with local shared clones at
the same pinned commits fixed preparation; the actual pinned S2a build exits 0.
The earlier full-history dependency download was stopped and the exact checker
pin fetched shallowly. Neither incident changed compiler sources or pins.

proof-results.json summarizes results. go-checks.json records exact commands,
exits and times; compiler-provenance.json records both compilers and dependencies.
Compressed type logs, all three census streams and driver logs, and both complete
oracle identity lists are retained. The identity observer records original Mocha
host-replay events without selecting tests or changing outcomes.

Core transpile outputs are retained in the two samemap-final-witnesses-core
JavaScript files. Their SHA-256 is `5dfe133e4cc0b9d06db5ec8a644a27988e6713901698b0413c75698597be860c`.
The adapter itself also checks complete core.ts JavaScript equality with comments
kept and removed before writing.

The README and LEDGER note record the pre-build count and exact property inventory.
CENSUS.md gives the grouped new first-stop table; census-results.json lists every
root, including the other six. This is a first-stop lowering census, not a claim
that TypeScript compiles natively.

Immediately before pushing, origin/main is fetched, the one adaptation commit is
rebased if needed, and the complete apply is rerun. Its keyed cache replay is
compared byte-for-byte, and the adapter is checked for zero edits and writes.
Final push-time logs stay outside this hashed adaptation directory at
/workspace/scratch/samemap-prepush.json and samemap-final-apply.log. That avoids
making the verification log itself change apply's input key after verification.
