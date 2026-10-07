# Child deadlines

Go implementation: `2ae227725a03db60a86b422a50dbfe1a5ac1a662`. Shell fix: `8f5c6e2`.
Baseline: `ac92bf0abf36b0fa442929b4e9a434445cb15a17`, origin/area/developer-tools.
Only devtools/child-deadlines was pushed. internal/oracle is unchanged.

## Audit

Each row names the original call site, including test and measurement children.
No os.StartProcess calls were found in the requested territory. LookPath does not start a child.

| Call site / children | Before deadline | After deadline |
| --- | --- | --- |
| gate main.go output: git, go version, clang version, node version, nproc and discovery | none | 30s probes; 10m Go list/test discovery |
| gate main.go shard runner: gofmt, go vet, go test | none externally; Go test had internal 60m timeout | 10m formatting/vet; 70m Go test |
| gate main_test.go selector subprocess go test | none | 10m |
| test262 run.go prepare/fallback go build | none | 10m |
| test262 cache.go node --version; run.go git rev-parse | none | 30s |
| test262 compiler.go compiler worker | none for process lifetime; 2m request timer | 70m process, 2m request, bounded close/reap |
| test262 run.go compiler, clang, native, Node | compiler/clang 2m; native/Node 15s; direct child cancellation | same limits, group kill and bounded reap |
| test262 RuntimeLibrary indirect clang/version/objects/ar | none | isolated operation group 10m |
| test262 performance_test.go go build/compiler | none | 10m / 30s |
| test262 compiler_test.go helper; edit_test.go location helper | none | 10m |
| test262 cache_test.go and corpus native Build indirect clang/ar | none | isolated test operation 2m |
| test262 measure.py check_output/run; measure_edits.py check_output/build/run | none | 10m per child |
| reducer/generator through fuzz run.go go build and clang | none | 10m / 2m |
| fuzz execute compiler, compiler JS backend | 30s direct child | 30s group |
| fuzz execute Node, native, backend Node, sanitizer leak run | 20s direct child | 20s group |
| fuzz RuntimeLibrary indirect clang/ar, including runtime_test.go | none | isolated operation group 10m |
| catalog check.sh Python supervisor | none | 4h aggregate |
| catalog check.py indirect git/go/probes via shared Python launcher | none | 10m per child; Go test 70m |
| setup utility wrappers: cat, awk, mkdir, install, mktemp, realpath, uname, ls, sort, dirname, ln, mv, grep, nproc, sha256sum, cut, head | none | 30s |
| setup flock | none | 10m |
| setup curl, tar, git submodule, go list/build/test | none | 10m |
| setup go/node/clang version, git config, setup-key Python, report git | none | 30s |
| setup sanitizer clang / executable | none | 120s / 15s |
| setup markdown dependency Python supervisor | none | 30m aggregate; its curl/tar/npm children 10m |
| setup four preparation shells | none | 30m aggregate |
| internal/leakcheck | absent from this branch | gap; landing from devtools/stage1-leaks |

Shared Go helper requires a context deadline, sets a new process group, kills the
whole group on expiry, names the child, bounds pipe draining and gives reaping at
most 2s. Runtime library operations run in a bounded re-exec group without editing
native or changing its library keys. Shell uses timeout --kill-after=1s and a TERM
supervisor that stays alive until group SIGKILL, including when the real child exits
first. It bypasses exported wrappers with command and explicitly preserves stdin.
Python uses new sessions, group SIGKILL and bounded pipe cleanup.

## Measurements and proof

Interleaved before/after, three rounds, same box and baseline. The table uses best
of three. A forever child has no natural completion: "before" is a watchdog
observation, not a completed run. All six after runs assert the child was named,
the tool returned before the watchdog and the grandchild heartbeat stopped.
Setup includes two consecutive capped probes and their kill-after grace.

| Tool | Child | Deadline | Before | After |
| --- | --- | --- | --- | --- |
| adamic-gate | go | 0.2s proof cap | hangs; watchdog stops at 1.5s | killed in 0.215s |
| adamic-test262 | go | 0.2s proof cap | hangs; watchdog stops at 1.5s | killed in 0.214s |
| adamic-reduce | go | 0.2s proof cap | hangs; watchdog stops at 1.5s | killed in 0.215s |
| adamic-fuzz | go | 0.2s proof cap | hangs; watchdog stops at 1.5s | killed in 0.215s |
| catalog | go | 0.2s proof cap | hangs; watchdog stops at 1.5s | killed in 0.265s |
| setup | go | 0.2s proof cap | hangs; watchdog stops at 1.5s | killed in 2.472s |

Instrument: `python3 internal/boundedrun/prove.py --repository /workspace/adamic --output /workspace/scratch/child-deadlines`.
Exact per-tool commands, per-run source commits, cache modes, and load averages
before and after are in [child-deadlines.json](child-deadlines.json). Go binaries
are from 2ae2277; final shell measurements are from 8f5c6e2.

Build-flags: `go build -buildvcs=false`; GOFLAGS empty; nproc=5;
cgroup cpu.max=`400000 100000`; go1.27.1 linux/amd64;
clang20.1.8 (LLVM 87f0227cb60147a26a1eeb4fb06e3b505e9c7261);
Node v24.19.0; hang proofs ADAMIC_GATE_UNCACHED=1, Go build cache warm.
Every sample includes its commit and load-before/load-after in the JSON.

| Loop | Before | After | Instrument |
| --- | --- | --- | --- |
| checked-in mini, uncached | 1.318s | 1.322s | `bin-VARIANT/adamic-test262 -adapt -json -jobs 4 -test262 cmd/adamic-test262/testdata/mini -profile PROFILE ""` |
| checked-in mini, warm | 1.068s | 0.968s | `bin-VARIANT/adamic-test262 -adapt -json -jobs 4 -test262 cmd/adamic-test262/testdata/mini -profile PROFILE ""` |

Normal comparison build-flags: the same toolchain/quota; GOMAXPROCS=4;
baseline ac92bf0 vs Go implementation 2ae2277; cache mode shown per row.
All twelve samples demand byte-identical JSON and ordered stderr, including
uncached and warm observations. Per-sample load averages and exact commands
are in the JSON. This is correctness evidence, not a speedup claim.

Deadline basis: setup observed 32.803s (historical cold 114s); a gate package
recorded 918s, and Go already permits 60m. Local per-test maxima before the
change were compiler .076s, clang .188s, native .005s and Node .084s; broader
recorded cold compiler/clang cases reach roughly 18s/19s. Existing 2m/15s/30s/20s
execution limits are retained. Build/download 10m and aggregate 30m allow ample
margin; catalogue 4h accommodates many 70m checks. Comments beside limits state
the measurements and rationale.

## Validation and mutants

- Scoped Go suite: internal/boundedrun, cmd/adamic-gate, cmd/adamic-test262,
  cmd/adamic-reduce, internal/fuzz passed. `go vet ./...` passed. Shared helper
  race test passed. Test output is in workspace logs.
- Filtered unchanged oracle: `go test -count=1 -v ./internal/oracle -run
  'TestNativeAgreesWithNode/internal/oracle/testdata/library_math_number_math.a'`
  passed, native misses=3 and Node misses=2.
- Actual catalogue entry 1 against ac92bf0 passed its recorded applies-and-fails check.
- Final shell validation: `bash -n cloud/setup.sh internal/boundedrun/shell.sh
  verify/catalog/check.sh`; `python3 internal/boundedrun/python_test.py`: six checks passed.
- Real final setup passed: go ready .053s; Node .051s; submodules .097s;
  markdown dependencies .125s; clang .249s; Go build 8.404s; complete 8.671s.
  Build-flags commit=2ae2277 with final shell working tree; nproc=5; quota/toolchain
  as above; cached=yes; warm-tests=false; load-before=0.00/0.17/0.71,
  load-after=0.72/0.32/0.76. Environment sourced from /workspace/adamic-tools/env.sh.

| Mutant | Check that failed |
| --- | --- |
| Kill only Go child PID instead of group | grandchild heartbeat in helper, gate, test262 command/worker and fuzz tests |
| Kill only Python child PID | test_python_group_deadline heartbeat |
| Remove Go bounded reaping select | TestWaitCannotHangInOutputWriter semantic 3s bound |
| Remove child name from Go error | TestDeadlineKillsGrandchild diagnostic assertion |
| Treat ordinary Python SIGKILL as timeout | test_signal_exit_is_not_deadline |
| Drop shared helper identity from Node cache key | TestNodeHelperIdentityCache stale observation |
| Remove shell --kill-after | test_shell_kill_after bounded outer watchdog; rerun on final shell |
| Remove shell TERM supervisor trap | test_shell_exited_leader_still_kills_grandchild; rerun on final shell |
| Allow exported utility wrapper recursion | test_shell_exported_wrapper_and_pipeline |

Each mutant was run, failed with exit 1, and restored. Raw logs are deliberately
uncommitted under /workspace/scratch/child-deadlines. Only this summary and
measurement JSON are committed.

Gaps: internal/leakcheck is absent. Full gate was not run; scoped tests plus the
filtered oracle and actual catalogue entry were run. Hanging network tools were
covered by shared mechanisms rather than separate injected proofs for every
curl/tar/npm/git call. Uninterruptible kernel I/O cannot be made to reap immediately;
the Go/Python parent wait remains bounded. No claim is made for descendants
that deliberately escape their process group.
