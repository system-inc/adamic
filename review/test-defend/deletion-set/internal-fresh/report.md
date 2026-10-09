Deletion-set replay on b8bcadb2c493173855f19d7e5c508b34f5eeb5b6.

The requested audit branch test-audit/internal-fresh is absent. The defender report identifies the matching audit at test-audit/fresh; that branch supplies M2 and M18. D4 comes from test-defend/internal-fresh. Frozen candidate and other failure lists are in mutant-list.json, written before baseline. M2 is supplemental and M18 is an empty-answer probe in the original audit; both are included because this replay requested every candidate-catching diff.

Full-package baseline with the candidate skipped passed in 15.634 wall seconds. M2, M18 and D4 completed in 20.279, 19.558 and 20.245 wall seconds. Every diff applied without modification. All runs used uncached gates and distinct fresh build caches. No environment narrowing, stale diffs, test-binary panics, witness-only catches or pin-only catches occurred. TestNodeFSFileResultsAreFresh independently checks confined filesystem result proofs and caught all three. All observed top-level failures and passes are in matrix.json, and logs retain complete output.

Conclusion: the candidate is deletable within the gathered evidence; no demonstrated mutant loses its last catcher. This does not authorize deleting or rewriting it.

Warm tool setup was skipped. npm ci ran before baseline. Pinned cohere submodule is 7945d102a6c18dd36adf9114a758ce646e8b2359. Disk began with /tmp 8.3GB free (8.8GB total) and /workspace 13GB free. The earlier /tmp/defend-flow scratch was removed; protected repo and tools were untouched. No disk-full failures occurred. Production files are restored.
