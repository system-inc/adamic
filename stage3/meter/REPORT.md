# Runner unit evidence, October 7

Base main: ef3d907ecdc4c771b016f7d9c52372def057a340.
Prerequisite files were read from origin/codex/stage3-base and
origin/codex/tsc-census and materialized only for local validation. They are not
included in this unit's commit. No in-flight adapters were imported.

Toolchain command: `bash cloud/setup.sh > /tmp/stage3-setup.log 2>&1`.
Go go1.27.1, clang 20.1.8, Node v24.19.0. Go, clang, Node and submodules
ready at 0s; cache warm and done at 96s. nproc 5; cgroup quota 400000/100000.

`stage3/meter/twice-daily.sh > /tmp/stage3-meter-run.log 2>&1` completed.
The timestamped report lists 81 files, including 78 source files and three JSON
inputs after normal upstream generation. Checker passes 0/78; lowering passes
0/78. Lowering was blocked by the checker, not observed failing independently.
No claim of native TypeScript compilation follows from this census.

`python3 -m unittest discover -s stage3/meter -p '*_test.py'` passes six tests,
with output in /tmp/stage3-meter-tests-final.log. `bash -n` passes the shell script.
A scratch copy of report.py removed Refused from the checker-pass kinds.
`test_checker_and_lowering_are_separate` fails: checker total 2 instead of 3,
exit 1, in /tmp/stage3-meter-mutant.log. Missing/duplicate file records, unknown
kinds, and incomplete whole-program coverage are independently rejected by tests.

The three seeds were observed with `node --disable-warning=ExperimentalWarning
oracle/node.mjs <fixture>` and `go run ./cmd/adamic build <fixture> -o <scratch>`.
The compiling returnTrue seed was also built with `--sanitize` and run with
`ASAN_OPTIONS=detect_leaks=1`: stdout is exactly "true\n", stderr empty, exit 0.
Its build log is /tmp/stage3-seed-sanitize-build.log and output files are
/tmp/stage3-seed-native.stdout and /tmp/stage3-seed-native.stderr.

Not done: fixtures_test.go, its small oracle hook, -update, and all three requested
runner mutants. Automatic approval review rejected both writes because the later
shared fixture instructions prohibit editing the runner. The unit-specific
assignment expressly assigns it here; a clarification request is pending.
No compiler/oracle source file was changed. No other fixture bucket was touched.
No cron job, message delivery, full uncached gate, or upstream test suite was run.

Filtered oracle command: `go test ./internal/oracle -run
'TestNativeAgreesWithNode/internal/oracle/testdata/(functions|library_object_order)\.a$'
-count=1 -timeout 30m -v > /tmp/stage3-filtered-oracle.log 2>&1`.
It passed functions.a, generic_functions.a, and library_object_order.a in 2.872s.
The cache reported three native hits, six native misses and six Node misses.
