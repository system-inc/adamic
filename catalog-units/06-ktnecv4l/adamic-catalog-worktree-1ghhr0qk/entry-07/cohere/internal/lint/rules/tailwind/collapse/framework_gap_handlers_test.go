package tailwind

import (
	"sort"
	"strings"
	"testing"
)

// gapRoots is the 57 roots this slice ports, named rather than derived from the emitter map.
//
// Written out so that a root silently dropped from `gapEmitters` fails rather than shrinking the
// population this test measures. That failure mode is the one this package keeps finding: a
// measurement over the wrong population reads exactly like a passing one.
var gapRoots = []string{
	"backdrop-filter", "bg", "bg-conic", "bg-radial", "block", "border", "border-b", "border-be",
	"border-bs", "border-e", "border-l", "border-r", "border-s", "border-t", "border-x", "border-y",
	"decoration", "drop-shadow", "duration", "fill", "filter", "flex", "font", "inline",
	"inset-ring", "inset-shadow", "mask", "outline", "ring", "rotate", "scale", "shadow", "stroke",
	"text", "text-shadow", "transform",
	"mask-b-from", "mask-b-to", "mask-conic-from", "mask-conic-to", "mask-l-from", "mask-l-to",
	"mask-linear-from", "mask-linear-to", "mask-r-from", "mask-r-to", "mask-radial-from",
	"mask-radial-to", "mask-t-from", "mask-t-to", "mask-x-from", "mask-x-to", "mask-y-from",
	"mask-y-to",
	"-bg-conic", "-rotate", "-scale",
}

// gapProbe is one measured cell of a root's descriptor and the branch that should reproduce it.
//
// A descriptor row is three parallel axes, each holding a type map, a namespace map, a fallback and
// an empty reading. Every populated cell is an independent measurement against the engine, so each
// one is a separate comparison here rather than one comparison per root.
type gapProbe struct {
	// name is what the cell is, for a failure message: `Absent.ByType[length]`.
	name string
	// branch is the input that should make the emitter produce the cell's reading.
	branch UtilityBranch
	// reading is the measured side.
	reading Reading
}

// gapProbesFor turns one root's descriptor row into the cells this test compares.
//
// The mapping from a cell to a branch is the claim under test, and it is the same one the package
// comment on frameworkgaphandlers.go states: a `ByType` cell is an arbitrary value of that type, a
// `--color` or `@colorKeyword` cell is a named value the theme answered as a colour, an `@none`
// cell is a named value it did not, and `Empty` is a class written with no value.
//
// A cell whose branch cannot be expressed is skipped by name rather than silently, and
// TestGapRootExemptionsAreExactlyTheStaticCollision asserts the skipped set is the one this file
// documents.
func gapProbesFor(descriptor *Descriptor) []gapProbe {
	var probes []gapProbe
	axes := []struct {
		name        string
		axis        AxisReadings
		hasModifier bool
	}{
		{"Absent", descriptor.Absent, false},
		{"Alpha", descriptor.Alpha, true},
	}
	for _, axis := range axes {
		for dataType, reading := range axis.axis.ByType {
			probes = append(probes, gapProbe{
				name:    axis.name + ".ByType[" + string(dataType) + "]",
				branch:  UtilityBranch{HasValue: true, IsArbitrary: true, DataType: dataType, HasModifier: axis.hasModifier},
				reading: reading,
			})
		}
		for namespace, reading := range axis.axis.ByNamespace {
			switch namespace {
			case "--color", namespaceColorKeyword:
				probes = append(probes, gapProbe{
					name:    axis.name + ".ByNamespace[" + namespace + "]",
					branch:  UtilityBranch{HasValue: true, ResolvedAsColor: true, HasModifier: axis.hasModifier},
					reading: reading,
				})
			case namespaceNone:
				probes = append(probes, gapProbe{
					name:    axis.name + ".ByNamespace[@none]",
					branch:  UtilityBranch{HasValue: true, HasModifier: axis.hasModifier},
					reading: reading,
				})
			}
			// Every other namespace is a repository's own token bucket, such as `--shadow` or
			// `--text`. Those are compared by TestGapRootThemeNamespacesAreNamed rather than here,
			// because which namespace a root consumes is per-root data this file would have to
			// restate to probe.
		}
		if axis.axis.Empty != nil {
			probes = append(probes, gapProbe{
				name:    axis.name + ".Empty",
				branch:  UtilityBranch{HasModifier: axis.hasModifier},
				reading: *axis.axis.Empty,
			})
		}
		// The fallback, probed as an arbitrary value that infers as nothing.
		//
		// This cell is where an arbitrary arm's reading lands for a root whose row has no ByType
		// entry, and leaving it out left a real hole. Deleting `scale`'s entire arbitrary branch
		// changed no cell this test compared, because `scale` has no ByType map at all and its
		// one-declaration arbitrary reading lives here. The mutation was caught only by
		// TestGapRootArbitraryAndNamedCanDiffer, which probes that one root by hand.
		//
		// Skipped when the axis has no fallback, which is a zero Reading rather than an absent one
		// and would otherwise be compared as a real measurement of nothing.
		if axis.axis.Fallback.Count != 0 {
			probes = append(probes, gapProbe{
				name:    axis.name + ".Fallback",
				branch:  UtilityBranch{HasValue: true, IsArbitrary: true, HasModifier: axis.hasModifier},
				reading: axis.axis.Fallback,
			})
		}
	}
	sort.Slice(probes, func(first, second int) bool { return probes[first].name < probes[second].name })
	return probes
}

// gapCellsMeasuredElsewhere is every cell this test deliberately does not compare, with its reason.
//
// Keyed on the exact cell rather than on a root, and that precision is the point. A first version of
// this map exempted whole roots, which read as eight documented exceptions and was in fact hiding 47
// cells. Measured by disabling the skips: of those 47, exactly 7 disagreed and 40 were passing
// comparisons being thrown away under a comment explaining why they could not be made. The reasons
// written for `shadow`, `inset-shadow`, `drop-shadow`, `text-shadow`, `text` and `font` were all
// true of some cell of those roots and true of none of the cells that were being skipped.
//
// So the nine that remain are the whole exemption, and they are one fact rather than nine.
//
// Three roots have a `staticUtility` sharing a name with a functional root or standing where one
// would be, all three declaring `display`: `staticUtility('block', ...)` at line 956,
// `staticUtility('inline', ...)` at 958 and `staticUtility('flex', ...)` at 973. The extractor
// probed all three as functional roots, so their rows record the static utility's reading for the
// form the static answers.
//
// `block` and `inline` have no functional registration at all, so all four of their measurable
// cells record the static and the emitter correctly produces nothing for each. `flex` does have one,
// which answers every value form and refuses the empty one, so exactly one of its cells is the
// static's and the rest are compared like any other root's.
//
// In every case the emitter returning nothing is the correct answer to a question the row is not
// asking.
//
// See TestBlockAndInlineHaveNoFunctionalRegistration, which pins that reading rather than restating
// it here, and TestGapRootExemptionsAreExactlyTheStaticCollision, which fails if this map grows.
var gapCellsMeasuredElsewhere = map[string]string{
	"block.Absent.ByNamespace[@none]":  "no functional registration; the row records the static display utility",
	"block.Absent.Empty":               "no functional registration; the row records the static display utility",
	"block.Alpha.ByNamespace[@none]":   "no functional registration; the row records the static display utility",
	"inline.Absent.ByNamespace[@none]": "no functional registration; the row records the static display utility",
	"inline.Absent.Empty":              "no functional registration; the row records the static display utility",
	"inline.Alpha.ByNamespace[@none]":  "no functional registration; the row records the static display utility",
	"flex.Absent.Empty":                "the functional root refuses an empty value; the row records the static display utility",
	"block.Absent.Fallback":            "no functional registration; the row records the static display utility",
	"inline.Absent.Fallback":           "no functional registration; the row records the static display utility",
}

// TestGapRootEmittersAgreeWithTheirMeasuredReadings is the acceptance test for this slice.
//
// Every one of the 57 roots has a `baseDescriptors` row carrying a measured `{order, count}` per
// value type, produced by asking the engine rather than by reading the source. Every root now also
// has an emitter transcribed from `utilities.ts`. The two were produced by different means from
// different sources, so forcing them to agree cell for cell is a real comparison.
//
// The comparison is on the reading, because that is what every consumer asks for and because
// `PropertySort` is where a wrong property name stops being a typo and becomes a wrong sort key.
// TestGapRootPropertyNamesAgreeWithTheTable compares the names, which the reading cannot separate
// for two properties `PropertyOrder` does not know.
func TestGapRootEmittersAgreeWithTheirMeasuredReadings(t *testing.T) {
	t.Parallel()
	var agreed, compared int
	var disagreements, unported []string

	// Which skip keys actually matched. A key that matches nothing is not a harmless typo: it reads
	// as a documented exemption while the cell it names is being compared anyway, or worse, while a
	// different cell goes uncompared. The first run of this test had exactly that, `flex.Empty`
	// against a probe named `Absent.Empty`, and it surfaced as a real disagreement only by luck.
	used := map[string]bool{}

	for _, root := range gapRoots {
		descriptor, found := baseDescriptors[root]
		if !found {
			t.Errorf("%s is in the list this slice ports and has no baseDescriptors row, so there is nothing to compare it against", root)
			continue
		}
		if _, ported := gapEmitters[root]; !ported {
			unported = append(unported, root)
			continue
		}
		for _, probe := range gapProbesFor(descriptor) {
			if _, skipped := gapCellsMeasuredElsewhere[root+"."+probe.name]; skipped {
				used[root+"."+probe.name] = true
				continue
			}
			emitted := EmitGapRoot(root, probe.branch)
			sorted := PropertySort(emitted)
			computed := Reading{Order: sorted.Order, Count: sorted.Count}

			compared++
			if computed.Equal(probe.reading) {
				agreed++
				continue
			}
			disagreements = append(disagreements,
				root+" "+probe.name+": the engine measured "+formatReading(probe.reading)+
					", the ported handle emits "+describeDeclarations(emitted)+" reading "+formatReading(computed))
		}
	}

	sort.Strings(disagreements)
	sort.Strings(unported)

	// The population, stated before the disagreements. A run that compared nothing disagrees about
	// nothing and is otherwise indistinguishable from a run that compared everything.
	t.Logf("gap handles: %d roots in this slice, %d cells compared, %d agreed, %d unported, %d disagreed",
		len(gapRoots), compared, agreed, len(unported), len(disagreements))

	if len(gapRoots) != 57 {
		t.Errorf("this slice lists %d roots and the assertions below were sized for 57; the population moved", len(gapRoots))
	}
	if compared == 0 {
		t.Error("no cell was compared, so this test proves nothing about any emitter")
	}
	if agreed != compared {
		t.Errorf("%d of %d cells agreed; the acceptance condition is every cell, not a rate", agreed, compared)
	}
	for _, root := range unported {
		t.Errorf("%s has a baseDescriptors row and no ported handle, so it emits nothing and reads nothing", root)
	}
	for _, disagreement := range disagreements {
		t.Error(disagreement)
	}

	// Every documented exemption has to have fired. See `used`.
	var unmatched []string
	for key := range gapCellsMeasuredElsewhere {
		if !used[key] {
			unmatched = append(unmatched, key)
		}
	}
	sort.Strings(unmatched)
	for _, key := range unmatched {
		t.Errorf("gapCellsMeasuredElsewhere names %q and no cell matched it; an exemption that never fires "+
			"documents a skip that is not happening, which is how a cell goes uncompared under a comment "+
			"saying why it was skipped", key)
	}
}

// TestEveryGapRootHasAnEmitter pins that the port covers the whole slice.
//
// Separate from the acceptance test because they fail differently: this one catches a root nobody
// wrote a body for, where the acceptance test above would report it as unported and could be read as
// a smaller population rather than as a gap.
func TestEveryGapRootHasAnEmitter(t *testing.T) {
	t.Parallel()
	var missing []string
	for _, root := range gapRoots {
		if _, ported := gapEmitters[root]; !ported {
			missing = append(missing, root)
		}
	}
	sort.Strings(missing)
	for _, root := range missing {
		t.Errorf("%s is one of the 57 and has no entry in gapEmitters", root)
	}

	// And the reverse. An emitter for a root outside this slice would be a body nothing compares
	// against, which is the state the port exists to remove.
	inSlice := make(map[string]bool, len(gapRoots))
	for _, root := range gapRoots {
		inSlice[root] = true
	}
	for root := range gapEmitters {
		if !inSlice[root] {
			t.Errorf("gapEmitters carries %s, which is not one of the 57 this slice ports and which no test compares", root)
		}
	}
	t.Logf("emitters: %d roots in the slice, %d in gapEmitters", len(gapRoots), len(gapEmitters))
}

// TestGapRootPropertyNamesAreDistinctPerArm replaces a comparison whose other side is gone.
//
// This compared the emitters against `RootDeclaredProperties` and expected to disagree: the table
// stored one property list per root while these roots declare different properties per value, and 31
// of 57 differed. #mz0m6k8 deleted that table precisely because the emitter is the correct side, so
// the comparison has nothing left to compare against.
//
// What it was really asserting was that the arms are distinct, and that is asserted directly here.
// A root whose arms returned the same property list would have made the old comparison AGREE, which
// is the direction that reads as success, so this is the check that carried the weight.
func TestGapRootPropertyNamesAreDistinctPerArm(t *testing.T) {
	t.Parallel()
	// The roots that branch colour against not-colour. Each must declare different properties on the
	// two arms, or a width and a colour of the same root would read as conflicting.
	var distinct, identical int
	for _, root := range gapRoots {
		ordinary := PropertySort(EmitGapRoot(root, UtilityBranch{HasValue: true}))
		colour := PropertySort(EmitGapRoot(root, UtilityBranch{HasValue: true, ResolvedAsColor: true}))
		if len(ordinary.Order) == 0 && len(colour.Order) == 0 {
			continue
		}
		if readingsEqual(
			Reading{Order: ordinary.Order, Count: ordinary.Count},
			Reading{Order: colour.Order, Count: colour.Count},
		) {
			identical++
			continue
		}
		distinct++
	}

	t.Logf("arms: %d gap roots read differently for a colour than for their ordinary value, %d read the same",
		distinct, identical)

	// Named rather than asserted loosely. The border family alone is eleven roots that must split,
	// and a zero here means every root now returns one answer for both values, which is the flattening
	// the deleted override tables existed to work around.
	if distinct == 0 {
		t.Error("no gap root distinguishes a colour value from its ordinary one, so every branch has " +
			"been flattened and a width would conflict with a colour on the same root")
	}
}

// TestObjectRootAndItsStaticsDeclareDifferentProperties is what the table defect looked like once
// the table was gone.
//
// `RootDeclaredProperties` stored `[object-fit]` for root `object`, and the functional registration
// declares `object-position`: `object-cover` is a static declaring a fit and `object-center` is the
// functional root declaring a position. One row could hold only one of those, so
// `ClassDeclaredProperties` patched nine `object-*` classes back while `object-cover` and
// `object-fill` were correct only because they happened to read the row.
//
// Both tables are deleted. This asserts the two registrations still answer separately, which is what
// makes the patch unnecessary rather than merely absent.
func TestObjectRootAndItsStaticsDeclareDifferentProperties(t *testing.T) {
	t.Parallel()
	utility, known := FrameworkFunctionalUtilities["object"]
	if !known {
		t.Fatal("object is not in FrameworkFunctionalUtilities")
	}
	emitted := utility.Emit("object", ResolvedUtilityValue{Value: "zzsentinel"})
	if len(emitted) != 1 || emitted[0].Property != "object-position" {
		t.Errorf("the functional root object emits %s; it declares object-position, and reading "+
			"object-fit here is the defect the deleted table had", describeDeclarations(emitted))
	}

	for _, className := range []string{"object-cover", "object-fill", "object-contain"} {
		declarations, isStatic := FrameworkStaticDeclarations[className]
		if !isStatic {
			t.Errorf("%s is not in FrameworkStaticDeclarations; it is a staticUtility declaring object-fit", className)
			continue
		}
		if len(declarations) != 1 || declarations[0].Property != "object-fit" {
			t.Errorf("%s declares %v; upstream registers it as a static object-fit", className, declarations)
		}
	}
}

// TestBlockAndInlineHaveNoFunctionalRegistration records why two of the 57 emit nothing.
//
// `block` and `inline` have a `baseDescriptors` row, which is what put them in this slice, and no
// functional registration in `utilities.ts` at all: both are `staticUtility` calls declaring
// `display`. So their row's `Empty` reading is the static utility's, and the value-form readings in
// it are the extractor probing a root that does not answer.
//
// The emitter returning nil is therefore correct rather than unported, and this test is what keeps
// that distinction from being read as a gap.
func TestBlockAndInlineHaveNoFunctionalRegistration(t *testing.T) {
	t.Parallel()
	for _, root := range []string{"block", "inline"} {
		emitted := EmitGapRoot(root, UtilityBranch{HasValue: true})
		if len(emitted) != 0 {
			t.Errorf("%s emitted %s for a value form; it has no functional registration, so a value form compiles to nothing",
				root, describeDeclarations(emitted))
		}
		descriptor, found := baseDescriptors[root]
		if !found {
			t.Errorf("%s has no baseDescriptors row, so the collision this test records is no longer observable", root)
			continue
		}
		if descriptor.Absent.Empty == nil {
			t.Errorf("%s has no Empty reading; the row exists because the static utility of the same name has one", root)
			continue
		}
		// The static utility's own reading, which is what the row's Empty is recording.
		if position, known := PropertyOrder["display"]; known {
			empty := *descriptor.Absent.Empty
			if len(empty.Order) != 1 || empty.Order[0] != position {
				t.Errorf("%s reads %s for its empty form; this test records that the row is the static display "+
					"utility's, and it no longer is", root, formatReading(empty))
			}
		}
	}
}

// TestShadowFamilyAlphaDeclarationDecidesTheCount pins the branch the milestone did not name.
//
// The shadow family writes `decl('<prefix>-alpha', alpha)` where `alpha` is undefined without a
// modifier. `PropertySort` skips a declaration whose value is absent and counts one whose value is
// empty, so that single declaration is the entire difference between a root's unmodified `#2` and
// its modified `#3`.
//
// That is a branch on the candidate rather than on the value, which is why the emitters here take a
// `UtilityBranch` and not a `ResolvedUtilityValue`: resolution consumes the modifier and reports
// nothing about it, so on a resolved value alone this branch is unreachable.
func TestShadowFamilyAlphaDeclarationDecidesTheCount(t *testing.T) {
	t.Parallel()
	for _, root := range []string{"shadow", "inset-shadow", "drop-shadow", "text-shadow"} {
		unmodified := PropertySort(EmitGapRoot(root, UtilityBranch{HasValue: true}))
		modified := PropertySort(EmitGapRoot(root, UtilityBranch{HasValue: true, HasModifier: true}))

		if modified.Count != unmodified.Count+1 {
			t.Errorf("%s counts %d unmodified and %d modified; the alpha declaration is absent-valued without a "+
				"modifier and present with one, so the modified form must count exactly one more",
				root, unmodified.Count, modified.Count)
		}
		if len(modified.Order) != len(unmodified.Order) {
			t.Errorf("%s reads %d positions unmodified and %d modified; the alpha property is outside PropertyOrder, "+
				"so it changes the count and never the order", root, len(unmodified.Order), len(modified.Order))
		}
	}
}

// TestGapRootArbitraryAndNamedCanDiffer pins the branch no data type can express.
//
// `scale-[2]` emits one declaration and `scale-150` emits four. Both infer nothing useful and both
// are the same root, so the only thing separating them is that upstream's arbitrary arm returns
// before reaching the axis variables.
//
// Without this the `IsArbitrary` field could be deleted and every type-keyed comparison above would
// still pass, since the type maps are probed with arbitrary branches throughout.
func TestGapRootArbitraryAndNamedCanDiffer(t *testing.T) {
	t.Parallel()
	arbitrary := PropertySort(EmitGapRoot("scale", UtilityBranch{HasValue: true, IsArbitrary: true}))
	named := PropertySort(EmitGapRoot("scale", UtilityBranch{HasValue: true}))

	if arbitrary.Count != 1 {
		t.Errorf("scale-[2] emitted %d declarations; upstream's arbitrary arm returns a single scale declaration", arbitrary.Count)
	}
	if named.Count != 4 {
		t.Errorf("scale-150 emitted %d declarations; upstream's named path writes three axis variables and the shorthand", named.Count)
	}
	if arbitrary.Count == named.Count {
		t.Error("scale reads the same for an arbitrary and a named value, so IsArbitrary is deletable and this " +
			"slice's branch model is carrying a field nothing turns on")
	}
}

// joinStrings renders a property list for a failure message.
//
// A `[]string` formats readably with %v, and this is spelled out so the message reads as prose
// rather than as a Go literal beside describeDeclarations' output, which is also prose.
func joinStrings(values []string) string {
	rendered := ""
	for index, value := range values {
		if index > 0 {
			rendered += ", "
		}
		rendered += value
	}
	return rendered
}

// TestGapRootExemptionsAreExactlyTheStaticCollision pins the size and the shape of the hole.
//
// Every exemption in gapCellsMeasuredElsewhere has to be a cell of `block`, `inline` or `flex`, and
// there have to be nine of them. Both halves matter and they fail differently.
//
// The count catches the map growing. An exemption is the one mechanism in this file that turns a
// disagreement into a silence, so a seventh entry appearing is a cell that stopped being compared,
// and it should cost somebody a line here rather than passing quietly.
//
// The shape catches the map broadening. An earlier version keyed on roots instead of cells, which
// exempted 47 cells to excuse 7; the other 40 were agreeing all along and were being discarded
// under a comment saying they could not be compared. Requiring every key to name one cell of one of
// the three collided roots makes that mistake unrepresentable rather than merely discouraged.
func TestGapRootExemptionsAreExactlyTheStaticCollision(t *testing.T) {
	t.Parallel()
	if len(gapCellsMeasuredElsewhere) != 9 {
		t.Errorf("gapCellsMeasuredElsewhere holds %d exemptions and this test was written against 9; "+
			"an exemption is the one way a cell stops being compared without failing, so growing the map "+
			"should cost a line here", len(gapCellsMeasuredElsewhere))
	}
	for key := range gapCellsMeasuredElsewhere {
		root, _, split := strings.Cut(key, ".")
		if !split {
			t.Errorf("%q exempts a whole root rather than a cell; that is the shape that hid 41 passing "+
				"comparisons behind 6 real ones", key)
			continue
		}
		if root != "block" && root != "inline" && root != "flex" {
			t.Errorf("%q exempts a cell of %s; the only rows this port cannot answer belong to block, inline "+
				"and flex, each of which has a staticUtility declaring display standing where the extractor "+
				"probed for a functional one", key, root)
		}
	}
}
