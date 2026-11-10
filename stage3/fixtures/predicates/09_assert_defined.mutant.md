Mutant evidence for predicates fixture 09's stage-0 record in status.json, and for stage3/drivers/parser/native-generic-assert-non-nullable.a's first line.

What moved: a call ends a predicate proof's path when its callee is a function declaration nothing can replace, named bare or through a namespace, whose one body never returns (neverCall, internal/lower/predicates_proof.go), and the flow prover ends a path there as at the prelude's panic (predicateFlowProof.statement, internal/lower/predicates.go). Fixture 09 is tsc's debug.ts assertIsDefined and checkDefined. assertIsDefined's failing branch calls `fail`, a top-level function declaration whose body throws, so `asserts value is NonNullable<T>` is proven by the body (#8n7eet3, Adamic's ruling Nov 10, applying rulings.md:134 and :146). Its record's first stop moves past the predicate: from Refused at 8:127 ("normal return has not narrowed value to NonNullable<T>") to Refused at 11:45, the next wall, where `stackCrawlMark || assertIsDefined` hands a void-returning function to the fixture's AnyFunction, `(...args: never[]) => unknown` (adamic/no-void-widening, H2's rule; tsc's own AnyFunction returns void and doesn't meet it). The driver, a reduction of the same function with no mark, moves from `refused` to `checked`. Neither program changed.

Mutant: the flow prover's never call removed. In predicateFlowProof.statement, the ExpressionStatement branch ends a path only at the prelude's panic again:

    if node.Kind == ast.KindExpressionStatement && p.l.isPanicCall(node.Expression()) {

Commands: go test ./stage3/fixtures -count=1 -run 'TestFixturesPredicates$' -v, and go test ./internal/acheck -count=1 -v.
Expected with the mutant: the first exits 1, red on exactly 09_assert_defined.a's stage0 check, Refused at 8:127 against the recorded refusal at 11:45; the second exits 1, naming native-generic-assert-non-nullable.a as refused against its first line's promise of checked. Restored, both pass.
Result, on Pupa at 9a66e808 (hand job 01M73DZ8WWSQ2GBJE0VFB0W6XQ), the mutant applied by replacing the neverCall call in predicates.go with `p.l.isPanicCall(node.Expression())` and reverted with `git checkout` after:

    fxhead-exit=0
    ok  	github.com/system-inc/adamic/stage3/fixtures	17.247s
    fxmutant-exit=1
        fixtures_test.go:493: gap changed: update status.json and check the native output; got {Outcome:Refused What:stage3/fixtures/predicates/09_assert_defined.a:8:127: Adamic 0.1 refuses a type predicate whose return is not proven (normal return has not narrowed value to NonNullable<T>); inline the check where you use it, or return a discriminant comparison on the unmodified parameter (adamic/no-type-predicate) Stdout:}; recorded {Outcome:Refused What:stage3/fixtures/predicates/09_assert_defined.a:11:45: Adamic 0.1 refuses a void-returning function at the assignment site seen as AnyFunction | undefined; declare the function's real return type (adamic/no-void-widening) Stdout:}
    --- FAIL: TestFixturesPredicates (21.78s)
    acmutant-exit=1
    --- FAIL: TestEveryProgramOutsideAPackageKeepsItsPromise (18.44s)

Restored, internal/acheck passed whole on the combined tree a824d156 (ac-exit=0, hand job before it on Pupa), where this driver's first line already promised checked.
