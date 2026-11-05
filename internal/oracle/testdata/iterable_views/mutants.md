# Mutants for the iterable view (#qk8rztp S2b-1b)

Each mutant is one exact line edit, run on a Linux box against the test it names, then put back from
a saved copy. The oracle rows run with `ADAMIC_GATE_UNCACHED=1`. The lower tests are in
internal/lower/iterable_views_test.go, and the oracle's is internal/oracle/iterable_views_test.go.

| Mutant | Edit | Red |
|---|---|---|
| An unproven origin accepted | lower/iterable_views.go, `checkIterableViews`: `if reaches.Unknown \|\| len(unproven) != 0 {` made `if false && (...) {` | TestIterableViewsOfOtherOriginsAreNotYet: the program's own iterable object, the function as a value, the object literal's method and the field each compile. The array row stays green, since the array stops at the call (shape_fit.go) |
| The held check dropped | the same function: `if !origin.held.readsAs(view.held) {` made `if false && ...` | TestIterableViewsOfOtherOriginsAreNotYet: both held-differently rows compile |
| A pair's halves not compared | `readsAs`: its last line made `return true` | TestIterableViewsOfOtherOriginsAreNotYet: the entries row read as a union compiles; the values row still stops on its element |
| The view's tuple not read | `iterableViewElement`: `if checker.IsTupleType(yielded) {` made `if false && ...` | TestIterableViewsAgreeWithNode, view_entries.a: NotYet, entries held as a pair the view doesn't read |
| The arm skipped | lower/object.go, `forOf`: `if held, isView := l.iterableViewElement(iterated); isView {` made `...; false && isView {` | TestIterableViewsAgreeWithNode, all four: NotYet "for...of over an object" |
| The view not recorded | `forOf`: the `l.iterableViews = append(...)` line made `_ = iterableView{...}` | TestIterableViewsOfOtherOriginsAreNotYet: every row but the array's compiles, since nothing proves the view |
| The pass not run | lower/lower.go: `if err := lowering.checkIterableViews(); err != nil {` made `if err := error(nil); err != nil {` | the same rows as the view not recorded |
| The entry hook not mapping a collection's iterator | `checkIterableViews`: `return site(collectionOrigin(value))` made `return site(iterableOrigin{name: "a value made by CollectionIterator"})` | TestIterableViewsAgreeWithNode, all four: NotYet "of unproven origin" |
| A MapIterator or SetIterator slot taken by the arm (equivalent) | `iterableViewElement`: `"MapIterator", "SetIterator"` added to the listed types | green, as it must be: libraryIteratorElement takes every MapIterator- or SetIterator-typed value whose element is one boxed slot before forOf's switch, and the arm asks for the same slot, so no such value reaches it. A generator declared to return one is stepped as a generator before either (declared_iterators.a) |
