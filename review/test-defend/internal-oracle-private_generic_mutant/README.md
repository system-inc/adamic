# Regex native dependency defense

Starting origin/main: 7b890befc8fc43fa57435c0959510d13349dd31c. Fetched the audit with its full remote refspec. Read REPORT.md, rows.json, friction-and-limits.md, summary.json, mutant-plan.json and S01.diff; the prior S01 edits a test helper to return false, which this defense forbids. The requested name still exists at internal/oracle/regexp_cycle_test.go:37. Current scope lists 198 top-level test functions.

Verdict: cannot-judge within the permitted production-mutation rules. No eligible mutant was planted and no unique failure is claimed. There are no standalone mutant diffs to replay. Removing the IR field was tested only as a compile-feasibility probe: lower/native still reference Regexps, so the test package cannot build. That probe is explicitly excluded from attempts and kills, and its scratch overlay was never applied to repository source. Neither a compile failure nor the old audit's skip proves an assertion catch.

The clean whole package cooked at 90.241 seconds. Before timeout, this row passed and no test-failure event or assertion error was recorded. Its isolated clean coverage run passed in 0.017 seconds. It executed zero production IR statements. Five registered regex fixtures, run through TestNativeAgreesWithNode under the same coverage target, passed in 1.623 seconds with 37.6% coverage. Thus the requested row has no exclusive covered production block; coverage-difference.json lists both sets. No claim of whole-package dynamic coverage is made. The whole-package timeout does not justify a deletion.

## Owner finding

The name promises that the regex cycle fixture has its native dependency. The body only asks whether a static IR struct has a field named Regexps. If not, it skips successfully; if present, it returns without any assertion. It does not check field contents, native lowering, emission, runtime linkage, or registration of the five fixture paths. The skip message still says native regex lowering/emission is on codex/stage1-css, although current main has the representation and the five fixtures execute successfully. This row currently functions as an optional dependency gate rather than a failure assertion.

## Brief feedback and time costs

- The 15 GB threshold is impossible on the 8.8 GB /tmp filesystem. /workspace is a different filesystem. Deleted only earlier named /tmp/defend-estree-scalars, /tmp/defend-estree-scalars-tmp and /tmp/u010-stage1-progress directories, never repository or tools. /tmp free space increased from 6.0 GB to 6.4 GB; /workspace began and remained at 7.5 GB after cleanup.
- The audit treated this as a construction row and allowed S01's harness edit. This defense forbids that very edit. Its actual predicate is entirely in _test.go; no production-function mutation can exercise its skip policy.
- Static type metadata has no Go statement coverage. Zero production counters do not imply the reflection helper was not executed. The helper is excluded from -coverpkg instrumentation because it is a test function.
- The permitted menu does not include a struct-field rename/removal, and removing the field also prevents the compiler from building. The logged layout experiment is feasibility evidence, not a compliant mutation or kill. Three unrelated regex production mutants would not be honest aimed attempts at this row.
- A whole-package coverage/matrix cannot fit 90 seconds. Compared the row with its five registered native regex fixtures instead; outside that subset remains unmeasured. Since the row's production coverage set is empty, it cannot acquire an exclusive production statement by broadening the comparison set.
- An initial fixture selector placed whole slash-separated paths inside one alternation. It selected no child tests. That failed selection and its zero coverage are retained; the corrected component-wise selector selected all five fixtures. Only the corrected profile is used for comparison.
- There is no failing test line to report. Inventing one from the compiler's field-removal build error would misstate the evidence.

Warm toolchain worked, so setup was skipped. nproc=5. npm ci in stage3/api ran before baseline; its own reported time is in npm-ci.log.gz. Total work was about 9 minutes, including the 90-second baseline, coverage compilation, and reading. No production source, tests, or harness were changed. No other-package test run, native semantic mutation matrix, deletion/rewrite, main push, or pull request occurred. Logs are retained losslessly as .log.gz files.
