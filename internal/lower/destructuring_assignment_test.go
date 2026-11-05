package lower

import "testing"

// tsc's Version constructor (semver.ts:60) writes a parsed version's fields into its own wider
// parameters with an object pattern, ({ major, minor, prerelease } = result): each field is read by
// name and boxed into its parameter. Held to Node natively and through the JavaScript backend.
func TestObjectDestructuringAssignmentAgrees(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNodeNative(t, "ObjectDestructuringAssignmentWiden")
}
