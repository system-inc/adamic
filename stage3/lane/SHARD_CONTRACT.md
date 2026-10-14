# Gate manifest contract

`shards.json` is exactly the developer-tools version-1 schema: `version` and
`units`, with `name`, `command` and `expect` in each unit. Commands are argv from
the repository root. The last argument is the unit-definition and runtime SHA-256 hash.
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
identity digest and runtime/unit-definition hash for each unit. All other unit suites from one
source file stay together. Watcher selections go through upstream's own
`--tests` parser and `runConsoleTests` Mocha grep, preserving setup/cleanup hooks,
original waits and the 40-second upstream timeout. The test-only entry function
avoids Hereby's build dependencies. Node and Python command walls, artifact
verification, baseline scanning and API checks all count toward 30 seconds.

The committed inputs hash binds only the complete runtime harness and unit
definition. The manifest never commits an artifact key or API digest. Each run
computes artifact.input_key() from its actual checkout and toolchain and requires
both that key and its complete input description to equal ready.json. A mismatch
names the checkout key and artifact ready key. The API digest comes from the
fetched shards-ready.json, whose artifact key must equal ready.json. Units rehash
the whole artifact and stock parser before execution. Changing an adaptation,
census README or evidence outside lane/oracle requires rebuilt products, with
no manifest regeneration. Changing the runtime or unit definition requires a new
committed unit hash. They also check
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
`TestProduct_Stage3LaneCompiler` and `TestProduct_Stage3LaneHarness`. Preparing apply's `TestProduct_Stage3Adapted99` first keeps unit wall times short. Every lane consumer calls the producer's own buildcache recipe. On a miss,
install builds apply's chain, compiler builds install, harness builds compiler,
and a shard builds harness. Dependencies are resolved inside the consumer's
build callback, so a hit does not prepare earlier products. First-tree ordering
is not required; later trees can use Loom's recorded read-set ordering. The gate supplies `ADAMIC_TYPESCRIPT_SOURCE` for apply. Loom
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
unchanged. `bind_products.py` updates only the existing inventory's runtime and unit-definition
hashes. It needs no artifact and does not repartition the inventory.

Only `TestShard_Stage3LaneFile001` is declared in Go in wave 1. It reads
`stage3-lane-file-001` from shards.json, gets or builds the final product, runs the
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

# First-tree product dependencies

`adaptedTree` calls the existing exported applyproducts.Product recipe at its
product-99 slot. No apply key or worker recipe is copied into lane.
TestStage3LaneAdaptedTreeMatchesProducer independently compares its returned
product directory with applyproducts.Get, the producer's final-product resolver.
A wrong-slot mutant must fail with the consumer and producer keys named.

Every lane product and shard resolves a missing dependency by its original
producer recipe. Workshop does not need a preexisting read set on a tree's first
build. Apply still needs the gate-provided pinned source and installed packages.
The first-tree chain is explicitly allowed to exceed the ordinary cold unit
budget and is measured with a hard 14-minute Go limit. Prepared units retain
90-second test limits and all existing shard acceptance checks.

# Generated Go units, wave 2

`generate_shards.py` generates all 151 `TestShard_` functions in
`shards_generated_test.go` from shards.json. Each function calls the common
manifest reader, gets or builds the final product and executes the manifest's argv and
expect counts with fresh t.TempDir results. `TestShardsMatchManifest` compares
canonical generated bytes (including the complete manifest fingerprint) and the
Go AST's complete TestShard_ namespace. Missing, added and renamed tests fail.
Run `python3 -B stage3/lane/generate_shards.py` after rebinding the manifest.

The 17 `TestStage3ApplyShardNN` units use test_apply.py's original selected_suite
partition, now containing 25 discovered cases. The historical SHARDS_REPORT.md
had 17 discovered cases; the partition count remains 17 and the current case
inventory is in python-units.json. `TestStage3LaneLandingLaneTests` and
`TestStage3LaneLaneTableTests` preserve the two original unittest classes from
test_check and test_table. `TestPythonUnitsMatchCommands` proves their exact
case union equals the old command discoveries and checks every Go wrapper is
present. Each unit also checks the actual executed Python case identities and
outcomes, not just the Python process exit. Apply units get or build product 99 first;
the Python child restores its verified archive directly via STAGE3_PRODUCT_STORE
and uses buildcache read mode, so a miss cannot trigger rebuilding.

Normal Go shards use private temporary results. For measured union evidence,
STAGE3_UNION_RESULTS names an existing directory: only after the normal verdict
passes does the Go caller retain report.json, observations.json and the raw
baseline/test logs beneath its uniquely created unit directory. It refuses a
preexisting unit directory. STAGE3_PYTHON_RESULTS similarly retains the observed
Python case records. Neither option changes manifest execution or expectations.

`measure_go_units.py RESULTS` runs every top-level lane Go test in its own fresh
go test process, with -count=1, -timeout 90s and an outer 90-second process limit.
Products must already exist. It records both Go test wall and complete process
wall, flags every test over 30 seconds and fails any over 60. Its resumable
coordinator is outside the unit tier. The report sorts all unit walls slowest
first. The whole-lane reference is a separate bounded run.sh --artifact command;
merge_shards.py compares the complete executed identities, outcomes and
multiplicities, counts, original sanction and raw baseline bytes. Missing shards
report missing test names. prove_wave2.py exercises generator, namespace,
Python-coverage and outcome/multiplicity mutants against the real Go checks.
