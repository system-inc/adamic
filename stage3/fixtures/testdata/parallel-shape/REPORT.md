# Fixture test parallel shape

Base: `af25357ea1c4ff7127a631c8917e88ac778868be` on `codex/stage3-trio-real-unit`.

All twelve directory tests now call their own `t.Parallel()` first, then `testFixtureDirectory(t, "<directory>")`. The helper no longer calls `t.Parallel()`. The guard accepts exactly those two statements, in that order, with the actual test parameter and a literal directory. Shared `sync.Once` hook preparation, fixture checks, manifests, and the serial baseline are unchanged.

## Validation

Measurements ran in separate processes pinned to CPUs 0-3, with `GOMAXPROCS=4`, `ADAMIC_GATE_UNCACHED=1`, and no shard selector. The first unit self-prepared the hook in an empty store; later units fetched the same SHA-256-verified keyed product. Go compiler artifacts and OS pages were retained. Times include the complete Go command, setup/fetch, and all subtests. No successful test observations were reused.

Each named command was:

```sh
go test ./stage3/fixtures/ -run '^<Name>$' -count=1 -timeout 90s -v > <Name>.log 2>&1
```

| Top-level test | Cold complete command (s) | Result |
|---|---:|---|
| `TestFixturesAssertions` | 17.359 | PASS |
| `TestFixturesCycles` | 3.286 | PASS |
| `TestFixturesEnums` | 3.946 | PASS |
| `TestFixturesHost` | 16.391 | PASS |
| `TestFixturesNamespaces` | 4.008 | PASS |
| `TestFixturesNestedFunctions` | 11.369 | PASS |
| `TestFixturesObjects` | 10.543 | PASS |
| `TestFixturesPredicates` | 3.438 | PASS |
| `TestFixturesReal` | 3.636 | PASS |
| `TestFixturesRecords` | 3.127 | PASS |
| `TestFixturesRunner` | 2.930 | PASS |
| `TestFixturesTaste` | 11.877 | PASS |
| `TestPrepareFixtureOracleHook` | 2.730 | PASS |
| `TestFixturePaths` | 2.122 | PASS |
| `TestFixtureShardManifest` | 2.121 | PASS |
| `TestTransformedNodeRunnerGuardHook` | 2.172 | SKIP |
| `TestTransformedNodeRunnerGuard` | 2.271 | PASS |
| `TestFixtureDirectoriesHaveTopLevelTests` | 2.087 | PASS |

`TestTransformedNodeRunnerGuardHook` skips when run alone; the enclosing guard exercises its accepted and rejected runner shapes.

`go test ./stage3/fixtures/ -count=1 -timeout 90s` passed: complete wall time **20.019 s**. No unit reached 90 seconds or exceeded 30 or 60 seconds.

## Guard mutants

All mutants changed compilable test source and ran only `TestFixtureDirectoriesHaveTopLevelTests`, with `-count=1 -timeout 90s`. Each exited 1 at the intended guard assertion. Every source change was restored in `finally`; the restored guard passed.

| Mutant | Catcher |
|---|---|
| `missing-unit` | `fixture directory assertions has no top-level test` |
| `duplicate-unit` | `has duplicate units` |
| `non-literal-directory` | `must name a literal fixture directory` |
| `extra-statement` | `must call its test parameter's Parallel() then testFixtureDirectory directly` |
| `missing-parallel` | `must call its test parameter's Parallel() then testFixtureDirectory directly` |
| `wrong-parallel-receiver` | `must call its test parameter's Parallel() first` |
| `reversed-statements` | `must call its test parameter's Parallel() first` |
| `wrong-helper-parameter` | `must pass its own test parameter` |

Exact commands, elapsed times, and exits are in `times.json`, `package.json`, and `mutants.json`. Full outputs are compressed under `logs/`. Lane checks are run on the committed branch before push, using the worker in-place invocation documented by `origin/cloud/merge-tree:cloud/integration/lane-checks.py`; the final response reports that result.

## Setup and scope

Exported `GOPROXY=https://proxy.golang.org|direct` before `timeout 240s bash cloud/setup.sh`. Setup exited 0: submodules 0.067 s; Node 0.092 s; Go 0.107 s; markdown dependencies 0.229 s; clang 0.493 s; Go build 70.652 s; test binaries deferred 70.800 s; cache warm 70.801 s; done 70.829 s. These are the script's elapsed timing lines. Toolchain source: `/workspace/adamic-tools/env.sh`. `nproc` was 5; cgroup quota 400000/100000 is four CPUs, and test-process affinity was four CPUs.

No new language fixtures or status updates, and no counts change. Only the two requested Go test files and fixture testdata evidence changed. The full repository gate and unrelated packages were not run, apart from the integration lane's analyzer and vet checks.
