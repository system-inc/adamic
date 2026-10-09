package tailwind

import (
	"fmt"
	"sort"
	"strings"
	"testing"
)

// Whether the descriptor rows are still carrying something the ported handle bodies cannot compute.
//
// `baseDescriptors` is one of the two tables in this package that is not Tailwind's own data, and its
// stated reason is that a root's reading partitions on the value's resolved type while no
// registration shape carries both halves. That was true when it was written and the emitting half was
// not ported. It is ported now, across three slices and 242 roots, so the claim is measurable rather
// than argued and this measures it.
//
// # What a cell is, and how the branch that reaches it is built
//
// A row holds a type map, a namespace map, a fallback and an empty reading, per modifier axis. Each
// is an independent measurement, so this compares cells rather than roots, the way
// frameworkgaphandlers_test.go does.
//
// The branch per cell is read off `Lookup`'s own precedence rather than guessed, which four earlier
// attempts at this measurement got wrong in four different ways and every one of them looked like an
// emitter defect:
//
//   - A `ByType` cell is reached by an arbitrary value whose inferred type is the key, so the branch
//     is arbitrary. Sending a named value instead put `bg`, `text` and `font` on the wrong arm.
//   - A `--color` or `@colorKeyword` cell is reached through theme resolution rather than inference,
//     so `ResolvedAsColor` is what selects it, not `DataTypeColor`. Missing that reported 31 cells.
//   - A `Fallback` is a named value that matched no namespace and inferred as nothing. Whether that
//     lands on the colour arm is the row's own claim: a root whose fallback equals its `--color`
//     reading resolves an unmatched name as a colour, which is the eleven-root border family.
//   - `scale` has an empty type list, so nothing infers and `Lookup` returns `Fallback` for an
//     arbitrary value too. Its fallback is the arbitrary arm's reading rather than the named one.
//
// Each of those is a fact about how a class reaches a cell. Reconstructing them is the work this
// measurement is, and getting one wrong reads exactly like a port that cannot answer.
func TestDescriptorRowsAgreeWithTheEmitters(t *testing.T) {
	t.Parallel()
	type counter struct{ compared, agreed int }
	counts := map[string]*counter{"ByType": {}, "ByNamespace": {}, "Fallback": {}, "Empty": {}}
	var disagreements []string

	record := func(kind, label string, want Reading, nodes []*Node) {
		if len(nodes) == 0 {
			return
		}
		counts[kind].compared++
		got := PropertySort(nodes)
		if fmt.Sprint(got) == fmt.Sprint(want) {
			counts[kind].agreed++
			return
		}
		disagreements = append(disagreements, fmt.Sprintf("%s %s: row %v, emitters %v", kind, label, want, got))
	}

	for root, descriptor := range baseDescriptors {
		if _, isGapRoot := gapEmitters[root]; !isGapRoot {
			continue
		}
		// Only the unmodified axis is compared here. The Alpha and Themed axes are the same rows
		// read through a modifier, and `TestModifierAxesAgreeWithTheEmitters` below covers them.
		axis := descriptor.Absent

		for dataType, want := range axis.ByType {
			branch := UtilityBranch{HasValue: true, IsArbitrary: true, DataType: dataType}
			branch.ResolvedAsColor = dataType == DataTypeColor
			record("ByType", root+"/"+string(dataType), want, EmitGapRoot(root, branch))
		}

		for namespace, want := range axis.ByNamespace {
			branch := UtilityBranch{HasValue: true, ResolvedNamespace: namespace}
			branch.ResolvedAsColor = namespace == "--color" || namespace == namespaceColorKeyword
			if namespace == namespaceNone {
				branch.ResolvedNamespace = ""
			}
			record("ByNamespace", root+"/"+namespace, want, EmitGapRoot(root, branch))
		}

		record("Fallback", root, axis.Fallback, EmitGapRoot(root, fallbackBranchFor(descriptor)))

		if axis.Empty != nil {
			record("Empty", root, *axis.Empty, EmitGapRoot(root, UtilityBranch{}))
		}
	}

	sort.Strings(disagreements)
	var totalCompared, totalAgreed int
	for _, kind := range []string{"ByType", "ByNamespace", "Fallback", "Empty"} {
		totalCompared += counts[kind].compared
		totalAgreed += counts[kind].agreed
		t.Logf("%-12s %3d compared, %3d agreed", kind, counts[kind].compared, counts[kind].agreed)
	}
	t.Logf("descriptor cells: %d compared, %d agreed, %d disagreed", totalCompared, totalAgreed, totalCompared-totalAgreed)

	for _, disagreement := range disagreements {
		if descriptorCellsTheEmittersCannotReach[strings.SplitN(disagreement, ":", 2)[0]] {
			continue
		}
		t.Errorf("%s", disagreement)
	}

	if totalCompared == 0 {
		t.Fatal("no cell was compared, so this test measured nothing")
	}
}

// fallbackBranchFor builds the branch that reaches a row's `Fallback`.
//
// A fallback is what a named value gets when it matched no namespace and inferred as nothing. Two
// facts about the row decide which arm that is, and both are read off the row rather than assumed.
//
// A root whose fallback equals its `--color` reading resolves an unmatched name as a colour. Upstream
// that is the colour lookup running before the width lookup and returning, so the eleven border roots
// read `border-notacolor` as a colour. Eleven roots depend on this and reporting them as
// disagreements was the third wrong version of this measurement.
//
// A root with an empty type list has nothing to infer against, so `Lookup` returns `Fallback` for an
// arbitrary value as well as a named one, and the arbitrary arm is what produced the stored reading.
// `scale` and `-scale` are the two.
func fallbackBranchFor(descriptor *Descriptor) UtilityBranch {
	branch := UtilityBranch{HasValue: true}

	if colorReading, hasColor := descriptor.Absent.ByNamespace["--color"]; hasColor {
		branch.ResolvedAsColor = fmt.Sprint(colorReading) == fmt.Sprint(descriptor.Absent.Fallback)
	}
	if len(descriptor.TypeList) == 0 {
		branch.IsArbitrary = true
	}
	return branch
}

// Cells the emitters answer differently from the row, each measured rather than waved through.
//
// One entry, and it is dead data rather than a fact the emitters lack.
//
// `text/--drop` stores `{[] 1}`. The generator reached it by compiling `text-shadow-lg`, whose value
// is a key in the `--drop` namespace, and filing the result under root `text`. The parser reads that
// class as root `text-shadow`, which has its own descriptor row whose fallback is `{[] 1}`, the same
// reading. So the cell is a second copy of a row that already exists, attributed to the wrong root.
//
// Measured: of the seven keys the `--drop` namespace holds in this repository, spelled as `text-*`
// classes, 0 parse as root `text` and 7 parse as root `text-shadow`. No class can reach this cell
// through `Lookup`'s precedence, so no answer depends on it and the emitters disagreeing with it
// changes nothing.
//
// Deleting it belongs to whoever owns `generate_descriptor_base`, since the row is printed rather
// than written, and the same probe would reproduce it. Recorded here with the measurement so the
// disagreement is a known-dead cell rather than an unexplained one.
//
// An entry here is a claim that a cell is genuinely unreachable, not a way to quiet a failure. The
// test below asserts each one still disagrees, so an entry that stops being needed fails rather than
// sitting.
var descriptorCellsTheEmittersCannotReach = map[string]bool{
	"ByNamespace text/--drop": true,
}

// The exemption map holds only cells that genuinely still disagree.
func TestDescriptorExemptionsAreStillNeeded(t *testing.T) {
	t.Parallel()
	if len(descriptorCellsTheEmittersCannotReach) == 0 {
		t.Fatal("the map is empty, so this test measured nothing")
	}

	for cell := range descriptorCellsTheEmittersCannotReach {
		parts := strings.Fields(cell)
		if len(parts) != 2 {
			t.Errorf("%q is not a `kind root/key` cell name", cell)
			continue
		}
		rootAndKey := strings.SplitN(parts[1], "/", 2)
		if len(rootAndKey) != 2 {
			t.Errorf("%q does not name a root and a key", cell)
			continue
		}

		descriptor, hasRow := baseDescriptors[rootAndKey[0]]
		if !hasRow {
			t.Errorf("%s exempts a root with no descriptor row", cell)
			continue
		}
		want, hasCell := descriptor.Absent.ByNamespace[rootAndKey[1]]
		if !hasCell {
			t.Errorf("%s exempts a namespace the row does not carry", cell)
			continue
		}

		// The same branch the measurement builds, not a simpler one. A first version of this guard
		// omitted `ResolvedAsColor`, so a `--color` exemption was compared on the wrong arm, always
		// disagreed, and the guard passed whatever the exemption claimed. A check that cannot fire
		// reads exactly like a check nothing violates.
		branch := UtilityBranch{HasValue: true, ResolvedNamespace: rootAndKey[1]}
		branch.ResolvedAsColor = rootAndKey[1] == "--color" || rootAndKey[1] == namespaceColorKeyword
		if rootAndKey[1] == namespaceNone {
			branch.ResolvedNamespace = ""
		}
		nodes := EmitGapRoot(rootAndKey[0], branch)
		if len(nodes) == 0 {
			t.Errorf("%s exempts a cell the emitters decline, which needs no exemption", cell)
			continue
		}
		if fmt.Sprint(PropertySort(nodes)) == fmt.Sprint(want) {
			t.Errorf("%s now agrees, so the exemption is stale and should be deleted", cell)
		}
	}
}

// No class can reach the `text/--drop` cell, which is why the emitters disagreeing with it is free.
//
// The exemption above claims the cell is dead. This measures that claim against the live parser
// rather than restating it: every key in the `--drop` namespace, spelled as a `text-*` class, must
// parse as some other root. A key that parsed as `text` would reach the cell and the disagreement
// would stop being harmless.
func TestNoClassReachesTheTextDropCell(t *testing.T) {
	t.Parallel()
	forEachEngineSystem(t, func(t *testing.T, system *LoadedDesignSystem) {

		keys := system.Theme().KeysInNamespaces([]string{"--drop"})
		if len(keys) == 0 {
			t.Fatal("this theme declares no --drop keys, and every theme over the framework's does, so nothing reached the cell and the test proves nothing")
		}

		var reachedText int
		for _, key := range keys {
			parsed := ParseCandidate("text-"+key, system)
			if len(parsed) == 0 {
				continue
			}
			if parsed[0].Root == "text" {
				reachedText++
				t.Errorf("text-%s parses as root `text`, so it reaches the text/--drop cell the exemption calls dead", key)
			}
		}

		t.Logf("--drop keys spelled as text-*: %d checked, %d reaching root text", len(keys), reachedText)
	})
}

// The Alpha and Themed axes, which are the same rows read through a modifier.
//
// A modifier changes which arm several of these roots take: the shadow family writes its alpha
// declaration with a real value rather than an absent one, and `text` turns a font size into a font
// size plus a line height. `UtilityBranch` carries `HasModifier` for exactly that, so the axes are
// reachable and this measures whether they agree.
//
// Split from the unmodified measurement rather than folded into it, because the two axes have
// different populations and a combined total would hide one of them being empty.
func TestModifierAxesAgreeWithTheEmitters(t *testing.T) {
	t.Parallel()
	type counter struct{ compared, agreed int }
	counts := map[string]*counter{"Alpha": {}, "Themed": {}}
	var disagreements []string

	for root, descriptor := range baseDescriptors {
		if _, isGapRoot := gapEmitters[root]; !isGapRoot {
			continue
		}

		for _, axisCase := range []struct {
			name     string
			readings AxisReadings
			// hasModifier is whether a class on this axis writes a modifier the handle body reacts
			// to, which is not the same as whether it wrote one at all.
			//
			// The Alpha axis is a real alpha, `shadow-md/50`, and the shadow family writes its
			// `--tw-shadow-alpha` declaration with a value rather than absent, which PropertySort
			// counts. The Themed axis is `/none`, a `--leading` key: it means something to `text`
			// and nothing to `shadow`, so a root that does not consult the theme for its modifier
			// emits what it would with no modifier at all. Measured: `shadow/--shadow` reads `#3` on
			// Alpha and `#2` on Themed, and `#2` is the unmodified emission.
			//
			// Which roots do is per-root and is `rootConsultsTheThemeForItsModifier` below.
			hasModifier bool
		}{
			{name: "Alpha", readings: descriptor.Alpha, hasModifier: true},
			{name: "Themed", readings: descriptor.Themed, hasModifier: rootConsultsTheThemeForItsModifier[root]},
		} {
			for dataType, want := range axisCase.readings.ByType {
				branch := UtilityBranch{HasValue: true, IsArbitrary: true, DataType: dataType, HasModifier: axisCase.hasModifier}
				branch.ResolvedAsColor = dataType == DataTypeColor
				nodes := EmitGapRoot(root, branch)
				if len(nodes) == 0 {
					continue
				}
				counts[axisCase.name].compared++
				got := PropertySort(nodes)
				if fmt.Sprint(got) == fmt.Sprint(want) {
					counts[axisCase.name].agreed++
					continue
				}
				disagreements = append(disagreements,
					fmt.Sprintf("%s %s/%s: row %v, emitters %v", axisCase.name, root, dataType, want, got))
			}

			for namespace, want := range axisCase.readings.ByNamespace {
				branch := UtilityBranch{HasValue: true, ResolvedNamespace: namespace, HasModifier: axisCase.hasModifier}
				branch.ResolvedAsColor = namespace == "--color" || namespace == namespaceColorKeyword
				if namespace == namespaceNone {
					branch.ResolvedNamespace = ""
				}
				nodes := EmitGapRoot(root, branch)
				if len(nodes) == 0 {
					continue
				}
				counts[axisCase.name].compared++
				got := PropertySort(nodes)
				if fmt.Sprint(got) == fmt.Sprint(want) {
					counts[axisCase.name].agreed++
					continue
				}
				disagreements = append(disagreements,
					fmt.Sprintf("%s %s/%s: row %v, emitters %v", axisCase.name, root, namespace, want, got))
			}
		}
	}

	sort.Strings(disagreements)
	var totalCompared, totalAgreed int
	for _, name := range []string{"Alpha", "Themed"} {
		totalCompared += counts[name].compared
		totalAgreed += counts[name].agreed
		t.Logf("%-8s %3d compared, %3d agreed", name, counts[name].compared, counts[name].agreed)
	}
	t.Logf("modifier-axis cells: %d compared, %d agreed, %d disagreed", totalCompared, totalAgreed, totalCompared-totalAgreed)

	// The same dead cell the unmodified measurement exempts, on both modifier axes. `text/--drop` is
	// a `text-shadow` reading filed under root `text`, and no class reaches it on any axis.
	for _, disagreement := range disagreements {
		if strings.Contains(disagreement, "text/--drop") {
			continue
		}
		t.Errorf("%s", disagreement)
	}
	if totalCompared == 0 {
		t.Fatal("no modifier cell was compared, so this test measured nothing")
	}
}

// Roots whose handle body reacts to a themed modifier, which is `/none` in Tailwind 4.3.3.
//
// `none` is a `--leading` key, so it means something to a root that consults `--leading` for its
// modifier and nothing to one that does not. `text/lg/none` is a font size whose line height the
// modifier supplied, so it emits both declarations; `shadow-md/none` writes no alpha, because
// `shadow` reads its modifier as an alpha and `none` is not one, and its Themed readings equal its
// unmodified ones.
//
// `UtilityBranch` carries `HasModifier` as a bool and cannot express the difference. That is a real
// limit of the port rather than a probe detail: an emitter for a root that reacted to both kinds
// would need to know which was written. Measured across every gap root with a Themed axis, exactly
// one does, so a per-root set expresses it exactly today and
// `TestThemedModifierRootsAreTheOnesThatDiffer` fails if a second appears.
var rootConsultsTheThemeForItsModifier = map[string]bool{
	"text": true,
}

// The themed-modifier set holds exactly the roots whose Themed readings differ from their unmodified
// ones.
//
// A root in the set whose axes agree is a stale entry; one outside it whose axes differ is a missing
// entry, and either would make the measurement above pass while comparing the wrong arm.
func TestThemedModifierRootsAreTheOnesThatDiffer(t *testing.T) {
	t.Parallel()
	var checked int

	for root, descriptor := range baseDescriptors {
		if _, isGapRoot := gapEmitters[root]; !isGapRoot {
			continue
		}
		if len(descriptor.Themed.ByNamespace) == 0 && len(descriptor.Themed.ByType) == 0 {
			continue
		}
		checked++

		// Both maps, because a root's themed difference can live in either. `text` reads the same
		// for every namespace it carries and differs on four types, so a namespace-only comparison
		// called it unchanged and the entry stale.
		differs := false
		for namespace, themed := range descriptor.Themed.ByNamespace {
			if unmodified, has := descriptor.Absent.ByNamespace[namespace]; has && fmt.Sprint(themed) != fmt.Sprint(unmodified) {
				differs = true
			}
		}
		for dataType, themed := range descriptor.Themed.ByType {
			if unmodified, has := descriptor.Absent.ByType[dataType]; has && fmt.Sprint(themed) != fmt.Sprint(unmodified) {
				differs = true
			}
		}

		if differs && !rootConsultsTheThemeForItsModifier[root] {
			t.Errorf("%s reads differently on the Themed axis and is not in the set, so its cells are compared on the unmodified arm", root)
		}
		if !differs && rootConsultsTheThemeForItsModifier[root] {
			t.Errorf("%s is in the set and reads the same on both axes, so the entry is stale", root)
		}
	}

	t.Logf("gap roots with a Themed axis: %d checked, %d in the set", checked, len(rootConsultsTheThemeForItsModifier))

	if checked == 0 {
		t.Fatal("no root carried a Themed axis, so this test measured nothing")
	}
}

// The rows are computable and the table is not deletable, and the difference is `TypeList`.
//
// 396 of 399 cells agree with the emitters, so every reading a row stores can be produced from a
// ported handle body. That answers "is the stored reading redundant" and does not answer "can the
// table go", because a cell is only reachable once something has decided which cell a class lands in,
// and that decision is `InferDataType` against the root's own ordered type list.
//
// The order is the whole of it. `InferDataType` returns the first match, so:
//
//	bg      InferDataType("3px") = position   list starts [percentage url position length]
//	border  InferDataType("3px") = length     list starts [length line-width percentage ratio]
//
// One value, two roots, different types, decided by nothing but list order. An emitter cannot supply
// it: it answers "given this type, what is emitted", which is the arm after the decision. Asked with
// each of `bg`'s ten types it returns a property for every one, and four distinct properties across
// them, with nothing distinguishing the order they should be tried in.
//
// So deleting `baseDescriptors` means moving the type lists somewhere else rather than dropping them,
// and `TypeList` is read at three live sites: `Lookup`'s arbitrary arm, `bareReading`'s inference
// step, and `gapTypeListFor`, which hands it to `UtilityBranchFor` so the rule's own computation
// infers the same way. A table of 78 ordered lists is what those three need, whatever it is called.
//
// This test pins the fact rather than the conclusion: if a future change makes the order derivable,
// it fails and the conclusion above is due a re-read.
func TestTypeListOrderDecidesTheCellAndEmittersCannotSupplyIt(t *testing.T) {
	t.Parallel()
	background, hasBackground := baseDescriptors["bg"]
	border, hasBorder := baseDescriptors["border"]
	if !hasBackground || !hasBorder {
		t.Fatal("bg and border are the two roots this contrast is built on and one is missing")
	}

	const sharedValue = "3px"
	backgroundType := InferDataType(sharedValue, background.TypeList)
	borderType := InferDataType(sharedValue, border.TypeList)

	if backgroundType == borderType {
		t.Errorf("bg and border both infer %q as %s; the contrast this rests on is gone and the "+
			"conclusion that TypeList is not derivable needs re-measuring", sharedValue, backgroundType)
	}
	t.Logf("%q infers as %s for bg and %s for border, decided by list order alone",
		sharedValue, backgroundType, borderType)

	// And the emitters answer every one of those types, which is why they cannot rank them.
	var answered int
	for _, dataType := range background.TypeList {
		if len(EmitGapRoot("bg", UtilityBranch{HasValue: true, IsArbitrary: true, DataType: dataType})) > 0 {
			answered++
		}
	}
	if answered != len(background.TypeList) {
		t.Errorf("the emitter answered %d of bg's %d types; this test assumes it answers all of them, "+
			"which is what makes the ordering unrecoverable from it", answered, len(background.TypeList))
	}
	t.Logf("bg's emitter answers %d of %d types, so nothing in it ranks them", answered, len(background.TypeList))
}
