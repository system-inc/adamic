Mutant evidence for objects fixture 30's stage-0 record in status.json.

What moved: optionalValue looks through an arrow's expression body (#6c4swvt G4), so `getCompilerHost: () => ({name: "compiler"})` returns a fresh literal and no field of ResolutionCacheHost is hidden. The fixture's refusal at 8:73 (adamic/no-optional-widening) no longer fires, and it now stops at its own census reason, a call through ?. at 6:12. The fixture is unchanged.

Mutant: optionalValue's arrow case disabled (`if true || body == nil || ...`), on top of this change; the change's status.json stays.
Command: GOFLAGS=-buildvcs=false go test -count=1 -v ./stage3/fixtures -run 'TestFixturesObjects$'
Result with the mutant: exit 1, red in objects on exactly 30_host_optional_method.a/stage0, back at the recorded refusal. Restored, checked with cmp, the same command passes (darwin/arm64, the stage0 record only).

    fixtures_test.go:472: gap changed: update status.json and check the native output; got {Outcome:Refused What:stage3/fixtures/objects/30_host_optional_method.a:8:73: Adamic 0.1 refuses optional property getCompilerHost in ResolutionCacheHost absent from structural source { name: string; }, which can hide fields; declare getCompilerHost on the source type, or build a fresh object with known fields (adamic/no-optional-widening) Stdout:}; recorded {Outcome:NotYet What:stage3/fixtures/objects/30_host_optional_method.a:6:12: stage 0 can't lower a call through ?. (an optional call) yet Stdout:}
--- FAIL: TestFixturesObjects/objects/30_host_optional_method.a/stage0
