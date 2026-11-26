# Mutants of enumNeverTrapped's never argument (#xfpsqtn, 1)

Each mutant replaces exactly one line, checked to match once, runs `go test -run TestNeverEnumArgument
./internal/oracle/`, and is put back from a copy of the head file, checked byte for byte. Each is red beside the
same pins green at head.

| Mutant | Line replaced | Red |
|---|---|---|
| M1 | `checkNeverArguments` (never_parameter.go): the acceptance `if l.enumNeverTrapped(argument) {` made `if false {` | TestNeverEnumArgumentAgreesWithNode/binder, reverse, preferred, statistic, two_paths and TestNeverEnumArgumentTrapsBeforeTheCall/binder_out, reverse_out, preferred_out, statistic_out, two_paths_out_switch, two_paths_out_if, each with today's stop at its assertNever argument |
| M2 | `enumNeverTrapped` (never_read.go): the `&& l.enumNeverReadGap(node) == nil` test dropped | TestNeverEnumArgumentStops/write_between, whose stop at 20:21 moves to value()'s gap words, "a read the checker narrowed to never across code that may write it" |
| M3 | `enumNeverCheck` (enum_never.go): the call of the trap's helper returned as the value it was given, `return value` | TestNeverEnumArgumentTrapsBeforeTheCall/binder_out, reverse_out, preferred_out, statistic_out, two_paths_out_switch, two_paths_out_if; the in-range rows stay green |

M2 shows the gap test is in the helper for the stop's words, not its safety: without it the argument still
stops, at value()'s enumNeverReadGap, which is the same judgment made again.

Measured by Loom's proof job J-m375sc6 at 4ba0792c (code 176e57a4): the pins green at head, M1 to M3 each red as
listed, each file put back and checked byte for byte. The job kept M3's subtest names but not its failure words, so
which of stdout, stderr or the exit moved under it isn't recorded here.
