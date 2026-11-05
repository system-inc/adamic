package lower

import (
	"errors"
	"strings"
	"testing"
)

// tsc's Version constructor (semver.ts:60) writes a parsed version's fields into its own wider
// parameters with an object pattern, ({ major, minor, prerelease } = result): each field is read by
// name and boxed into its parameter. Held to Node natively and through the JavaScript backend.
func TestObjectDestructuringAssignmentAgrees(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNodeNative(t, "ObjectDestructuringAssignmentWiden")
}

// A field a class serves by a getter is read through the getter, once per pattern, never from a
// slot: the count shows each read ran.
func TestObjectDestructuringAssignmentReadsAGetter(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNodeNative(t, "ObjectDestructuringAssignmentGetter")
}

// The value is evaluated whole before any target is written, so a swap through a held object
// swaps.
func TestObjectDestructuringAssignmentSwaps(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNodeNative(t, "ObjectDestructuringAssignmentSwap")
}

// A literal written in the assignment itself is stored as the pattern it goes into, number | string,
// while its own type says number, so reading its fields by that type read a box as a number on
// native ("two 5.27e-310"). It isn't lowered yet.
func TestObjectDestructuringAssignmentOfALiteralIsNotYet(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `let first: number | string = 1;
let second: number | string = 'two';
({ first, second } = { first: second, second: first });
console.log(`+"`${first} ${second}`"+`);
`)
	var notYet *NotYet
	if !errors.As(err, &notYet) || !strings.Contains(err.Error(), "main.a:3:22: stage 0 can't lower destructuring a value made in the assignment itself yet") {
		t.Errorf("got %v, want the literal's NotYet", err)
	}
}
