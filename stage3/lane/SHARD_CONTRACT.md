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
look below `cache_root()/adamic-stage3-artifacts/INPUTS_HASH`, rooted at
`ADAMIC_BUILD_CACHE_DIR` on a runner.

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

# Chained Go products, wave 1

`stage3/lane` declares `TestProduct_Stage3LaneInstall`,
`TestProduct_Stage3LaneCompiler` and `TestProduct_Stage3LaneHarness`. Prepare
apply's `TestProduct_Stage3Adapted99` first. Each lane product fetches its
predecessor and fails if it is missing; no consumer silently builds a missing
predecessor. The gate supplies `ADAMIC_TYPESCRIPT_SOURCE` for apply. Loom
installs the tree NodeProjects `stage3/typescript` and `stage3/api` from their
lockfiles with npm 11.11.0; the lane reads `stage3/typescript/node_modules`. Product steps never
fetch packages or source. Install materializes the supplied packages after
checking their hidden lockfile, versions and the adapted tree's lockfile.

The final product itself has `ready.json`, `tree/`, `api/` and
`shards-ready.json`. It also contains `artifact.tgz`, a regular-file blob that
archives those four entries, plus phase metadata and an empty npm config.
Current buildcache stores regular-file blobs rather than implicitly archiving
symlinks: installed `.bin` aliases become relative shell wrappers which execute
the original target in its own directory. No compiler or test source is changed
by that packaging step. The original package bytes, including alias targets,
are pinned by lane/node_modules.sha256, measured from the npm 11.11.0 install,
and validated before copying. A mismatch names that pin and the reference npm
version. The hidden npm lockfile is included, matching the API pin precedent. The pin
participates in every product key and in the artifact input identity. A shard
runner needs the committed pin and fetched artifact, not a second upstream install.

The artifact's identity excludes both physical NodeProject installs;
`api/node_modules.sha256` pins API bytes, verified by apply and the final
producer; `lane/node_modules.sha256` pins the tree TypeScript install. This avoids binding a runner's extra lockfile marker. All other
previous artifact inputs remain bound, and the supplied upstream install's
full digest is now bound too through lane/node_modules.sha256. The tree
TypeScript package.json and lockfile remain direct identity inputs. Paths are absent from both digests. `shard.py`
and its payload, input, inventory, count, baseline and deadline checks are
unchanged. `bind_products.py ARTIFACT` updates the existing inventory's artifact,
runtime and unit hashes after this identity change; it does not repartition it.

Only `TestShard_Stage3LaneFile001` is declared in Go in wave 1. It reads
`stage3-lane-file-001` from shards.json, fetches the final product, runs the
manifest argv from the repository root with fresh STAGE3_RESULTS, and checks
both command success and the recorded counts. Missing or empty manifests fail.
The other 150 Go shard declarations are intentionally left for wave 2.

Planner follow-up: [READS.md](READS.md) inventories the additional Loom unit-key
reads, separately from buildcache's product inputs. The npm 11.11.0 pin includes
hidden npm metadata. A clean npm 11.9.0 install with the same lockfile and gate
flags has exactly the same digest and passes deliberately: the check binds
bytes, not the npm version string. An install with changed bytes fails naming
stage3/lane/node_modules.sha256. Neither physical NodeProject install is hashed
into the artifact identity; their declared byte pins and lockfiles are, and the
producer validates the bytes before publishing. Shard consumers fetch products
and need no installed packages.

The root .fast-gate-lockfile-sha256 runner marker is excluded from the pin and
from the copied product payload, so it cannot change bytes under the same key.
