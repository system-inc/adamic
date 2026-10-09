# Top-level stage 3 fixture tests

Base: `709de60d4d2e68987553b1703eebef69bad34e57`. Branch: `codex/test-split-stage3-fixtures-toplevel`.

The original `TestFixtures` is now 11 named directory tests calling the same helper. All Node byte comparisons, provenance, diagnostics, native sanitizer/leak checks, update rules, shard assignment, and shared `sync.Once` hook preparation remain intact. The guard reads active test declarations and fails for missing or duplicate units, missing manifests, empty directories, and units naming absent directories.

The measured baseline parent took **21.934 s** with an empty hook store. Its observed 173 fixture identities exactly equal the new directory union, without omissions or duplicates. Both use the original status files unchanged.

## Measurement

Every command ran in a fresh process with `GOMAXPROCS=4`, CPU affinity 0-3, `-count=1`, `-parallel=4`, `ADAMIC_GATE_UNCACHED=1`, and no `ADAMIC_TEST_SHARD`. Cgroup quota is 400000/100000 (four CPUs). Unpinned `nproc` printed 5; pinned `nproc` printed 4. Toolchain reused from setup on this same box: `/workspace/adamic-tools/env.sh`, Go 1.27.1, Node 24.19.0.

Cold means fresh test observations after keyed build-product fetch, as requested. Go compilation artifacts and OS pages are retained. The first pass also exercised automatic preparation with an empty per-unit hook store. The final pass fetched and SHA-256-verified that keyed hook; no test result was reused. Times below are complete command wall times, including Go compilation/linking, hook fetch/preparation, and test execution.

```sh
source /workspace/adamic-tools/env.sh
GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=1 ADAMIC_STAGE3_BUILD_STORE=<unit-store> \
  go test ./stage3/fixtures/... -run '^<Name>$' -count=1 -parallel=4 -timeout 90s -v > <unit>.log 2>&1
```

All 17 final units exited 0. None reached 90 s or exceeded the 60 s budget. The standalone transform hook intentionally skips without its subprocess environment; its enclosing guard executes all supported and mutant runner shapes.

| Top-level test | Fixtures | Cold after keyed fetch (s) | Empty hook store (s) | 90 s failure |
|---|---:|---:|---:|---|
| `TestFixturesAssertions` | 20 | 6.973 | 10.852 | No |
| `TestFixturesCycles` | 10 | 3.515 | 6.566 | No |
| `TestFixturesEnums` | 13 | 4.245 | 7.018 | No |
| `TestFixturesHost` | 25 | 6.607 | 8.669 | No |
| `TestFixturesNamespaces` | 12 | 3.846 | 6.165 | No |
| `TestFixturesNestedFunctions` | 11 | 3.852 | 6.370 | No |
| `TestFixturesObjects` | 25 | 3.997 | 6.706 | No |
| `TestFixturesPredicates` | 11 | 3.189 | 6.056 | No |
| `TestFixturesRecords` | 19 | 3.332 | 6.100 | No |
| `TestFixturesRunner` | 3 | 2.726 | 5.804 | No |
| `TestFixturesTaste` | 24 | 4.753 | 7.319 | No |
| `TestPrepareFixtureOracleHook` | - | 2.323 | 5.099 | No |
| `TestFixturePaths` | - | 2.129 | 2.129 | No |
| `TestFixtureShardManifest` | - | 1.901 | 2.126 | No |
| `TestTransformedNodeRunnerGuardHook` | - | 1.824 | 2.025 | No |
| `TestTransformedNodeRunnerGuard` | - | 1.824 | 2.023 | No |
| `TestFixtureDirectoriesHaveTopLevelTests` | - | 1.977 | 1.977 | No |

The first guard run correctly found the manifestless `enum-init-reach` directory. It is owned by `internal/oracle/enum_initialization_reach_test.go`, was never in the original status-manifest suite, and is explicitly excluded to preserve the original 173-fixture union. Any other new directory containing `.a` files or `status.json` without a unit fails. There is no `real` status directory at this base. The initial discovery result is preserved in `measurement-progress.log`; the final guard passes.

## Mutants and union

| Planted change | Observed catcher |
|---|---|
| Scratch `runner/03_return_true.a` recorded stdout changed to `PLANTED FAILURE\n` | Only `TestFixturesRunner` exited 1, at `runner/03_return_true.a/node`: `recorded Node byte comparison failed`. All other 10 directory tests exited 0 against the same mutant root. |
| New scratch `unregistered` fixture directory | `TestFixtureDirectoriesHaveTopLevelTests` exited 1: `fixture directory unregistered has no top-level test`. |
| Rename `TestFixturesAssertions` to a non-test helper | The directory guard exited 1: `fixture directory assertions has no top-level test`. Source restored in `finally`; guard then passed. |
| Prepared hook overwritten with corrupt bytes | `TestFixturesRunner` exited 1 at automatic preparation: `oracle build product hash mismatch`. Original bytes restored; runner then passed. |

`fixture-union.json` contains all original, baseline-observed, and new-observed identities: 173 each, equal sets, no duplicate execution. `proofs.json` records the exact mutant commands and exits. Every mutant test also used `-timeout 90s` and four CPUs. All full logs are compressed under `logs/`; stdout/stderr went directly to files.

The before measurement used Go `-overlay` with the exact `709de60d` fixture test file and `-run '^TestFixtures$' -timeout 90s`; see `baseline.json` and `logs/baseline.log.gz`. No whole-package confirmation or full gate was run.

## Scope

Changed only `_test.go` files and evidence under `stage3/fixtures/testdata`. No new language fixture, no status update, and no counts change. The gate dispatcher, apply, lane, other stage 3 packages, and the externally owned enum-init oracle tests were not changed or rerun in this follow-up. Existing README examples still name the old parent because non-test documentation is outside this task's allowed territory.
