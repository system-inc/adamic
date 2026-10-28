Mutant evidence for assertions fixture 19's stage-0 record in status.json.

What moved: this change splits the checked field alias NotYet into its three cases (internal/lower/interface_cast.go, checkedFieldAliasGap). Fixture 19 (tsc's identifier kind read through a checked view) stops at the same node, 19_identifier_kind.a:25:13, and its message moves from "a checked field alias requiring an optional, accessor, or representation conversion" to the optional case, "a checked field alias of an optional field or through ?.". The fixture is unchanged.

Mutant: the change's lowering undone. internal/lower/interface_cast.go goes back to its parent's bytes (309cfed694^); the change's status.json stays.
Command: go test -p 1 ./stage3/fixtures -count=1 -timeout 20m -run 'TestFixturesAssertions$' -v
Result with the mutant: exit 1, red on exactly 19_identifier_kind.a. Restored, the same command passes, 65 assertions subtests green (Golem, linux/amd64, under the suite lock, Oct 27).

    fixtures_test.go:472: gap changed: update status.json and check the native output; got {Outcome:NotYet What:stage3/fixtures/assertions/19_identifier_kind.a:25:13: stage 0 can't lower a checked field alias requiring an optional, accessor, or representation conversion yet Stdout:}; recorded {Outcome:NotYet What:stage3/fixtures/assertions/19_identifier_kind.a:25:13: stage 0 can't lower a checked field alias of an optional field or through ?. yet Stdout:}
--- FAIL: TestFixturesAssertions (18.10s)
    --- FAIL: TestFixturesAssertions/assertions (0.03s)
        --- FAIL: TestFixturesAssertions/assertions/19_identifier_kind.a (0.84s)
            --- FAIL: TestFixturesAssertions/assertions/19_identifier_kind.a/stage0 (0.00s)
