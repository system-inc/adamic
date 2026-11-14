# Mutants for an array literal bound to a constraint that declares nothing (#x856a4v)

One edit to constraint_relation.go, run on a house box, then put back. Measured at 1cd41901, Loom job J-xa31sa5.

| Mutant | Old text | New text | Red |
|---|---|---|---|
| The declaresNothing arm dropped | `if l.declaresNothing(array) {` | `if false && l.declaresNothing(array) {` | internal/oracle's TestBoundArrayEmptyConstraint: e01_append_pairs, e02_append_object_element and e05_object_constraint, .a and .ts, each back at judgeBoundArray's NotYet at the array literal |

e04_unknown_constraint stays green under the mutant, since judgeBinding returns before judgeBoundArray for an unknown target: the arm is defense there. e03_hidden_element, a constraint that declares `0`, keeps its NotYet at the literal at the head and under the mutant.
