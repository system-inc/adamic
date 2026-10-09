package tailwind

import (
	"sort"
	"testing"
)

// TestMultiDeclarationEmittersAgreeWithTheMeasuredReading is the acceptance test for the 152.
//
// Every row of `FrameworkMultiDeclarationUtilities` carries a `Reading` measured against the shipped
// engine, and every root now carries an emitter transcribed from `utilities.ts` or, for five of
// them, from `compat/legacy-utilities.ts`. The two were produced by different means from different
// sources, so `PropertySort(Emit(root, value))` against the stored reading is a real comparison.
//
// It is a stronger comparison than the 33-root one, because a reading here carries a count that
// exceeds its order length on 30 roots and a latched order on six. An emitter that dropped a
// declaration, emitted one too many, or lost a `--tw-sort` changes the count or the order and fails
// here, where a property-name comparison would see nothing.
func TestMultiDeclarationEmittersAgreeWithTheMeasuredReading(t *testing.T) {
	t.Parallel()
	const sentinel = "zzsentinel"

	var agreed int
	var disagreements, unported []string

	for root, utility := range FrameworkMultiDeclarationUtilities {
		emitted := utility.Emit(root, "", ResolvedUtilityValue{Value: sentinel})
		if len(emitted) == 0 {
			unported = append(unported, root)
			continue
		}

		sorted := PropertySort(emitted)
		computed := Reading{Order: sorted.Order, Count: sorted.Count}
		if readingsEqual(utility.Reading, computed) {
			agreed++
			continue
		}
		disagreements = append(disagreements,
			root+": measured "+formatReading(utility.Reading)+", ported handle emits "+
				describeDeclarations(emitted)+" reading "+formatReading(computed))
	}

	sort.Strings(disagreements)
	sort.Strings(unported)

	t.Logf("ported multi-declaration handles: %d roots in the table, %d compared, %d agreed, %d unported, %d disagreed",
		len(FrameworkMultiDeclarationUtilities), agreed+len(disagreements), agreed, len(unported), len(disagreements))

	if len(FrameworkMultiDeclarationUtilities) != 152 {
		t.Errorf("the table holds %d roots and this comparison was written against 152; the population moved and the assertions below were sized for the old one", len(FrameworkMultiDeclarationUtilities))
	}
	if agreed != 152 {
		t.Errorf("%d of 152 roots agreed; the acceptance condition is every root, not a rate", agreed)
	}
	for _, root := range unported {
		t.Errorf("%s is in the table and has no ported handle, so it emits nothing and reads nothing", root)
	}
	for _, disagreement := range disagreements {
		t.Error(disagreement)
	}
}

// TestMultiDeclarationLiteralsAgreeWithTheirMeasuredReading covers the branching handlers.
//
// A root-defined keyword is answered before the handle body runs and can emit a different list.
// `LiteralReadings` measured what each one reads; this checks the emitter reproduces it, including
// the four whose list genuinely differs from their root's.
//
// The population matters more than usual here. 34 roots declare literals and most of them agree with
// their root, so a comparison that only visited the four divergent ones would still report a rate of
// one and would miss an emitter that broke every agreeing literal.
func TestMultiDeclarationLiteralsAgreeWithTheirMeasuredReading(t *testing.T) {
	t.Parallel()
	const sentinel = "zzsentinel"

	var agreed, divergent int
	var disagreements []string

	for root, utility := range FrameworkMultiDeclarationUtilities {
		for literal, measured := range utility.LiteralReadings {
			emitted := utility.Emit(root, literal, ResolvedUtilityValue{Value: sentinel})
			if len(emitted) == 0 {
				disagreements = append(disagreements, root+"-"+literal+": emits nothing, measured "+formatReading(measured))
				continue
			}
			if !measured.Equal(utility.Reading) {
				divergent++
			}
			sorted := PropertySort(emitted)
			computed := Reading{Order: sorted.Order, Count: sorted.Count}
			if readingsEqual(measured, computed) {
				agreed++
				continue
			}
			disagreements = append(disagreements,
				root+"-"+literal+": measured "+formatReading(measured)+", ported handle emits "+
					describeDeclarations(emitted)+" reading "+formatReading(computed))
		}
	}

	sort.Strings(disagreements)
	t.Logf("literals: %d compared, %d agreed, %d of them read differently from their root, %d disagreed",
		agreed+len(disagreements), agreed, divergent, len(disagreements))

	if agreed+len(disagreements) < 60 {
		t.Errorf("only %d literals were compared; the table declares far more, so this measured the wrong population", agreed+len(disagreements))
	}
	// Named rather than left implicit: these four are the whole reason literal emitters exist, and a
	// count that drops to zero means the divergence stopped being exercised.
	if divergent != 4 {
		t.Errorf("%d literals read differently from their root; four do at 4.3.3 (divide-none, ease-initial, transition-none, translate-none), so the branch coverage moved", divergent)
	}
	for _, disagreement := range disagreements {
		t.Error(disagreement)
	}
}

// TestMultiDeclarationLiteralsAreOnlyListedWhenTheyDiffer keeps the override map honest.
//
// A literal whose declaration list matches its root's needs no entry, and adding one anyway would be
// a second copy of the same fact that nothing forces to stay in step. So an override that agrees
// with its root's ordinary path is a defect rather than harmless duplication.
func TestMultiDeclarationLiteralsAreOnlyListedWhenTheyDiffer(t *testing.T) {
	t.Parallel()
	const sentinel = "zzsentinel"

	listed := 0
	for root, byLiteral := range frameworkMultiLiteralEmitters {
		utility, known := FrameworkMultiDeclarationUtilities[root]
		if !known {
			t.Errorf("%s has literal emitters and is not in the table", root)
			continue
		}
		for literal, emitter := range byLiteral {
			listed++
			ordinary := PropertySort(utility.Emit(root, "", ResolvedUtilityValue{Value: sentinel}))
			override := PropertySort(emitter(ResolvedUtilityValue{Value: sentinel}))
			if readingsEqual(Reading{Order: ordinary.Order, Count: ordinary.Count}, Reading{Order: override.Order, Count: override.Count}) {
				t.Errorf("%s-%s is listed as an override and reads the same as its root, so the entry restates a fact rather than correcting one", root, literal)
			}
			if _, measured := utility.LiteralReadings[literal]; !measured {
				t.Errorf("%s-%s has an override and no measured reading, so nothing checks it", root, literal)
			}
		}
	}
	t.Logf("%d literal overrides listed", listed)
	if listed != 4 {
		t.Errorf("%d overrides listed; four literals diverge at 4.3.3", listed)
	}
}

// TestSortLatchIsObservableInTheMultiTable pins the behaviour six roots depend on.
//
// A `--tw-sort` naming a known property latches the order and lets every later declaration count
// without contributing. Without the latch these roots read every property they declare, which is a
// different sort key rather than a slightly wrong one.
//
// Asserted on the readings directly, because the latch is the one part of this port where removing
// a declaration and removing the redirect produce different failures, and a test that only compared
// counts would not tell them apart.
func TestSortLatchIsObservableInTheMultiTable(t *testing.T) {
	t.Parallel()
	for _, root := range []string{"divide", "divide-x", "divide-y", "space-x", "space-y", "placeholder"} {
		utility, known := FrameworkMultiDeclarationUtilities[root]
		if !known {
			t.Errorf("%s is not in the table", root)
			continue
		}
		emitted := utility.Emit(root, "", ResolvedUtilityValue{Value: "zzsentinel"})

		// All six of these wrap their declarations in a nested rule, so the leading `--tw-sort` is
		// inside it. Descending here rather than asserting on `emitted[0]` matches what
		// `PropertySort` does, which is the behaviour this test is about.
		leading := emitted
		if len(leading) == 1 && leading[0].Kind == KindRule {
			leading = leading[0].Nodes
		}
		if len(leading) == 0 || leading[0].Property != "--tw-sort" {
			t.Errorf("%s does not lead with a --tw-sort declaration, so its order cannot latch", root)
			continue
		}
		sorted := PropertySort(emitted)
		if len(sorted.Order) != 1 {
			t.Errorf("%s reads %d positions; the latch should hold it to one", root, len(sorted.Order))
		}
		if sorted.Count <= len(sorted.Order) {
			t.Errorf("%s reads %d positions at count %d; a latched root counts more declarations than it positions", root, len(sorted.Order), sorted.Count)
		}
	}

	// `size` is the counterexample and it is asserted here rather than described, because it is what
	// makes the latch a property of PropertyOrder rather than a flag on the emitter.
	size := FrameworkMultiDeclarationUtilities["size"]
	emitted := size.Emit("size", "", ResolvedUtilityValue{Value: "zzsentinel"})
	sorted := PropertySort(emitted)
	if len(sorted.Order) != 2 {
		t.Errorf("size reads %d positions; its --tw-sort names `size`, which PropertyOrder does not know, so it does not latch and width and height both contribute", len(sorted.Order))
	}
}

// TestMultiDeclarationCountsExceedOrderWhereMeasured is the population check for the harder half.
//
// 30 roots read fewer positions than declarations, either because a declaration's property is
// outside `PropertyOrder` or because a `--tw-sort` latched. That gap is the thing a property-name
// table could not express and the reason these roots carried a measured reading, so it is asserted
// rather than left to the acceptance test to imply.
func TestMultiDeclarationCountsExceedOrderWhereMeasured(t *testing.T) {
	t.Parallel()
	const sentinel = "zzsentinel"

	measuredGap, emittedGap := 0, 0
	for root, utility := range FrameworkMultiDeclarationUtilities {
		if utility.Reading.Count > len(utility.Reading.Order) {
			measuredGap++
		}
		sorted := PropertySort(utility.Emit(root, "", ResolvedUtilityValue{Value: sentinel}))
		if sorted.Count > len(sorted.Order) {
			emittedGap++
		}
	}

	t.Logf("roots reading fewer positions than declarations: %d measured, %d emitted", measuredGap, emittedGap)
	if measuredGap != emittedGap {
		t.Errorf("%d roots are measured with a count above their order length and %d emit that way; the ported handles disagree with the table about which roots have invisible declarations", measuredGap, emittedGap)
	}
	if measuredGap < 25 {
		t.Errorf("only %d roots show the gap; 30 do at 4.3.3, so this measured the wrong population", measuredGap)
	}
}
