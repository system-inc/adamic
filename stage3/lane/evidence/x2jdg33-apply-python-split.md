Split the 23 apply tests into three Go wrappers, on origin/main 23ff023b.
Exact-once coverage and the unchanged stage3 file census pass.
All three wrappers pass; all six executed mutants are caught and restored.
The single checkout case is 14.303 s alone, so the six-second target is not achievable by this permitted wrapper-only split.
No assertions, Python classes, census file, or shard files changed.

Task: #x2jdg33. Branch: typescript/stage3-apply-python-split.

Measurements precede edits. From /workspace/adamic/stage3, each command was
`timeout 120 python3 -B -m unittest -v NAME`, separately, with a fresh STAGE3_CACHE,
PYTHONDONTWRITEBYTECODE=1 and ADAMIC_TYPESCRIPT_SOURCE=/tmp/x2jdg33-typescript
(the pinned 050880ce59e30b356b686bd3144efe24f875ebc8 checkout). Times are process wall
seconds including Python startup. All exit 0. Full outputs:
/tmp/x2jdg33-alone-NAME.log; timing ledger /tmp/x2jdg33-times.json.
Setup ran concurrently with these observations, so these are not unloaded-machine guarantees.

| Python test or class (prefix test_apply.) | Alone seconds |
|---|---:|
| ApplyCheckoutTests.test_ordinary_apply_leaves_git_clean | 14.303 |
| ProductTests.test_build_with_pinned_packages_never_runs_npm | 0.130 |
| ProductTests.test_build_without_api_packages_installs_nothing | 0.124 |
| ProductTests.test_corrupt_product_never_builds | 0.136 |
| ProductTests.test_existing_output_is_refused | 0.134 |
| ProductTests.test_input_key_covers_all_inputs | 0.207 |
| ProductTests.test_manifest_mutant_is_rejected | 0.136 |
| ProductTests.test_missing_product_builds_and_reports | 0.125 |
| ProductTests.test_new_adaptation_directory_changes_hash | 0.142 |
| ProductTests.test_only_manifest_absence_is_a_cache_miss | 0.147 |
| ProductTests.test_packages_come_installed_from_the_lockfile | 0.139 |
| ProductTests.test_packages_digest_pins_every_byte | 0.136 |
| ProductTests.test_payload_manifest_mutant_is_rejected | 0.139 |
| ProductTests.test_payload_mutant_is_rejected | 0.134 |
| ProductTests.test_publish_hook_describes_local_payloads | 0.137 |
| ProductTests.test_restore_fetches_verified_product | 0.151 |
| ProductTests.test_shards_partition_every_case | 0.138 |
| ProductTests.test_symlinked_directory_digests_as_its_link | 0.134 |
| ProductTests.test_typescript_source_is_the_pinned_gate_input | 0.146 |
| ProductTests.test_unconfigured_cold_cache_builds_without_remote_read | 0.137 |
| ProductTests.test_unkeyed_parser_override_is_refused | 0.128 |
| ProductTests.test_write_table_is_explicit | 0.136 |
| ProofShardTests.test_every_indexed_read_file_has_one_shard | 0.186 |
| ApplyCheckoutTests | 15.002 |
| ProductTests | 0.327 |
| ProofShardTests | 0.166 |

| New Go wrapper | Tests | Alone seconds reported by Go |
|---|---:|---:|
| TestApplyCheckoutPython | 1 | 15.72 |
| TestApplyProductPython | 21 | 0.31 |
| TestApplyProofShardPython | 1 | 0.17 |

Groups use the exact method names in the table, partitioned by their existing class.
There is no setUpClass or setUpModule hook, hence zero repeated shared-setup cost.
Keeping each class together also avoids extra Python launches (roughly 0.12 s each).
The checkout class contains just one test; dividing classes cannot divide that test's work.
Question resolved without pausing: how to meet six seconds without changing assertions
or leaving territory? Chosen: isolate the single slow test and report the unmet bound,
without altering its behavior. This does not establish a fix for its loaded-box timeout.

The literal runPython module remains test_apply for the existing source census.
runPython expands it to the current wrapper's explicit selector list. The new coverage
check reads actual wrapper declarations/calls, discovers unittest IDs without running them,
and compares multiplicities. Thus a new method in an existing class is red too.
The table's 23 method rows are the exact union: Checkout=1, Product=21, ProofShard=1,
no duplicates. The coverage log prints every full ID. Existing census reports
14 suites, 12 wrapped, 2 listed; test_apply.py remains wrapped.

Commands from /workspace/adamic (source /workspace/adamic-tools/env.sh first):

```
export GOPROXY='https://proxy.golang.org|direct'
timeout 600 bash cloud/setup.sh > /tmp/x2jdg33-setup.log 2>&1
timeout 180 bash -c 'source /workspace/adamic-tools/env.sh; npm ci --ignore-scripts --prefix stage3/api > /tmp/x2jdg33-api.log 2>&1'
timeout 120 git clone --depth 1 --branch v6.0.3 https://github.com/microsoft/TypeScript.git /tmp/x2jdg33-typescript > /tmp/x2jdg33-source.log 2>&1
# Measurements, from stage3; runner invokes the per-name command above:
timeout 400 python3 -B /tmp/x2jdg33-measure.py > /tmp/x2jdg33-measure.log 2>&1
# Each wrapper alone and mutants, each invocation writes its own complete log:
timeout 500 python3 /tmp/x2jdg33-proofs.py > /tmp/x2jdg33-proofs.log 2>&1
# Within that runner, NAME is each wrapper or coverage test, with GOWORK=off:
ADAMIC_TYPESCRIPT_SOURCE=/tmp/x2jdg33-typescript timeout 180 go test ./stage3/lane -run '^(NAME)$' -count=1 -v -parallel=1 > /tmp/x2jdg33-LABEL.log 2>&1
# Final run uses normal workspace, after submodule setup:
ADAMIC_TYPESCRIPT_SOURCE=/tmp/x2jdg33-typescript timeout 180 go test ./stage3/lane -run '^(TestApply.*Python|TestApplyPythonGroupsCoverEveryTestOnce|TestEveryStage3PythonSuiteRunsOrSaysWhy|TestPythonSuiteGapsNamesWhatNoWrapperRuns)$' -count=1 -v -parallel=1 > /tmp/x2jdg33-final.log 2>&1
timeout 180 go vet ./stage3/lane > /tmp/x2jdg33-vet.log 2>&1
gofmt -l stage3/lane/python_test.go
git diff --check
```

Proofs (all restored independently; mutant command exits 1, restored coverage exits 0):

| Mutant | Observed failure | Log under /tmp/x2jdg33- |
|---|---|---|
| Flip assertEqual(CASES, expected) to assertNotEqual in ProductTests.test_shards_partition_every_case | Exactly TestApplyProductPython fails (0.29 s); Checkout passes (13.69 s), ProofShard passes (0.17 s) | mutant-assertion.log |
| Delete TestApplyProductPython wrapper | Coverage names TestApplyProductPython and all 21 uncovered full IDs | mutant-dropped.log |
| Add empty ProductTests.test_unassigned_mutant | Coverage names test_apply.ProductTests.test_unassigned_mutant, covered 0 times | mutant-new.log |
| Duplicate ProductTests.test_shards_partition_every_case selector | Coverage names full ID, covered 2 times | mutant-duplicate.log |
| Replace that selector with ProductTests.test_missing_mutant | Coverage names undiscovered test and omitted real test | mutant-unknown.log |
| Rename product selector-table key to MissingProductWrapper | Coverage names missing wrapper and empty TestApplyProductPython; product wrapper fails before Python | mutant-empty.log |

Alone Go logs: /tmp/x2jdg33-go-TestApply{Checkout,Product,ProofShard}Python.log.
Restored coverage and census: /tmp/x2jdg33-restored.log (23 names, 3 wrappers, 0.154 s package).
Initial combined green run: /tmp/x2jdg33-green.log (16.782 s package).
The first normal-workspace test attempt failed before compilation because setup had not yet
fetched cohere/TypeScript/tsc/go.mod. GOWORK=off allowed focused package tests to proceed;
no dependency files were edited. First formatting attempt lacked gofmt on PATH; sourcing
the printed environment fixed it. Full repository tests were not run for this wrapper-only unit.

Setup exited 0. nproc=5, cgroup CPU quota 4, Go 1.27.1, clang 20.1.8,
Node 24.19.0. Timing lines (elapsed seconds):

```
setup: node ready (0.068s)
setup: go ready (0.086s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.480s)
setup: markdown dependencies installed (npm ci, integrity verified); step-duration=0.855s
setup: markdown dependencies ready (1.034s)
setup: submodules ready (180.146s)
setup: shared cache ready (181.866s)
setup: go build ready (450.081s)
setup: test binaries deferred (use --warm-tests) (450.215s)
setup: build cache warm (450.217s)
setup: build-flags commit=23ff023b9ed986834d20404d43e2c363667edf9b nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false gocacheprog=on load-before=0.05 0.01 0.00 1/136 714 load-after=8.67 5.59 2.43 1/170 12777
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (450.342s)
setup: source /workspace/adamic-tools/env.sh
setup: logs /tmp/adamic-gate/setup.gIPb25
```

Normal-workspace final checks exited 0: all three wrappers, exact-once coverage,
and both existing census tests, package 38.997 s. During setup build load, the
checkout wrapper was 36.69 s, product 0.88 s, shard proof 0.58 s. These loaded
observations do not replace the separate alone timings above. Vet exited 0
without diagnostics; gofmt and git diff --check printed nothing. The sixth
mutant command is recorded in /tmp/x2jdg33-empty-mutant.py and ran via
`timeout 190 python3 /tmp/x2jdg33-empty-mutant.py > /tmp/x2jdg33-empty-proof.log 2>&1`.
After it was restored, the normal-workspace coverage/census rerun exited 0,
package 0.862 s, log /tmp/x2jdg33-restored-final.log (`timeout 30 go test
./stage3/lane -run '^(TestApplyPythonGroupsCoverEveryTestOnce|TestEveryStage3PythonSuiteRunsOrSaysWhy)$'
-count=1 -v`). No automatic approval review refusal occurred.
