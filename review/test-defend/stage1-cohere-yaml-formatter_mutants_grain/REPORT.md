# YAML formatter-mutant family defense

Defense: cannot-judge under the permitted mutation scope. This is not a production-mutant survival finding, a three-attempt not-defended verdict, or a deletion recommendation. No test, harness, oracle, source fixture or production source was altered. No production mutation matrix or unique catch is claimed.

Starting origin/main: bfe0553300773c0b37db2c10df97adeb909705f8. Prior audit and full current test list are preserved. All six target members still exist in formatter_mutants_grain_test.go. Current package has 72 top-level Test names, the same count as the prior audit. The whole package was selected for the baseline, including every current name, but not every row finished before timeout; complete outside-family outcomes are unknown. The prior audit's requested 26 names are unchanged.

## Code under test and oracle

CODE UNDER TEST, as named in the audit: the family's own construction and formatterMutantSurvived survivor guard, not Go cohere or ordinary formatter correctness. The guard is defined in stage1/cohere/yaml/formatter_mutants_grain_test.go:245. The execution family calls it at line 285; the only other call is TestFormatterMutantsPlantedFailure at line 489. All are test/harness code. The family compiles and executes six deliberately altered copies of the production TypeScript formatter, but their differences are the family's inputs.

ORACLE: Go cohere actually formats the corpus and supplies expected bytes. The family requires each built-in mutant's native and Node output stream to differ from those bytes somewhere, with successful execution and no unexpected stderr. That is an aggregate external-run inequality oracle. The separate planted-survivor row uses synthetic self-written bytes.

The audit W1 deleted the survivor condition and returned nil. Its six witnesses passed and its separate planted-survivor test failed. W1 is a harness weakening, not a production mutant. Its original diff and report are copied with prior- prefixes. The audit's witness rule explicitly allowed that harness edit. This defense brief explicitly says never mutate the oracle, harness or test, without an exception for witnesses. Repeating W1 would violate that rule. Neither a broken formatter nor a compiler/preparation error proves that a witness notices removal of its guard. A production change that makes one built-in source mutant equivalent could make the family fail, but does not establish a broken working behavior that uniquely guards the audited comparison. Such a result would require separate witness-specific criteria and cannot silently replace the requested proof.

## Coverage and current-session observations

The requested family command was run with -coverpkg=github.com/system-inc/adamic/stage1/cohere/yaml. Its six members passed in 89.838 binary seconds. The only other guard caller, TestFormatterMutantsPlantedFailure, also passed under that coverage option in 0.009 seconds. Both profiles contain only mode: set, and both outputs report coverage: [no statements]. Go production coverage cannot instrument the guard in a _test.go file. coverage-diffs.json therefore marks exclusive lines unmeasurable, not absent. It is not a zero-exclusive-coverage claim. No native or V8 port profile substitutes for coverage of this Go guard, and no test was moved into production to instrument it.

Guard observation command: go run /workspace/yaml-guard-observation.go > guard-observation.log 2>&1. guard-observation.go.txt contains a byte-for-byte copy of the existing function, extracted without modifying its repository definition, plus a separate main that passes three synthetic streams. Results:

identical: missed mutant
unrelated wrong answer: <nil>
empty answer: <nil>

This directly observes the guard's accepted domain. It is not a mutation, not a full native-port empty-answer run, not a claim that an empty executable satisfies the command wrapper, and not a basis for a unique-kill verdict.

## Baseline, budget and narrowing

Warm /workspace/adamic-tools/env.sh worked, so setup was skipped. npm ci --prefix stage3/api ran before the baseline. nproc: 5. Whole-package baseline command: timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/yaml/ -run . > baseline.log 2>&1. The binary timed out at its own 90s limit, with package fail elapsed 90.705s and panic: test timed out after 1m30s. This is an incomplete over-budget baseline, not a recorded semantic assertion failure. It was not extended. The unchanged selected family was then narrowed and passed. Coverage work briefly overlapped the end of the whole baseline; no source edits existed during either run, but contention limits interpretation of timing.

The whole baseline recorded TestComposeMatchGo skipping its optional library oracle before timeout. Full skip inventory and remaining package outcomes were not obtained. The target family has no optional YAML/Prettier library gate; its narrowed pass runs its native and source Node variants against Go cohere. All bounded outcomes and raw logs are provided. No other packages ran. No native build-only time is separated from the 89.838-second family run. No fresh mutant build caches were needed because no compiler or port was mutated.

## Brief friction and owner finding

1. The mutation target identified by the audit is explicitly prohibited by this defense's no-harness/no-test-edit rule. The audit had an explicit witness exception; the current brief does not. This prevents a faithful comparison-removal defense. The twin rule is not relevant: native and Node are two sides inside each same family member, not separate twin rows.
2. Go coverage does not instrument the audited test-only check. Supplying ordinary package coverage cannot prove exclusivity here. The only other direct guard caller was measured separately; unrelated package rows cannot reach this unexported test helper by production calls.
3. The package's whole baseline exceeded the test-binary budget; narrowing was necessary. The chosen family narrowly fits, so its timing should not be treated as a cheap witness.
4. The prior untrue label is a weakened-check result with zero production mutants. Treating it as evidence that no production mutant can fail these tests would overstate the audit. It also differs from production subsumption, so there is no named subsumer to compare.

Name/assertion finding: TestFormatterMutants does assert successful native and Node execution and disagreement with Go cohere. It does not require the intended affected case to differ, nor require unaffected cases to agree. Any unrelated wrong stream passes its survivor guard, as observed above. It does not itself assert that the guard rejects a deliberately surviving mutant; that assertion belongs to the separate TestFormatterMutantsPlantedFailure. These scope limits should be considered by the owner, but do not prove that deleting or rewriting the execution family is safe. No such change was made.
