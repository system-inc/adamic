Mutant evidence for predicates fixture 09's stage-0 record in status.json, and for stage3/drivers/parser/native-generic-assert-non-nullable.a's first line.

What moved: a call ends a predicate proof's path when its callee is a function declaration nothing can replace, named bare or through a namespace, whose one body never returns (neverCall, internal/lower/predicates_proof.go), and the flow prover ends a path there as at the prelude's panic (predicateFlowProof.statement, internal/lower/predicates.go). Fixture 09 is tsc's debug.ts assertIsDefined and checkDefined. assertIsDefined's failing branch calls `fail`, a top-level function declaration whose body throws, so `asserts value is NonNullable<T>` is proven by the body (#8n7eet3, Adamic's ruling Nov 10, applying rulings.md:134 and :146). Its record moves from Refused at 8:127 ("normal return has not narrowed value to NonNullable<T>") to Compiles, printing what Node prints. The driver, a reduction of the same function, moves from `refused` to `checked`. Neither program changed.

Mutant: the flow prover's never call removed. In predicateFlowProof.statement, the ExpressionStatement branch ends a path only at the prelude's panic again:

    if node.Kind == ast.KindExpressionStatement && p.l.isPanicCall(node.Expression()) {

Commands: go test ./stage3/fixtures -count=1 -run 'TestFixturesPredicates$' -v, and go test ./internal/acheck -count=1 -v.
Expected with the mutant: the first exits 1, red on exactly 09_assert_defined.a's stage0 check, outcome Refused at 8:127 against the recorded Compiles; the second exits 1, naming native-generic-assert-non-nullable.a as refused against its first line's promise of checked. Restored, both pass.
Result: not yet run; the house run's output for both commands, with and without the mutant, belongs here verbatim before this lands.
