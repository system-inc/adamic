# Gate manifest contract

`shards.json` is exactly the developer-tools version-1 schema: `version` and
`units`, with `name`, `command` and `expect` in each unit. Commands are argv from
the repository root. The last argument is the unit's SHA-256 inputs hash.
`expect` uses `pass`, `fail` and `skip`. Every successful command exits zero,
including the unit which validates the single sanctioned upstream failure.

This measured manifest is for Linux x64 with Node v24.19.0. Its 151 units comprise
134 deterministically balanced file shards, 16 individually selected watcher
cases, and the byte-comparison unit. Only `stage3-lane-byte-comparisons` expects a
failure, the original Public APIs failure. It owns every suite from publicApi.ts,
checks the exact API diff and applies the unchanged stock-parser declaration
sanction. Every other unit requires no baseline mismatch and no failure.

Before starting test units, produce or restore the shared build as an artifact:

```sh
python3 stage3/lane/prepare_shards.py /absolute/shared/artifact-cache > /tmp/stage3-artifact.log 2>&1
export STAGE3_ARTIFACT=/absolute/shared/artifact-cache/INPUTS_HASH
```

The producer runs apply, locked install, compiler build and harness build once.
It is not a test unit. It preserves the stock TypeScript 6.0.3 API parser beside
the build. A cache hit verifies all payload bytes. A gate must prepare or restore
this artifact before launching manifest units; units fail if it is absent and
never silently build it. `STAGE3_ARTIFACT` names the directory containing
`ready.json`, `tree/`, `api/` and `shards-ready.json`. Without the variable, units
look below `$HOME/.cache/adamic-stage3-artifacts/INPUTS_HASH`.

Each unit receives `STAGE3_RESULTS` and creates its named directory below it.
The literal `$STAGE3_RESULTS` argv entry is expanded by the command itself.
Use a common results root, or collect each named result directory beneath one
root before merging. Every unit may have its own `STAGE3_CACHE`; no unit uses
that cache to build or fetch the shared artifact. Tests use fresh processes with
Node's compile cache disabled, private mutable outputs and the verified artifact.
They never reuse a previous test result. Existing result directories are refused.

`shard-plan.json` contains the tasks, exact anchored watcher grep, expected
identity digest and inputs hash for each unit. All other unit suites from one
source file stay together. Watcher selections go through upstream's own
`--tests` parser and `runConsoleTests` Mocha grep, preserving setup/cleanup hooks,
original waits and the 40-second upstream timeout. The test-only entry function
avoids Hereby's build dependencies. Node and Python command walls, artifact
verification, baseline scanning and API checks all count toward 30 seconds.

The inputs hash binds the prepared input tree, stock parser bytes, complete
runtime harness and the unit definition. Different build inputs, tool versions,
platform, harness, task assignment or expectations require a new hash. Units
rehash the whole artifact and stock parser before execution. They also check
counts, selected test identities, task completion, failure titles, baseline
paths, upstream exit and elapsed wall. No acceptance predicate in check.py or
expected.json has been removed or relaxed.

The gate wiring belongs to developer tools on `devtools/fast-gate`, in
`cloud/fast-gate/run.py stage3()`. This branch does not modify it.

# Executed union and merge

The merge is a separate script, not a parallel manifest unit with missing
predecessor dependencies. After all units finish, it consumes their actual
observations and a same-artifact unsharded lane measurement:

```sh
bash stage3/lane/run.sh /absolute/new/whole --artifact "$STAGE3_ARTIFACT" > /tmp/stage3-whole.log 2>&1
python3 stage3/lane/merge_shards.py "$STAGE3_RESULTS" --whole /absolute/new/whole --output /absolute/new/merge.json > /tmp/stage3-merge.log 2>&1
```

The unsharded command is an explicit reference measurement, outside the gate
unit tier. Ordinary run.sh still applies and builds as before. With `--artifact`
it verifies and shares the same producer's tree, keeps private baselines, runs
all tests with eight original upstream workers, and applies the unchanged lane
checker. Producer install/build exits and times are labeled as producer evidence,
not represented as newly executed unit phases.

The merge compares actual test identities with multiplicity, summed counts,
failure titles, baseline path sets and the complete diff bytes. Its reference
must carry the same artifact hash and pass the existing full lane checker. A
missing, repeated, swapped or extra test cannot be replaced by an equal total.
Any red unit, input mismatch, wall at or above 30 seconds, count mismatch or
unsanctioned failure rejects the merged verdict.

The complete listing in `evidence/case-shards/tests-and-units.jsonl.gz` assigns
all 106367 observed test results once. The reproducible measurement coordinator
is `measure_shards.py`; it invokes the actual manifest commands, gives each its
own STAGE3_CACHE, records complete process wall times and preserves every log.
Its coordinator wall is not a test unit. `--only-prefix` supports targeted proof
measurements without altering the manifest or skipping a test in normal runs.

`--mutant-runner`, and `--mutant-case` with `--mutant-content`, are explicit proof
options to shard.py. They change only a private test input and record its SHA-256.
They leave normal gate commands and expectations intact. Failure text, stacks,
timing and baseline diffs are preserved per unit by the step12 observer.
