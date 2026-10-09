Partial review of devtools/gate-reverse-deps, rebased onto devtools/fast-gate commit 96b1d93317a692927e022437746a046e0885a82d. Nothing was deployed or promoted. Valid x1 walls below are pre-rebase; the repaired rebased x1 pair was interrupted by environment replacement and has no verdict. Standalone stage 3 compiler probe dispatch and fixture-level oracle dependency selection remain incomplete.

The gate reverses go list -deps -test -json ./..., including regular, internal-test and external-test imports, and takes the transitive closure from changed-file owners. ForTest normalizes external test variants; synthetic test binaries are excluded. Existing embed and ancestor ownership remain. Compiler changes select the whole oracle and clear deferrals.

The checked-in map is cloud/fast-gate/compiler-dependencies.json, JSON version 1. packages maps repository-relative Go package directories to lists of compiler package directories; declarations cover all tests and helpers in that package. drivers maps stage3/drivers directories to dependency lists, recorded-probe references and an explicit coverage status. The candidate map takes precedence; the tools map is the fallback for older candidates. The map contains 67 conservative package declarations and four driver groups (parser, scanner, tsc, tsc-entry). Compiler dependencies cover all seven stages. A declaration can overselect a process launcher that does not actually run the compiler.

The census scans all tracked Go package sources and rejects undeclared compiler markers or process launch capabilities, including os/exec with a computed executable name. It rejects compiler tests outside stage 1 and deleted existing declarations. It also rejects an undeclared process-launching source under stage3/drivers. This is a lexical census; arbitrary indirect invocations from other script locations remain outside its scope.

Stage 3's agreed coverage is the existing compiler probes with recorded expectations. The map references parser's front10/build/report.json, scanner's clean-records-rebase/witnesses/results.json, and tsc-entry's main-efe9f404/stops.json with compiler-tips/area-b68b2fe1 streams; tsc shares the tsc-entry probes. The gate reads dependencies and reports selected driver groups, but does not yet execute these probe manifests. Full native proofs remain in their own lane. Existing stage 3 landing-lane execution is separate.

Budget measurements

Instrument: Python 3 time.perf_counter around subprocess.run for the full fast-gate process. Go -json Elapsed, decoded by 96b1d933 testUnits, measures completed units; the inclusive comparison below adds each leaf subtest's recorded ancestor setup time. Toolchain: Go 1.27.2, native clang 19 (ASAN witness verified), WASI SDK 27, GOMAXPROCS=2, two test slots, four quota CPUs with five affinity CPUs. Build caches and pinned npm were warm, Go test results uncached (-count=1), and each gate started with the same timing-history seed. Valid runs were sequential, single observations; x1 is from the pre-rebase runner and the ordinary control is from 96b1d933. GOFLAGS=-buildvcs=false accommodates the scratch dependency copy, with submodule git links repaired before the retained after/control runs.

The limits are under 300 seconds for a small-change gate and under 30 seconds for every unit including setup. x1 is c7835cb3f1a0a9abb4f932d0353887dab1fbb907 against main 54cbc125422d4e1d64c1ffe782445b2cbc2bc5b8. The ordinary control is a synthetic comment-only cmd/lint-registry/main.go change, 4303cecd61732acecb091bd07ca1fd8bd669622c against x1. The ordinary before/after runners include the 96b1d933 budget code. The valid x1 before/after measurements predate that rebase; their Go events are decoded with its unit instrument for the 30-second comparison. These are not final rebased x1 measurements.

| Change | Before wall | After wall | Before selection | After selection | Verdict |
| --- | ---: | ---: | --- | --- | --- |
| x1 | 487.227s | 275.226s | 31 packages, whole oracle | 75 packages, whole oracle; four driver groups inventoried | both red |
| Ordinary tool comment | 15.125s | 15.377s | one package + 46 smoke fixtures | one package + 46 smoke fixtures | both green |

| Run | Wall above 300s | Max completed unit including setup | Completed units over 30s |
| --- | ---: | ---: | ---: |
| x1-before-paired | 187.227s | 257.38s | 1659 |
| x1-after-final-valid | 0.000s | 144.56s | 1655 |
| control-before-rebased | 0.000s | 0.72s | 0 |
| control-after-rebased | 0.000s | 0.83s | 0 |

The x1 before gate first failed at tests: github.com/system-inc/adamic/internal/oracle TestElementAccessCompilerAreaBoundaries/paths. The after gate first failed at stage3: stage 3 tier 1 failed: stage3-lane exit 1 (logs: stage3-lane.log). These are time-to-red measurements. Interrupted units are unmeasured, so neither a short red wall nor a unit ledger certifies complete green coverage. The budget report includes strict over-30 units; new-unit, drift and burn-down classifications are available for the rebased ordinary control. A rebased x1 classification remains unmeasured.

I retained the whole oracle. The old x1 selector already selects it because internal/oracle/counts_test.go changes beyond fixture entries. x1 also changes shared lower/flow/IR code and both native and JavaScript emitters; no smaller unaffected fixture set has been demonstrated. Fixture-level dependencies are not yet in the map, so narrowing now would discard potentially affected coverage. This is an unresolved part of the requested five-minute budget work; no tests were dropped to fit a ceiling.

The x1 selector itself chooses 10 packages before and 71 after; final gate selections include coverage-rule additions. The original selection contains none of the eight closed-gap subtests' packages; the reverse selection contains all eight. The actual provisioned, warm-build, uncached 13-package gap-family replay took 16.059215s, exit 1, and reconfirmed all eight named failures. Its maximum recorded unit was 0.89s. Raw events are in gate-reverse-deps-x1-family.jsonl.gz.

Validation: 60 unittest tests pass in 4.912s, including the rebased budget checks and 19 executed reverse-dependency/census source mutants caught by assertions. New mutants remove computed process detection, restrict the census to stage 1, accept a deleted declaration, disable script census or driver selection, and accept invalid driver dependencies. Tests use fake processes and temporary git trees; no Python dependencies were added.

Exact commands, package lists, budget classifications and unit observations are in gate-reverse-deps-budget.json; raw fast.json, Go events, runner logs and the unittest log are in gate-reverse-deps-budget-logs.tar.gz. The valid pre-rebase x1 observations are explicitly labeled; disk failures and the interrupted rebased run are excluded.
