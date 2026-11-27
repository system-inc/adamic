# Mutants of enumStoredType's tuple element (#xfpsqtn, 2)

Each mutant replaces exactly one line of internal/lower/enum_never.go, checked to match once, runs `go test -run
TestNeverEnumTuple ./internal/oracle/`, and is put back from a copy of the head file, checked byte for byte. Each is
red beside the same pins green at head.

| Mutant | Line replaced | Red |
|---|---|---|
| M1 | `enumStoredType`: the tuple branch's `return l.tupleElementStoredType(stored, node.AsElementAccessExpression().ArgumentExpression)` made `return nil` | TestNeverEnumTupleAgreesWithNode/first, rest, pair and TestNeverEnumTupleTrapsBeforeTheCall/first_out, rest_out, pair_out, each stopped at its switch's scrutinee (first.ts:21:10, rest.ts:23:10, pair.ts:18:10, and the same in each _out) in main's words, "a switch's unmatched case on a value the checker narrowed across code that may write it"; TestNeverEnumTupleStops/variable_key, whose stop moves to the switch over endings[0] at variable_key.ts:13:10 |
| M2 | `tupleElementStoredType`: the fixed element's `return elements[index]` made `return elements[0]`, the first element's type | TestNeverEnumTupleAgreesWithNode/pair and TestNeverEnumTupleTrapsBeforeTheCall/pair_out, each stopped at the switch over entry[1] (pair.ts:18:10, pair_out.ts:18:10) in M1's words |
| M3 | `tupleElementStoredType`: the rest's `return elements[fixed]` made `return nil` | TestNeverEnumTupleAgreesWithNode/rest and TestNeverEnumTupleTrapsBeforeTheCall/rest_out, each stopped at the switch over endings[1] (rest.ts:23:10, rest_out.ts:23:10) in M1's words |

M2 is the first element's type rather than the union's, since that is the one line. The pair pins catch it at the
switch, before the trap is reached; were a reading by Phase or by Phase | Level let through the switch, pair_out's 5,
which is Phase.Check and no member of Level, would still not name Level in the trap.

Measured by J-wv3q19n at 86c5ec75: the pins green at head, and every mutant above red with the subtests named, each
put back and checked byte for byte.
