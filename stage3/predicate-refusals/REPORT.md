Built narrow helper-summary and array-storage refusals, fixed optional-chain directions, and pinned both existing proof guards toward roadmap step 09 (#fxspptb).
Commits: 2c2557125a3ea6b5949b6c04af94da33250e0c63 (unit) and d38ec576a3cd796abca300995ffd5963e776df7c (one-line sanitizer portability fix), based on main 031a1259.
Final focused lower and oracle tests passed; all eleven programs held to Node; counts refresh passed and added three rows.
Six revert mutants failed their exact fixtures: helper index, filter representation, find representation, optional-chain container, missing else, parameter write.
Limits: boxed array conversion is still pending; the existing overload contract makes p14/p27 stop loudly where Node continues; dead interface-construction code is unchanged; macOS was not executed.

# Predicate refusals

Delivery branch: `compiler/miscompile-predicates`. Label: predicate-refusals. This serves roadmap step 09 by stopping the step 04 miscompiles before unsafe representation reaches the backends. Main was fetched again before delivery and remained `031a1259bc7973934792dc6cb1bd4074fc2204b9`. The branch contains only own commits on that main tip; no other worker's branch was imported.

## Changes

`predicateVerifier.expression` now checks the helper declaration's actual predicate ParameterIndex before using its summary. The existing supported call shape has one argument at index 0; a summary for any other parameter fails with the helper call path and a fix to pass the tested value at the predicate parameter index. The outer refusal pass preserves its existing diagnostic and caller fix; a separate direct-proof test asserts the more specific helper diagnostic.

`arrayVisit` refuses a predicate filter from boxed union slots to a different element representation as NotYet. Predicate find refuses incompatible result storage as NotYet before constructing unsafe backend IR. Both messages name the source path and suggest an explicit narrowed copy/result in a loop. Boolean callbacks with no narrowing claim and compatible representations remain admitted. Finding undefined in packed optional-number storage is explicitly retained: the initial counts refresh exposed that correct case, it was added to the control fixture, and the existing maybe_number_slots.a also passes against Node. The final counts refresh is green.

`predicateUseDirections` includes property/element optional-chain containers as separate references, using the same pinned-checker flow comparison as the argument itself. The p14 container read now consumes the true direction, which carries the named overload check. The false direction remains unobservable. No IR, backend, native runtime, lower.go, emit.go, native.go, or oracle_test.go was edited.

The existing missing-else and overload-parameter-write guards needed fixtures, not new compiler guards. p26 stays refused. p27 retains its named call-site check. Assumption: the unit's explicit requirement to pin loss of the p27 call-site check means the existing overloaded-predicate checked-failure contract remains in force for p14/p27. Their full Node output is asserted independently, and all three Adamic modes must produce the exact named exit-70 diagnostic before the false claim is consumed. These checked failures are not described as agreement with Node or as compile-time refusals.

## Every program: before and after

All eleven supplied programs are pinned verbatim under internal/oracle/testdata/predicate_refusals/, with their supplied oct8_predicates_ filenames. The IDs below identify those filenames. All before results were reproduced with the unmodified compiler on the pinned main; all Node outputs had exit 0 and empty stderr. The baseline run deliberately records miscompiles and fails the three native build cases. Sanitized execution was not reached for those build failures.

| Program | Node output | Observed before | After |
|---|---|---|---|
| p05_helper_second_parameter | number NaN; number 42 | JS/release/sanitized exit 70 with generic union backstop before output | Refused at 5:3: unproven predicate return; direct summary fails at helper call 5:10 with parameter-index fix |
| p06_helper_second_parameter_undefined | undefined? abc; no; undefined? undefined; yes | All Adamic modes print undefined? undefined for the first line, exit 0 | Refused at 5:3; direct helper proof fails at 5:10 |
| p12_find_undefined_target | found undefined | JS agrees; release clang rejects assignment from adamic_string* to adamic_object* | NotYet at 5:15: find predicate result representation conversion |
| p13_filter_wrong_parameter_helper | 3; item a; item undefined; item bc | All Adamic modes print item undefined three times, exit 0 | Refused at 8:27: unproven predicate argument g; direct helper proof fails at 5:10 |
| p14_overload_optional_chain_container | claimed box: box a; direct: box; claimed box: undefined; direct: undefined | All Adamic modes print first two lines then generic undefined backstop exit 70; true direction unobservable | All modes print first two lines then named overload 1 of isText predicate x check, exit 70; true checked, false unobservable |
| p21_find_inferred_number | found 2; a\|bc | JS agrees; release clang rejects adamic_heap* assigned to adamic_maybe_number | NotYet at 2:15: find predicate result representation conversion |
| p22_filter_inferred_narrowing | a\|bc; 1,2; total 6 | JS agrees; release total 9.3578556356678e-310; sanitized total 1.348956142254303e-309; both exit 0, empty stderr | NotYet at 2:17: filter predicate element representation conversion |
| p23_find_declared_predicate | found 2 | JS agrees; release clang rejects adamic_heap* assigned to adamic_maybe_number | NotYet at 5:15: find predicate result representation conversion |
| p24_filter_proven_number | 2; 2; 2,4 | JS agrees; release second line 4.64848914938803e-310; sanitized 6.70460608812573e-310; both joins also garbage, exit 0, empty stderr | NotYet at 5:17: filter predicate element representation conversion |
| p26_assert_nested_if_without_else | value abc; value undefined | Already Refused at 3:5 | Same refusal pinned; missing-else mutant incorrectly admits it |
| p27_overload_parameter_rebound | text a; text 41 | All Adamic modes print text a then named overload result check, exit 70 | Same exact check pinned; both true and false directions checked |

Representation-conversion fixes remain later work. The float values are the observed outputs, not deterministic golden values. Exact before/after output and diagnostics are saved in logs/miscompile-predicates-before.log.gz and the final fixture logs.

The executable control fixture also agrees with Node in JS, release and sanitized native, and passes the oracle leak check. It covers a helper whose tested value occupies its true parameter at index 0 even with a second optional parameter; optional-number filter and find, including find(undefined); and a boolean filter over a boxed union without a narrowing claim.

## Every mutant

Run `source /workspace/adamic-tools/env.sh` then `python3 stage3/predicate-refusals/mutants.py`, with output redirected to a log. The runner operates serially, restores each source in finally, verifies a named fixture actually failed, and rejects a build-only failure. All six final mutants returned test exit 1. Compiler sources were restored and the complete targeted fixture set rerun afterward.

| Mutant | Selector under TestPredicateMiscompileRefusals | What caught it |
|---|---|---|
| 1-helper-parameter-index: remove helper slot guard | p(05\|06\|13)_ | Expected path-bearing Refused becomes nil |
| 2-filter-union-representation: bypass union conversion refusal | p(22\|24)_ | Expected filter NotYet becomes nil; before run independently demonstrates the native garbage |
| 3-find-result-representation: bypass find conversion refusal | p(12\|21\|23)_ | p12/p23 expected NotYet becomes nil, before any clang call; p21 reaches the next filter stop, so its exact find diagnostic assertion fails |
| 4-optional-chain-container: omit containers from observed references | p14_ | Named overload check is replaced by generic undefined backstop; direction report loses its checked true direction |
| 5a-missing-else: treat missing else as terminating | p26_ | Expected refusal becomes nil |
| 5b-parameter-write: ignore writes to overload parameter | p27_ | Named check disappears; generic union backstop replaces it and explain metadata incorrectly calls both directions proven |

The find mutant is caught at lowering rather than by -Werror, as required. The p27 assertion distinguishes the named check from any generic runtime abort; checking only exit 70 would miss this mutant.

## Commands and outputs

All test output was redirected to files, never piped from a running test. Adjacent logs are compressed without altering their contents.

Toolchain: `export GOPROXY='https://proxy.golang.org|direct'`; `bash cloud/setup.sh`; `source /workspace/adamic-tools/env.sh`. Setup passed: node 0.022s, go 0.032s, submodules 0.057s, markdown ready 0.082s (step 0.010s), clang 0.212s, Go build 42.568s, deferred test binaries 42.717s, warm cache 42.718s, done 42.743s. nproc 5; cgroup CPU quota 4. Setup required no workaround.

```sh
PREDICATE_MISCOMPILE_BASELINE=1 go test ./internal/oracle -run '^TestPredicateMiscompileRefusals$' -count=1 -v
go test ./internal/lower -run '^(TestPredicateSummaryParameterIndex|TestPredicateBodiesAreProven|TestPredicateBodyProof|TestPredicateCallbackContracts|TestPredicateOverloadCallback|TestPredicateUseRegions|TestPredicateUsesBelongToEachCall|TestPredicateOverloadRuntime|TestIndirectPredicateOverloadIsPending)$' -count=1 -v
go test ./internal/oracle -run '^TestPredicateMiscompileRefusals$|^TestNativeAgreesWithNode$/internal/oracle/testdata/predicate_refusals/' -count=1 -v
python3 stage3/predicate-refusals/mutants.py
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -args -update-counts
go run ./cmd/adamic c internal/oracle/testdata/predicate_refusals/oct8_predicates_p14_overload_optional_chain_container.a --explain-checks
go test ./internal/lower -run '^TestPredicateOverloadRuntime$' -count=1 -v
```

Final lower tests PASS 5.025s. Final oracle fixtures PASS 0.704s. Final counts refresh PASS 39.149s and changes only three new executable rows. The previous refresh failed only on the compatible optional-number find case and drove the narrower rule; that failure log is retained. The narrowed-control run additionally held the existing maybe_number_slots.a to Node, PASS 0.849s. The one-line sanitizer-default test run PASS 7.916s on Linux. No whole package or full gate was run.

Actual p14 --explain-checks output:

```text
true: checked; a narrowed read consumes this direction
false: unobservable (proven); no narrowed read in the false region
adamic: predicate checks: proven 1 checked 1 unobservable 1
```

## Out of scope

Removed the single line forcing ASAN_OPTIONS=detect_leaks=1 in predicates_overload_test.go, in its own commit d38ec576. The test now inherits platform sanitizer defaults. Linux runtime tests pass; no macOS runner was available, so macOS execution is not claimed. Dead interface_cast.go construction-enforcement code and its comments require more than a one-line cleanup and remain untouched. No runtime helpers were added.
