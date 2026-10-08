Historical Unit 4 report. The wave13 difference was subsequently fixed; see
`AWAIT_NEW_LANDING_REPORT.md` for its removal and the current 60 matching cases.

The incomplete compiler gate compares every selected input on typescript-go,
Node and sanitized native. `incomplete_test.go` prepares the healthy builds,
then starts its workers before the parent joins the parallel queue. It schedules
22,497 cases from the pinned 77-file compiler corpus on
`min(runtime.NumCPU(), 8)` workers. This box starts five workers. Each manifest
contains at most 64 cases and 8 MiB of input; one unusually large input is kept
whole. Every manifest child now has a 120-second CPU budget and a 60-minute wall
backstop, following the progress-guard correction below.

`incomplete_shards_test.go` requires ordered case headers, compares every case's
body without stopping at the first differing case, and names missing, extra, duplicate and changed cases. The full gate also
checks the planned count, the generated count, a completion ledger and the final
compared count. It rejects `ADAMIC_RECOVERY_START`, so the full gate cannot
silently become a suffix of the corpus. A failure retains inputs and all three
answers under `ADAMIC_RECOVERY_ARTIFACTS`, along with a runnable manifest.

`shared_test.go` shares the healthy oracle and sanitized/release port builds only
within one test process. Every scratch mutant directory builds independently.
Clang keeps the standard flags: sanitized builds use `-O1` with address and
undefined-behavior sanitizers; release builds use `-O2`. Native build children now have
a ten-minute CPU budget and a 60-minute wall backstop. The merged lint entry requires its checker archive;
`checker_build_shared_test.go` links the real sanitized archive and uses the
landed split builder with one clang worker. A private compiler proxy guards
each clang call without changing the parent environment. Successful shard answers use the test's temporary
directory and are removed after comparison.

`testdata/batch_node.mjs` formats each console call with Node's formatter and
writes bounded chunks, including a final flush on normal completion or exit.
It is used by the incomplete gate only. `nodes.ts` reuses fixed field text and
common depth text; greater depths retain ordinary number formatting. Paired
77-file checks require identical output. The measurements are in
`validation/unit4/node-buffer-timing.json`, `writer-timing.json` and
`depth-timing.json`. The slower Map-based offset-cache experiment was discarded.

The one-byte and dropped-case experiments changed `main.ts` in scratch copies.
Both real test commands exited 1 and named
`testdata/recovery/corePublic.ts-cut-37.ts.txt`. Their source changes and exact
changed/dropped counts are recorded in `validation/unit4/mutants.json`; the
logs are `mutant-byte.log` and `mutant-drop.log`. The temporary probe tests and
mutant sources were never committed.

`validation/unit4/before-times.csv` records all Unit 3 top-level durations.
`after-times.csv` sorts all final top-level durations by active time, calculated
as run-to-pause plus continue-to-pass. `summarize.py` regenerates that table and
pass/fail/skip counts from `parser-final.jsonl`. These times overlap and do not
sum to the package wall time; parent tests include their parallel children's
completion. Corpus workers can run while their parent is paused; the requested
active-time formula excludes that queue interval, while package wall time and
child timers continue to include it. Unit 3 did not record package-boundary load readings; `before-box.json`
marks them unavailable rather than substituting late benchmark readings.

`testdata/lint_cases/wave13_top_level_await_new.ts` and its expected-difference
sidecar remain unchanged. The collection contains 58 cases: 57 agree and this
one retains its documented difference.

Run the full parser gate with the pinned, clean TypeScript checkout in
`ADAMIC_TYPESCRIPT_SOURCE`, `ADAMIC_PARSER_BENCH=1`, and
`go test ./stage1/typescript/parser -count=1 -json -v -failfast=false -timeout=90m`.
The opt-in benchmark flag avoids the two performance-test skips. Also run
`go test ./stage1/cohere/lint -run '^TestRulesAgree$' -count=1 -v -timeout=90m`
and `go test ./stage1/typescript/parser -run '^TestLintCases$' -count=1 -v -timeout=90m`
with the same TypeScript source. Final measurements and counts are recorded in
`validation/unit4/report.json` and `after-box.json`.

The final guard correction uses a small local adapter in `child_guard_test.go`,
following `stage1/cohere/estree/deadlines_test.go`. It can be replaced with
`internal/childguard` when that helper lands. The helper re-execs this test
binary, sets `RLIMIT_CPU`, then execs the target with the same PID. Quiet builds
and corpus children spend their CPU budget only while doing work; queueing and
CPU contention do not spend it. The long wall backstop catches children blocked
without using CPU. Guard failures include the command and input and say stalled.
Normal builds include runtime compilation inside the guarded helper; checker
builds also guard each compiler call. The manifest size bound remains 8 MiB
with 120 CPU seconds of headroom. The prior longest manifest used 27 wall seconds.

`TestParserChildGuard` plants an infinite loop and runs a healthy child with one
CPU burner pinned to each available core. The healthy child shares a burner’s
core. In `validation/progress-guards/proofs.log`, the hang is killed after
1.025 wall seconds and names `testdata/recovery/corePublic.ts-cut-37.ts.txt`
as stalled. The healthy child survives 6.046 wall seconds while using 2.412
CPU seconds under its 3-second CPU budget, with all five cores occupied. Neither
proof asserts a performance target. Final gate events, wall times, counts and
box readings are in `validation/progress-guards`.

The final merge includes `origin/area/stage1-lint` at `c8f6d74f`, which contains
the requested `c4bdc23f`. The top-level-await expected difference is unchanged.

The completed parser rerun took 936.966 package seconds (939.102 seconds
including the go command), with 176 passes, zero failures and zero skips. All
22,497 inputs still matched on all three implementations. This is a measured
time, not a performance assertion. On host `0adeb267a186`, `nproc` is 5 and
the CPU quota is 4 (`400000 100000`). Load before the complete rerun was
0.58 / 4.13 / 4.09; after it was 5.54 / 6.40 / 6.05. The longest native build
was 219.143 seconds. An earlier interactive run lost its executor connection
after 13,978 comparisons; `parser-interrupted.jsonl` preserves that incomplete
run. The detached `run_gates.py` recorder then restarted the entire corpus
and writes each gate’s exit code, wall time and box readings.

`TestRulesAgree` passed in 92.761 package seconds (94.970 command seconds),
with one pass, no failures and no skips. The standalone `TestLintCases` passed
in 22.149 package seconds (24.321 command seconds), with 59 passes, no
failures and no skips. Its 58 input files retain 57 exact matches and the one
documented wave13 difference. All gate commands exited zero with the clean
pinned TypeScript checkout set. `report.json` contains exact command wall times
and each gate’s before/after load; `after-times.csv` lists every parser top-level
test using run-to-pause plus continue-to-pass timing.
