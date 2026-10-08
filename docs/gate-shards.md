# Full main gate across SSH boxes

`cloud/full-gate-main.sh` supports package shards of **the existing full main gate**
(`cloud/fast-gate/run.py --full`). The current fleet is home and server: both have 64 CPUs and the same chip.
Run the coordinator from the Mac checkout of the gate tools, with its existing push credential
and integration messaging setup:

```sh
export ADAMIC_FULL_GATE_BOXES="home server"
cloud/full-gate-main.sh <full-main-sha>
# Omit the SHA to watch new main commits continuously.
```

Each alias must already provide `~/adamic-tools/env.sh`, `~/full-gate/tree`,
`~/full-gate/tools`, and `~/full-gate/weights.txt`. The coordinator checks out the same source
and tools SHAs under each box's existing full-gate lock, discovers packages with `go list ./...`,
and requires identical package universes and weights on every box. It never provisions a box.
Multi-box mode uses all CPUs; `ADAMIC_FULL_GATE_SHARE` remains a single-box setting.
An unset box list retains the single-box runner. A one-entry list uses that alias.

The plan depends only on the package universe, parsed weights, and CPU counts from
`nproc --all`. Packages sort by descending estimated seconds, then import path. Each goes to
the box with the smallest projected total seconds divided by CPUs, ties by alias. Unknown
packages use one estimated second. Zero weights are allowed; negative and nonfinite weights
fail planning. `plan.json` records the complete inputs, assignments, estimated totals and
CPU-normalized totals; `coordinator.log` also prints it. Build, vet, coverage, the separate
WASI invocation, stage3 and catalog run only on the largest box (ties by alias). Census runs
there once after collecting every shard's test log. Package tests keep the existing full-gate
flags, including `-count=1`, the three-hour package timeout, and the TestWASI exclusion.
An empty shard explicitly records a successful empty tests stage and never tests `.`.

Published logs retain `gate-logs/<sha12>/<stamp>/full-main`: aggregate `test.jsonl`,
`full.json`, `status.txt`, `first-failure.txt` on red, `plan.json`, `census/`, and
`shards/<index>/` with each box's evidence. Live snapshots publish a red as soon as it is
observed; that first observed failure remains sticky and names its shard and alias. Other
shards continue for triage. SSH failures, collection failures, missing reports, incorrect
source/tool/package identities, unreported stages, or missing census cannot produce green.
`ADAMIC_FULL_GATE_TIMEOUT` sets a positive per-shard wall limit in seconds (default 86400),
in addition to the existing per-package timeout. The Mac coordinator must stay running;
an interrupted run cannot publish a green final report. A completed red returns to the loop;
network blips are retried by the loop. Heartbeat rows on `records/full-gate-heartbeat`
(`documentation/velocity/full-gate-heartbeat.csv`) record start, running every ten minutes,
and idle. The `box` field names both aliases (`home server`) for a sharded run.

### Equivalence proof and first live run

No distributed equivalence is claimed before the following comparison passes on one main.
Whole packages have no repeated subtest parents across boxes. Each package's ordinary
invocation appears once, and the separate WASI invocation appears once, just as in the
unsharded full gate. Therefore the proposed partition preserves the invocation multiset;
the raw event comparison below checks the measured pass/fail/skip multiplicities per package,
including subtests, examples and fuzz seeds. It does not deduplicate events or package outcomes.

1. On the Mac, fetch and check out `codex/full-gate-shards`. Record its full tools SHA and
   choose one full main SHA. Stop another full-main coordinator before using its checkouts.
2. Confirm the two aliases and their provisioned inputs locally; use the same weights
   and toolchains. Run `ADAMIC_FULL_GATE_BOXES="home server"
   cloud/full-gate-main.sh <sha>`. Check `plan.json` for CPUs 64/64, exactly-once package
   assignments, and home as the once-only owner. Save the published branch name and final result.
3. Run the same main with the same tools in single-box mode:
   `ADAMIC_FULL_GATE_BOXES=home cloud/full-gate-main.sh <sha>`. Save its separate timestamp
   and published branch. Both runs must finish; investigate any red before enabling the loop.
4. Extract `test.jsonl` from each published branch into separate local directories. For
   large logs extract `test.jsonl.gz` and decompress it. The aggregate log is at the root
   of each published branch. Keep the plan and full reports beside the logs for audit.
5. Compare:

   ```sh
   python3 cloud/fast-gate/shards.py compare \
     --whole /path/to/whole/test.jsonl \
     --sharded /path/to/sharded/test.jsonl
   # Alternatively --sharded accepts all individual shards' test.jsonl paths.
   ```

   Exit 0 prints equality; exit 1 lists every differing package/action count. Malformed JSON
   fails rather than being silently ignored. Preserve both logs and the comparator output
   as the equivalence evidence. This comparison checks counts; assess gate verdicts separately.
6. Test a deliberately unavailable alias with a one-shot multi-box run; expect a named red,
   never green. Then enable the continuous loop with the production alias list.

## Earlier subtest-sharding tool

The separate [`cmd/adamic-gate`](../cmd/adamic-gate) tool provides `plan`, `shard`,
`merge`, and `compare` commands for splitting individual tests and audited subtests.
Its implementation and tests document its discovery, repeated-parent accounting,
resume support and coverage rules. It is not used by `cloud/full-gate-main.sh`;
follow the package-sharding procedure above for the full main gate.
