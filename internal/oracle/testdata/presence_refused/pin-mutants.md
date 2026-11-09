# Refusal pin mutants

Test: TestAbsentOptionalOwnFieldWriteRefusalPin. All four new pins pass with the overlay inactive.

The external Go overlay changes lower.notYet only when its computed Where
contains optional_field_write.a:9:. The reason mutant replaces What with
"mutant changed refusal reason"; the location mutant preserves What and
replaces the line with 999, preserving the file and column. The same compiled
test binary runs all four pins for every mutant. Each mutant exits 1, with
exactly this pin failing and the other three passing. No production file is
changed in the commit.

## Reason mutant

Log: /tmp/5jnptpn-mutant-presence_refused-reason.log

Output line:

    presence_refused_test.go:61: want located refusal "writing a possibly absent optional own field" at optional_field_write.a:9:: /workspace/oracle-refusal-pins/internal/oracle/testdata/presence_refused/optional_field_write.a:9:1: stage 0 can't lower mutant changed refusal reason yet

## Location mutant

Log: /tmp/5jnptpn-mutant-presence_refused-location.log

Output line:

    presence_refused_test.go:61: want located refusal "writing a possibly absent optional own field" at optional_field_write.a:9:: /workspace/oracle-refusal-pins/internal/oracle/testdata/presence_refused/optional_field_write.a:999:1: stage 0 can't lower writing a possibly absent optional own field yet

Reproduction details and shared overlay are in ../oracle-refusal-pin-report.md.
