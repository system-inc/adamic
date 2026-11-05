# Mutants for the generator fixtures (#qk8rztp S1c)

Each mutant is one exact edit, run on the Mac with `ADAMIC_GATE_UNCACHED=1` against
`TestGeneratorsAgreeWithNode` (internal/oracle/generators_test.go) unless it names another test, then
put back from the commit. "Red" names the fixtures that went red and how. On the Mac the leak check is
the counted build (allocations against frees); LeakSanitizer is the box's.

| Mutant | Edit | Red |
|---|---|---|
| Kill without `NULL` (M-c) | runtime/generator.c, `adamic_generator_kill`: the `cell->value.reference = NULL;` line dropped | every fixture but drop_start.a, ASan heap-use-after-free at finish or the drop walk; return_from_loop.a prints freed bytes natively |
| Throw leaves `frame->error` in place | native/generator.go, the throw terminal: `frame->error = NULL;` dropped, so finish releases the error the step's caller then reads | uncaught.a, split.a, body_throw.a, ASan heap-use-after-free; split.a prints `caught caught fore the yield` natively |
| The step's thrown test removed | native/generator.go, `generatorNext`: `checkThrown` never emitted | uncaught.a prints `1`, `after` and exits 0; split.a prints `after` with no catch and leaks 2 values; body_throw.a prints `consumed` where Node catches |
| `ResumeMayThrow` forced false | lower/exceptions.go: the generator arm of the fixed point never sets it | the same three, the same way |
| The entry's parameter retain dropped | native/generator.go, the entry: `adamic_retain(...)` around a reference parameter dropped | retained_param.a (`sy again sy again` natively, ASan use after free), and seven more under ASan |
| The kill set takes every counted cell (Diagnostics' must 2) | native/generator.go: the live test made always false | every fixture but drop_start.a, which never resumes: UBSan null member access on resume; also TestGeneratorKillSetReleasesOnlyDeadCells (internal/native) |
| The kill set skipped (M-b) | native/generator.go: the `adamic_generator_kill` line not emitted | green here, as Memory ruled: the drop walk frees a dead cell either way, so it's a counts row, which regenerates only on Linux. Red in TestGeneratorKillSetReleasesOnlyDeadCells (internal/native), which reads the resume |
| The drop walk frees nothing before the first step | runtime/generator.c, `adamic_generator_free_children`: cells skipped while status is start | drop_start.a, a counted leak of 2 |
| The drop walk frees nothing while suspended | the same walk, cells skipped while suspended | return_from_loop.a, break_map.a (9), labeled_exit.a, consumer_throw.a, counted leaks |
| A scope exit skips the generator | native/emit_ownership.go, `releaseScopes`: a name ending `_generator` not released | every fixture that finishes but drop_start.a, a counted leak of at least the frame |
| A labeled exit skips the generator | native/labels.go, `labeledJump`: the same filter, on a labeled break or continue only | labeled_exit.a alone, a counted leak of 10 |

Wave 2, run against the six fixtures it added:

| Mutant | Edit | Red |
|---|---|---|
| Throw leaves `frame->error` in place (Diagnostics' must 3) | as above | first_throw.a prints `RangeError: RangeError: ` and freed bytes natively; it and uncaught_spread.a and uncaught_from.a, ASan heap-use-after-free |
| `ResumeMayThrow` forced false | as above | uncaught_spread.a and uncaught_from.a print `1`, `after` and exit 0; first_throw.a prints an empty line where Node catches, and leaks 2 |
| The step's thrown test removed | as above | green on all six. The collect helper is a function whose body can throw, so its caller tests the pending exception after the helper's call, and the helper's partial array is let go of there; the test inside the helper only stops it pushing sooner. The `ResumeMayThrow` row is the one that reaches the caller's test |
| A for-of routes a declared iterator by type | lower/object.go, `forOf`: the generator arm skipped for a call typed SetIterator or MapIterator | declared_iterators.a, which then stops NotYet (planIteration's generator arm), never compiling as a Set's iterator |
| A spread routes a declared iterator by type | lower/collections.go, `iterated`: the same, so libraryIteratorElement reads the call's declared type | declared_iterators.a, NotYet at the call (the call guard in expression.go) |
| Array.from routes a declared iterator by type | lower/from.go, `arrayFrom`: the same | declared_iterators.a, NotYet at the call |

The other wave 2 pins have their mutants in their tests: generatorStepped taken out of each void
rule (TestGeneratorVoidReturnIsNeverReadBySpreadOrArrayFrom, internal/lower), the consumer's
generator local given a checker type (TestNoSlotTheIteratorArmWalksHoldsAGenerator), the return
arm's word put back to `resumed` (TestGeneratorRuntimeReturnWhileRunning, internal/native), and
the resume emitted with no source lines (TestDebugGeneratorSourceLines, internal/native).

Two mutants the brief names don't apply to this slice:

- **IterationCell's references flag false (M-f).** A generator's frame never holds an IterationCell:
  NormalizeAsync marks one only for a captured loop binding, and lowering stops a generator whose
  body a closure captures (generators.go, checkGeneratorBody). The flag's branch in generator.go's
  `counted` is unreachable, so the mutant is equivalent. It becomes live when S3 lets a closure
  capture a frame cell.
- **The done test ordered before the thrown test.** The thrown test is part of the step's own
  statement and the done test is the next statement, so the only reordering that changes anything
  puts the test after the loop's `break`, which is the thrown test removed for the step that threw.

Wave 3 (S2a, generic generators lowered per instance), run against the three fixtures it added:

| Mutant | Edit | Red |
|---|---|---|
| Every instance shares one key | lower/generic.go, `instantiateFunctionSignature`: `key += "," + l.genericTypeKey(concrete)` made `key += ","`, so a second instantiation reuses the first's function | generic.a alone, NotYet at 15:11: the number for-of finds the string instance, held differently from what it yields. The other two instantiate one generator once |
| The entry's retain of a closure parameter dropped | native/generator.go, the entry: `if of.IsReference() {` made `if of.IsReference() && of != ir.Closure {` | generic_map.a, ASan heap-use-after-free in the resume, reading the mapper `ending` made, which nothing but the frame held; and generic_throw.a, whose named mapper's value adapter is released once too often. No fixture before these passed a function to a generator |
| A call through a function value never throws | lower/exceptions.go, `throwsOut`: `ir.CallClosure` taken out of its arm, so a generator whose body throws only through its mapper isn't `ResumeMayThrow` | generic_throw.a alone, natively and sanitized: prints `consumed` and an empty line where Node catches, and leaks 2 |

The lowering's own pins have their mutants in their tests (internal/lower): generatorCallee's
generic arm taken out (TestGenericGeneratorsLowerPerInstance and every road row), call()'s generator
guard at the generic branch taken out (TestGenericGeneratorRoadsAreNotYet's value rows, two of them
on checkGeneratorRoads' compiler bug), and the walk stopping a generic generator before the void
rules (the void instantiation row of TestGeneratorVoidReturnIsNeverReadBySpreadOrArrayFrom, which
must be a void rule's refusal). The forms' stops at lowering, once rule 2500 is gone, have theirs in
TestGeneratorFormsStopAtLowering and TestUninstantiatedGenericGeneratorsAreWalkedButNotLowered. Two
mutants read by testing are equivalent on these
programs: generatorStepped reading `T` from the declaration (the spread's array is `void[]`, refused
whatever the exemption says), and generatorSignature reading its result without `concrete`
(representation reads a bare `T` through the instance's substitution; only an unread instance's
words change, `T` for `unknown`, red on that row alone).

Wave 3b (Lower's and Compiler's pins on an unused class, after rule 2500 went), each one exact line,
run by `s2a-proof.sh` with the one test it must turn red:

| Mutant | Edit | Red |
|---|---|---|
| The walk skips a module's classes | lower/refusals.go, `refuse`: `module.AsNode().ForEachChild(visit)` (the line after `l.refusal.deferred = nil`) made to skip a `ClassDeclaration` | TestUnusedClassGeneratorMethodIsStillWalked: the any cast in the unused class's generator method is never judged, and the program lowers |
| signature lets a generator method through | lower/functions.go, `signature`: `if !loweredGenerator(declaration) {` made `if false && !loweredGenerator(declaration) {` | TestGeneratorMethodUseIsDecidedAcrossModules (the dropped rows compile, the for-of rows stop in other words) and the used half of TestGeneratorMethodBothWays |
| The overload check taken out | lower/functions.go, `signature`: `if implementation := generatorImplementation(declaration); implementation != nil {` made `...; false && implementation != nil {` | TestOverloadedGeneratorMethodStopsAtItsImplementation (f): the first overload is written as an ordinary method and stops in other words |
| signature lets an object literal's generator method through (e) | as the second row | TestGeneratorFormsStopAtLowering's object literal method rows: the dropped ones compile, the for-of ones stop in other words |
| The walk stops every generator method | lower/refusals.go, `asyncGeneratorRefusal`: `if asyncGenerator(node) {` made `if asyncGenerator(node) \|\| generatorDeclaration(node) && node.Kind == ast.KindMethodDeclaration {` | unused_class.a, refused before its static block can run, and the unused half of TestGeneratorMethodBothWays |

S2b-1a (new Map and new Set of a generator's call), each one exact line, run by `s2b1a-proof.sh`
with the one test it must turn red:

| Mutant | Edit | Red |
|---|---|---|
| The pair's key and value swapped | lower/generators.go, `generatorPairTypes`: `return key, value, keyKnown && valueKnown` made `return value, key, keyKnown && valueKnown` | new_map.a, NotYet: its `[string, number]` pairs read as held otherwise than the Map's |
| The step's throw never tested | lower/exceptions.go: the generator arm of the fixed point made `if false && functions[index].Generator && ...`, so no step is `ResumeMayThrow` | new_map_throw.a: the throw is never seen, the Map is built from the two pairs before it and printed |
| The collected pairs not let go of | native/emit_ownership.go, `releaseScopes`: a local whose name ends `_collected` not released | new_map_throw.a, a counted leak of the half-built Map, the pairs the helper had collected when the step threw |
| The helper's generator not let go of on the throw exit | native/exceptions.go, `jumpThrown`: a function's throw exit, `e.releaseScopes(e.functionDepth)`, made, in a `collect_generator` helper only, to skip the local whose name ends `_generator` (the scopes trimmed of it for that one call) | new_map_throw.a, a counted leak of the frame the helper held when the step threw, the label it was building still in it (and under LeakSanitizer on the box) |
| new Set doesn't route a generator by origin | lower/collections.go, `iterated`: its generator arm, `} else if isGenerator {`, made `} else if false && isGenerator {` | new_set.a, NotYet as destructuring (planIteration's generator arm) |
| The collected pairs dropped | lower/generators.go, `newMapFromGenerator`: the MapNew made from an empty array, its pairs unused | TestNewMapAndNewSetStepAGeneratorsCall (internal/lower), and new_map_entries.a prints an empty Map |
| generatorPairTypes' tuple test taken out | lower/generators.go: `if len(arguments) == 0 \|\| !checker.IsTupleType(arguments[0]) {` made `if len(arguments) == 0 {` | TestNewMapOfAGeneratorsCallStopsWhereItsYieldsArentItsPairs, its union row |
| newMapFromGenerator's held-otherwise test taken out | lower/generators.go: `if pairKey != key \|\| pairValue != value {` made `if false && (pairKey != key \|\| pairValue != value) {` | TestNewMapOfAGeneratorsCallStopsWhereItsYieldsArentItsPairs, its held-otherwise row |

S2b-1c (new Map and new Set as stepped consumers, Adamic's ruling on #qk8rztp), each one exact line,
run by `s2b1c-proof.sh` with the one test it must turn red:

| Mutant | Edit | Red |
|---|---|---|
| The new Map and new Set arm dropped | lower/generators.go, `steppedSite`: `case parent.Kind == ast.KindNewExpression && l.collectionConstructor(parent) && ...` made `case false && ...` | new_map_plain.a and new_set_plain.a, refused again by the void rules; TestGeneratorVoidReturnIsNeverReadByNewMapOrNewSet's lowering rows |
| Widened to any Generator-typed expression at a stepped site | lower/generators.go, `generatorStepped`: its first test, `if node.Kind != ast.KindCallExpression {`, made to return `steppedSite` and a Generator type for any expression | TestGeneratorVoidReturnIsNeverReadByNewMapOrNewSet's stored rows, which reach lowering and stop there instead of a void rule |
| The yielded-type test dropped | lower/generators.go, `generatorStepped`: `return len(arguments) == 3 && !l.isVoidValue(arguments[0]) && !l.holdsVoid(arguments[0])` made `return len(arguments) == 3` | TestGeneratorVoidReturnIsNeverReadByNewMapOrNewSet's void rows: the void-yielding Set and both void pairs reach lowering |
| The constructor judged by its name | lower/generators.go, `collectionConstructor`: the library-symbol test made a name test | TestGeneratorVoidReturnIsNeverReadByNewMapOrNewSet's user-class row |
| steppedSite sees through a cast | lower/generators.go, `steppedSite`: `for site.Parent != nil && site.Parent.Kind == ast.KindParenthesizedExpression {` made to climb an `AsExpression` too | survives, equivalent: on Pupa (job 01M6N04CHA7P2ZCMK8M0AV6NDP) the cast row stops before any void rule judges the call, so seeing through the cast admits nothing. The row is held by its own stop, not by this mutant |

S2b-1b (a view stepped in a generator's body: tsc's mapIterator and mapDefinedIterator over a
collection's iterator), each one exact line, run by `s2b1b-proof.sh` with the one test it must turn
red:

| Mutant | Edit | Red |
|---|---|---|
| The held iterator's frame cell not counted | native/generator.go, `generatorFunction`'s `counted`: `return e.program.Locals[local].IterationCell \|\| e.program.Locals[local].Type.IsReference()` made to answer false for a local whose name starts `iterator` (the view's held iterator, libraryIteratorLoop's local) | map_iterator_entries.a and map_defined_iterator.a: the cell is never released at finish or by the drop walk, a counted leak of the iterator, its next closure and its state, and of the collection the state keeps (and under LeakSanitizer on the box) |
| The view arm skipped | lower/object.go, `forOf`: `if held, isView := l.iterableViewElement(iterated); isView {` made `...; false && isView {` | map_iterator_entries.a, NotYet "for...of over an object" at its for-of; map_defined_iterator.a is held as a union and still lowers |
| The union view arm skipped | lower/object.go, `forOf`'s union case: `held, isView := l.iterableViewElement(iterated)` followed by `; isView = false` | map_defined_iterator.a, NotYet "for...of over a union of differently held members" |

S2c-1 (a generator's call into a one-read Iterable<T> parameter), each one exact line, run by
`s2c1-proof.sh` with the one test it must turn red:

| Mutant | Edit | Red |
|---|---|---|
| The one-read check dropped | lower/iterable_views.go, `generatorParameter`: `if uses.reads != 1 {` made `if false && uses.reads != 1 {` | TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad (internal/lower): the next-through-the-parameter, typeof, next-inside-the-loop, labeled-break, identity and default rows compile |
| Widened to a stored generator | lower/expression.go, `call`'s plain generator arm: `if l.iterableArgument(node) {` made `if true {` | the same test's stored-generator row: the stored call goes through IterableOf into a local, and checkGeneratorRoads says so as a compiler bug instead of the stored-generator stop |
| The close at the break dropped | native/emit_ownership.go, `releaseScopes`: the `adamic_iterator_close` line made `(void)` of the local | view_outer_loop.a, natively: the second and third passes resume the generator and print its next values, where Node prints nothing |
| The dispatch's generator arm removed | lower/iterable_views.go, `dispatchView`: the kind test's condition, `ir.Unary{Operator: ir.Not, Operand: ir.IsGenerator{Value: held}}`, made `ir.BooleanConstant{Value: true}` | every fixture with a view a generator feeds (array_from_view.a, elementwise_view.a, view_outer_loop.a, view_throw.a): the loud default panics, a compiler bug naming the site |
| IterableOf counted fresh | native/region.go, `freshValue`: IterableOf's `return false` made `return true` | TestIterableOfIsNeverFresh (internal/native) |
| A for-of head not counted as a write (Lower's must) | lower/local_uses.go: `"ForOf.Local": {role: localWrite},` made `{why: "mutant: no local"}` | TestLocalUseCountSeesEveryWrite (internal/lower): the for-of head row counts no write |
| A pattern's binding left unclassified | lower/local_uses.go: the `"Binding.Local"` entry taken out | TestLocalFieldsClassifyEveryIntField: Binding.Local holds an int and isn't listed |
| checkGeneratorRoads' count ignored (Compiler's guard) | lower/generators.go: `if counted != allowed {` made `if false && counted != allowed {` | TestGeneratorRoadsRefuseAGeneratorBoxedOrCastIntoAnObject: a generator boxed or cast into an Object slot passes |

q3's retain around a step and a return (runtime/generator.c) has no mutant row: no program reaches
it yet, since no slot but the consumer's holds a generator, so nothing a resume runs can let go of
the last holder. It becomes live with stored generators (S2c-4 on).
