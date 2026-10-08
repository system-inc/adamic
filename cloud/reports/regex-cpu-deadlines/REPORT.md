Built: regex native deadline checks use child CPU time and separate generous hang caps.
Commits: branch codex/regex-cpu-deadlines, based on origin/main 855d114e9b37776ec3739f25d63dbf4da968d02e.
Commands/results: full native package passes in 121.163s; focused tests pass; uncached sixteen-fixture regex oracle passes in 5.205s; loaded-box step-limit test passes.
Mutants: an infinite native fixture fails at the wall cap; disabling the CPU check fails the CPU-overrun witness.
Uncovered: whole-repository gate and WebAssembly were not run; two named timing tests are absent on this main.

## Change

TestRegExpNativeStepLimit keeps exit 70 and the exact instruction-step-limit message. It now gives the child a five-minute wall cap only to stop hangs and checks ProcessState.UserTime() + SystemTime() against ten seconds after it exits. Sanitizer startup/shutdown is included in CPU time; scheduling delays are excluded.

runRegexCases previously shared one two-minute wall deadline across both bounded and unlimited runs. Each run now has its own five-minute hang cap and a two-minute CPU budget. TestRegExpBytecodeRandomNode's external Node guard is increased from thirty seconds to five minutes. No production runtime changes.

An audit of internal/native and internal/oracle found no other regex-specific wall-clock checks in this main. TestRegExpNativeTiming and TestRegExpLongBacktrackNode do not exist in this snapshot, so there is no change to those names. The ordinary shared oracle harness is unchanged. Assumption: preserve the old two-minute native corpus limit as a CPU budget per run.

## Evidence

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
GOFLAGS=-buildvcs=false go test ./internal/native -run 'TestRegExp(NativeStepLimit|Child)' -count=1 -v
python3 internal/native/testdata/run-regexp-cpu-deadline-probes.py
GOFLAGS=-buildvcs=false go test ./internal/native -count=1 -timeout=30m
ADAMIC_GATE_UNCACHED=1 GOFLAGS=-buildvcs=false go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/regexp' -count=1 -v -timeout=30m
GOFLAGS=-buildvcs=false go vet ./internal/native ./internal/oracle
```

Focused gate: 5.078s. The correct step-limit child uses 3.775ms CPU against ten seconds and retains exit/message checks. The CPU-overrun fixture executes a real 100ms clock-counted C loop; its observed CPU is 106.744ms against a ten-millisecond demonstration budget. The nonreturning sanitized C fixture is killed and reaped at a 250ms demonstration wall cap.

The proof runner uses Go overlays without modifying source files. Its hang mutation inserts an infinite loop into the actual step-limit fixture and shortens only the demonstration wall cap to 250ms; the step-limit test fails with the wall-cap error. Production keeps five minutes. Its second mutation disables CPU-budget enforcement; TestRegExpChildCPUBudget then fails because the real overrun escaped. Neither is a build-error kill. The unchanged step-limit test runs alongside eight busy-loop processes, uses 3.589ms CPU against ten seconds, and passes in 0.508s. The runner kills and reaps exactly its eight hogs in finally.

The sixteen oracle fixtures include the twelve existing regex method fixtures and four regex cycle fixtures. All agree with Node on both backends, with Linux sanitizers/leak checks; native result cache was bypassed. The first oracle filter selected no tests; this was noticed and corrected before recording the successful gate.

Setup: nproc 5, cgroup quota four CPUs, Go 1.27.1, clang 20.1.8, Node 24.19.0. Setup ready timings: submodules 0.071s, clang 0.174s, Go build 28.482s, total 28.739s. Setup ran in the existing primary checkout; the main-based worktree shares its exact cohere pin 7945d102a6c18dd36adf9114a758ce646e8b2359 and toolchain. GOFLAGS=-buildvcs=false avoids Go's VCS stamp error for that shared submodule symlink.

Initial checks hit a full root filesystem (27GB Go build cache). Those failures were not certified as mutants. Removed 6,461,571,965 bytes of disposable cache entries older than one hour, then reran all checks. No source or proof artifacts were removed.

Final full native package gate: PASS, 121.163s. Vet and diff check: exit zero, no output. TestRuntimeStaticsAreListed is absent from this main; its explicit filter reported no tests, not a statics certification. This test-only unit adds no runtime C or statics. The full repository gate was not run under the worker exception.
