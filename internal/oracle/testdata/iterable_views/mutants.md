# Mutants for the iterable view (#qk8rztp S2b-1b)

Each mutant is one exact line edit, run on a Linux box against the test it names, then put back from
a saved copy. The oracle rows run with `ADAMIC_GATE_UNCACHED=1`. The lower tests are in
internal/lower/iterable_views_test.go, and the oracle's is internal/oracle/iterable_views_test.go.

| Mutant | Edit | Red |
|---|---|---|
| An unproven origin accepted | lower/iterable_views.go, `checkIterableViews`: `if reaches.Unknown \|\| len(unproven) != 0 {` made `if false && (...) {` | TestIterableViewsOfOtherOriginsAreNotYet: the program's own iterable object, the function as a value, the object literal's method, the field, the host callback's parameter, the record's and the Map's value, and the string each compile. The array row stays green, since the array stops at the call (shape_fit.go) |
| The held check dropped | the same function: `if !origin.held.readsAs(view.held) {` made `if false && ...` | TestIterableViewsOfOtherOriginsAreNotYet: both held-differently rows compile |
| A pair's halves not compared | `readsAs`: its last line made `return true` | TestIterableViewsOfOtherOriginsAreNotYet: the entries row read as a union compiles; the values row still stops on its element |
| The view's tuple not read | `iterableViewElement`: `if checker.IsTupleType(yielded) {` made `if false && ...` | TestIterableViewsAgreeWithNode, view_entries.a: NotYet, entries held as a pair the view doesn't read |
| The arm skipped | lower/object.go, `forOf`: `if held, isView := l.iterableViewElement(iterated); isView {` made `...; false && isView {` | TestIterableViewsAgreeWithNode, every fixture but view_union.a: NotYet "for...of over an object" |
| The view not recorded | `forOf`: the `l.iterableViews = append(...)` line made `_ = iterableView{...}` | TestIterableViewsOfOtherOriginsAreNotYet: every row but the array's compiles, since nothing proves the view |
| The pass not run | lower/lower.go: `if err := lowering.checkIterableViews(); err != nil {` made `if err := error(nil); err != nil {` | the same rows as the view not recorded |
| The entry hook not mapping a collection's iterator | `checkIterableViews`: `return site(collectionOrigin(value))` made `_ = value; return site(iterableOrigin{name: "a value made by CollectionIterator"})`, so `value` stays used and vet passes | TestIterableViewsAgreeWithNode, all six: NotYet "of unproven origin"; and TestIterableViewsOverCollectionIteratorsLower |
| The union-held slot's unbox unchecked | lower/object.go, `forOf`'s union case: the `unboxed := ir.Narrow{...}` line with `Checked: true` made `Checked: false` | TestIterableViewHeldAsAUnionIsReadThroughACheckedUnbox: no checked unbox in the loop. No proven program reaches the failing arm, so this pin reads the IR |
| The union case skipped | `forOf`'s union case: `held, isView := l.iterableViewElement(iterated)` followed by `; isView = false` | TestIterableViewsAgreeWithNode, view_union.a: NotYet "for...of over a union of differently held members" |
| An unproven origin accepted, the string row | as the first row | TestIterableViewsOfOtherOriginsAreNotYet, the string handed to an Iterable<string> slot: it compiles, and only the checked unbox would stop it at run time |
| A MapIterator or SetIterator slot taken by the arm (equivalent) | `iterableViewElement`: `"MapIterator", "SetIterator"` added to the listed types | green, as it must be: libraryIteratorElement takes every MapIterator- or SetIterator-typed value whose element is one boxed slot before forOf's switch, and the arm asks for the same slot, so no such value reaches it. A generator declared to return one is stepped as a generator before either (declared_iterators.a) |
| A function value's parameters taken as known | lower/view_flow.go, `newAllocationFlowGraph`'s MakeClosure arm: `graph.unknownParameters[parameter+1] = true` made `_ = parameter` | TestIterableViewsOfOtherOriginsAreNotYet's program arrow row: the arrow's one call, through a local whose closure is constant (ClosureTargets), hands its parameter a Map's values(), so the view is proven and compiles. Inferred, not run: it is red only if ClosureTargets names the arrow for that call, and if it doesn't, the call's arguments never reach the parameter and the row still stops |

## A hand-driven next() through a view (#qk8rztp S2c-2a)

The lower rows are internal/lower/iterable_hand_next_test.go's; run by s2c2a-proof.sh.

| Mutant | Edit | Red |
|---|---|---|
| The hand view's origins not proven | lower/iterable_views.go, `checkIterableViews`: `if reaches.Unknown \|\| len(unproven) != 0 {` made `if !view.hand && (...) {` | TestHandDrivenNextStopsOnEveryOtherRoad, the arrow row: its parameter's callers are unseen, and it compiles |
| next(x) let through | lower/expression.go, `callClosure`: `if hand && len(...Arguments.Nodes) != 0 {` made `if false {` | the same test's next(x) row: it stops otherwise, or compiles |
| A for-of beside a hand next let through | lower/iterable_views.go, `checkIterableViews`: the mixed test made `&& false` | TestHandDrivenNextStopsOnEveryOtherRoad, the mixed rows in .a and .ts: they compile |
| No hand next let past the guard | lower/object.go, `property`: `node != l.handNext` made `true` | the same fixture: every next() is refused, as on main |
| IteratorYieldResult's fields admitted in iteratorResultField (dropped from the head) | an earlier draft let them past prototype.go's owner test | it survived on Golem (job 01M6Q811): hand_next.a's done and value reads lower on main by a road that doesn't reach that guard, so the head doesn't change it. Which road isn't traced yet |
| The narrowing check dropped (Adamic's) | none: the narrowing isn't a check of this slice's | an unnarrowed or done-arm `status.value` is `T \| any`, which is any, and the any rules stop it as they stop any value; the two value rows pin that it stops |

## A hand-driven next() through a view a generator feeds (#qk8rztp S2c-2b)

Run by s2c2b's h3-proof.sh. The lower rows are internal/lower/iterable_hand_next_test.go's, and the
oracle's are this directory's hand_next_generator.a, hand_next_generator_throw.a, hand_next_mixed.a
and hand_next_maybe_string.a. The closed defaults' rows are in internal/ir, internal/worker and
internal/native, and the await's witness is internal/oracle/testdata/reuse_await.

| Mutant | Edit | Red |
|---|---|---|
| M1: the arm retains the yielded value | native/runtime/generator.c, `generator_result`: `return adamic_iterator_result(kind, present, yielded, tag);` made `if (kind == adamic_field_reference && present) { adamic_retain(yielded.reference); } return ...` | hand_next_generator.a: every diagnostic outlives its result, and the counts leak |
| M3: the result built before the throw test | `adamic_iterator_next`: `if (adamic_thrown != NULL) {` made `if (adamic_thrown != NULL && generator_result(generator, kind, element_type, yielded) != NULL) {` | hand_next_generator_throw.a: the throw's step builds a done result nothing lets go of |
| The exemption widened to any call | lower/iterable_views.go, `generatorParameter`: `hand && isNext {` made `hand && (isNext \|\| fmt.Sprintf("%T", node) == "ir.Call") {` | TestHandDrivenNextOverAGeneratorStops, the forward row: the count passes it, and checkGeneratorRoads' hand road stops it as a compiler bug instead of NotYet |
| The exemption dropped | the same line made `false && hand && isNext {` | TestHandDrivenNextOverAGeneratorLowers: every row stops, read as many times as it has nexts |
| MaybeBoolean admitted | `handNextHolds`: `\|\| of == ir.MaybeBoolean` added to its return | Only a unit test can turn it red: TestHandNextHoldsOnlyNumbersBooleansAndReferences, the MaybeBoolean row. No program reaches handNextHolds with a maybe yield, since the source-level stop is in front of it: generatorSignature stops a `boolean \| undefined` or `number \| undefined` yield at the declaration, "(each yielded value must be one present slot)", pinned by TestHandDrivenNextOverAGeneratorStops' two maybe-yield rows, which stay green under this mutant |
| A Defined over any operand taken | `handNextReceiver`: `receiver = defined.Value` made `return property.Object, defined.Value != nil` | Only a unit test can turn it red: TestHandNextReceiverIsTheParametersOwnRead, the holder.it row, on built IR. The source-level stop is in front of it: `holder.it.next()` is no hand next, since handDrivenNext takes only a receiver that names a parameter, so it never reaches the matcher, pinned by TestHandDrivenNextOverAGeneratorStops' holder.it row, which stays green under this mutant |
| The rewrite skipped | `handView`: the `return ir.IteratorNext{...}` line made `_ = iterator; return node` | hand_next_generator.a: S2c-2a's CallClosure steps a generator frame by a next field it hasn't got |
| IteratorNext never throws | ir/generator.go, `IteratorNextMayThrow`: `return p.Functions[next.Function].ResumeMayThrow` made `return false && ...` | hand_next_generator_throw.a: nothing tests for the throw, and the `NULL` result is read |
| The inert default opened | ir/inert.go, `Inert`: its last line made `return true \|\| ...` | TestInertFailsClosedOnAMadeUpExpression |
| Purity's default opened | worker/purity.go: the default arm's test made `false && isExpression && ...` | TestCheckPureRefusesAnOperationNoReviewNames |
| Purity's step not followed | the GeneratorNext arm's `if err := inspect(node.Function); err != nil {` made `if err := error(nil); err != nil {` | TestCheckPureInspectsAGeneratorStepsBody |
| Reuse's default opened | native/reuse.go, `reachesGlobal`: `return !ir.Inert(expression)` made `return false && ...` | TestReuseTakesAnOperationNoReviewNamesAsTouchingTheGlobal |
| Reuse's await reaching nothing, main's behavior (Memory's question 1) | native/reuse.go, `reachesGlobal`: `ir.Await` added to the case of an encoding and the host operations, which reach nothing | TestReuseAwaitAgreesWithNode, reuse_await/global_across_await.a, sanitized: swap's write frees a1 while grow's suspended concatenation holds it unretained, and ASan reports heap-use-after-free when it resumes. If the scheduler runs swap only after grow resumes, the row stays green and the awaits want reordering until swap's write lands inside grow's suspension |
| Reuse's step not followed | the GeneratorNext arm's `return touches(...)` made `return false` | TestReuseFollowsAGeneratorStepIntoItsBody |
| A switch's IteratorNext case dropped | flow/infer.go, `callsBack`: `case ir.GeneratorNext, ir.IteratorNext:` made `case ir.GeneratorNext:` | TestEverySwitchOverGeneratorNextNamesIteratorNext (internal/ir), naming infer.go's line |
