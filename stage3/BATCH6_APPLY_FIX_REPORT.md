# Batch 6 apply-products investigation

Candidate 8eaa9ed3 was checked out into a scratch branch and merged with origin/main b8bcadb2c493173855f19d7e5c508b34f5eeb5b6 (scratch merge 693f35a1). The fix branch codex/stage3-batch6-fix starts at 8eaa9ed3 and separately merges that main. Every fix-branch commit ends with Task: #aq4nkjm.

## Observed reproductions

All commands carried timeout limits before launch and wrote complete output to files. The lane invokes `bash stage3/apply.sh <results>/adapted-tree`; the fast gate invokes the entire `python3 stage3/test_apply.py`, not only the last shard printed in its tail.

- Merged candidate: `timeout 90 env ADAMIC_TEST_SHARD=17/18 python3 stage3/test_apply.py` passed its one ProofShardTests case, 0.043 s; log /tmp/batch6-candidate-shard.log. All 17 focused ProductTests and ProofShardTests also passed.
- Main b8bcadb2: the exact 17/18 invocation exited 0 but selected zero tests. Its equivalent named proof and focused cases passed: 16 tests in 0.268 s, /tmp/batch6-main-focused.log.
- Main cold lane apply: `timeout 300 env STAGE3_CACHE=/tmp/batch6-main-cache STAGE3_PRODUCT_STORE=/tmp/batch6-main-store ADAMIC_BUILD_CACHE_DIR=/tmp/batch6-main-products bash stage3/apply.sh /tmp/batch6-main-lane/adapted-tree`, exit 0, /tmp/batch6-main-apply.log.
- Candidate clean lane apply: same invocation with candidate paths and timeout 240, exit 0, /tmp/batch6-candidate-apply-retry.log. Products 10/40 from the initial attempt were already prepared; products 70/99 were rebuilt. Adapt44 and nonnull both applied successfully.
- Candidate full gate test command with GOFLAGS=-buildvcs=false -trimpath, ADAMIC_GATE_UNCACHED=1, ADAMIC_TEST_WASI=1, ADAMIC_ORACLE_WASI=1 and ADAMIC_GATE_COHERE=1, warm prepared products and timeout 90: 18 tests passed in 18.914 s, /tmp/batch6-candidate-full-tests-retry.log.

The first local candidate run exhausted this 8.8 GB workspace during concurrent cold builds, raising shutil.Error with [Errno 28] No space left on device. It is discarded for gate attribution. The early full Python attempt reached its outer timeout 90 while waiting for the cold chain; it is also discarded. Completed prior-task trees and named scratch artifacts were removed before the isolated retry.

## Confirmed apply-products defect and fix

Python product_key hashes every permission bit, while Go buildcache.Key hashes only executable status for files and no permissions for directories or symlinks. A source checkout's read/write permissions can therefore change the Python manifest name without invalidating the Go product. Reusing that product fails in fetch_product -> restore_product, with this exact observed error:

`ProductMissing: product 36331e9d30187c13dcb2fad156ad1e37ca3cddb07233b2aa5a0fecbdc9a4edc0 missing from /tmp/batch6-candidate-products/6afb7f4611b9dcb1b934c4c8292327e4f93c2cc1ea02ff6c9db297b7bd5b7cc9/products`

Its preceding FileNotFoundError names that directory's `products/36331e9d30187c13dcb2fad156ad1e37ca3cddb07233b2aa5a0fecbdc9a4edc0.manifest`. Log: /tmp/batch6-mode-key-error.log. The actual cached manifest was keyed 7a36781b22a645bd2d1937d5db1ed4881ab08b488d397625284c209b3bff3e58. Changing only source.json from mode 0600 to 0644 reproduced this disagreement; Go reused the same final product.

The production fix in stage3/apply.py retains executable status for regular files and ignores checkout read/write permissions, matching buildcache.Key. Compiler source/adaptations are unchanged. The regression changes real file and directory permissions in a private keyed input tree and also proves executable status remains an input.

Attribution: this confirmed cache/manifest mismatch belongs to apply-products (946ba8d6), and does not require adapt44 or nonnull. Neither reported gate failure reproduces under normal local checkout conditions. The published gate record contains only tails, so its actual underlying apply error and any failing test assertion remain unknown; the permission mismatch is a demonstrated possible explanation, not a claimed observation of the gate error. No evidence singles out nonnull, adapt44, or their merge as the cause.

## Mutants and validation

Original permission-sensitive key: the new regression fails with `checkout permissions changed the manifest key`, /tmp/batch6-permissions-mutant.log. A second in-memory mutant removes executable status from the real key function: the regression fails with `executable input mode was omitted`, /tmp/batch6-executable-mutant.log. No source mutation overlaps product preparation. Fixed focused cases: 18 tests passed in 1.732 s, /tmp/batch6-fix-focused.log.

Final validation: a fresh four-product chain and lane apply passed with timeout 240, /tmp/batch6-fix-apply.log. Exact ADAMIC_TEST_SHARD=17/18 passed one selected case in 0.009 s, /tmp/batch6-fix-shard.log; adding the regression changes shard membership, so the complete run also verifies the named ProofShardTests case. All 19 Python cases passed in 14.509 s under the gate flags and timeout 90, /tmp/batch6-fix-full-tests.log. Changing the real source.json permissions again preserved the manifest key and found the final cold product's manifest, /tmp/batch6-fixed-permission-reuse.log.

Root integration checks after committing use `timeout 90 bash -c 'git fetch -q origin main devtools/fast-gate cloud/merge-tree && git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -'` with GOWORK=off and the toolchain environment sourced, /tmp/batch6-fix-lane-checks.log. No full TypeScript oracle lane is repeated: this task requests apply diagnosis. No new .a fixtures, counts rows, or PR. Without the gate's complete apply.log and test failure body, definitive attribution of that specific run remains uncovered.

## Toolchain

`export GOPROXY='https://proxy.golang.org|direct'; timeout 360 bash cloud/setup.sh` failed with Go linker `mapping output file failed: no space left on device` for stage3/ledger/checker-259, its rerun, stage3/census/tool and stage3/drivers/parser/probe. Log /tmp/batch6-setup.log. Observed timings: Node ready 0.028 s; Go ready 0.031 s; markdown install skipped 0.013 s, ready 0.096 s; submodules ready 0.107 s; clang ready 0.245 s. Workaround: free named prior-task artifacts and source the existing validated /workspace/adamic-tools/env.sh. Go 1.27.1, Node 24.19.0, clang 20.1.8; nproc=5, cgroup 4 CPUs.
