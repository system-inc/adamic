# Refusal pin mutants

Test: TestRecordR24StringConversionRefusalPin. All four new pins pass with the overlay inactive.

The external Go overlay changes lower.notYet only when its computed Where
contains r24.a:4:. The reason mutant replaces What with
"mutant changed refusal reason"; the location mutant preserves What and
replaces the line with 999, preserving the file and column. The same compiled
test binary runs all four pins for every mutant. Each mutant exits 1, with
exactly this pin failing and the other three passing. No production file is
changed in the commit.

## Reason mutant

Log: /tmp/5jnptpn-mutant-records_refuse-reason.log

Output line:

    records_test.go:235: want located refusal "String conversion of an object, array, map or function (ToPrimitive is not lowered)" at r24.a:4:: /workspace/oracle-refusal-pins/internal/oracle/testdata/records_refuse/r24.a:4:20: stage 0 can't lower mutant changed refusal reason yet

## Location mutant

Log: /tmp/5jnptpn-mutant-records_refuse-location.log

Output line:

    records_test.go:235: want located refusal "String conversion of an object, array, map or function (ToPrimitive is not lowered)" at r24.a:4:: /workspace/oracle-refusal-pins/internal/oracle/testdata/records_refuse/r24.a:999:20: stage 0 can't lower String conversion of an object, array, map or function (ToPrimitive is not lowered) yet

Reproduction details and shared overlay are in ../oracle-refusal-pin-report.md.
