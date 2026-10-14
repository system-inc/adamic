All lane consumers now build missing products through their producers' original recipes and keys.
Products: typescript/stage3-lane-products-4 at 924b31bbf102e9c7aa43dc80364be34a2fbb9471; child typescript/stage3-lane-shards-3 tip is recorded in delivery.
Fresh install succeeded in 231.668250s, apply 99 reused in 1.222069s, fresh shard 001 built all nine products and passed in 365.355186s.
Producer-key mutant and retained lane mutants caught; runtime binding, manifest checks and the retained 106367-execution union pass.
Only stage3/lane changed; applyproducts is untouched; union records are retained, not a new whole/all 151 execution on the fresh artifact.

# One recipe per product

adaptedTree calls the existing exported applyproducts.Product(t,5), the producer's current 99 slot. Independent applyproducts.Get returns its own final 99 directory; the comparison test requires exact canonical directory/key equality. Selecting70 instead fails naming both directories. This uses existing producer APIs, with no copied Inputs, key, flags, worker code or apply edits. All three lane products use the same laneProduct implementation for their own declarations and consumers. Predecessors are resolved inside buildcache.Product callbacks, so hits avoid earlier products and misses hold their own dependency chain.

Every generated TestShard calls laneProduct for harness. Every apply Python wrapper calls adaptedTree before starting the Python child; that child still restores the verified prepared archive in read mode. The lane-only unittest classes do not consume a product. There are no remaining applyproducts.Fetch or fetchLane calls in Go code.

# Exact cold and verification command walls

nproc 5, CPU quota 4, Go1.27.1, Node24.19.0, npm11.9.0. Cold caches were created as absent directories and removed only after their proof records were saved. The fresh install and fresh shard are separate caches. First-tree long paths use the user-authorized exception: go test -timeout14m and an860-second subprocess bound. Prepared units retain90s limits. Every command below has its actual exit and monotonic process wall, including Go startup/compile overhead.

| Command | Exit | Process seconds |
|---|---:|---:|
| `["go", "test", "./stage3/lane", "-run", "^TestProduct_Stage3LaneInstall$", "-count=1", "-timeout", "14m", "-v"]` | 0 | 231.668250 |
| `["go", "test", "./stage3/applyproducts", "-run", "^TestProduct_Stage3Adapted99$", "-count=1", "-timeout", "90s", "-v"]` | 0 | 1.222069 |
| `["go", "test", "./stage3/lane", "-run", "^TestStage3LaneAdaptedTreeMatchesProducer$", "-count=1", "-timeout", "90s", "-v"]` | 0 | 1.872135 |
| `["go", "test", "./stage3/lane", "-run", "^TestStage3LaneAdaptedTreeMatchesProducer$", "-count=1", "-timeout", "90s", "-v"]` | 1 | 1.920472 |
| `["go", "test", "./stage3/lane", "-run", "^TestShard_Stage3LaneFile001$", "-count=1", "-timeout", "14m", "-v"]` | 0 | 365.355186 |
| `["go", "test", "./stage3/lane", "-run", "^TestProduct_Stage3LaneInstall$", "-count=1", "-timeout", "90s", "-v"]` | 0 | 32.376089 |
| `["go", "test", "./stage3/lane", "-run", "^TestProduct_Stage3LaneCompiler$", "-count=1", "-timeout", "90s", "-v"]` | 0 | 49.800664 |
| `["go", "test", "./stage3/lane", "-run", "^TestProduct_Stage3LaneHarness$", "-count=1", "-timeout", "90s", "-v"]` | 0 | 31.381698 |
| `["go", "test", "./stage3/lane", "-run", "^TestShard_Stage3LaneFile001$", "-count=1", "-timeout", "90s", "-v"]` | 0 | 20.349785 |
| `["python3", "stage3/lane/shard.py", "stage3-lane-file-001", "$STAGE3_RESULTS", "--inputs-hash", "23b8f505d961f4900852c9a178388c30e16c93f3053e6926b3c4f045ad6e1001"]` | 1 | 3.673945 |
| `["go", "test", "./stage3/lane", "-run", "^TestShard_Stage3LaneFile001$", "-count=1", "-timeout", "90s", "-v"]` | 1 | 1.370911 |
| `["python3", "-B", "stage3/lane/prove_union.py", "/workspace/scratch/hefhych-recut/measure", "/workspace/scratch/hefhych-recut/full-union/whole", "/tmp/hefhych-review-cache/local/c5f9d634f8ea40b4f488abf0325d5abec10791f25ebad12b6725712ff2f8aaff", "/workspace/scratch/hefhych-ordering/union-proof"]` | 0 | 37.679660 |

Fresh install Go test wall: 229.11s; full command 231.668250s. Fresh shard Go test wall: 362.24s; full command 365.355186s. Its actual shard execution after products existed was 18.179303s, with 273 passes. Logs require misses for every apply boundary10/20/40/65/70/99 and lane install/compiler/harness. The following apply 99 command logged the same79d532961a35 key as a hit with no misses. The successful producer-comparison test independently checked full canonical key paths.

# Prepared tests and all remaining recorded walls

| Log / command family | Measured process seconds and exit |
|---|---|
| prepared-go.log | 8.677793s; exit 0 |
| python-go.log | 30.206516s; exit 0 |
| child-python.log | 11.999469s; exit 0 |
| wave2-mutants.log | 13.331317s; exit 0 |
| guard-mutants.log | 0.912521s; exit 0 |
| product-mutants.log | 17.511737s; exit 0 |
| artifact-mutants.log | 0.148715s; exit 0 |
| owned-file-mutant.log | 56.321059s; exit 0 |
| prepared-shards.log | 33.627163s; exit 0 |

All individual Go walls from these prepared runs:

| Test | Go wall seconds | Outcome |
|---|---:|---|
| TestShardsMatchManifest | 0.07 | PASS |
| TestStage3LaneAdaptedTreeMatchesProducer | 0.76 | PASS |
| TestProduct_Stage3LaneHarness | 0.77 | PASS |
| TestProduct_Stage3LaneInstall | 0.71 | PASS |
| TestProduct_Stage3LaneCompiler | 0.79 | PASS |
| TestShardManifestRejectsMissingAndEmpty | 0.01 | PASS |
| TestUnionIdentityOutcomes | 0.09 | PASS |
| TestPythonUnitsMatchCommands | 0.15 | PASS |
| TestStage3LaneHarnessTracksInputs | 6.07 | PASS |
| TestStage3LaneCompilerTracksInputs | 6.76 | PASS |
| TestStage3LaneInstallTracksInputs | 6.78 | PASS |
| TestStage3LaneLandingLaneTests | 9.89 | PASS |
| TestStage3ApplyShard08 | 0.40 | PASS |
| TestStage3ApplyShard16 | 0.19 | PASS |
| TestStage3ApplyShard15 | 0.21 | PASS |
| TestStage3ApplyShard14 | 0.18 | PASS |
| TestStage3ApplyShard13 | 0.17 | PASS |
| TestStage3ApplyShard12 | 0.17 | PASS |
| TestStage3ApplyShard11 | 0.17 | PASS |
| TestStage3ApplyShard10 | 0.18 | PASS |
| TestStage3ApplyShard09 | 0.17 | PASS |
| TestStage3ApplyShard03 | 0.19 | PASS |
| TestStage3ApplyShard07 | 0.23 | PASS |
| TestStage3ApplyShard06 | 0.24 | PASS |
| TestStage3ApplyShard05 | 0.19 | PASS |
| TestStage3ApplyShard04 | 0.17 | PASS |
| TestStage3ApplyShard01 | 0.17 | PASS |
| TestStage3ApplyShard02 | 0.17 | PASS |
| TestStage3ApplyShard00 | 15.69 | PASS |
| TestStage3LaneLaneTableTests | 0.20 | PASS |
| TestShard_Stage3LaneByteComparisons | 12.71 | PASS |
| TestShard_Stage3LaneFile010 | 19.83 | PASS |

# Mutants and catchers

| Mutant | Catcher | Exit |
|---|---|---:|
| Consumer chooses adapted 70 instead of producer 99 | TestStage3LaneAdaptedTreeMatchesProducer: adapted product key differs from apply producer | 1 |
| Census README edit, rebuild products without regenerating manifests | Shard001 passes 273 cases with the rebuilt artifact | 0 |
| Edited tree with pre-edit artifact | Shard refuses and names both keys | 1 |
| Edit shard.py | Actual Go shard refuses: shard harness inputs changed | 1 |
| delete-generated-test | generated shard declarations differ | 1 |
| name-drift | generated shard declarations differ | 1 |
| added-manifest-unit | generated shard declarations differ | 1 |
| extra-Go-shard | manifest/test names differ | 1 |
| delete-Python-wrapper | Python Go unit names differ | 1 |
| added-old-command-case | Python units differ from old command case discovery | 1 |
| dropped-executed-Python-case | Python cases differ | 1 |
| ignore-outcomes | AssertionError | 1 |
| ignore-multiplicity | AssertionError | 1 |
| unanchored-case-grep | test_watcher_grep_selects_exactly_one_case | 1 |
| ignored-unit-budget | test_budget_can_fail_with_all_other_checks_green | 1 |
| swapped-case-identity | test_inventory_detects_swapped_case_with_unchanged_counts | 1 |
| ignored-unit-input-hash | test_input_hash_guard_detects_changed_definition | 1 |
| ignored-stock-parser-corruption | test_stock_parser_corruption_is_rejected | 1 |
| ignored-producer-phase | test_failed_producer_phase_is_rejected | 1 |
| ignored-ready-input-identity | test_changed_ready_input_identity_is_rejected | 1 |
| ignored-payload-corruption | test_payload_corruption_is_rejected_before_execution | 1 |
| ignored-exact-diff-bytes | test_merge_rejects_newline_tampering_with_all_other_checks_green | 1 |
| ignored-executed-union | test_merge_rejects_changed_union_despite_equal_counts | 1 |
| ignored-input-bytes | test_changed_proof_invalidates_input_key | 1 |
| ignored-tool-version | test_tool_version_invalidates_key | 1 |
| ignored-payload-corruption | test_corrupted_artifact_is_rejected | 1 |
| ignored-mid-build-input-change | test_changing_inputs_during_preparation_is_rejected | 1 |
| ignored-phase-exit | test_failed_preparation_never_publishes | 1 |
| `['go', 'test', './stage3/lane', '-run', '^TestStage3LaneInstallTracksInputs$', '-count=1', '-timeout', '90s', '-v']` | Respective TracksInputs/ArtifactTests/ProductPackageTests assertion; full output retained | 1 |
| `['go', 'test', './stage3/lane', '-run', '^TestStage3LaneCompilerTracksInputs$', '-count=1', '-timeout', '90s', '-v']` | Respective TracksInputs/ArtifactTests/ProductPackageTests assertion; full output retained | 1 |
| `['go', 'test', './stage3/lane', '-run', '^TestStage3LaneHarnessTracksInputs$', '-count=1', '-timeout', '90s', '-v']` | Respective TracksInputs/ArtifactTests/ProductPackageTests assertion; full output retained | 1 |
| `['python3', '-B', '-m', 'unittest', 'test_artifact.ArtifactTests.test_upstream_installed_bytes_invalidate_key']` | Respective TracksInputs/ArtifactTests/ProductPackageTests assertion; full output retained | 1 |
| `['python3', '-B', '-m', 'unittest', 'test_artifact.ArtifactTests.test_installed_parser_is_keyed_by_its_byte_pin']` | Respective TracksInputs/ArtifactTests/ProductPackageTests assertion; full output retained | 1 |
| `['python3', '-B', '-m', 'unittest', 'test_products.ProductPackageTests.test_changed_installed_bytes_are_refused_before_copy']` | Respective TracksInputs/ArtifactTests/ProductPackageTests assertion; full output retained | 1 |
| Owned test-file break in a private copy | Actual shard 001 fails with planted error; shard002 passes; product payload unchanged | 1 / 0 |
| Missing manifest | Actual Go shard fails | 1 |
| Empty manifest | Actual Go shard fails | 1 |
| Wrong report counts / nonzero command exit | Actual Go result predicates catch each independently | 1 / 1 |
| Drop retained shard 001 from union | merge_shards names 273 missing executions; restore passes | 1 / 0 |

The real runtime proof restores the census and shard source in finally. Every generator/product-key/acceptance mutant is restored before normal execution. Owned-file proof uses a private tree and verifies the product unchanged. Full commands and individual command seconds are in the packed logs.

# Union and payload provenance

The complete retained 151-shard union was rechecked against its original same-artifact whole execution. It passed 106367 identities/outcomes/multiplicities, counts and raw diffs. Drop-shard 001 failed with 273 named missing executions and no extras; restoration passed. This is a recheck of retained executions, not a fresh whole/all 151 run. Fresh shard 001 on the new cold build is separate.

An attempted byte-identical-artifact gate rejected combining fresh and retained executions. Exactly two files differ between fresh and retained product trees: .git/index and .git/logs/HEAD. Every non-.git path, mode, symlink target and byte matches, with tree hash b03264a1526a19d6b81c962be6b45e73de3729e865ed9a05a66785a8f6c50bf8. Both full input descriptions and API markers match. Full payload verification remains unchanged, including .git; no exclusion was added. The parallel apply git-read change is outside this branch and was not edited or tested here.

```json
{
  "status": "pass",
  "errors": [],
  "counts": {
    "pass": 106366,
    "fail": 1,
    "skip": 0
  },
  "whole_counts": {
    "pass": 106366,
    "fail": 1,
    "skip": 0
  },
  "units": 151,
  "tests": 106367,
  "artifact": "331995f853472db52bdecf6875eac0018b95f752a2882f13ac6d15476cd4ecbe",
  "failed_tests": [
    "unittests:: Public APIs for typescript.d.ts should be acknowledged when they change"
  ],
  "baseline_diffs": [
    "api/typescript.d.ts"
  ],
  "whole_verdict": "pass",
  "missing_tests": [],
  "extra_tests": []
}
```

# Setup, failures and coverage limits

```text
submodule cache: off
v24.19.0
go version go1.27.1 linux/amd64
setup: go ready (0.024s)
setup: node ready (0.024s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.012s
setup: markdown dependencies ready (0.080s)
setup: submodules ready (0.082s)
clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.190s)
setup: shared cache ready (1.159s)
setup: go build ready (19.225s)
setup: test binaries deferred (use --warm-tests) (19.315s)
setup: build cache warm (19.317s)
setup: build-flags commit=5f38748e9453e7ca5cb4dc549d0b0b8d2c585662 nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false gocacheprog=on load-before=0.25 0.06 0.02 1/911 139173 load-after=1.71 0.40 0.13 1/915 139794
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (19.344s)
setup: source /workspace/adamic-tools/env.sh
setup: logs /tmp/adamic-gate/setup.BpsjKh

```

The first prepared-check launcher exited127 before running a test because /usr/bin/time is absent. It was replaced with a monotonic Python timer and all checks were rerun successfully. Prior disposable task cache directories were removed to provide inodes; the retained union artifact and all proof logs remain.

No Python runtime, acceptance predicate, manifest definition or generated shard body was relaxed. Only Go product consumption changed. Manifest generation/match, all 19 Python wrappers, all 77 lane Python fixtures, product/key checks, byte comparisons, file 010, and the listed mutants were rerun. No actual Workshop trace, full compiler gate, parallel applyproducts-3 validation, new full whole/all 151 execution, PR or Loom submission is claimed. The owner will re-cut these branches onto the parallel apply change.
