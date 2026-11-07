package lower

import (
	"errors"
	"strings"
	"testing"
)

// tsc's Version constructor (semver.ts:60) writes a parsed version's fields into its own wider
// parameters with an object pattern, ({ major, minor, prerelease } = result): each field is read by
// name and boxed into its parameter. Held to Node natively and through the JavaScript backend.
func TestObjectDestructuringAssignmentAgrees(test *testing.T) {
	test.Parallel()
	lowersAndAgreesWithNodeNative(test, "ObjectDestructuringAssignmentWiden")
}

// A field a class serves by a getter is read through the getter, once per pattern, never from a
// slot: the count shows each read ran.
func TestObjectDestructuringAssignmentReadsAGetter(test *testing.T) {
	test.Parallel()
	lowersAndAgreesWithNodeNative(test, "ObjectDestructuringAssignmentGetter")
}

// The value is evaluated whole before any target is written, so a swap through a held object
// swaps.
func TestObjectDestructuringAssignmentSwaps(test *testing.T) {
	test.Parallel()
	lowersAndAgreesWithNodeNative(test, "ObjectDestructuringAssignmentSwap")
}

// A literal written in the assignment itself is stored as the pattern it goes into, number | string,
// while its own type says number, so reading its fields by that type read a box as a number on
// native ("two 5.27e-310"). It isn't lowered yet.
func TestObjectDestructuringAssignmentOfALiteralIsNotYet(test *testing.T) {
	test.Parallel()
	_, err := lowerSource(test, `let first: number | string = 1;
let second: number | string = 'two';
({ first, second } = { first: second, second: first });
console.log(`+"`${first} ${second}`"+`);
`)
	var notYet *NotYet
	if !errors.As(err, &notYet) || !strings.Contains(
		err.Error(),
		"main.a:3:22: stage 0 can't lower destructuring a value made in the assignment itself yet",
	) {
		test.Errorf("got %v, want the literal's NotYet", err)
	}
}

// A held source's field narrowed by an assignment before the pattern (held.x = 5, so the checker
// reads held.x as number) is stored as its declared number | string. A field held in a union slot
// isn't read by a pattern yet, so the program stops there and never reads the box as the narrowed
// number (#fsb6n82's class, Lower's probe; Node prints 5 c).
func TestObjectDestructuringAssignmentOfAUnionFieldIsNotYet(test *testing.T) {
	test.Parallel()
	_, err := lowerSource(test, `let first: number | string = 1;
let second: number | string = 'b';
const held: { x: number | string; y: number | string } = { x: 2, y: 'c' };
held.x = 5;
({ x: first, y: second } = held);
console.log(`+"`${first} ${second}`"+`);
`)
	var notYet *NotYet
	if !errors.As(err, &notYet) || !strings.Contains(
		err.Error(),
		"main.a:5:4: stage 0 can't lower a destructured field of type string | number yet",
	) {
		test.Errorf("got %v, want the union field's NotYet", err)
	}
}
