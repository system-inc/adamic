package tailwind

import (
	"sort"
	"testing"
)

// TestEmitterSlicesDoNotOverlap pins the assumption emitFunctionalRoot's dispatch rests on.
//
// It tries three tables in order and returns the first hit. That is only safe because the three
// partition the roots: a root in two of them would make the dispatch order a silent precedence, and
// which slice answered would depend on the order somebody wrote the ifs in.
func TestEmitterSlicesDoNotOverlap(t *testing.T) {
	t.Parallel()
	owner := map[string]string{}
	for root := range FrameworkFunctionalUtilities {
		owner[root] = "single-declaration"
	}
	for root := range FrameworkMultiDeclarationUtilities {
		if previous, taken := owner[root]; taken {
			t.Errorf("%s is in both %s and multi-declaration; the dispatch order in emitFunctionalRoot "+
				"would decide which body answers", root, previous)
		}
		owner[root] = "multi-declaration"
	}
	for root := range gapEmitters {
		if previous, taken := owner[root]; taken {
			t.Errorf("%s is in both %s and the gap slice; the dispatch order in emitFunctionalRoot "+
				"would decide which body answers", root, previous)
		}
		owner[root] = "gap"
	}
	t.Logf("emitter slices partition %d functional roots", len(owner))
}

// equalPropertySets compares two property lists as sets.
//
// Order is not part of the answer here. A declaration list has an emission order and a conflict
// comparison asks only which properties are shared, so two lists differing in order are the same
// answer to this question. The deleted tables were themselves generated sorted, which would make an
// order-sensitive comparison fail on a difference that means nothing.
func equalPropertySets(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	leftSorted := append([]string(nil), left...)
	rightSorted := append([]string(nil), right...)
	sort.Strings(leftSorted)
	sort.Strings(rightSorted)
	for index := range leftSorted {
		if leftSorted[index] != rightSorted[index] {
			return false
		}
	}
	return true
}

// TestComputedPropertiesMakeTheDistinctionsTheOverrideTablesHeld is what survives the deletion.
//
// Three tables were deleted on a measurement that can no longer be re-run, because one side of it is
// gone: `ClassDeclaredProperties` reproduced 13 of 13 and `RootColorProperties` 15 of 15 against the
// computed answer, and `StaticDeclaredProperties` 839 of 839 with 8 declines that were this
// repository's own `@utility` blocks leaking into a framework table.
//
// A measurement that cannot be re-run is a claim, so this pins the distinctions themselves rather
// than the comparison. Each case below is one the deleted tables needed a separate entry for and the
// emitter derives, and each would silently stop being made if a branch were flattened: the class
// still resolves, still reports no finding on correct code, and starts failing to report a real one.
func TestComputedPropertiesMakeTheDistinctionsTheOverrideTablesHeld(t *testing.T) {
	t.Parallel()
	functional := func(root, value string, resolution ValueResolution) []string {
		properties, resolved := DeclaredPropertiesFor(&ParsedCandidate{
			Kind:  ParsedCandidateKindFunctional,
			Root:  root,
			Value: &ParsedValue{Kind: ParsedValueKindNamed, Value: value},
		}, resolution)
		if !resolved {
			t.Errorf("%s-%s computes nothing", root, value)
		}
		return properties
	}

	// The fifteen colour arms. Each root declares a different property for a colour than for its
	// ordinary value, which is what `RootColorProperties` recorded once per root.
	for root, wantColorProperty := range map[string]string{
		"border": "border-color", "border-b": "border-bottom-color", "border-t": "border-top-color",
		"border-x": "border-inline-color", "border-y": "border-block-color",
		"decoration": "text-decoration-color", "outline": "outline-color",
		"stroke": "stroke", "text": "color",
	} {
		colour := functional(root, "red-500", ValueResolution{IsColor: true})
		if !containsProperty(colour, wantColorProperty) {
			t.Errorf("%s-red-500 computes %v, want %q", root, colour, wantColorProperty)
		}
		ordinary := functional(root, "4", ValueResolution{})
		if containsProperty(ordinary, wantColorProperty) {
			t.Errorf("%s-4 computes %v, which includes the colour property %q; the two arms have been "+
				"flattened and a width would now conflict with a colour", root, ordinary, wantColorProperty)
		}
	}

	// The `font` namespace split, which is neither a colour nor an inferable type and is the branch
	// the three `font-*` overrides recorded.
	family := functional("font", "mono", ValueResolution{Namespace: "--font"})
	weight := functional("font", "medium", ValueResolution{})
	if !containsProperty(family, "font-family") {
		t.Errorf("font-mono computes %v, want font-family", family)
	}
	if !containsProperty(weight, "font-weight") {
		t.Errorf("font-medium computes %v, want font-weight", weight)
	}
	for _, property := range family {
		if containsProperty(weight, property) {
			t.Errorf("font-mono and font-medium share %q, so `font-medium font-mono` reads as a "+
				"conflict on correct code, which is the defect the overrides existed to prevent", property)
		}
	}

	// `text` with and without a modifier, which is the correction this migration made rather than
	// preserved: the deleted root row said `{font-size, line-height}` for both, and upstream emits
	// the line height only when a modifier is written.
	unmodified, resolved := DeclaredPropertiesFor(&ParsedCandidate{
		Kind:  ParsedCandidateKindFunctional,
		Root:  "text",
		Value: &ParsedValue{Kind: ParsedValueKindArbitrary, Value: "10px"},
	}, ValueResolution{})
	if !resolved {
		t.Fatal("text-[10px] computes nothing")
	}
	modified, resolved := DeclaredPropertiesFor(&ParsedCandidate{
		Kind:     ParsedCandidateKindFunctional,
		Root:     "text",
		Value:    &ParsedValue{Kind: ParsedValueKindArbitrary, Value: "10px"},
		Modifier: &ParsedModifier{Kind: ParsedModifierKindNamed, Value: "6"},
	}, ValueResolution{})
	if !resolved {
		t.Fatal("text-[10px]/6 computes nothing")
	}
	if containsProperty(unmodified, "line-height") {
		t.Errorf("text-[10px] computes %v, which includes line-height. Upstream returns "+
			"[decl('font-size', value)] with no modifier, and claiming line-height makes the class "+
			"fail to conflict with leading-6 when it should", unmodified)
	}
	if !containsProperty(modified, "line-height") {
		t.Errorf("text-[10px]/6 computes %v, want line-height alongside font-size", modified)
	}
}

// TestCustomPropertiesAreStrippedFromDeclaredProperties pins a strip the corpus cannot check.
//
// Two classes that both write a `--tw-*` variable are not in conflict about anything an author sees.
// `shadow-lg` and `ring-1` both write into the box-shadow variable chain, and reporting them as
// conflicting would be a finding on correct code. The deleted conflict tables stripped custom
// properties for that reason while the ordering tables deliberately keep them, and
// `DeclaredPropertiesFor` has to keep making that split now that one emitter serves both.
//
// This exists because removing the strip was mutated and SURVIVED the whole suite. The corpus went
// from 11,288 resolved to 11,292 and still reported zero findings, because no class list in it
// happens to pair two classes whose only shared property is a `--tw-*` one. A rule can be wrong in a
// way real code does not exercise, and a corpus measures the code somebody wrote rather than the
// code somebody could write, so this pairs the classes directly.
func TestCustomPropertiesAreStrippedFromDeclaredProperties(t *testing.T) {
	t.Parallel()
	// Roots whose emitters write a `--tw-*` variable alongside a real property. Each must report only
	// the real one, or a class sharing the variable becomes a false conflict.
	for _, probe := range []struct {
		root      string
		wantReal  string
		wantNoVar string
	}{
		{root: "blur", wantReal: "filter", wantNoVar: "--tw-blur"},
		{root: "grayscale", wantReal: "filter", wantNoVar: "--tw-grayscale"},
		{root: "leading", wantReal: "line-height", wantNoVar: "--tw-leading"},
		{root: "tracking", wantReal: "letter-spacing", wantNoVar: "--tw-tracking"},
		{root: "scale-x", wantReal: "scale", wantNoVar: "--tw-scale-x"},
	} {
		properties, resolved := declaredPropertiesOfRoot(probe.root)
		if !resolved {
			t.Errorf("%s computes nothing", probe.root)
			continue
		}
		if !containsProperty(properties, probe.wantReal) {
			t.Errorf("%s computes %v, want the real property %q", probe.root, properties, probe.wantReal)
		}
		for _, property := range properties {
			if len(property) > 1 && property[0] == '-' && property[1] == '-' {
				t.Errorf("%s computes %v, which includes the custom property %q. Two classes both "+
					"writing a --tw-* variable are not in conflict about anything an author sees, and "+
					"reporting them would be a finding on correct code", probe.root, properties, property)
				break
			}
		}
	}

	// The pairing the strip exists to prevent, asserted directly. Both roots write into the
	// box-shadow variable chain and neither shares a real property with the other.
	shadow, shadowResolved := declaredPropertiesOfRoot("shadow")
	ring, ringResolved := DeclaredPropertiesFor(&ParsedCandidate{
		Kind:  ParsedCandidateKindFunctional,
		Root:  "ring",
		Value: &ParsedValue{Kind: ParsedValueKindNamed, Value: "1"},
	}, ValueResolution{})
	if !shadowResolved || !ringResolved {
		t.Fatal("shadow or ring computes nothing")
	}
	for _, property := range shadow {
		if len(property) > 1 && property[:2] == "--" {
			t.Errorf("shadow computes the custom property %q", property)
		}
	}
	for _, property := range ring {
		if len(property) > 1 && property[:2] == "--" {
			t.Errorf("ring computes the custom property %q", property)
		}
	}
}
