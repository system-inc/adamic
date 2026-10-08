Built: library-area regex timing checks use child CPU budgets, with separate generous hang caps.
Commits: branch codex/regex-cpu-deadlines-area from area/library bfcddd1c; b79535c2 merged with merge commit 8f5371ba.
Commands/results: full native package PASS 339.514s; focused oracle PASS 211.726s; loaded area oracle PASS 463.468s.
Mutants: native and area nonreturning fixtures caught by hang caps; both disabled CPU checks caught; native and area loaded runs pass.
Uncovered: full oracle/repository gate and macOS were not run; no production runtime changes.

## Implementation

Both requested test names live in internal/oracle on this area snapshot: TestRegExpNativeTiming in regexp_native_fixes_test.go and TestRegExpLongBacktrackNode in regexp_long_backtrack_test.go.

The four timing fixtures keep five trials and Node stdout comparisons. Each native trial checks ProcessState.UserTime() + SystemTime() against three seconds; native best-of-five logs CPU time. Node timing controls retain wall observations with a five-minute cap, and their best value is initialized from the first actual observation rather than a three-second sentinel.

Long-backtrack sanitized, release and Linux leak-check runs each keep the old three-minute limit as a CPU budget, with a separate five-minute wall cap. All stdout, stderr and exit comparisons remain. The process group is killed for a hang, preserving the existing external-tool cleanup behavior. The macOS leaks wrapper is an external tool and uses only the generous wall cap: the release binary was already CPU-checked directly. This macOS path was not run.

The merge resolution preserves area/library's Linux-only LeakSanitizer environment and the inherited native corpus CPU budgets. The area V8 Node oracle's thirty-second wall cap is widened to five minutes. No production runtime C or lowering changes; no new runtime statics.

## Gates and observations

```sh
source /workspace/adamic-tools/env.sh
GOFLAGS=-buildvcs=false go test ./internal/native ./internal/oracle -run 'TestRegExp(NativeStepLimit|Child|Oracle|NativeTiming|LongBacktrackNode)$|TestRegExp(Child|Oracle)' -count=1 -v -timeout=30m
python3 internal/native/testdata/run-regexp-cpu-deadline-probes.py
export WASI_SYSROOT=/workspace/adamic-tools/wasi-sdk/share/wasi-sysroot
GOFLAGS=-buildvcs=false go test ./internal/native -count=1 -timeout=30m
GOFLAGS=-buildvcs=false go vet ./internal/native ./internal/oracle
```

The focused native package passed in 6.670s and the focused oracle package in 211.726s. Normal long-backtrack CPU observations: sanitized 58.787731s; release 13.414238s; leak detection 63.192231s. Four timing fixtures agree with Node; the quadratic trials take approximately 0.45s CPU each against three seconds.

Under eight simultaneous busy loops, the inherited step-limit test passes, and both area tests pass in 463.468s. Long-backtrack uses 70.894737s CPU sanitized, 14.29606s release, and 70.63281s with LeakSanitizer, all below 180 seconds. Its complete three-mode test takes 442.60s wall time under contention; each child has a separate hang cap. All four timing fixtures pass their CPU budgets and Node comparisons; their combined test takes 20.81s. The proof runner kills and reaps exactly its eight hogs in finally. The full native package overlaps that loaded run and passes in 339.514s.

The initial area mutant attempted replacing the C build input and left a Go local unused, causing a build failure. It was rejected by the proof runner and is not a certified kill. The corrected overlay retains evaluation of native.C(program), compiles a real infinite-loop C fixture with sanitizers, and fails TestRegExpNativeTiming/class_string_duplicates only at its 250ms demonstration wall cap. Production remains five minutes. Disabling the actual oracle CPU-budget condition fails the real 100ms CPU-loop witness. The inherited actual-step-limit hang and disabled-native-CPU-check mutants also both fail at their intended checks. Four certified kills, no build-error kills.

Shared oracle helper tests additionally demonstrate a nonreturning sanitized C fixture killed/reaped at a short cap and a 100ms clock-counted C fixture rejected by a ten-millisecond CPU budget. These test the same cap/budget implementation used by both area tests.

Environment and setup are unchanged from b79535c2's logged setup: Go 1.27.1, clang 20.1.8, Node 24.19.0, nproc five, CPU quota four. The worktree shares cohere pin 7945d102a6c18dd36adf9114a758ce646e8b2359; GOFLAGS=-buildvcs=false avoids VCS stamping on the shared submodule symlink. The cold area Go test build was rebuilt before execution. Attached logs contain all final observations. The full repository and full oracle package gates were not run under the worker exception; no production behavior changed.

The closure-convention ancestry poll remains active at three-minute intervals. Last observed main 74fb6490 still does not contain 22fd701a; the replacement-callback branch remains untouched by this unit.
