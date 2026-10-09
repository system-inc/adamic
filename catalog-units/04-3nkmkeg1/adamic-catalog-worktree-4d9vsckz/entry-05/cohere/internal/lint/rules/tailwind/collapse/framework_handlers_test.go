package tailwind

import (
	"sort"
	"testing"
)

// TestEmittedDeclarationsAgreeWithTheMeasuredProperty is the acceptance test for the ported handles.
//
// Every row of `FrameworkFunctionalUtilities` carries a `Property` that was measured against the
// engine before any handle body was ported, and every root now carries an emitter transcribed from
// `utilities.ts`. The two were produced by different means from different sources, so forcing them
// to agree class for class is a real comparison rather than a table checking itself.
//
// The comparison is on the reading rather than on the property string, because the reading is what
// every consumer of this package actually asks for, and because `PropertySort` is where a wrong
// property name stops being a typo and starts being a wrong sort key. A property absent from
// `PropertyOrder` contributes no position, so two different unknown properties would compare equal
// as strings and equal as readings alike; TestEmittedPropertyNamesAreTheMeasuredOnes is what covers
// those six roots.
func TestEmittedDeclarationsAgreeWithTheMeasuredProperty(t *testing.T) {
	t.Parallel()
	// A sentinel rather than a resolved value, to make the point that the emitters are being asked
	// for a shape. If any of these bodies started reading its value, the reading would be computed
	// from a string that is not CSS, and this test would be the place it showed.
	const sentinel = "zzsentinel"

	var agreed int
	var disagreements, unported []string

	for root, utility := range FrameworkFunctionalUtilities {
		emitted := utility.Emit(root, ResolvedUtilityValue{Value: sentinel})
		if len(emitted) == 0 {
			unported = append(unported, root)
			continue
		}

		measured := PropertySort([]*Node{Declaration(utility.Property, sentinel)})
		computed := PropertySort(emitted)

		measuredReading := Reading{Order: measured.Order, Count: measured.Count}
		computedReading := Reading{Order: computed.Order, Count: computed.Count}
		if readingsEqual(measuredReading, computedReading) {
			agreed++
			continue
		}
		disagreements = append(disagreements,
			root+": the measured property "+utility.Property+" reads "+formatReading(measuredReading)+
				", the ported handle emits "+emitted[0].Property+" reading "+formatReading(computedReading))
	}

	sort.Strings(disagreements)
	sort.Strings(unported)

	// The population, stated before the disagreements. A run that compared nothing disagrees about
	// nothing, and would otherwise be indistinguishable from a run that compared everything.
	t.Logf("ported handles: %d roots in the table, %d compared, %d agreed, %d unported, %d disagreed",
		len(FrameworkFunctionalUtilities), agreed+len(disagreements), agreed, len(unported), len(disagreements))

	if len(FrameworkFunctionalUtilities) != 33 {
		t.Errorf("the table holds %d roots and this comparison was written against 33; the population moved and the assertions below were sized for the old one", len(FrameworkFunctionalUtilities))
	}
	if agreed != 33 {
		t.Errorf("%d of 33 roots agreed; the acceptance condition is every root, not a rate", agreed)
	}
	for _, root := range unported {
		t.Errorf("%s is in the table and has no ported handle, so it emits nothing and reads nothing", root)
	}
	for _, disagreement := range disagreements {
		t.Error(disagreement)
	}
}

// TestEmittedPropertyNamesAreTheMeasuredOnes compares the strings rather than the readings.
//
// The reading comparison above is the one that matters to consumers, and it is blind to a
// distinction it cannot see: two properties that `PropertyOrder` does not know both read `[]#1`. Two
// of these roots declare such a property, `perspective` and `perspective-origin`, so a mutation
// swapping one of their emitters for the other survives the reading test and is caught here. The
// count is measured by the test itself rather than asserted from this comment, which is why the
// assertion below is that it is not zero rather than that it is two.
//
// Both tests are kept rather than folded together. The string comparison alone would be the weaker
// test, since it never runs `PropertySort` and so never touches the code path the port exists to
// feed.
func TestEmittedPropertyNamesAreTheMeasuredOnes(t *testing.T) {
	t.Parallel()
	var compared, unknownToPropertyOrder int
	for root, utility := range FrameworkFunctionalUtilities {
		emitted := utility.Emit(root, ResolvedUtilityValue{Value: "zzsentinel"})
		if len(emitted) != 1 {
			t.Errorf("%s emitted %d declarations; every root in this table emits exactly one", root, len(emitted))
			continue
		}
		compared++
		if _, known := PropertyOrder[utility.Property]; !known {
			unknownToPropertyOrder++
		}
		if emitted[0].Property != utility.Property {
			t.Errorf("%s: the ported handle declares %q, the measured table says %q", root, emitted[0].Property, utility.Property)
		}
		if !emitted[0].ValuePresent {
			t.Errorf("%s emitted a declaration with an absent value, which PropertySort skips entirely, so the root would read nothing", root)
		}
	}

	t.Logf("property names: %d roots compared, %d declare a property PropertyOrder does not know", compared, unknownToPropertyOrder)
	if compared != 33 {
		t.Errorf("compared %d roots of 33", compared)
	}
	// Named rather than asserted loosely, because it is the reason this test exists beside the
	// reading one. If it reaches zero, every root reads distinguishably and this test is redundant.
	if unknownToPropertyOrder == 0 {
		t.Error("no root declares a property outside PropertyOrder, so the reading comparison already covers every emitter and this test is measuring nothing")
	}
}

// TestStaticValuesEmitTheirOwnProperty covers the branch the ordinary path never reaches.
//
// A static value is answered before the handle body runs and carries its own property, so the
// emitter looks it up rather than inheriting the root's. Across this wave no static entry diverges
// from its root, which is asserted in frameworkutility_test.go, so this test cannot show the lookup
// mattering on real data. What it can show is that the lookup runs at all: a static name resolves
// through the static emitter and an unrecognised one falls through to the handle body.
func TestStaticValuesEmitTheirOwnProperty(t *testing.T) {
	t.Parallel()
	utility, known := FrameworkFunctionalUtilities["order"]
	if !known {
		t.Fatal("order is not in the table")
	}

	static := utility.Emit("order", ResolvedUtilityValue{IsStaticValue: true, StaticValueName: "first", Value: "ignored"})
	if len(static) != 1 || static[0].Property != "order" || static[0].Value != "-9999" {
		t.Errorf("order-first emitted %s; upstream stores a single order declaration of -9999 under that name and returns it without the handle body running", describeDeclarations(static))
	}

	// An unrecognised static name is the fall-through, and it must reach the handle body rather than
	// emit nothing. Upstream cannot produce this state, since resolution only sets IsStaticValue for
	// a name the map holds; it is asserted here so the fall-through is not a silent nil.
	fallen := utility.Emit("order", ResolvedUtilityValue{IsStaticValue: true, StaticValueName: "zznotastatic", Value: "7"})
	if len(fallen) != 1 || fallen[0].Property != "order" || fallen[0].Value != "7" {
		t.Errorf("an unrecognised static name emitted %s; it should fall through to the handle body", describeDeclarations(fallen))
	}
}

// describeDeclarations renders an emitted list as property and value pairs.
//
// A `[]*Node` formats as a slice of pointers, so a failure printing one names nothing a reader can
// act on. Both halves are shown because the static branch is the one place in this file where the
// value is asserted rather than passed through.
func describeDeclarations(nodes []*Node) string {
	if len(nodes) == 0 {
		return "nothing"
	}
	rendered := ""
	for index, node := range nodes {
		if index > 0 {
			rendered += ", "
		}
		rendered += node.Property + ": " + node.Value
	}
	return rendered
}

// TestUnportedRootEmitsNothing pins the deliberate absence of a fallback.
//
// `Emit` returns nil for a root with no emitter rather than rebuilding a declaration from
// `Property`. That choice is what makes the acceptance test able to fail: with a fallback, a root
// whose emitter was never written would agree with the table perfectly, because both sides would be
// reading the same field.
func TestUnportedRootEmitsNothing(t *testing.T) {
	t.Parallel()
	utility := FrameworkFunctionalUtilities["cursor"]
	// Reported through describeDeclarations rather than with %v, because a `[]*Node` formats as a
	// pointer and a failure that names no property tells the next reader nothing about what fired.
	if nodes := utility.Emit("zznotaroot", ResolvedUtilityValue{Value: "pointer"}); nodes != nil {
		t.Errorf("an unported root emitted %s; a fallback to Property would make the acceptance test compare a field against itself", describeDeclarations(nodes))
	}
}
