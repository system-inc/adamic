# Lane hashes from the tree

Base main tip: `80c2d3282e44b527555d59690bb413cc883d2bf2`.
Only stage3/lane changes. stage3/apply.py, adaptations, cloud/ and loom are unchanged.

The accepted preflight reproduced all 398 committed unit hashes byte for byte
with the original artifact key and runtime. The new shared `unit_inputs_hash`
helper uses that same digest. The old artifact key was box-specific because
`artifact.input_key()` includes installed stage3 files, tool versions and platform.
`shard.py:adapted_inputs` never compared that committed key on the product path:
it returned STAGE3_ADAPTED_TREE after harness, API and runtime checks.

The plan retains the same 398 definitions, assignments, expectations and API
digest. Only artifact and per-unit inputs_hash are removed, with runtime refreshed
for the changed harness. Commands drop --inputs-hash. Runner reports retain the
key and hash computed from the current checkout. Shared artifacts still require
matching ready key and inputs, successful phases and intact payload and API.
Merge recomputes report identities with artifact.input_key and the same helper.
Its API check uses the checkout's pinned parser, avoiding an override before the
key calculation (input_key deliberately rejects TSC_ADAPT_TYPESCRIPT).

## Mutants

`python3 -B stage3/lane/prove_hash_mutants.py` runs each retained patch in memory
and requires exactly one assertion failure from its named test, with no errors:

| Mutant file | Test in test_shards.ShardTests |
|---|---|
| stale-runner-hash-mutant.patch | test_tree_hash_reports_change_after_adaptation_edit |
| accept-other-tree-mutant.patch | test_merge_rejects_report_key_from_another_tree |
| artifact-in-plan-mutant.patch | test_plan_bytes_ignore_adaptation_edit |

The runner test edits an adaptation README in a scratch tree and checks both
reported identities against the tree and against their previous values. The
merge test holds the report hash and all other evidence valid but changes its
artifact key. The regeneration test compares serialized plan and manifest bytes
after an adaptation edit, with fixed tool responses and no dependency on a box.
The existing five shard guard mutants are also caught.

## Regeneration property

An isolated scratch copy edited stage3/adapt/99-hash-proof/README.md and executed
both shard_plan.py and generate_shards.py before and after. All three files match
each other and the delivered files byte for byte:

| File | SHA-256 before and after |
|---|---|
| shard-plan.json | 5be9071c7d78d0241f4d86a171c4b71589e73d809d8d74c2bf7310687ec1eec0 |
| shards.json | c6aac4b07abc5164021e2382fccef265b7a0047de40bd33b2151b6ce5185a647 |
| shards_generated_test.go | f9e1eb2afbdd50f7c6afe90ea155d9216079da5f494ce5d32d0988805a1b9a1b |

Evidence is in evidence/hash-from-tree/.

## Checked-in patch-set.md readers

Repository-wide grep was inspected to exclude output-tree tables, generated
fixture tables, archived evidence and prose references. The readers of the
checked-in stage3/patch-set.md are:

- stage3/lane/artifact.py:31 selects it in the broad stage3 scan, then :20 reads its bytes.
- stage3/test_apply.py:57 names it and :60 copies the checked-in file into the test checkout.
  Reads of that scratch copy occur at :66, :71 and :81.
- stage3/adapt/75-optional-widening/coverage/run.py:29 reads the checked-in bytes;
  :32 copies the checked-in file (the path is assigned at :28).

stage3/apply.py:393 writes the checked-in table only with --write-table; it reads
the output copy. stage3/lane/run.py:81 reads the output tree's table. Neither is
a reader of the checked-in table. No patch-set.md file was changed.

Three historical adaptation evidence scripts also read the removed fields:
`stage3/adapt/40-explicit-any/evidence/brands-undefined-3/rebind-plan.py:9`,
`stage3/adapt/40-explicit-any/evidence/brands-undefined-3/summarize-shards.py:10`,
and `stage3/adapt/40-explicit-any/evidence/timer-slot/compare-oracles.py:13`.
They remain unchanged under the required stage3/lane-only scope and require their
historical plan schema if replayed. Active lane readers use computed identities.

## Validation

- `python3 -B stage3/lane/generate_shards.py --check`: passed.
- `GOFLAGS='-parallel=4 -timeout=30m' go test -count=1 ./stage3/lane`: passed,
  1281.593 seconds, with the adapted-tree product already prepared. The longer
  package timeout accommodates the complete package on this four-CPU worker.
- The new Python cases run through the existing TestLaneShardsPython wrapper.
  They run no Go or compiler subprocess. No new Go test declarations were needed.
- Shared artifact preparation completed all producer phases successfully in
  273.433 seconds. The real artifact was accepted by the new identity guard.

The accepted preflight's per-unit hashes, mutant failures, property hashes and
successful Go/setup logs are retained under evidence/hash-from-tree/.

The separate manifest run used STAGE3_ADAPTED_TREE with the main-tip gate-built
product and concurrency 2. All 398 commands and lane units passed. Its 106,367
observations contain exactly 106,366 passing tests plus the unchanged standing
Public APIs failure, zero skipped tests, and only api/typescript.d.ts as a
baseline difference. The executed identity multiset matches the original full
inventory. Every report's artifact key and unit hash was recomputed and checked
against the current tree.

Maximum unit wall: 16.551606 seconds. Maximum complete command wall: 16.740110
seconds. Total measured wall: 1908.575734 seconds. The worker has four CPU quota
cores (cpu.max 400000 100000); this measurement overlapped the Go package check.
All 398 units and commands remained below 30 seconds.

`lane-summary.json` records counts, identity checks, input versions and timings;
`reports.json.gz` retains every actual unit report. Full per-unit observations,
logs and the ledger remain at /workspace/scratch/sharded-lane. The shared build
log is /workspace/scratch/prepare.log, and the full Go log is
/workspace/scratch/go-lane.log. The standalone 20-case shard suite also passed;
the required full Go check subsequently passed with the same production code.
