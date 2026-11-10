Mutant evidence for objects fixture 16's stage-0 record in status.json.

What moved: an object pattern's default now lowers (#smqr4vn, destructuredDefault in internal/lower/collections.go), so Version.with's major, minor and patch defaults lower, and the fixture's first stop moves from the first defaulted name (15:13, "a destructured name that isn't plain") to prerelease = this.prerelease (18:13, "a destructured name held otherwise than its field"): prerelease is string | readonly string[], a union destructuring doesn't read yet, default or not. The fixture is unchanged.

Mutant: in internal/lower/collections.go's destructureFrom, `if tuple && declared.Initializer != nil {` made `if declared.Initializer != nil {`, on top of this change, so every object pattern's default stops again (as "a tuple element default"); the change's status.json stays.
Command: go test -p 1 -count=1 -timeout 60m ./stage3/fixtures -run 'TestFixturesObjects$' -v
Result with the mutant: exit 1, red in objects on exactly 16_version_defaults.a. Restored, checked with git diff --exit-code, the same command passes (Codex 01a1c62f, linux, at e2189100).

    fixtures_test.go:493: gap changed: update status.json and check the native output; got {Outcome:NotYet What:stage3/fixtures/objects/16_version_defaults.a:15:13: stage 0 can't lower a tuple element default yet Stdout:}; recorded {Outcome:NotYet What:stage3/fixtures/objects/16_version_defaults.a:18:13: stage 0 can't lower a destructured name held otherwise than its field yet Stdout:}
--- FAIL: TestFixturesObjects (48.95s)
    --- FAIL: TestFixturesObjects/objects (0.00s)
        --- FAIL: TestFixturesObjects/objects/16_version_defaults.a (0.32s)
            --- FAIL: TestFixturesObjects/objects/16_version_defaults.a/stage0 (0.00s)
