# Refusal pin mutants

Test: TestComputedFieldsWriteRefusalPin. All four new pins pass with the overlay inactive.

The external Go overlay changes lower.notYet only when its computed Where
contains fields.a:12:. The reason mutant replaces What with
"mutant changed refusal reason"; the location mutant preserves What and
replaces the line with 999, preserving the file and column. The same compiled
test binary runs all four pins for every mutant. Each mutant exits 1, with
exactly this pin failing and the other three passing. No production file is
changed in the commit.

## Reason mutant

Log: /tmp/5jnptpn-mutant-notyet_element_access-reason.log

Output line:

    element_access_fields_test.go:33: want located refusal "assigning an element of a value" at fields.a:12:: /workspace/oracle-refusal-pins/internal/oracle/testdata/notyet_element_access/fields.a:12:5: stage 0 can't lower mutant changed refusal reason yet

## Location mutant

Log: /tmp/5jnptpn-mutant-notyet_element_access-location.log

Output line:

    element_access_fields_test.go:33: want located refusal "assigning an element of a value" at fields.a:12:: /workspace/oracle-refusal-pins/internal/oracle/testdata/notyet_element_access/fields.a:999:5: stage 0 can't lower assigning an element of a value yet

Reproduction details and shared overlay are in ../oracle-refusal-pin-report.md.
