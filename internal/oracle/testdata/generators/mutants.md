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
