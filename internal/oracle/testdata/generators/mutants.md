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
| The walk stops every generator method | lower/refusals.go, `asyncGeneratorRefusal`: `if asyncGenerator(node) {` made `if asyncGenerator(node) \|\| generatorDeclaration(node) && node.Kind == ast.KindMethodDeclaration {` | unused_class.a, refused before its static block can run, and the unused half of TestGeneratorMethodBothWays |
