package tailwind

import (
	"sort"
	"strings"
	"testing"
)

// The composing answer is computed from two emissions, and this holds it against the generated table
// it replaces, class for class rather than as a rate.
//
// The table was produced by compiling two values through the real Tailwind engine and comparing
// declaration text. `ComposesFor` performs the same comparison against the ported handle bodies, so
// the two are independent measurements of one fact and are forced to agree here.
func TestComposesForAgreesWithTheGeneratedTable(t *testing.T) {
	t.Parallel()
	var agreed, disagreed, unanswerable int
	var disagreements []string

	for root := range composingRootsTheGeneratorFound {
		composes, ok := ComposesFor(root)
		if !ok {
			unanswerable++
			continue
		}
		if composes {
			agreed++
			continue
		}
		disagreed++
		disagreements = append(disagreements, root)
	}

	sort.Strings(disagreements)
	t.Logf("composing roots: %d in the table, %d answered, %d agreed, %d disagreed, %d unanswerable",
		len(composingRootsTheGeneratorFound), agreed+disagreed, agreed, disagreed, unanswerable)

	if disagreed != 0 {
		t.Errorf("%d roots the table calls composing are computed as conflicting: %s",
			disagreed, strings.Join(disagreements, ", "))
	}
	if agreed == 0 {
		t.Fatal("no root was answered, so this test measured nothing")
	}
}

// The control, and it is the whole test.
//
// A derivation that answers "composing" for everything scores 82 of 82 against the table above and
// is worthless. This asserts the other direction on roots the table deliberately omits: `px-4` and
// `px-8` genuinely collide, and a computation that cannot say so has not learned anything.
//
// The controls are read out of the two emitter tables rather than hand-listed, so a root added
// upstream joins the population instead of being missed by a list nobody updates.
func TestComposesForRefusesRootsTheTableOmits(t *testing.T) {
	t.Parallel()
	var checked, wronglyComposing int
	var wrong []string

	for root := range FrameworkMultiDeclarationUtilities {
		if composingRootsTheGeneratorFound[root] || composingRootsTheGeneratorNeverReached[root] {
			continue
		}
		composes, ok := ComposesFor(root)
		if !ok {
			continue
		}
		checked++
		if composes {
			wronglyComposing++
			if len(wrong) < 20 {
				wrong = append(wrong, root)
			}
		}
	}
	for root := range FrameworkFunctionalUtilities {
		if composingRootsTheGeneratorFound[root] || composingRootsTheGeneratorNeverReached[root] {
			continue
		}
		composes, ok := ComposesFor(root)
		if !ok {
			continue
		}
		checked++
		if composes {
			wronglyComposing++
			if len(wrong) < 20 {
				wrong = append(wrong, root)
			}
		}
	}

	sort.Strings(wrong)
	t.Logf("non-composing controls: %d answered, %d wrongly computed as composing", checked, wronglyComposing)

	if wronglyComposing != 0 {
		t.Errorf("%d roots the table omits are computed as composing: %s",
			wronglyComposing, strings.Join(wrong, ", "))
	}
	if checked == 0 {
		t.Fatal("no control was answered, so this test measured nothing")
	}
}

// The value-partitioned roots answer composition too, and their non-composing siblings are the
// control that says the answer is not a constant yes.
//
// This slice takes a `UtilityBranch` rather than a resolved value, and the branch carries a `Value`
// so two emissions of one root can differ. A first version of that rewrite keyed each family's
// shared constant on the property name alone and was wrong on two roots, both caught here rather
// than by reading: `bg` writes a gradient name into `background-image` for a named non-colour value,
// and `filter` writes the chain only when the class is the bare `filter`. Keying on the root's use
// of the property fixed both.
func TestComposesForAnswersGapRootsAndTheirControls(t *testing.T) {
	t.Parallel()
	var composingAnswered, composingAgreed int
	var controlAnswered, controlWrong int
	var wrong []string

	for root := range gapEmitters {
		composes, ok := ComposesFor(root)
		if !ok {
			continue
		}
		if composingRootsTheGeneratorFound[root] {
			composingAnswered++
			if composes {
				composingAgreed++
			}
			continue
		}
		controlAnswered++
		if composes {
			controlWrong++
			if len(wrong) < 20 {
				wrong = append(wrong, root)
			}
		}
	}

	sort.Strings(wrong)
	t.Logf("gap roots: %d composing answered and %d agreed, %d controls answered and %d wrongly composing",
		composingAnswered, composingAgreed, controlAnswered, controlWrong)

	if composingAnswered != composingAgreed {
		t.Errorf("%d of %d composing gap roots computed as conflicting",
			composingAnswered-composingAgreed, composingAnswered)
	}
	if controlWrong != 0 {
		t.Errorf("%d non-composing gap roots computed as composing: %s", controlWrong, strings.Join(wrong, ", "))
	}
	if composingAnswered == 0 || controlAnswered == 0 {
		t.Fatal("one side of the comparison was empty, so this test measured nothing")
	}
}

// Roots the computation calls composing and the generated table omits, because the generator's probe
// values never reached them.
//
// Absence from the generator's table means one of two things and they are not the same: the generator
// measured the root and found it conflicting, or the generator never measured it. The table cannot
// tell them apart, and this map is where that difference is written down.
//
// # mask-radial, measured against the live engine rather than argued
//
//	mask-radial emits utility CSS for 0 of the generator's 15 probe values
//	mask-linear emits for 15 of 15
//	mask-conic  emits for 15 of 15
//
// The probe list is the numeric and named scales. `mask-radial` takes a radial size or shape, so
// `mask-radial-4` compiles to nothing at all. The generator's loop takes the first probe that
// compiles, and on `continue`ing past all fifteen it skips the root entirely rather than recording it
// as unmeasured.
//
// Reached with values it accepts, the engine's own answer is that it composes:
//
//	mask-radial-[3px] and mask-radial-[50%] both emit
//	  mask-image: var(--tw-mask-linear), var(--tw-mask-radial), var(--tw-mask-conic)
//	  mask-composite: intersect
//
// Identical shared declarations, the value only in `--tw-mask-radial-size`, which is exactly the
// shape its two siblings have and both of them are in the table.
//
// This is the fifth instance of a failure the generator's own comment names: "A check that cannot
// fire is indistinguishable from a check nothing violates, and the same is true of a probe that
// cannot reach its subject." It lists four found before this one. The right repair is a probe list
// that reaches `mask-radial`, which belongs to whoever owns the generator; this map records the
// finding rather than silently agreeing with a hole.
var composingRootsTheGeneratorNeverReached = map[string]bool{
	"mask-radial": true,
}

// The exemption map must stay a record of measured findings rather than a place to put disagreements.
//
// Two assertions. Every entry must be a root the computation actually calls composing, so an entry
// cannot sit here excusing a root the computation agrees is conflicting. And every entry must be
// absent from the generator's table, so an entry cannot shadow a root the table already answers.
//
// A previous slice of this port shipped an exemption map that read as careful and was hiding 47
// comparisons, 7 of which disagreed. Keeping this one honest means asserting its shape rather than
// trusting its comments.
func TestTheUnreachedMapHoldsOnlyMeasuredFindings(t *testing.T) {
	t.Parallel()
	if len(composingRootsTheGeneratorNeverReached) == 0 {
		t.Fatal("the map is empty, so this test measured nothing")
	}
	for root := range composingRootsTheGeneratorNeverReached {
		composes, ok := ComposesFor(root)
		if !ok {
			t.Errorf("%s is exempted but is not answerable, so the exemption hides nothing real", root)
			continue
		}
		if !composes {
			t.Errorf("%s is exempted but computes as conflicting, which the control test would pass anyway", root)
		}
		if composingRootsTheGeneratorFound[root] {
			t.Errorf("%s is exempted but the table already carries it, so the exemption is dead", root)
		}
	}
}

// The 82 rows the generated `ComposingRoots` held, captured here when it was deleted.
//
// The table was produced by compiling two values of each root through the real Tailwind engine and
// comparing declaration text. `ComposesFor` performs the same comparison against the ported handle
// bodies, so keeping the generator's answers as a fixture means the two remain independent
// measurements of one fact rather than one fact read twice.
//
// This is a test fixture and not a table the rule reads. It never grows: a root added upstream is
// answered by the emitters, and the only thing this pins is that the port did not lose an answer the
// generator had. `4f23e9f` and `7a72125` both found generated tables carrying one repository's own
// `@utility` blocks as though they were Tailwind's, which is what a live table costs and a frozen
// fixture does not.
var composingRootsTheGeneratorFound = map[string]bool{
	"-backdrop-hue-rotate": true,
	"-bg-conic":            true,
	"-hue-rotate":          true,
	"-mask-conic":          true,
	"-mask-linear":         true,
	"-rotate-x":            true,
	"-rotate-y":            true,
	"-rotate-z":            true,
	"-scale":               true,
	"-scale-x":             true,
	"-scale-y":             true,
	"-scale-z":             true,
	"-skew":                true,
	"-skew-x":              true,
	"-skew-y":              true,
	"-translate":           true,
	"-translate-x":         true,
	"-translate-y":         true,
	"-translate-z":         true,
	"backdrop-blur":        true,
	"backdrop-brightness":  true,
	"backdrop-contrast":    true,
	"backdrop-grayscale":   true,
	"backdrop-hue-rotate":  true,
	"backdrop-invert":      true,
	"backdrop-opacity":     true,
	"backdrop-saturate":    true,
	"backdrop-sepia":       true,
	"bg-conic":             true,
	"blur":                 true,
	"border-spacing":       true,
	"border-spacing-x":     true,
	"border-spacing-y":     true,
	"brightness":           true,
	"contrast":             true,
	"drop-shadow":          true,
	"grayscale":            true,
	"hue-rotate":           true,
	"inset-ring":           true,
	"inset-shadow":         true,
	"invert":               true,
	"mask-b-from":          true,
	"mask-b-to":            true,
	"mask-conic":           true,
	"mask-conic-from":      true,
	"mask-conic-to":        true,
	"mask-l-from":          true,
	"mask-l-to":            true,
	"mask-linear":          true,
	"mask-linear-from":     true,
	"mask-linear-to":       true,
	"mask-r-from":          true,
	"mask-r-to":            true,
	"mask-radial-from":     true,
	"mask-radial-to":       true,
	"mask-t-from":          true,
	"mask-t-to":            true,
	"mask-x-from":          true,
	"mask-x-to":            true,
	"mask-y-from":          true,
	"mask-y-to":            true,
	"ring":                 true,
	"rotate-x":             true,
	"rotate-y":             true,
	"rotate-z":             true,
	"saturate":             true,
	"scale":                true,
	"scale-x":              true,
	"scale-y":              true,
	"scale-z":              true,
	"scrollbar-thumb":      true,
	"scrollbar-track":      true,
	"sepia":                true,
	"shadow":               true,
	"shadow-":              true,
	"skew":                 true,
	"skew-x":               true,
	"skew-y":               true,
	"translate":            true,
	"translate-x":          true,
	"translate-y":          true,
	"translate-z":          true,
}

// Asking a gap root for a value must not change what the next caller sees.
//
// `EmitGapRoot` rewrites declaration values when the branch carries one, and every other caller in
// the package reads property names off the sentinel. If the rewrite mutated the emitter's nodes
// rather than copying them, a composition query would leave a real value behind for whoever asked
// next, and every reading in the package is downstream of that.
//
// Measured rather than assumed: mutating `withResolvedValue` to rewrite in place instead of copying
// leaves this test passing, because every emitter here allocates its nodes per call and none returns
// a package-level slice. So the copy is defensive today and this test does not currently catch its
// removal, which is recorded rather than dressed up as a caught mutation.
//
// Both stay. The copy is what makes the rewrite correct against an emitter that caches, and this is
// what would notice if one started to; the alternative is a defect that surfaces as a wrong reading
// somewhere else in the package, with nothing pointing back here.
func TestAskingForAValueLeavesTheSentinelIntact(t *testing.T) {
	t.Parallel()
	var checked int
	for root := range gapEmitters {
		branch := UtilityBranch{HasValue: true, DataType: DataTypeLength}

		before := EmitGapRoot(root, branch)
		if len(before) == 0 {
			continue
		}
		beforeValues := make([]string, 0, len(before))
		for _, node := range before {
			beforeValues = append(beforeValues, node.Value)
		}

		valued := branch
		valued.Value = "zzsomeoneelsesvalue"
		EmitGapRoot(root, valued)

		after := EmitGapRoot(root, branch)
		if len(after) != len(before) {
			t.Errorf("%s emitted %d declarations before a valued call and %d after", root, len(before), len(after))
			continue
		}
		for index, node := range after {
			if node.Value != beforeValues[index] {
				t.Errorf("%s declaration %d carried %q before a valued call and %q after",
					root, index, beforeValues[index], node.Value)
			}
		}
		checked++
	}

	t.Logf("gap roots re-read after a valued call: %d", checked)
	if checked == 0 {
		t.Fatal("no root was checked, so this test measured nothing")
	}
}

// The selector shapes the deleted `RootSelectorShapes` held, captured when it was deleted.
//
// Generated by compiling each root through the real engine and reading the selector back.
// `SelectorShapeForRoot` reads the same fact off the ported handle bodies, so this keeps the two as
// independent measurements rather than one fact read twice, the same way
// `composingRootsTheGeneratorFound` does.
var rootSelectorShapesTheGeneratorFound = map[string]string{
	"-space-x":    ":where(.CLASS > :not(:last-child))",
	"-space-y":    ":where(.CLASS > :not(:last-child))",
	"divide":      ":where(.CLASS > :not(:last-child))",
	"divide-x":    ":where(.CLASS > :not(:last-child))",
	"divide-y":    ":where(.CLASS > :not(:last-child))",
	"placeholder": ".CLASS::placeholder",
	"space-x":     ":where(.CLASS > :not(:last-child))",
	"space-y":     ":where(.CLASS > :not(:last-child))",
}

// Every shape the generator measured is reproduced by reading the ported handle bodies.
func TestSelectorShapesAgreeWithTheGeneratedTable(t *testing.T) {
	t.Parallel()
	var agreed int
	for root, expected := range rootSelectorShapesTheGeneratorFound {
		actual := SelectorShapeForRoot(root)
		if actual != expected {
			t.Errorf("%s: the generator measured %q and the emitters produce %q", root, expected, actual)
			continue
		}
		agreed++
	}

	t.Logf("selector shapes: %d in the generated table, %d agreed", len(rootSelectorShapesTheGeneratorFound), agreed)

	if agreed == 0 {
		t.Fatal("no shape was compared, so this test measured nothing")
	}
}

// The control: a root the generator did not list must read as a bare class.
//
// Without this, an emitter that wrapped everything would agree with all eight above and be wrong
// about the other 234. The population is every root the emitters answer rather than a list, so a
// root added upstream is checked rather than needing to be added here.
func TestUnwrappedRootsReadAsABareClass(t *testing.T) {
	t.Parallel()
	var checked, wrong int
	var wrongNames []string

	check := func(root string) {
		if rootSelectorShapesTheGeneratorFound[root] != "" {
			return
		}
		checked++
		if shape := SelectorShapeForRoot(root); shape != ".CLASS" {
			wrong++
			if len(wrongNames) < 15 {
				wrongNames = append(wrongNames, root+" -> "+shape)
			}
		}
	}
	for root := range FrameworkMultiDeclarationUtilities {
		check(root)
	}
	for root := range FrameworkFunctionalUtilities {
		check(root)
	}
	for root := range gapEmitters {
		check(root)
	}

	sort.Strings(wrongNames)
	t.Logf("roots the table did not list: %d checked, %d wrapped", checked, wrong)

	if wrong != 0 {
		t.Errorf("%d roots emit under a non-bare selector and the generator did not list them: %s",
			wrong, strings.Join(wrongNames, ", "))
	}
	if checked == 0 {
		t.Fatal("no root was checked, so this test measured nothing")
	}
}

// A wrapped root still declares its properties, and this is the check that was missing.
//
// Wrapping the six roots in a rule node dropped `DeclaredPropertiesFor` from 11,288 resolved corpus
// classes to 11,163, because it walked only top-level nodes and found no declarations inside the
// wrapper. The whole suite stayed green: every acceptance test compares readings, and `PropertySort`
// already descended, so nothing measured the one traversal that did not.
//
// 125 classes silently stopped being answered, which is the exact shape of failure this rule keeps
// producing, a rule going quiet rather than wrong. The placement counts caught it and no test did,
// so this is that test.
func TestWrappedRootsStillDeclareTheirProperties(t *testing.T) {
	t.Parallel()
	expected := map[string][]string{
		"space-x":     {"margin-inline-start", "margin-inline-end"},
		"space-y":     {"margin-block-start", "margin-block-end"},
		"divide":      {"border-color"},
		"divide-x":    {"border-inline-style", "border-inline-start-width", "border-inline-end-width"},
		"divide-y":    {"border-bottom-style", "border-top-style", "border-top-width", "border-bottom-width"},
		"placeholder": {"color"},
	}

	for root, want := range expected {
		candidate := &ParsedCandidate{
			Kind:  ParsedCandidateKindFunctional,
			Root:  root,
			Value: &ParsedValue{Kind: ParsedValueKindNamed, Value: "4"},
		}
		properties, canCompute := DeclaredPropertiesFor(candidate, ValueResolution{})
		if !canCompute {
			t.Errorf("%s declares nothing, so a wrapped root is being declined", root)
			continue
		}

		found := make(map[string]bool, len(properties))
		for _, property := range properties {
			found[property] = true
		}
		for _, property := range want {
			if !found[property] {
				t.Errorf("%s does not declare %s; got %v", root, property, properties)
			}
		}
	}
}
