# Mutants for a predicate parameter passed as its checked closure (#jqf1drp)

One edit to internal/lower/predicate_parameter_value.go, run on a house box, then put back. Measured at 4395b82f, Loom job J-v3gw1cj.

| Mutant | Old text | New text | Red |
|---|---|---|---|
| The parameter passed straight through, with no closure | `if value.Type() != ir.Closure \|\| !l.checkedPredicateParameterPassed(node) {` | `if value.Type() != ir.Closure \|\| !l.checkedPredicateParameterPassed(node) \|\| true {` | TestPredicateParameterPassedAsCheckedClosure's honest and conditional ("want one checked closure, got 0"), and TestPredicateParameterLieStopsAtTheClosure's lie and escape, whose second run, with the forwarder's check taken out of the IR, no longer stops at the closure |

## Why the lie pin edits the IR

No program reaches the closure's check before another check stops it. A lying predicate is checked where it becomes a value: isCircleLie passed directly, through a conditional, a `??`, a returned or stored function, or a checked arrow, stops in its forwarder's words ("predicate isCircleLie used as a value") on every backend, at the head and under the mutant alike (J-xqd016h). A cast to a lying predicate type is refused by adamic/no-unchecked-cast, and a non-predicate passed for a predicate parameter is a TypeScript error (J-eadh6v8). So the closure's own stop is proven by taking the forwarder's check out of the lowered program, and today the closure is defense in depth behind each predicate value's own check, there for any road that hands the parameter an unchecked one. The mutant is caught by the closure count on the honest programs and by the second run on the lying ones.
