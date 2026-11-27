# Mutants of enumStoredType's tuple element (#xfpsqtn, 2)

Each mutant replaces exactly one line of internal/lower/enum_never.go, checked to match once, runs `go test -run
TestNeverEnumTuple ./internal/oracle/`, and is put back from a copy of the head file, checked byte for byte. Each is
red beside the same pins green at head.

| Mutant | Line replaced | Red |
|---|---|---|
| M1 | `enumStoredType`: the tuple branch's `return l.tupleElementStoredType(stored, node.AsElementAccessExpression().ArgumentExpression)` made `return nil` | TestNeverEnumTupleAgreesWithNode/first, rest, pair and TestNeverEnumTupleTrapsBeforeTheCall/first_out, rest_out, pair_out, each stopped at its switch's scrutinee as main stops it; TestNeverEnumTupleStops/variable_key, whose stop moves to the switch over endings[0] |
| M2 | `tupleElementStoredType`: the fixed element's `return elements[index]` made `return elements[0]`, the first element's type | TestNeverEnumTupleAgreesWithNode/pair and TestNeverEnumTupleTrapsBeforeTheCall/pair_out, judged by Phase where entry[1] is a Level |
| M3 | `tupleElementStoredType`: the rest's `return elements[fixed]` made `return nil` | TestNeverEnumTupleAgreesWithNode/rest and TestNeverEnumTupleTrapsBeforeTheCall/rest_out, each stopped at its switch's scrutinee |

M2 is the first element's type rather than the union's, since that is the one line. pair_out is what catches either:
its 5 is Phase.Check and no member of Level, so a reading by Phase, or by Phase | Level, whose first defined member
is Phase's, would not name Level in the trap.

Not yet measured by a proof job.
