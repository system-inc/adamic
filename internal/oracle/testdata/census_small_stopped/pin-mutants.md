# Refusal pin mutants

Test: TestCensusSmallDynamicReceiverRefusalPin. All four new pins pass with the overlay inactive.

The external Go overlay changes lower.notYet only when its computed Where
contains census_small_optional_stopped.a:11:. The reason mutant replaces What with
"mutant changed refusal reason"; the location mutant preserves What and
replaces the line with 999, preserving the file and column. The same compiled
test binary runs all four pins for every mutant. Each mutant exits 1, with
exactly this pin failing and the other three passing. No production file is
changed in the commit.

## Reason mutant

Log: /tmp/5jnptpn-mutant-census_small_stopped-reason.log

Output line:

    census_small_test.go:70: want located refusal "a function expression with a this parameter (dynamic receivers are not implemented)" at census_small_optional_stopped.a:11:: /workspace/oracle-refusal-pins/internal/oracle/testdata/census_small_stopped/census_small_optional_stopped.a:11:22: stage 0 can't lower mutant changed refusal reason yet

## Location mutant

Log: /tmp/5jnptpn-mutant-census_small_stopped-location.log

Output line:

    census_small_test.go:70: want located refusal "a function expression with a this parameter (dynamic receivers are not implemented)" at census_small_optional_stopped.a:11:: /workspace/oracle-refusal-pins/internal/oracle/testdata/census_small_stopped/census_small_optional_stopped.a:999:22: stage 0 can't lower a function expression with a this parameter (dynamic receivers are not implemented) yet

Reproduction details and shared overlay are in ../oracle-refusal-pin-report.md.
