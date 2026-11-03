# The plan's runtime field is checked against the tree

shard.py refuses every lane unit when `shard-plan.json`'s `runtime` isn't `shard_plan.runtime_hash()`, and `RUNTIME` holds
`sanctioned-api.json`. 99c's first cut (chg_ynyzx2ka, 524d2104) regenerated the sanction without moving the field, and all 48
lane units went red in Loom. `generate_shards.py --check` compared only the Go declarations with shards.json, so the cut's
checks passed. Now `--check` also compares the runtime field, and `test_plan_runtime_is_this_trees_runtime_hash` in
test_shards.py (run by TestLaneShardsPython) holds it.

Mutant: one newline appended to `stage3/lane/sanctioned-api.json` on main 8012e48d, the plan left as it was.

```text
$ python3 -B -m unittest test_shards.ShardTests.test_plan_runtime_is_this_trees_runtime_hash
AssertionError: 'b38a379acd26a09bdf89b450bedd0193d4dc2a2ede9a87461fbd90371f314202' != 'd8c20cc44ebce2cc23339aab01fa9204ac86015f57e6cefdd56596a031d2082a'
$ python3 -B generate_shards.py --check
shard-plan.json runtime differs from shard_plan.runtime_hash(); a file in RUNTIME changed, such as sanctioned-api.json, so the plan must move with it
```

Both pass with the file restored.
