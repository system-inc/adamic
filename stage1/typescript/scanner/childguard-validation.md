# Stage 1 analysis child guards

Validated 2026-10-08 on `stage1-guards/analysis`, based on `origin/main` at `efe9f404` (contains shared guard `a32c2187`). Setup: `GOPROXY=https://proxy.golang.org|direct`, `bash cloud/setup.sh`; Go 1.27.1, Node 24.19.0, clang 20.1.8.

All five helpers retain their signatures. Both callers of each `bounded` helper now run through childguard, including the combined-output cohere oracle. Byte comparisons, stderr checks, exit-code checks, mutants, refusal checks, native sanitizers and leak checks are unchanged. The original unit left the one-minute fixture probes in helpers and comments unchanged; the integration follow-up below replaces the helper refusal probe. The comments probe remains unchanged. `childguard.Run` forwards guarded output through its pipe writers into the existing destination files.

## Output cadence and chosen windows

All sites select the shared defaults with `Options{}`: FirstOutput 30 minutes, Stall 2 minutes, Ceiling 60 minutes. Output cadence was measured during the loaded real tests using an uncommitted Go overlay of childguard. It records both streams at the same watched Write callbacks used by the guard. The gap below is the largest idle interval after first output, including the conservative last-output-to-exit interval; startup silence is recorded separately. Thus 120 seconds exceeds three times every measured gap.

| Site | Helper | Children observed | Maximum gap (s) | 3x gap (s) | Maximum first-output delay (s) | Stall |
| --- | --- | ---: | ---: | ---: | ---: | --- |
| `stage1/typescript/scanner/scanner_test.go` | `execute` | 39 | 0.960090 | 2.880270 | 1.145189 | 2 min |
| `stage1/cohere/lint/helpers/helpers_test.go` | `run` | 25 | 1.787958 | 5.363875 | 18.397458 | 2 min |
| `stage1/cohere/lint/helpers/comments/comments_test.go` | `run` | 21 | 3.401114 | 10.203342 | 35.130984 | 2 min |
| `stage1/cohere/suppression/suppression_test.go` | `bounded` | 35 | 0.956818 | 2.870453 | 2.809581 | 2 min |
| `stage1/cohere/static_single_assignment/static_single_assignment_test.go` | `bounded` | 31 | 0.708968 | 2.126905 | 1.982873 | 2 min |

Go oracle builds in scanner, helpers and comments produce no output on success; scanner also has a silent clang profiling build (31.129s under load). They rely on FirstOutput until completion. Scanner `--count` Go/native/Node commands buffer their single count until exit. Cohere `go test` children in suppression and static single assignment emit their final summary after the internal test completes, so their work before that summary also relies on FirstOutput. Comments oracle calls with buffered answers and short witness/count commands similarly use FirstOutput before their first response. The longest observed silent Go build was 20.576s; the longest delay before a first byte across output-producing children was 35.131s.

## Planted hangs

Only scratch copies of the five test files were modified, selected with Go overlays; no hang or shortened window is committed. Each Run path was invoked through its actual helper. Both CombinedOutput paths were exercised as well. A child printed an initial progress marker and then looped without further output; the production default Stall killed it and each helper reported `stalled: no output for 2m0s`. The tests intentionally fail through the existing fatal-error checks, and the harness verifies that every failure names stalled. Separate silent-from-start and post-output probes used 200ms FirstOutput/Stall windows to exercise both guard branches. All probes ran while the same 10 CPU burners remained active.

| Site | Production Stall kill / test time (s) | Silent startup kill with 200ms window (s) | Real loaded package time (s) |
| --- | ---: | ---: | ---: |
| `stage1/typescript/scanner/scanner_test.go` | 120.020 | 0.220 | 335.285 |
| `stage1/cohere/lint/helpers/helpers_test.go` | 120.130 | 0.210 | 245.908 |
| `stage1/cohere/lint/helpers/comments/comments_test.go` | 120.010 | 0.220 | 470.808 |
| `stage1/cohere/suppression/suppression_test.go` | 120.110 | 0.220 | 85.043 |
| `stage1/cohere/static_single_assignment/static_single_assignment_test.go` | 120.150 | 0.220 | 93.383 |
| `stage1/cohere/suppression` CombinedOutput | 120.060 | 0.210 | 85.043 |
| `stage1/cohere/static_single_assignment` CombinedOutput | 120.020 | 0.240 | 93.383 |

The production kill suite took 274.187s including compilation and two sequential kill tests in each bounded package. Load averages at its start/end: `19.87 10.48 5.53` / `8.55 10.17 6.72`. Accelerated post-output kill test times ranged from 0.21s to 0.26s.

## Full package runs

Command: `go test -json -count=1 -timeout=0 ./stage1/typescript/scanner ./stage1/cohere/lint/helpers ./stage1/cohere/lint/helpers/comments ./stage1/cohere/suppression ./stage1/cohere/static_single_assignment`. The loaded run adds the scratch childguard cadence overlay; it changes only measurement, not guard windows or termination behavior.

Scanner inputs explicitly enabled: `ADAMIC_SCANNER_BENCH=1`, `ADAMIC_SCANNER_PROFILE_DIR=/tmp/stage1-guards-analysis-final/profile`, `ADAMIC_SCANNER_PROFILE_SNAPSHOTS=/tmp/stage1-guards-analysis-final/profile`. This runs artifacts, snapshot agreement and all five throughput rounds. The pinned TypeScript compiler corpus is included. Cohere uses all checked-in witness/catalog/consumer fixtures, suppression's default 2500 generated cases, and static single assignment's default 2000 generated functions plus React graphs. Counts include test and subtest terminal events, excluding the package terminal event and the separate expected-failure hang probes. No tests skipped in either real run.

| Package | Unloaded pass/fail/skip | Unloaded wall (s) | Loaded pass/fail/skip | Loaded wall (s) |
| --- | --- | ---: | --- | ---: |
| `stage1/typescript/scanner` | 10/0/0 | 136.398 | 10/0/0 | 335.285 |
| `stage1/cohere/lint/helpers` | 8/0/0 | 95.032 | 8/0/0 | 245.908 |
| `stage1/cohere/lint/helpers/comments` | 10/0/0 | 162.112 | 10/0/0 | 470.808 |
| `stage1/cohere/suppression` | 17/0/0 | 49.693 | 17/0/0 | 85.043 |
| `stage1/cohere/static_single_assignment` | 17/0/0 | 52.746 | 17/0/0 | 93.383 |

`nproc=5`; cgroup `cpu.max=400000 100000`. Unloaded: zero burners; total invocation wall 175.068s; load before `0.00 0.13 1.72`, after `2.98 2.70 2.55`. Loaded: 10 Python CPU burners (2x nproc), removed in a finally block; total invocation wall 516.098s; load before `2.98 2.70 2.55`, after `10.33 10.28 7.09`. Package wall times exclude compilation before a package starts. The loaded run also overlapped scratch kill-test compilation and probes, making the measured cadence conservative.

Raw JSON, per-child cadence records, scratch source/overlays and metrics remain under `/tmp/stage1-guards-analysis-final/`: `unloaded.json`, `loaded.json`, `cadence/`, `kill-production.json`, `kill-production.metrics`, `kill-silent.metrics`, `kill-progress.metrics`. No shared-guard, module, setup, submodule or unrelated package files are committed. No sites remain unconverted.

## Integration follow-up: message refusal probe

Integration reported `gate-logs/670c401c898b/20261008T105432Z/fast`, with case 6 killed and empty stderr. This path was still the explicitly retained one-minute `context.WithTimeout` inside `TestMessageRefusalsMatchGo`; it did not invoke childguard, so this diagnostic is not a childguard 120-second Stall termination. A plain `signal: killed` alone does not distinguish a context cancellation from an external SIGKILL. The integration kill was not reproduced here, and no claim is made that a healthy child was observed exceeding 120 seconds.

The test was run alone with `nproc=5`, 20 CPU burners (4x nproc), and the same sanitized native build and exact Go panic/exit-code comparisons. The original fixed-deadline version passed all 10 refusal cases in 45.00s (99.792s including compilation). A scratch guarded diagnostic version passed in 40.13s (74.477s including compilation). All 20 native/Node probes returned exit 70 and the exact expected stderr. Scratch measurements record downstream Write callbacks when using childguard; only end-to-end timings from the original direct-exec instrumentation are used because the buffer's ReadFrom optimization bypassed its Write callback.

Case 6 is `{"Definitions":{},"Cases":[{"Kind":"message","Rule":"structure/consistency-require-matching-file-name","Id":"requireMatchingFileName","Values":{},"Choices":null}]}` with the checked-in catalog. `main.ts` parses the inputs and constructs an empty choices list; `PolicyMessage.render` in `policy_message.ts:46` observes one required phrase but zero chosen options and calls panic. Native `adamic_panic` writes the exact diagnostic then calls `_exit(70)`. At 4x load this native case completed in 0.293305s, first output at 0.287915s; Node completed in 0.501434s, first output at 0.451661s. Thus the missing-choice input did not hang in this reproduction.

Across the guarded probes, maximum first-output silence was 1.063245s. The entire interval from first output to completion was at most 0.143567s, a conservative upper bound on any subsequent silent gap; 120s gives much more than 3x headroom. These refusal children have no progress output before the final panic, so the fix routes their startup through the default 30-minute FirstOutput window. It removes the independent 60-second context, retains the default 120-second Stall and 60-minute Ceiling, and preserves every existing panic-text and exit-70 assertion. This change removes a performance-dependent deadline from the reported failure path rather than increasing Stall without evidence. External SIGKILL and nonmatching/refusal-less port failures still fail the test.

The actual changed file was then tested without an overlay: `go test -v -count=1 -timeout=0 -run=^TestMessageRefusalsMatchGo$ ./stage1/cohere/lint/helpers`. It passed in 36.89s, with total invocation wall 68.502s, 20 burners, load before `7.05 7.08 4.27 21/191 16414` and after `16.83 10.00 5.48 21/192 16669`. Raw diagnostic copies, overlays, per-child timings and final logs remain under `/tmp/stage1-guards-refusal/`. The comments one-minute JSX probe remains unchanged.
