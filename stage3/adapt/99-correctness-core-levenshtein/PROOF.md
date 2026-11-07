# Post-apply checkpoint

| Check | Result |
| --- | --- |
| Real-call count | 798965 calls, 0 differences, all 398 accepted shards |
| Unchanged full lane control | Pass; identities, counts and raw baseline diff match |
| Owner contract | All 29 checks pass |
| Full main apply | Exit 0 |
| Full main plus this adapter apply | Exit 0 |
| Source scope | All 746 sources compared; exact pinned owner is the only change |
| 79-root types, full paired replay, JavaScript/declarations, census and mutant | Pending final proof run |

Both complete apply logs, results and frozen adapter hashes are packed in
apply-evidence.tar.gz. This checkpoint preserves the completed count and
implementation before the remaining proofs. The initial report remains
historical in initial-report.tar.gz. No completed proof is claimed for the
pending items. No earlier adapter or compiler source changes.
