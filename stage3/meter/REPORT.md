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

The runner and its small oracle hook are now complete following the user's
ownership clarification. The only internal/oracle addition is the 35-line
stage3_hook_test.go; existing helpers and oracle_test.go are unchanged. No other
fixture bucket or production compiler file was changed. No cron job, message
delivery, full uncached gate, or upstream TypeScript suite was run.

Filtered oracle command: `go test ./internal/oracle -run
'TestNativeAgreesWithNode/internal/oracle/testdata/(functions|library_object_order)\.a$'
-count=1 -timeout 30m -v > /tmp/stage3-filtered-oracle.log 2>&1`.
It passed functions.a, generic_functions.a, and library_object_order.a in 2.872s.
The cache reported three native hits, six native misses and six Node misses.

## Completed fixture gate

`go test ./stage3/fixtures -count=1 -timeout 10m -v` passed all three seeds
in 4.937s, with output in /tmp/stage3-runner-baseline.log. The compiling seed's
native check calls the existing sanitizer and leak helpers uncached.
`go vet ./stage3/fixtures ./internal/oracle` passes, with output saved in
/tmp/stage3-runner-vet.log.

The audit command is `python3 /tmp/stage3-runner-audit.py`, with its summary in
/tmp/stage3-runner-audit.log. Scratch inputs are under
/tmp/stage3-runner-audit-gcalu7y1. Each Go test invocation ran to completion with
stdout/stderr redirected to its own log, not piped.

| Mutant | Named check and observed catch |
| --- | --- |
| Recorded Node stdout true\n changed by one byte to truf\n | TestFixtures/runner/03_return_true.a/node: recorded Node byte comparison failed; exit 1 |
| Recorded Compiles changed to NotYet, diagnostic unchanged | TestFixtures/runner/03_return_true.a/stage0: gap changed: update status.json and check the native output; exit 1 |
| Scratch runtime/string_build_impl.h changes ADAMIC_STRING("true") to ADAMIC_STRING("truf") | TestFixtures/runner/03_return_true.a/native: native versus Node byte comparison failed, native stdout truf\n, Node true\n; both exit 0, empty stderr |
| Refused fixture's diagnostic appended with !, outcome unchanged | TestFixtures/runner/02_is_string.a/stage0: gap changed; exit 1 |

The native mutant ran through a Go -overlay file used to rebuild the oracle hook.
The live runtime was not edited. Node remained true\n, the stage0 outcome and
diagnostic remained Compiles/empty, clang and ASan/UBSan/leaks passed, and only
the native byte comparison failed. The other two requested mutants each failed
only their intended check. Detailed logs: /tmp/stage3-runner-node-mutant.log,
/tmp/stage3-runner-stage0-mutant.log, /tmp/stage3-runner-native-mutant.log and
/tmp/stage3-runner-diagnostic-mutant.log.

## Update evidence

A scratch compiling record changed to NotYet with diagnostic "old" was refreshed
by -update, exit 0. An independent byte comparison verified that the entire
status file changed only at that stage0 value; a normal gate on the refreshed
record then passed. The real seed status.json was unchanged throughout.

Three update refusals each returned exit 1 and left the complete status file
byte-for-byte unchanged: stale recorded Node, the one-byte native runtime mutant
with stale stage0, and a stale outcome on the noncompiling nested-function gap.
The latter proves -update does not bless an unexecuted native program. Logs:
/tmp/stage3-runner-update-success.log, /tmp/stage3-runner-updated-baseline.log,
/tmp/stage3-runner-update-node-rejected.log,
/tmp/stage3-runner-update-native-rejected.log and
/tmp/stage3-runner-update-gap-rejected.log.

Not covered: other workers' incoming fixture buckets were not present here;
nonerasable enum/namespace Node runner changes are not merged into this base.
This gate uses the current oracle/node.mjs and does not weaken checker options
or substitute Adamic-generated JavaScript for Node's source oracle.

Final baseline after aligning Node/native working directories:
`go test ./stage3/fixtures -count=1 -timeout 10m -v` passed in 1.888s,
log /tmp/stage3-runner-final-baseline.log. The complete mutant/update audit was
rerun afterward and all nine assertions passed. The combined touched-package
and filtered oracle gate passed (fixtures 2.115s, oracle 0.671s), log
/tmp/stage3-runner-final-gate.log. It selected TestFixtures, the dormant hook,
and functions.a, generic_functions.a and library_object_order.a. No full gate
was attempted for this harness-only change.


## Landing runner follow-up

The all-bucket Linux gate on origin/land/stage3 d9fc3df passes 173 fixtures
in the supplied 11 status buckets: 0 fail, 0 skip, six native checks, 41.717s.
The per-bucket table and complete commands are in ../fixtures/README.md;
JSON counts are in runs/20261007T025701Z.runner-landing/fixtures.json.
Setup completed in 271s: Go ready 0s, clang/Node ready 0s, submodules ready 19s,
build cache warm 271s; nproc 5, cgroup quota 400000/100000, 17.6 GB.
Setup log: /tmp/stage3-landing-setup.log. Vet passes in
/tmp/stage3-landing-vet.log. The transform/path/platform audit passes in
/tmp/stage3-landing-audit.log, with each mutant returning exit 1 at its intended
check and the foreign-platform skip probe returning exit 0 with its reason.
Only the requested platform fields on three host entries were added; no
recorded behavior, stage0 diagnostic, fixture source, compiler or oracle source
was edited. macOS and the complete repository integration gate were not run.

The filtered oracle gate also passes functions.a, generic_functions.a and
library_object_order.a in 0.657s, log /tmp/stage3-landing-oracle.log. Commands
used -count=1 and redirected test output to files. Ordinary fixtures still use
the unmodified oracle/node.mjs; the temporary transform runner follows the
mode used by codex/flag-enums and codex/namespaces-tsc.
