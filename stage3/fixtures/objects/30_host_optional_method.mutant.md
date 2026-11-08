Mutant evidence for objects fixture 30's stage-0 record in status.json.

What moved: an optional method the program declares lowers when it's called through `?.(` (#tvq1eqm, internal/lower/optional_method_call.go), so the fixture's census reason, `resolutionHost.getCompilerHost?.()` at 6:12, lowers, and so does `host.getCurrentDirectory?.()` at 16:75. The fixture compiles: its native run, sanitized and leak-checked through TestStage3FixtureHook, is held to the row's recorded Node observation. The fixture is unchanged.

Before that, it stopped at 6:12 with `stage 0 can't lower a call through ?. (an optional call) yet`, after optionalValue learned to look through an arrow's expression body (#6c4swvt G4), which moved it off its refusal at 8:73 (adamic/no-optional-widening).

Mutant: the change undone, optionalMethodSignature finding nothing (`if true || call.QuestionDotToken == nil || call.Expression.Kind != ast.KindPropertyAccessExpression {`), on top of this change; status.json stays.
Command: GOFLAGS=-buildvcs=false go test -count=1 -v ./stage3/fixtures -run 'TestFixturesObjects$'
Result with the mutant: exit 1, red in objects on exactly TestFixturesObjects/objects/30_host_optional_method.a/stage0, "gap changed", got NotYet at 6:12 (a call through ?. (an optional call)) where Compiles is recorded. It ran as mutant 8 (ONLY=8) of the head's proof script, optcall-proof.sh, on Home in job 01M6WQVAGY45JY1YPT3FYSKQFP at bb7da386 (linux/amd64); restored, the tree was clean and the same command passes.
