Mutant evidence for objects fixture 25's stage-0 record in status.json.

What moved: rule 2500 says NotYet for a generator function, never Refused (#qk8rztp S1b, docs/rulings.md open question 17), at the same ordinal. The fixture still stops at 3:1, its `function*`, with the new kind and text. Its census reason now names the construct, "a generator function", rather than the refusal's old "yield (generators)". The fixture is unchanged.

Mutant: generatorGap in internal/lower/refusals.go returns the old Refused (its what and fix as on main 02b0112f), on top of this change; the change's status.json stays.
Command: GOFLAGS=-buildvcs=false go test -count=1 -v ./stage3/fixtures -run 'TestFixturesObjects$'
Result with the mutant: exit 1, red in objects on exactly 25_map_generator.a/stage0, back at the old refusal. Restored, checked with cmp, the same command passes (darwin/arm64, the stage0 record only).

    fixtures_test.go:476: gap changed: update status.json and check the native output; got {Outcome:Refused What:stage3/fixtures/objects/25_map_generator.a:3:1: Adamic 0.1 refuses a generator function; use an explicit iterator object; suspended frames need ownership and cancellation rules before generators can be compiled without a collector (docs/user-iterators.md) Stdout:}; recorded {Outcome:NotYet What:stage3/fixtures/objects/25_map_generator.a:3:1: stage 0 can't lower a generator function (its counted frame is being built, #qk8rztp) yet Stdout:}

--- FAIL: TestFixturesObjects/objects/25_map_generator.a/stage0

## S2a: generic generators lowered per instance (#qk8rztp)

What moved: mapIterator is a generic top-level `function*`, which now lowers once per instantiation, where it's called, so rule 2500 no longer stops it at 3:1. The spread instantiates it at number, and its body lowers as far as its for-of over `iter`, an `Iterable<T>` the body can't step until the erased view is built (S2b), and stops there, at 4:21, NotYet, "for...of over an object". The census reason names that construct. The fixture is unchanged, and the record still checks the outcome and the exact text.

Mutant: loweredGenerator in internal/lower/generators.go refuses type parameters again (`len(node.TypeParameters()) == 0 &&` put back), on top of this change; the change's status.json stays.
Command: GOFLAGS=-buildvcs=false go test -count=1 -v ./stage3/fixtures -run 'TestFixturesObjects$'
Result with the mutant: exit 1, red in objects on exactly 25_map_generator.a/stage0, back at 3:1, rule 2500's generator stop. Restored, checked with cmp, the same command passes (darwin/arm64).

Then rule 2500 is deleted (#qk8rztp, Lower's open question 7): generator forms not built stop at lowering. The record doesn't move, since mapIterator is lowered. The same mutant, on top of that change, reads Refused at 8:17, the spread, by adamic/no-void-read (a generator no longer lowered isn't exempt from the void rules), and is red on exactly 25_map_generator.a/stage0 again.

## S2b-1b: a for-of over an Iterable<T> slot is a view (#qk8rztp)

What moved: mapIterator's for-of over `iter` lowers as a view, proven after lowering to be handed only a collection's iterator (internal/lower/iterable_views.go). The body no longer stops at 4:21. The spread hands it an array, `[1, 2, 3]`, which stops where it goes into `Iterable<number>`, at 8:29, NotYet, "an array going where a plain object type is held" (shape_fit.go). The census reason names that construct. The fixture is unchanged. This record was written before any run: the box run of S2b-1b's proof says whether the argument's stop fires first, or the view's proof at 4:21 ("a view over an iterable of unproven origin"), and this record is held to what it prints.

Mutant: forOf's view arm skipped (`false && isView`, internal/lower/object.go), on top of this change; the change's status.json stays.
Command: GOFLAGS=-buildvcs=false go test -count=1 -v ./stage3/fixtures -run 'TestFixturesObjects$'
Expected with the mutant: red on exactly 25_map_generator.a/stage0, back at 4:21, "for...of over an object".
