Four of five rows defended by package-unique production mutants.
All seven mutant runs completed all 41 tests; no narrowing or skips.
CompilerStartupMeasurement remains not defended after three capture faults.

Starting commit: 23a19b8b48662dfb6a19baa00b8efbb52ab5275e.
Production selectors were restored after runs; standalone diffs contain no switches.
Every standalone diff passed go vet and git apply --check against the clean starting tree.
CODE UNDER TEST and ORACLE are recorded per row in code-and-oracles.json and report.json.
Coverage profiles instrument the code under test, cmd/adamic-test262, for each row and its prior subsumer.
Detailed coverage differences: coverage-differences.json and coverage/*-difference.txt.
Complete observed pass/fail lists: matrix.json. Raw logs: D01.log through D07.log.
Commands and wall timings: baseline-times.json, coverage-times.json, runs.json.
Each mutant uses a separate ADAMIC_BUILD_CACHE_DIR. Measurement opt-in is enabled throughout.
Setup skipped because the warm toolchain worked; nproc=5; npm ci stage3/api succeeded in 0.951s.
Clean baseline: 75.729s binary, 78.542s command wall. No red baseline and no cooked steps.

Brief feedback and limits:
* The prior audit was on an older commit with 37 tests. Current main has 41, including new capture and deadline tests; all were included.
* A normal git fetch did not retrieve the audit branch under this checkout's limited fetch configuration; an explicit ref fetch was needed.
* Exclusive statement coverage cannot represent different cache identity edit histories. EditCacheSeparation's runtime-key history is distinct from LoweringSourceEdit's context history, despite reaching shared statements. D04 tests that semantic difference.
* StartupMeasurement has zero exclusive covered lines relative to LargeCompilerOutputIsComplete. Three honest shared-capture attempts were made; every one was caught by another row too. Lack of exclusive lines alone was not treated as a verdict.
* StartupMeasurement is opt-in and asserts exact C output, but no startup or performance threshold. Its oracle is the compiler's own in-process output, not an external authority. This defense does not establish a performance guarantee.
* D06 makes the implicit full-write slice bound explicit and off by one. This is an off-by-one bound mutation, not an inserted statement.
* Full-package runs take about a minute, largely due to CPU deadline tests. Seven full matrices consume most of the budget; per-row coverage adds instrumentation build cost.
* The 90-second step rule and 120-second outer timeout distinguish test time from compilation ambiguously. We retained the prescribed 90-second binary timeout and 120-second compilation backstop. No step reached either limit.
* No tests were rewritten, weakened, or deleted. No packages outside the requested package were tested, and repo-wide uniqueness remains unknown.
* The report's subsumed_by is the prior audit's named subsumer, retained for provenance; it is disproved for the four defended rows by the new mutants.
