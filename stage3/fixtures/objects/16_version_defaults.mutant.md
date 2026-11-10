Mutant evidence for objects fixture 16's stage-0 record in status.json.

What moved: an object pattern's default now lowers (#smqr4vn, destructuredDefault in internal/lower/collections.go), so Version.with's major, minor and patch defaults lower, and the fixture's first stop moves from the first defaulted name (15:13, "a destructured name that isn't plain") to prerelease = this.prerelease (18:13, "a destructured name held otherwise than its field"): prerelease is string | readonly string[], a union destructuring doesn't read yet, default or not. The fixture is unchanged.

Mutant: in internal/lower/collections.go's destructureFrom, `if tuple && declared.Initializer != nil {` made `if declared.Initializer != nil {`, on top of this change, so every object pattern's default stops again (as "a tuple element default"); the change's status.json stays.
Command: go test -p 1 -count=1 -timeout 60m ./stage3/fixtures -run 'TestFixturesObjects$' -v
Result with the mutant: MEASURED_RESULT
