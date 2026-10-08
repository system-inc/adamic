368e0634: keep only c41c0e06, recheck every callClosure table site, and preserve the any-return boundary.
Commits: 27ab3b85 merges c41 only; 368e0634 fixes negative fixture placement and adds the final ranked boundary; both pushed.
Validation: oracle, focused lower, single-assignment flow and counts refresh PASS; five positive fixtures match Node in both backends with sanitized native.
Mutants: all ten final mutants fail for their intended assertions, including the new guessed-any-ABI mutant.
Limits: 24 sites still encounter earlier rollback boundaries; T and any remain named stops; callable-field declarations still require the other worker's parameter hook.

The withdrawn ffe428ab views merge was aborted before any commit or push. The branch has c41c0e06 as an ancestor and does not have ffe428ab as an ancestor. No force-push or revert was needed. The c41 merge had one additive counts.md conflict; both sets of rows were retained. Refreshing counts subsequently reordered those rows without changing any numeric count.

The October 8 ruling governs: `!` stays refused in `.a`. c41's existing controls confirm that refusal. Its checked `.ts` behavior remains distinct. No new production lowering or runtime change was made in this group. The earlier views-conflict approval requests no longer apply because that merge was withdrawn.

All 65 distinct owned-function signatures from the pinned table raw CSV were replayed, not just the original example pairs. [Per-site observations](c41-rechecks.csv) record the next stops.

| Table reason | Sites | Statement survived replay | Earlier rollback boundary | Exact target stop remains |
| --- | ---: | ---: | ---: | ---: |
| a call returning void \| undefined | 40 | 35 | 5 | 0 |
| passing union of differently held members to a function value | 23 | 4 | 19 | 0 |
| a call returning T | 1 | 0 | 0 | 1 |
| a call returning any | 1 | 0 | 0 | 1 |

This certifies 39 statements surviving the guarded standalone measurement replay, not 39 complete programs or a decrease measured by a full census. A site is counted as surviving only when its source offset lies inside an attempted body's range, it is outside all rollback boundaries, and no finding remains unpaired with a rollback boundary. Those observations are on a checker-rejected entry-root program. Earlier findings elsewhere in the selected unit remain visible. The absence of the requested signature alone is insufficient to certify a statement.

Binder 587:13 survives. Binder 585:13 still meets `a value of type Path` at 585:73. Checker 3467:42 and 3494:42 remain blocked by `__String` at 3464:78 and 3492:74. WatchUtilities 540:22 reproduces `a call returning T`. Sys 378:63 reproduces `a call returning any`.

The final ranked kind is the timer host's any result. Its new `.a` fixture calls the callback and prints `tick` on Node, while lowering preserves `a call returning any`. `docs/0.1.md` says any has no proven type and requires unknown plus narrowing. No guessed return ABI was accepted. The mutant assigning Number to that any result compiles as a mutant, but fails TestCallableAnyResultStaysNotYet with `want the unrepresented any return stop`. Native and JavaScript execution are intentionally unavailable for this negative control; accepting it is not certified.

The c41 flow test exposed a pre-existing issue in this unit: deliberate NotYet controls in testdata's top level were picked up as runnable flow inputs. Move only this unit's four negative controls into `internal/oracle/testdata/notyet_call_boundaries/`, and put the new any control there too. Own test registrations and mutant selectors still assert every boundary. The flow test then passes without an exemption or any edit to flow_test.go.

Commands and outputs, with test output redirected to logs:

```bash
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/notyet_|TestOptionalVoidValueStillNotYet|TestUnionCallableSourceObservations|TestPossibleNonNullAdamicAssertionsAreRefused|TestImpossibleNonNullFixturesAreRefused|TestCheckedNonNullCounts' -count=1 -v > /tmp/notyet-c41-group-tests.log 2>&1
go test ./internal/lower -run 'TestNonNullAssertion|TestLibraryMethodValue|TestCensusOptional' -count=1 > /tmp/notyet-c41-lower.log 2>&1
go test ./internal/flow -run TestEveryFunctionIsInSingleAssignment -count=1 > /tmp/notyet-c41-flow-restored.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/notyet_|TestOptionalVoidValueStillNotYet|TestUnionCallableSourceObservations|TestCallableAnyResultStaysNotYet' -count=1 -v > /tmp/notyet-c41-boundaries.log 2>&1
python3 internal/oracle/testdata/run-notyet-call-mutants.py > /tmp/notyet-c41-final-mutants.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts > /tmp/notyet-c41-final-counts.log 2>&1
```

Group oracle PASS 1.704s, lower PASS 0.727s, restored flow PASS 18.227s, final oracle PASS 1.185s (native hits 0/misses 15, Node hits 0/misses 16), counts PASS 25.372s. The initial flow run failed on the Location negative fixture, as recorded in /tmp/notyet-c41-flow.log; it was repaired by relocating controls, not weakening lowering. No full package or gate was run.

The nine existing mutants retain the catchers listed in REPORT.md. The tenth, accept-any-return, fails the explicit any-return test. All ten were rerun after relocation, exited nonzero for their intended assertion, and production sources were restored. The runner rejects compiler build failures as mutation evidence.

Replay command is unchanged from REPORT.md. The all-site runner is saved at /tmp/notyet-c41-rechecks/run.py, its log is run.log, and its complete per-site JSON/log pairs are 0 through 64 in that directory. It reads /tmp/notyet-void-roots.csv, deduplicates where/reason for the four exact kinds, and saves results.json. classified.json records conservative classifications; c41-rechecks.csv is its repository summary. Compiler and adapted-source pins remain as recorded in REPORT.md.

Remaining ownership limit: `functions.go:signature` has the union callable parameter guard, changed by codex/notyet-generic-returns-t in fd0e1225. It remains untouched. This blocks the two original callable-field fixtures at `a function value taking Name/Location`. The boxed method adapters already pass. The ranked callClosure list is otherwise exhausted; unresolved T needs concrete specialization context, and any remains outside the language contract.
