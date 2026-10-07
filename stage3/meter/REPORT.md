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


## Two-line meter follow-up

Branched from origin/area/stage3 ce83c6a56b1b820a0fa77e440d98ccfacfcd192c.
Only stage3/meter changed. Every invocation fetches main and area, snapshots
their respective apply scripts and adaptations, and measures both trees with
one ordinary census binary. The compiler and both tree commits are recorded.
Existing JSON fields remain the area observation; trees.main and trees.area
contain both complete observations with the new checker measures.

Completed fresh run: runs/20261007T053859Z.fVzNrN/report.md and report.json.
Main c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06: whole program 0/78,
own file 22/78. Area 2a622d263fc4f1e8cad2f8498367c98f1e2f3712:
whole program 1/79, own file 25/79. Area includes hostErrors.ts, absent from
main. Whole program uses the existing loaded-program checker gate. Own file
attributes primary diagnostic locations across all census observations,
deduplicates repeats, and counts global/external findings separately.

Commands and observed results, with outputs retained in that run directory:

```sh
bash cloud/setup.sh > /tmp/meter-two-lines-setup.log 2>&1
source /workspace/adamic-tools/env.sh
bash stage3/meter/twice-daily.sh > /tmp/meter-two-lines-run.log 2>&1
CENSUS_BINARY=/tmp/adamic-gate/stage3-meter.CjYpiQ/census python3 -m unittest discover -s stage3/meter -p '*_test.py' > /tmp/meter-two-lines-tests.log 2>&1
go vet ./stage3/census/tool > /tmp/meter-two-lines-vet.log 2>&1
bash -n stage3/meter/twice-daily.sh
git diff --check
```

Setup: Go/clang/Node ready at 0s; submodules ready at 20s; cache warm and
complete at 274s. nproc 5; cgroup quota 400000/100000; 17.6 GB.
Both apply/census phases completed and the paired JSON/table were emitted.
Eleven tests pass, no skips, including the real checker probe. Vet, shell
syntax and diff whitespace checks pass. See setup.log, tests.log and vet.log.

The real probe changes an imported dependency's number initializer to a string.
Before: both files pass both measures. After: neither loaded program passes,
but only dependency.a changes its own-file result, leaving main.a passing.

Scratch report.py mutants were run independently without editing live code:

| Mutant | Check that caught it |
| --- | --- |
| Own-file result copied from root's whole-program result | test_planted_diagnostic_changes_only_its_own_file and test_type_error_in_dependency_changes_only_dependency_own_file fail their attribution assertions |
| Area census reused for main | test_pair_preserves_area_fields_and_prints_four_numbers_first fails the four-number comparison |
| Malformed diagnostic format silently ignored | test_malformed_diagnostic_is_rejected fails because no ValueError is raised |

Every mutant returned nonzero solely from the expected assertion, without a
setup error. Complete assertion traces are in mutants.log. Raw census JSONL
is retained locally compressed in main/area, excluded from Git as before.
No compiler, runner, bucket fixture or adaptation was edited. The complete
integration gate, native compiler execution and cron installation were not
part of this meter-only change.


## Latent lowering meter follow-up

Branched from origin/area/stage3 e39a299323cad76aea44ec71ca7b740130323d4c.
Only stage3/meter changed. The twice-daily script builds the existing latent
census with its scratch Go overlay, then runs it after each tree's ordinary
checker census with LATENT_ASSERT_NO_OUTPUT=1. No production compiler files
or latent tool files were edited, and no backend is invoked.

Fresh run: runs/20261007T062959Z.tb10Z0/report.md and report.json.
Main b8fb957aa839a9e8cb0b54279dd9864fa317bd30: checker whole 0/78,
own 22/78; latent NotYet 2059, Refused 2140.
Area e39a299323cad76aea44ec71ca7b740130323d4c: checker whole 1/79,
own 25/79; latent NotYet 2075, Refused 2104.
Both latent ledgers are labeled "measured on a checker-rejected program".
Each has four SkippedDependency sites, zero errors and zero panics.
Top ten reasons follow the checker table. JSON retains full reason counts and
adds latent_lowering to each tree; the root retains the area observation.
Counts use unique (kind, where, reason, text) sites, matching the latent tool.

Commands, each with output redirected to its own log:

```sh
bash cloud/setup.sh > /tmp/meter-lowering-setup.log 2>&1
source /workspace/adamic-tools/env.sh
bash stage3/meter/twice-daily.sh > /tmp/meter-lowering-run.log 2>&1
CENSUS_BINARY=/tmp/adamic-gate/stage3-meter.JeZTWq/census LATENT_CENSUS_BINARY=/tmp/adamic-gate/stage3-meter.JeZTWq/latent-census python3 -m unittest discover -s stage3/meter -p '*_test.py' > /tmp/meter-lowering-tests.log 2>&1
go vet ./stage3/census/tool ./stage3/census/latent/tool > /tmp/meter-lowering-vet.log 2>&1
bash -n stage3/meter/twice-daily.sh
git diff --check
```

Both complete checker/latent phases pass, including the no-IR and disabled
production-loader guards on both real trees. Nineteen tests pass with no skips.
Vet, shell syntax and whitespace checks pass. Setup: Go/clang/Node ready at 0s;
submodules at 18s; cache warm and done at 260s; nproc 5, cgroup quota
400000/100000, 17.6 GB. Toolchain remains Go 1.27.1, clang 20.1.8, Node 24.19.0.

The preservation audit imports report.py from the base commit, renders the
same two raw checker censuses, and compares every existing JSON field and the
entire checker headline/table prefix against the new report. They match exactly.
The latent headers' full diagnostics also equal the ordinary whole-program
checker diagnostics exactly on both trees. See checker-preservation.log.

The real planted-NotYet test starts with a rejected function, an eligible
function with an existing NotYet, an eligible non-null assertion producing
Refused, and an eligible target function. LATENT_MUTANT_FUNCTION=target plants
one extra overlay-only NotYet. Only "NotYet: latent planted extra NotYet"
increases, by one; every other reason and Refused total stays fixed.
A synthetic duplicate event also proves headline deduplication.
A scratch report.py mutant attributes every finding to "existing". Both named
planted-NotYet tests fail only their reason-count assertions, without setup
errors. The live report code and compiler remain unchanged by the mutant.
See tests.log and mutants.log for the passing baseline and expected catches.

Limits: observations on checker-rejected programs are not successful lowering.
Diagnosed function bodies and dependencies can be skipped; each eligible unit
can stop at its first lowering error. Final module order, ownership and backend
passes remain outside this measurement. The full integration gate and native
execution were not run for this meter-only change. No cron or messaging was
installed, and only the named meter branch is pushed.


Ownership addition: owners.json contains the supplied worker/branch assignments.
Exact reason text wins over the longest prefix; the explicit "seen as" entry
matches the phrase inside variance reasons. Unmatched rows print OWNER BLANK
and are listed first under Unowned, grouped by exact reason across both trees.
The first ten are visible; a collapsible table retains every remaining row.
Top-ten tables are grouped by exact reason, with owner and separate kind counts.
Owner grouping changes no checker number, latent total or exact reason count.
The new unknown-owner test checks OWNER BLANK and placement under Unowned;
a separate paired-report assertion checks Unowned precedes the latent tables.
An independent scratch mutant claims an unmatched owner, and that test catches
it only through the owner assertion. Prefix, variance and cross-kind grouping
checks also pass. The ownership-aware suite passes all nineteen tests, no skips.


Unowned refinement: the Markdown table now shows only reasons with at least
10 unique sites on either tree, ordered by the larger per-tree total. The final
line gives the number of lower-count reasons and their sites summed across
both trees. JSON retains every unowned reason. This recorded run has
26 displayed unowned rows. Existing checker fields, the checker table,
latent totals and exact reason counts are unchanged; recorded compiler
provenance is retained while rerendering the completed measurements.

Three >=30-site families were assigned: condition handling to
codex/taste-not-soundness (the pinned taste feature), generic returns to
01a1143c (the return family at latent REPORT.md:537 and :542), and
reading CharacterCodes to compiler/stage3-front (the enum family at :1141).
BinaryExpression with value operands, values of type T and values of type any
remain OWNER BLANK because the report does not establish their worker/branch
ownership. The remaining large families are not assigned by guesswork.

All twenty existing/new tests pass with both real census binaries enabled,
log refinement-tests.log. The threshold test covers 9 sites on each tree
(still omitted), 10 on either tree, a 6 NotYet plus 4 Refused total, sorting,
and the exact tail summary; the full JSON list remains intact. A scratch
>=9 threshold mutant fails only that test's row-exclusion assertion, with
trace in threshold-mutant.log. No census rerun or compiler change was needed
for this presentation and ownership refinement.
