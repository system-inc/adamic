Mutant evidence for objects fixture 25's stage-0 record in status.json.

What moved: rule 2500 says NotYet for a generator function, never Refused (#qk8rztp S1b, docs/rulings.md open question 17), at the same ordinal. The fixture still stops at 3:1, its `function*`, with the new kind and text. Its census reason now names the construct, "a generator function", rather than the refusal's old "yield (generators)". The fixture is unchanged.

Mutant: generatorGap in internal/lower/refusals.go returns the old Refused (its what and fix as on main 02b0112f), on top of this change; the change's status.json stays.
Command: GOFLAGS=-buildvcs=false go test -count=1 -v ./stage3/fixtures -run 'TestFixturesObjects$'
Result with the mutant: exit 1, red in objects on exactly 25_map_generator.a/stage0, back at the old refusal. Restored, checked with cmp, the same command passes (darwin/arm64, the stage0 record only).

    fixtures_test.go:476: gap changed: update status.json and check the native output; got {Outcome:Refused What:stage3/fixtures/objects/25_map_generator.a:3:1: Adamic 0.1 refuses a generator function; use an explicit iterator object; suspended frames need ownership and cancellation rules before generators can be compiled without a collector (docs/user-iterators.md) Stdout:}; recorded {Outcome:NotYet What:stage3/fixtures/objects/25_map_generator.a:3:1: stage 0 can't lower a generator function (its counted frame is being built, #qk8rztp) yet Stdout:}
--- FAIL: TestFixturesObjects/objects/25_map_generator.a/stage0
