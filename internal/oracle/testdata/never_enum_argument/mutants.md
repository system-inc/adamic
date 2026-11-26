# Mutants of enumNeverTrapped's never argument (#xfpsqtn, 1)

Each mutant replaces exactly one line, checked to match once, runs `go test -run TestNeverEnumArgument
./internal/oracle/`, and is put back from a copy of the head file, checked byte for byte. Each is red beside the
same pins green at head.

| Mutant | Line replaced | Red |
|---|---|---|
| M1 | `checkNeverArguments` (never_parameter.go): the acceptance `if l.enumNeverTrapped(argument) {` made `if false {` | TestNeverEnumArgumentAgreesWithNode/binder, reverse, preferred, statistic, two_paths and TestNeverEnumArgumentTrapsBeforeTheCall's six, each with today's stop |
| M2 | `enumNeverTrapped` (never_read.go): the `&& l.enumNeverReadGap(node) == nil` test dropped | TestNeverEnumArgumentStops/write_between, whose stop moves to value()'s gap words |
| M3 | `enumNeverCheck` (enum_never.go): the call of the trap's helper returned as the value it was given, `return value` | TestNeverEnumArgumentTrapsBeforeTheCall's six: the read passes through, and assertNever's own throw is what stops |

M2 shows the gap test is in the helper for the stop's words, not its safety: without it the argument still
stops, at value()'s enumNeverReadGap, which is the same judgment made again.
