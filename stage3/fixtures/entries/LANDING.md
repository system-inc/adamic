Built: registered the entries bucket and compared complete checked stops on native and Node.
Commits: based on e61c738ee3f5cb140b21945fce64f355631f78d0, over main 80640d7a8b8e456040758f4aa33791e45eceb4ce; landing commit is on typescript/stage3-fixtures-entries.
Commands and outputs: 30 fixture-package test commands exit 0, oracle acceptance 12/12 pass, cold entries wall 25.120s.
Mutants: three independent runner runs exit 1 at the intended assertions; six wrong-header mutants rejected.
Not covered: the unavailable cloud/merge-tree lane checker, the full repository gate, macOS, or a new stock TypeScript audit.

# Entries landing evidence

Linux, October 17, 2026. All six fixture bodies are byte-for-byte unchanged below their first-line headers. Only 01, 04, 05 and 06 needed header changes. Every byte of the original source Node stdout/stderr/exit observations is retained. The older files under evidence/ and the original README measurements are historical October 8 evidence.

The imported observation and refreshed observation for each fixture:

| Fixture | Before (October 8 record) | After on the landing compiler | Runtime held to Node |
| --- | --- | --- | --- |
| 01_scanner_keywords.a | Refused | Compiles | same raw-source stdout, stderr and exit 0 |
| 02_hidden_same_type.a | NotYet | Compiles | same raw-source stdout, stderr and exit 0 |
| 03_hidden_other_type.a | NotYet | CheckedStop | 15 stdout bytes, pinned stderr, exit 70; same generated JavaScript on Node |
| 04_alias_add_key.a | Refused | Compiles | same raw-source stdout, stderr and exit 0 |
| 05_getter.a | NotYet | Refused | compile refusal; raw source Node still exit 0 |
| 06_symbol_key.a | NotYet | Refused | compile refusal; raw source Node still exit 0 |

Four baselines compile: three finish and one stops at its inserted check. The getter and symbol fixtures refuse by name. Their complete current diagnostics are:

```text
stage3/fixtures/entries/05_getter.a:7:5: Adamic 0.1 refuses Object enumeration with a getter or accessor; use own data properties
stage3/fixtures/entries/06_symbol_key.a:4:30: Adamic 0.1 refuses Object enumeration with symbol keys; use own string-keyed data properties
```

The two original Refused diagnostics, now closed, were:

```text
stage3/fixtures/entries/01_scanner_keywords.a:95:5: Adamic 0.1 refuses an index signature; use a Map, which keeps keys in the order they were added
stage3/fixtures/entries/04_alias_add_key.a:4:34: Adamic 0.1 refuses an index signature; use a Map, which keeps keys in the order they were added
```

For 03 the raw source does not throw: Node exposes the hidden string and finishes. The intentional checked contract therefore uses generated JavaScript running on Node as the stop oracle, while still checking the original raw-source golden independently. Both backends must produce exactly:

```text
stdout: "before entries\n"
stderr: "adamic: panic: Object enumeration key 'hidden': actual string, declared number\n"
exit: 70
```

The shared stage0 shape adds optional stdout for CheckedStop. Omission still means exactly empty stdout. Native and checked Node must each equal the complete pin and each other, including exit 70. This is generic runner behavior. TestFixturesEnums passes its existing empty-stdout CheckedStop with no metadata change. A failed comparison cannot be accepted by -update. The bucket checker now records this shape too and refuses to overwrite an existing raw Node golden, even in --record mode.

## Test commands and walls

Every named fixture-package test ran separately via /tmp/entries-tests.py. Each row used the following command, with its own NAME and log; wall includes Go startup, product preparation and execution:

```sh
go test ./stage3/fixtures -run '^NAME$' -count=1 -timeout 90s -v > /tmp/entries-tests/NAME.log 2>&1
```

| Test | Wall seconds | Result |
| --- | ---: | --- |
| TestFixturesAssertions | 10.863 | exit 0 |
| TestFixturesCycles | 11.337 | exit 0 |
| TestFixturesEntries | 9.386 | exit 0 |
| TestFixturesEnums | 13.611 | exit 0 |
| TestFixturesHost | 24.364 | exit 0 |
| TestFixturesNamespaces | 11.728 | exit 0 |
| TestFixturesNestedFunctions | 20.414 | exit 0 |
| TestFixturesObjects | 18.346 | exit 0 |
| TestFixturesPredicates | 11.016 | exit 0 |
| TestFixturesReal | 11.742 | exit 0 |
| TestFixturesRecords | 7.719 | exit 0 |
| TestFixturesRunner | 9.084 | exit 0 |
| TestFixturesTaste | 25.912 | exit 0 |
| TestFixturePaths | 3.356 | exit 0 |
| TestFixtureShardManifest | 3.283 | exit 0 |
| TestExecuteGivesChildrenTheToolchainsCompiler | 3.178 | exit 0 |
| TestProduct_FixtureOracleHook | 6.778 | exit 0 |
| TestProduct_FixtureRuntimesSanitized0 | 12.645 | exit 0 |
| TestProduct_FixtureRuntimesSanitized1 | 13.031 | exit 0 |
| TestProduct_FixtureRuntimesSanitized2 | 10.204 | exit 0 |
| TestFixtureOracleProductTracksInputs | 6.698 | exit 0 |
| TestFixtureOracleProductIgnoresGoCacheLocation | 9.295 | exit 0 |
| TestTheFixtureOracleHookWritesNoBytecode | 3.241 | exit 0 |
| TestTheFixtureOracleHooksPythonReadsNoSiteDirectory | 3.229 | exit 0 |
| TestTransformedNodeRunnerGuardHook | 3.535 | exit 0; dormant child hook skips alone |
| TestTransformedNodeRunnerGuard | 3.396 | exit 0 |
| TestNodeRunnerModeGuardHook | 3.305 | exit 0 |
| TestNodeRunnerModeGuardRejectsBothCalls | 3.131 | exit 0 |
| TestNodeRunnerModeGuardRejectsNeitherCall | 3.244 | exit 0 |
| TestFixtureDirectoriesHaveTopLevelTests | 3.132 | exit 0 |

This runs all 13 manifest directory gates, the 189-fixture union, and every test in the changed Go package. The dormant child hook is exercised by TestTransformedNodeRunnerGuard. No real fixture skipped on this Linux run. The slowest directory command was TestFixturesTaste at 25.912s.

The separate cold run used a new empty product directory and disabled the shared product store. Required cloud/setup.sh had already warmed the Go compiler cache. Oracle hook preparation was a miss (8.45s), and native runtime/sanitizer builds were included. This is a cold fixture/product measurement, not a claim that installing and compiling the entire toolchain takes under 60 seconds:

```sh
ADAMIC_BUILD_STORE=off ADAMIC_BUILD_CACHE_DIR=/tmp/entries-cold-products ADAMIC_GATE_UNCACHED=1 timeout 120 go test ./stage3/fixtures -run '^TestFixturesEntries$' -count=1 -timeout 90s -v > /tmp/entries-cold.log 2>&1
# exit 0; package 21.626s; complete command wall 25.120s
```

Other commands, each timed with Bash TIMEFORMAT and redirected to its named log:

| Command | Exit and output | Wall seconds | Log |
| --- | --- | ---: | --- |
| timeout 120 go test ./internal/oracle -run '^TestEntriesAcceptance$' -count=1 -timeout 90s -v | 0; 12/12 baseline/source-mutant cases pass; package 15.905s | 19.763 | /tmp/entries-oracle.log |
| timeout 300 python3 stage3/fixtures/entries/check.py --record --scratch /tmp/entries-record | 0; six baselines, six source mutants, all 84 scanner rows | 89.397 | /tmp/entries-record.log |
| timeout 300 python3 -B stage3/fixtures/entries/check.py --scratch /tmp/entries-verify | 0; same observations verified without recording | 94.427 | /tmp/entries-verify.log |
| timeout 120 python3 -B stage3/fixtures/entries/check-headers.py /tmp/entries-fast-gate.py --scratch /tmp/entries-headers | 0; six headers accepted, six wrong headers rejected | 4.165 | /tmp/entries-headers.log |
| timeout 120 go vet ./stage3/fixtures | 0; no diagnostics | 21.886 | /tmp/entries-vet.log |

The fetched, unchanged header predicate is from devtools/fast-gate 348ac0ff8936d366936238165e53967af6a4b93e. Its captured observations and individual wrong-header failures are in /tmp/entries-headers/a-check.json and header-mutants.json. One wrong refusal reason was tested for each of the six fixture names above, preserving every body byte.

The existing six source-mutant acceptance controls also all passed: remove await from the scanner (83 entries), change 02 hidden to a string (checked exit 70), restore 03 hidden to a number (exit 0), write a string to 04 late (checked exit 70), replace 05 getter with data (getter count 0), and replace 06 symbol with a string key (hidden entry included). TestEntriesAcceptance additionally checks the proven/checked lowering classification. The three successful baseline programs run through the shared ASan/UBSan and leak hook; panic exits are not claimed leak-clean. No internal/oracle file was changed.

## Independent runner mutants

```sh
timeout 420 python3 -B stage3/fixtures/entries/check-runner-mutants.py --scratch /tmp/entries-runner-mutants > /tmp/entries-runner-mutants.log 2>&1
# verifier exit 0, wall 55.988s; each nested go test has -count=1 -timeout 90s
```

| Mutant, each its own run | Test exit | Wall seconds | Catcher |
| --- | ---: | ---: | --- |
| different-line | 1 | 24.057 | native stdout "different line\n", exit still 70; native versus checked Node and pinned native byte comparisons fail |
| missing-line | 1 | 23.282 | native stdout empty, exit still 70; the same two stdout comparisons fail |
| never-stops | 1 | 8.596 | safe hidden number, source Node golden refreshed only in scratch; native exits 0 and fails the existing pinned native stop and stage0 checks |

The first two use scratch Go overlays of emit_statements.go to change actual emitted C output. They leave source Node and the JavaScript backend unchanged. No repository compiler file is edited, no result is postprocessed, and no mutant is counted on a build error. The verifier requires the intended assertion text. Each complete log and overlay remains under /tmp/entries-runner-mutants/<name>/. The non-stopping fixture retains the CheckedStop promise but changes the hidden value to a valid number, so the raw Node check passes and only the stop promise fails.

## Setup, initial observations and unavailable lane check

Ran export GOPROXY='https://proxy.golang.org|direct', timeout 600 bash cloud/setup.sh > /tmp/entries-setup.log 2>&1, then sourced /workspace/adamic-tools/env.sh in every build/test shell. Setup exited 0. nproc is 5; cgroup cpu.max is 400000 100000, a four-CPU quota. Go 1.27.1, Node 24.19.0, clang 20.1.8. The timing lines were:

```text
setup: go ready (0.082s)
setup: node ready (0.089s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.485s)
setup: markdown dependencies installed (npm ci, integrity verified); step-duration=1.119s
setup: markdown dependencies ready (1.311s)
setup: submodules ready (187.135s)
setup: shared cache ready (194.410s)
setup: go build ready (520.245s)
setup: test binaries deferred (use --warm-tests) (520.359s)
setup: build cache warm (520.362s)
setup: build-flags commit=e61c738ee3f5cb140b21945fce64f355631f78d0 nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false gocacheprog=on load-before=0.28 0.08 0.03 1/137 684 load-after=7.61 6.13 2.86 2/224 9488
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (520.415s)
setup: source /workspace/adamic-tools/env.sh
setup: logs /tmp/adamic-gate/setup.XYoZ1n
```

The first fetch needed an explicit compiler/loom-entries fetch because the shallow checkout did not have e61c738e yet. The branch was created at that exact commit. An early test attempt saw the not-yet-cloned TypeScript go.mod and failed before execution. /usr/bin/time was absent (exit 127), so measurements use Bash time. Another early Go command hit its outer 120s compilation limit before a test binary ran. The pre-refresh run overlapped setup and intentionally failed all six stale stage0 records and the 03 native/raw-source mismatch; its test package took 54.491s, command wall 118.069s including initial test compilation. That failure is /tmp/entries-before-refresh.log. No final test reached 90 seconds.

The required pre-push lane fetch was attempted twice:

```sh
timeout 120 git fetch -q origin main devtools/fast-gate cloud/merge-tree
# exit 128: fatal: couldn't find remote ref cloud/merge-tree
```

git ls-remote confirmed that cloud/merge-tree is absent; git show origin/cloud/merge-tree:cloud/integration/lane-checks.py likewise failed with invalid object name. No empty-script execution is represented as a lane pass. The available main and fast-gate branches were fetched separately; main still names 80640d7a8b8e456040758f4aa33791e45eceb4ce. The choice was to complete all available validation and publish only the authorized task branch, reporting the missing lane source to the dispatching Circle. No main push, force push or pull request was made.

All test output went to files and the results/logs were read. git diff --check and gofmt checks are clean. Full repository tests, the whole oracle package, macOS, and a fresh stock TypeScript source audit were not run. Historical stock audit hashes that include the old headers remain historical; the fixture bodies, upstream spans and raw Node goldens remain unchanged.
