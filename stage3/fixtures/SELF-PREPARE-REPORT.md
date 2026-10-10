Built automatic keyed fixture-hook preparation, a named preparation unit, and shared sync.Once.
Base: origin/main 031a1259bc7973934792dc6cb1bd4074fc2204b9; code commits ff4005c8, 2fa42859, 41b68b7a.
Exact clean-checkout gate command exits 0 in 18.956 s; deleted product with manifest retained exits 0 in 18.467 s.
Corrupt product and compiler-failure mutants fail loudly; all 173 fixtures pass and the planted mismatch fails only shard 7/8.
All final measured units stay under 30 s; cold oracle compilation is included, but production build products remain hash-cached.

## Fix and clean checkout

TestFixtures calls build-hook.py --prepare itself when run alone. Within the
complete package, nonparallel TestPrepareFixtureOracleHook prepares before the
parallel fixture workers. Both use one sync.Once result in that process. Across
processes, the existing input-hash lock prepares once and every fetch verifies
the binary's SHA-256. Valid warm products are not rebuilt. Missing binary bytes
are rebuilt even if their valid action manifest remains. Invalid manifests or
corrupt bytes fail loudly. No test observation is cached and no check is loosened.
Preparation failures now include the complete compiler output from the build log.

The detached worktree /workspace/adamic-fixture-self-prepare was at 41b68b7a,
with pinned nested submodules and a clean git status before package measurement.
The toolchain and locked stage3/api dependencies were installed, and only the
fixtures test executable was built before observations. No oracle hook
--prepare command ran beforehand; the selected product store had no manifests
or binaries. The exact command, from repository root, was:

    go test ./stage3/fixtures/...

GOMAXPROCS=4, ADAMIC_GATE_UNCACHED=1 and an isolated ADAMIC_STAGE3_BUILD_STORE
were set. Go result caches were cleared before each invocation without clearing
build products. All test output went to files. The second cold invocation removed
the actual SHA-named binary while retaining its action manifest. The warm control
confirmed the product modification time did not change. The corruption mutant
replaced binary bytes, required a nonzero exit and the specific hash-mismatch
message, and confirmed those bytes had not been used or overwritten.

nproc=5; cgroup cpu.max=400000 100000, a four CPU quota. Go 1.27.1,
Node 24.19.0, clang 20.1.8. GOPROXY was exported before setup. Workspace setup
completed in 76.608 s (Go .031, Node .028, Markdown .080, submodules .083,
clang .230, Go build 76.437, build cache 76.578). Clean-checkout setup completed
in 172.420 s (Go .021, Node .024, submodules .086, clang .194, Markdown .817,
Go build 172.387, build cache 172.390); its full timing log is retained.
The first clean setup attempt was made while nested TypeScript initialization
was still running and failed with "Unable to find current revision in submodule
path 'cohere/TypeScript'". After initialization, setup was rerun successfully.

## Exact package invocations and named preparation

| Measurement | Before (s) | Final (s) | Exit |
| --- | ---: | ---: | ---: |
| Empty hook store, exact package command | 38.160 | 18.956 | 0 |
| Product deleted, valid manifest retained | 18.380 | 18.467 | 0 |
| Warm product, no rebuilding | 15.736 | 15.915 | 0 |
| Corrupt prepared product | 2.419 | 2.427 | 1 |
| Named TestPrepareFixtureOracleHook, missing product | not present | 3.028 | 0 |
| preparation-failure | not present | 0.432 | 1 |
| recovery-after-build-failure | not present | 18.465 | 0 |

The original first package run passed but took 38.160 s with additional first-use
build products absent. That is an observed over-budget aggregate, not an omitted
failure or a claim that the new named preparation alone eliminates all startup
cost. Preparation is now an independently dispatchable unit, and the fixture
unit remains shardable. The final rows use available production build products,
as the standing budget rule requires; no successful fixture observation is reused.
The named preparation includes test-only compilation when those objects are absent.

## Every final fixture unit, with colder test compiler objects

An isolated hard-linked snapshot of Go's build cache retained production compiler
products. Before each unit, the oracle test package export objects and hook binary
were removed from that snapshot/store. GOFLAGS=-x provided compiler evidence:
every hook preparation log contains the actual compilation of internal/oracle's
test package, rather than a reused test-only object. The compiled fixture test
binary was dispatched fresh from the package directory. Production native/runtime
build products remained in their existing keyed caches. This is a cold test
observation after build products, not a flushed filesystem or empty toolchain.

| Unit | Missing hook only (s) | Missing hook and oracle test compiler objects (s) | Exit |
| --- | ---: | ---: | ---: |
| `TestPrepareFixtureOracleHook` | 3.028 | 6.930 | 0 |
| `TestFixtures ADAMIC_TEST_SHARD=0/8` | 4.897 | 9.008 | 0 |
| `TestFixtures ADAMIC_TEST_SHARD=1/8` | 5.156 | 8.748 | 0 |
| `TestFixtures ADAMIC_TEST_SHARD=2/8` | 4.696 | 7.647 | 0 |
| `TestFixtures ADAMIC_TEST_SHARD=3/8` | 4.569 | 7.981 | 0 |
| `TestFixtures ADAMIC_TEST_SHARD=4/8` | 5.334 | 8.893 | 0 |
| `TestFixtures ADAMIC_TEST_SHARD=5/8` | 4.689 | 8.064 | 0 |
| `TestFixtures ADAMIC_TEST_SHARD=6/8` | 5.604 | 9.209 | 0 |
| `TestFixtures ADAMIC_TEST_SHARD=7/8` | 5.556 | 8.993 | 0 |
| `TestFixtures (all fixtures)` | not separately measured | 19.361 | 0 |
| `TestFixturePaths` | not separately measured | 0.008 | 0 |
| `TestFixtureShardManifest` | not separately measured | 0.010 | 0 |
| `TestTransformedNodeRunnerGuard` | not separately measured | 0.028 | 0 |
| `TestTransformedNodeRunnerGuardHook` | not separately measured | 0.006 | 0 |

The dormant runner-guard hook skips normally; its guard test exercises it.
The full fixture parent, all eight shards, manifest, path tests, runner guard and
new preparation entry point are included above. Each final unit is below 30 s.
The preparation test is nonparallel so the normal package command performs it
before fixture workers. Independently dispatched TestFixtures shards retain the
automatic preparation fallback and also meet the budget with cold test-only objects.

## Mutants and retained proof

- Corrupt prepared binary: exact package exits 1 at oracle build product hash mismatch.
- Compiler mutation returns 42 and prints planted oracle hook build failure:
  TestPrepareFixtureOracleHook exits 1 with both the exit and complete compiler
  message. A subsequent exact package invocation rebuilds and passes in 18.465 s.
- Recorded Node stdout mismatch at runner/03_return_true.a: union audit catches
  it only in shard 7/8, with all other shards passing. The complete disjoint
  union remains 173 fixtures; source Node, stage0, native, ASan/UBSan and leak
  checks are unchanged.
- The existing shard-selector guards are exercised by the union audit.

The first supplemental shard runner used the repository working directory for
a compiled test binary, so it failed before reaching the helper. The final runner
uses stage3/fixtures and supersedes that invalid measurement. The first compiler
wrapper attempt did not reach the compiler action; the final
mutant preserves the wrapper PATH by running the named preparation binary directly so it reaches the build action.
Neither harness attempt is counted as a killed semantic mutant or a budget result.

Complete commands, timing JSON, compiler traces, proof/case manifests, compressed
logs and executed proof scripts are in shard-evidence/self-prepare. The isolated
Go cache and executable bytes are not committed. go vet ./stage3/fixtures and
git diff --check passed. No new fixture or .a file was added, so counts.md is
unchanged. Other packages and the full gate were not rerun for confirmation.

The branch was rebased onto 031a1259. A history-only merge preserves the old
published tip as an ancestor so the requested foreground push is a fast-forward
without force. Neither main nor the gate command list was edited or pushed.
