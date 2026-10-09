u014: TestProgramCPUDeadline
Starting origin/main: 7b18d0576930caca4e22ce2eef92fcf563af52d0

CODE UNDER TEST, named before mutation: Adamic's runProgram CPU-limit shell wrapper and its command execution/capture path in cmd/adamic-test262/run.go.
ORACLE, named before mutation: the unchanged handwritten execution-field/stdout expectations and C clock() helper in run_test.go. The helper is built with clang; neither it nor the tests nor an external compiler oracle was mutated. Oracle kind is self. No external authority was checked or claimed.

Scope was taken from go test -list . ./cmd/adamic-test262/ at the base. The named row exists in run_test.go; it did not move or vanish. Its subtests are spin and saturated, not separate audit rows. The package list has 37 top-level Test functions. Only this row receives a verdict, while all 37 functions were observed in every whole-package matrix. Grouping other functions into families cannot change the two unique kills, because exactly this one function failed in both runs. TestCompilerHangHelper and TestRunnerLocationHelper are subprocess entries for other rows, not scoped rows.

Reached named package functions were runProgram, runCommand, runCommandWithLimit, limitedBuffer.Write, limitedBuffer.String and the init startup guard in compiler.go. functions.txt contains isolated-row coverage: runProgram/runCommand/String 100%, runCommandWithLimit 60%, Write 77.8%, init 20%; total package statement coverage for this row 1.8%. Startup guard is incidental preparation; it was not mutated.

The fixed menu selected four production mutants before checking any failures: floor CPU-limit rounding, invert SIGXCPU recognition, replace the two-minute wall backstop with the CPU duration, and change the exact capture-capacity boundary. Their independent diffs are M1.diff through M4.diff. Each applied against the base and separately passed go vet ./cmd/adamic-test262/. E_RUN_PROGRAM.diff is an independently vetted empty-answer probe, not a mutant and not used for verdicts. switch.diff is instrumentation only, for one compiled selector binary. Production source was restored before the final enabled uncached package check and vet.

Matrix command for ID:
ADAMIC_MUTANT=ID ADAMIC_TEST262_MEASURE=1 ADAMIC_GATE_UNCACHED=1 timeout 120 go test -json -count=1 -timeout 90s ./cmd/adamic-test262/ -run . > ID-matrix.log 2>&1
The opt-in compiler startup row is enabled in the clean enabled baseline, every matrix, and final restored run. Ordinary observation caches are bypassed; rows explicitly testing caches set their own settings and use their cache constructions. All matrix rows completed; no skips, panics or over-budget results, so no bounded matrix or isolated recovery was necessary. Raw exact test failures are in matrix.json and matrix.csv.

M2 and M3 each fail only TestProgramCPUDeadline. M2 is caught by spin, with TimedOut:false and Signal:CPU time limit exceeded. M3 is caught by saturated, which times out after 2.045062464s. The enabled clean saturated run completed its one CPU second after 2.582271724 wall seconds, beyond its 2s CPU budget. The spin assertion alone cannot distinguish a CPU timeout from an early wall timeout; the positive saturated case supplies that distinction in this session.

The empty runProgram result fails both scoped subcases, so vacuous=false and vacuous_subcases is empty. Probe failures in other package functions are also recorded, but never used as mutant kills or unique kills.

Survivors: M1 and M4. They are not equivalent candidates. observation-source.txt is a separate temporary observation entry; it was not used to decide verdicts and was removed before restoration. Run commands and times are in commands.jsonl. The compiled driver was run with empty, M1 and M4 selectors:
ADAMIC_MUTANT=ID /tmp/u014/observation.test -test.v -test.run '^TestAuditObservation$' > survivor-ID.log 2>&1
For a 1500ms runProgram request, shell ulimit -S -t prints 2 in control and 1 under M1. Capturing printf abc with capacity 3 exits 0 in control, but under M4 exits -1 with command output exceeded capture limit. Complete before/after logs are preserved.

Brief interpretations and costs:
- The reference file/commit in the brief is older than current origin/main; scope was verified at the actual fetched base, and the requested test still exists at the same path.
- C helper execution is part of a self-written test, not an external-run comparison against Node or an independently named reference runner. No outside authority value was checked.
- Whole-package baseline fit the 90s binary budget, so despite the unit being a slice, full-package matrices could establish package uniqueness. Repo-wide uniqueness remains for central replay.
- The default baseline skipped TestCompilerStartupMeasurement. The explicit opt-in was available, so an enabled clean baseline was added before mutations. All matrix and final enabled rows ran without skips.
- npm ci in stage3/api was mandatory though this scoped row uses clang and no Node dependencies. npm reported 451ms. Warm tools worked, setup was skipped, nproc was 5.
- CPU saturation is scheduling-dependent. The test itself only logs whether wall time exceeded the CPU budget; in the clean enabled run it did. The median uses three isolated package ok lines, not package-with-other-tests timing or outer go command time.
- Exactly-full output and fractional budgets are outside the scoped row's supplied inputs. The full package also missed both mutations, with direct changed-output witnesses.
- No instruction required clarification and no permission pause occurred. No compiler or stage1 port source was mutated, so no native product cache key or port rebuild was needed. Existing helper/program builds are part of test costs.

No other package tests were run, no full-repository gate or repo-wide uniqueness was claimed, no PR was opened, and main was not pushed. Measurements cover this Linux workspace, not macOS, blocking-child deadlines, other signals, spawn failures, descendant process cleanup or output limits generally. Only the selected mutations and entry probe are proven here. The final production source diff is empty.
