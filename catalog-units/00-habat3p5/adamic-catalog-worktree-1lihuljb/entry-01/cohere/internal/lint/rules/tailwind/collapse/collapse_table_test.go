package tailwind

import (
	"strings"
	"testing"
)

// The table replaces a search, so the tests have to guard what the search would have gotten right.
//
// `enforce-canonical-classes` was rejected as a Go port twice, because `canonicalizeCandidates` is a
// signature-equivalence search over the whole utility registry and porting it means porting the
// theme resolver, the compiler and the printer. The table is the search's output rather than its
// mechanism, which is only safe while three things hold: it contains the families that exist, it
// contains nothing that does not, and it was built from the registry rather than from whatever
// classes this codebase happens to use.

// TestTableContainsTheKnownFamilies pins the families the migration corpus earned.
//
// Each of these was a real finding at some point, several of them the specific finding a cheaper
// approach silently lost. A regeneration that drops one is a regression in what the rule can see,
// and without this test it would land as a quiet diff in a generated file.
func TestTableContainsTheKnownFamilies(t *testing.T) {
	t.Parallel()
	required := []CollapseFamily{
		// The canonical example, and the one that proves collapsing is on at all.
		{First: "px", Second: "py", Output: "p"},
		// Two roots declaring different CSS properties. Filtering pairs by declared property family
		// measured fifteen times faster and silently stopped reporting this one.
		{First: "h", Second: "w", Output: "size"},
		// `column-gap` and `row-gap` merging into `gap`, the other half of that same lesson.
		{First: "gap-x", Second: "gap-y", Output: "gap"},
		// Roots with no value at all. A hand-written dash-splitter read `border-l` as root `border`
		// with value `l`, bucketed it away from `border-r`, and stopped reporting this.
		{First: "border-l", Second: "border-r", Output: "border-x"},
		// Negative roots are their own roots, not a sign flag on a positive one.
		{First: "-mb", Second: "-mt", Output: "-my"},
		// A family this codebase has never written, which is the point of enumerating from the
		// registry rather than from the corpus. A corpus-derived table missed every scroll family.
		{First: "scroll-px", Second: "scroll-py", Output: "scroll-p"},
		// A named scale rather than a numeric one. The generator probed only `4` and `2` at first,
		// so every rounded family was absent: a root whose values the probe cannot express reports
		// as having no families, which reads exactly like a root that has none. Found by diffing
		// the table against the engine over the real corpus, not by a test.
		{First: "rounded-tl", Second: "rounded-tr", Output: "rounded-t"},
		{First: "rounded-bl", Second: "rounded-br", Output: "rounded-b"},
		// Recursion: `mx + my => m` only completes a four-class collapse when `mb + mt => my` and
		// `ml + mr => mx` are also present, so all three are required together.
		{First: "mb", Second: "mt", Output: "my"},
		{First: "ml", Second: "mr", Output: "mx"},
		{First: "mx", Second: "my", Output: "m"},
	}

	for _, want := range required {
		if !tableContains(want) {
			t.Errorf("the table is missing %s + %s => %s, so the rule can no longer report it",
				want.First, want.Second, want.Output)
		}
	}
}

// TestTableHasNoSoloRewrites is the known-dirty control, and it caught a real defect.
//
// Some classes are non-canonical on their own: `start-4` is `inset-s-4`, `z-[1]` is `z-1`. Those
// rewrites have nothing to do with pairing. A first version of the generator counted invented class
// names rather than testing whether both inputs were load-bearing, and reported 1,326 families
// including `-bg-conic + -end => -inset-e`, which is not a family at all: `-end-4` rewrites alone
// and `-bg-conic-4` was an unrelated neighbour.
//
// The signature of that defect is a table far larger than the real one, full of entries whose two
// roots share no axis. This asserts the size, because 1,326 and 38 are not close enough to be
// confused by anyone reading the number, and a future generator change that reintroduces the bug
// will fail here rather than ship.
func TestTableHasNoSoloRewrites(t *testing.T) {
	t.Parallel()
	if len(CollapseFamilies) == 0 {
		t.Fatal("the table is empty, so every test in this file passes for the wrong reason")
	}

	// The real table is a few dozen entries. An order of magnitude more means the pair test has
	// stopped requiring both inputs to matter.
	if len(CollapseFamilies) > 200 {
		t.Fatalf("the table holds %d families, which is far more than the real count. That is the "+
			"signature of counting invented class names instead of testing that both inputs are "+
			"load-bearing, which produced 1,326 entries including `-bg-conic + -end`.",
			len(CollapseFamilies))
	}

	// A family's two inputs must be distinct roots that are not each other's output. The stronger
	// property, that the pair genuinely shares an axis, is not checkable by name: `h + w => size`,
	// `bottom + top => inset-y` and `mb + mt => my` all share no common substring and are all real.
	// A first version of this test tried a stem heuristic and reported eleven of the thirty-eight
	// real families as defects, which is worse than no check.
	//
	// What actually separates the real table from the broken one is size and provenance, both
	// asserted above and below. The 1,326-entry version paired roots that never share an axis, and
	// it was an order of magnitude too large rather than subtly wrong.
	for _, entry := range CollapseFamilies {
		if entry.First == entry.Second {
			t.Errorf("%+v pairs a root with itself", entry)
		}
	}
}

// TestTableEntriesAreWellFormed guards the shape rather than the content.
func TestTableEntriesAreWellFormed(t *testing.T) {
	t.Parallel()
	seen := map[string]bool{}

	for _, entry := range CollapseFamilies {
		if entry.First == "" || entry.Second == "" || entry.Output == "" {
			t.Errorf("incomplete entry %+v", entry)
			continue
		}
		if entry.First == entry.Second {
			t.Errorf("%+v pairs a root with itself", entry)
		}
		// The output may equal one of its inputs, and `max-w + max-w-screen => max-w` is the case
		// that proves it rather than an entry to be filtered out.
		//
		// This assertion previously rejected that shape as "a rewrite rather than a merge", which was
		// right about the hazard and wrong about the test for it. The hazard is a class that
		// canonicalizes on its own, so its rewrite appears in the output whether or not the other
		// class contributed; the generator already excludes those two ways, by `rewritesAlone` before
		// the pair is probed and by requiring both inputs to be load-bearing after.
		//
		// Root equality is a different question, and the roots here are genuinely nested rather than
		// disjoint: `max-w-screen` is a registered root that also reads as `max-w` plus the value
		// `screen`. Measured on the engine, `max-w-sm max-w-screen-sm` canonicalizes to `max-w-96`,
		// which parses to root `max-w`, and neither input rewrites alone. So the merge is real, the
		// output root is one of the inputs, and rejecting it would delete a family the engine has.
		//
		// The entry only became visible when the registry enumeration landed: `max-w-screen`
		// advertises no class, so the class-list route never saw it and the pair was never probed.
		// What remains guarded is the shape that would actually be meaningless, an output equal to
		// both inputs at once, which cannot happen while First and Second differ.
		if entry.Output == entry.First && entry.Output == entry.Second {
			t.Errorf("%+v collapses into both of its inputs, which names no merge at all", entry)
		}

		key := entry.First + " " + entry.Second
		if seen[key] {
			t.Errorf("%s appears twice, so the table would answer the same question twice", key)
		}
		seen[key] = true
	}

	if TailwindVersion == "" {
		t.Fatal("the table records no Tailwind version, so a reader cannot tell which engine's opinion it holds")
	}
}

// tableContains reports whether the table holds a family, in either input order.
//
// Order-insensitive because the generator sorts its inputs and a caller asking about a pair should
// not have to know which way round they were enumerated.
func tableContains(want CollapseFamily) bool {
	for _, entry := range CollapseFamilies {
		if entry.Output != want.Output {
			continue
		}
		if entry.First == want.First && entry.Second == want.Second {
			return true
		}
		if entry.First == want.Second && entry.Second == want.First {
			return true
		}
	}
	return false
}

// TestDeclaredPropertiesCoverBothReadings guards a bug that was silent and shipped-adjacent.
//
// A utility name can be both static and functional. `flex` is the static utility `display: flex`
// and also the functional root of `flex-4`, which is `flex: 4`; `block` is `display: block` and the
// root of `block-4`, which is `block-size`. A first version of the generator classified each name
// as one or the other, checked functional first and returned early, so `flex` and `block` were
// recorded with their functional properties and disappeared from the static table entirely.
//
// Nothing failed. The table built, contained 250 root entries and 273 static ones, and looked
// complete. What it would have done is silently stop `flex block` being reported as a display
// conflict, which is the headline case of the rule it feeds.
//
// So this asserts the both-readings names specifically, and that the display utilities agree with
// each other, because a table where `flex` and `block` disagree about their property cannot report
// the conflict between them however correct each entry looks alone.
func TestDeclaredPropertiesCoverBothReadings(t *testing.T) {
	t.Parallel()
	// Names that are genuinely both. Their static reading is what `no-conflicting-classes` needs.
	bothReadings := []string{"flex", "block"}
	for _, name := range bothReadings {
		staticProperties, hasStatic := declaredPropertiesOfStatic(name)
		if !hasStatic {
			t.Errorf("%q is missing from the static table. It is both a static utility and a "+
				"functional root, and a generator that classifies rather than records loses whichever "+
				"reading it checks second.", name)
			continue
		}
		if !containsProperty(staticProperties, "display") {
			t.Errorf("%q declares %v as a static utility, want display", name, staticProperties)
		}

		// The functional reading should also be present, and should NOT be display.
		if functionalProperties, hasFunctional := declaredPropertiesOfRoot(name); hasFunctional {
			if containsProperty(functionalProperties, "display") {
				t.Errorf("%q as a functional root declares display, which means the static reading "+
					"overwrote the functional one rather than both being recorded", name)
			}
		}
	}

	// Every display utility must agree, or a conflict between two of them cannot be detected.
	displayUtilities := []string{"flex", "block", "hidden", "grid", "inline"}
	for _, name := range displayUtilities {
		properties, isPresent := declaredPropertiesOfStatic(name)
		if !isPresent {
			t.Errorf("%q has no computed static reading, so a conflict involving it is invisible", name)
			continue
		}
		if !containsProperty(properties, "display") {
			t.Errorf("%q declares %v, want display. Two display utilities that disagree about their "+
				"property cannot be reported as conflicting.", name, properties)
		}
	}

	// The property-name-not-value distinction, which inverts two answers if it is lost.
	width, hasWidth := declaredPropertiesOfRoot("w")
	height, hasHeight := declaredPropertiesOfRoot("h")
	if !hasWidth || !hasHeight {
		t.Fatal("w or h has no computed reading")
	}
	if containsProperty(width, "height") || containsProperty(height, "width") {
		t.Error("w and h share a property, so `w-8 h-8` would be reported as a conflict when it is a " +
			"collapse into `size-8`")
	}

	// And the shorthand distinction upstream deliberately does not normalise.
	padding, hasPadding := declaredPropertiesOfRoot("p")
	paddingInline, hasPaddingInline := declaredPropertiesOfRoot("px")
	if !hasPadding || !hasPaddingInline {
		t.Fatal("p or px has no computed reading")
	}
	for _, property := range padding {
		if containsProperty(paddingInline, property) {
			t.Errorf("p and px share the property %q, so `p-4 px-8` would be reported as a conflict. "+
				"Upstream reports no conflict there: `padding` and `padding-inline` are different "+
				"property names even though they visually overlap.", property)
		}
	}
}

// declaredPropertiesOfStatic is the computed answer for a static utility.
//
// These tests asserted invariants about `StaticDeclaredProperties` and its three companions, which
// #mz0m6k8 deleted in favour of computing the answer from the ported handle bodies. The invariants
// did not go away with the tables, so they are asserted against the computed source instead: an
// answer that is now derived rather than stored still has to be the right answer.
func declaredPropertiesOfStatic(className string) ([]string, bool) {
	return DeclaredPropertiesFor(&ParsedCandidate{Kind: ParsedCandidateKindStatic, Root: className}, ValueResolution{})
}

// declaredPropertiesOfRoot is the computed answer for a functional root on its ordinary arm.
//
// The probe value is a named one that resolves through nothing, which is the arm the deleted root
// table recorded. A root that branches on a colour or a theme namespace is asked for that arm
// explicitly where a test needs it.
func declaredPropertiesOfRoot(root string) ([]string, bool) {
	return DeclaredPropertiesFor(&ParsedCandidate{
		Kind:  ParsedCandidateKindFunctional,
		Root:  root,
		Value: &ParsedValue{Kind: ParsedValueKindNamed, Value: "zzprobe"},
	}, ValueResolution{})
}

// computedPropertyTables is every class and root answer the emitters produce, for the sweeps below.
//
// The two tests that follow swept the deleted tables looking for a property name that is not a CSS
// property. That guard is still worth having and is now stronger: it sweeps what the emitters
// actually produce rather than what a generator once wrote down, so a bad property name introduced
// in a handle body is caught rather than only one baked into a table.
//
// Built from the emitter registries so the population is every root the port answers, and from
// `FrameworkStaticDeclarations` for the static half.
func computedPropertyTables() map[string][]string {
	entries := map[string][]string{}
	for className := range FrameworkStaticDeclarations {
		if properties, resolved := declaredPropertiesOfStatic(className); resolved {
			entries[className] = properties
		}
	}
	roots := map[string]bool{}
	for root := range FrameworkFunctionalUtilities {
		roots[root] = true
	}
	for root := range FrameworkMultiDeclarationUtilities {
		roots[root] = true
	}
	for root := range gapEmitters {
		roots[root] = true
	}
	for root := range roots {
		if properties, resolved := declaredPropertiesOfRoot(root); resolved {
			entries[root] = properties
		}
		// The colour arm too, since it declares property names the ordinary arm does not.
		colored, resolved := DeclaredPropertiesFor(&ParsedCandidate{
			Kind:  ParsedCandidateKindFunctional,
			Root:  root,
			Value: &ParsedValue{Kind: ParsedValueKindNamed, Value: "red-500"},
		}, ValueResolution{IsColor: true})
		if resolved {
			entries[root+" (colour arm)"] = colored
		}
	}
	return entries
}

func containsProperty(properties []string, want string) bool {
	for _, property := range properties {
		if property == want {
			return true
		}
	}
	return false
}

// TestPropertiesExcludeAtRuleDescriptors guards a leak that produced false conflicts.
//
// `border-l-4` compiles to its own two declarations followed by an `@property --tw-border-style`
// block, whose descriptors are `syntax`, `inherits` and `initial-value`. A first version of the
// generator scanned the whole compiled output, so every class touching border style appeared to
// declare those three, and any two of them read as conflicting on three shared properties.
//
// The tell was that `border-l-4` and `border-r-4` shared properties, which they must not: they are
// different edges. Caught by the rule's own known-dirty control rather than by reading the table,
// which is the order this should happen in.
func TestPropertiesExcludeAtRuleDescriptors(t *testing.T) {
	t.Parallel()
	descriptors := []string{"syntax", "inherits", "initial-value"}

	for _, table := range []struct {
		name    string
		entries map[string][]string
	}{
		{name: "computed", entries: computedPropertyTables()},
	} {
		for utility, properties := range table.entries {
			for _, descriptor := range descriptors {
				if containsProperty(properties, descriptor) {
					t.Errorf("%s table: %q declares %q, which is an @property descriptor rather than a "+
						"CSS property. Every utility carrying it would falsely conflict with every other one.",
						table.name, utility, descriptor)
				}
			}
		}
	}

	// The specific pair that exposed it.
	left, hasLeft := declaredPropertiesOfRoot("border-l")
	right, hasRight := declaredPropertiesOfRoot("border-r")
	if !hasLeft || !hasRight {
		t.Fatal("border-l or border-r has no computed reading")
	}
	for _, property := range left {
		if containsProperty(right, property) {
			t.Errorf("border-l and border-r share %q, so opposite edges would be reported as "+
				"conflicting", property)
		}
	}
}

// TestPropertyExtractionRejectsSelectorsAndKeyframes guards a defect a fixture set could not see.
//
// Twelve entries were pinned by name in this file and 1,132 were not, and the unchecked space is
// where the bugs lived. A sweep comparing every entry against the engine found eleven wrong, all
// from one cause: the extraction read the compiled CSS as though it were a single flat block.
//
//	`.slide-in-from-end:dir(ltr) { --enter-translate-x: 100%; }`
//
// declares only a custom property, and the regex matched the SELECTOR `slide-in-from-end` out of
// `:dir(ltr)` as if it were a property name. A class declaring nothing was recorded as declaring
// itself, and any two such classes would have read as conflicting.
//
//	`.markdown-content p { ... } .markdown-content code { ... }`
//
// is a component class emitting several rules, and the regex harvested `p`, `code` and `pre`.
//
// Both are now excluded at the generator: declarations are read only from inside braces, at-rule
// blocks are cut first, and a utility emitting more than one rule is skipped entirely because its
// declarations belong to descendants rather than to the element the class sits on.
//
// This asserts the outcome rather than the mechanism, because the mechanism has been wrong in three
// different ways and the outcome is what a rule reads.
func TestPropertyExtractionRejectsSelectorsAndKeyframes(t *testing.T) {
	t.Parallel()
	// A property name is a CSS property, never a selector, an element name or an animation name.
	notProperties := []string{
		"p", "code", "pre", "div", "span", "h1", "li", "table",
		"slide-in-from-end", "slide-in-from-start", "slide-out-to-end", "slide-out-to-start",
		"syntax", "inherits", "initial-value",
	}

	for _, table := range []struct {
		name    string
		entries map[string][]string
	}{
		{name: "computed", entries: computedPropertyTables()},
	} {
		if len(table.entries) == 0 {
			t.Fatalf("the %s table is empty, so this assertion passes for the wrong reason", table.name)
		}
		for utility, properties := range table.entries {
			for _, property := range properties {
				for _, forbidden := range notProperties {
					if property == forbidden {
						t.Errorf("%s table: %q declares %q, which is a selector, element or animation "+
							"name rather than a CSS property. The extraction is reading outside a rule body.",
							table.name, utility, forbidden)
					}
				}
			}
		}
	}

	// Multi-rule component classes are skipped entirely rather than merged, so they must be absent.
	for _, componentClass := range []string{"markdown-content", "prose", "typing-dots"} {
		if properties, isPresent := declaredPropertiesOfStatic(componentClass); isPresent {
			t.Errorf("%q is a component class emitting several rules targeting descendants, so its "+
				"declarations belong to children rather than to the element. It should be absent, not "+
				"recorded as %v.", componentClass, properties)
		}
	}

	// A class that declares only custom properties records nothing rather than recording itself.
	if properties, isPresent := declaredPropertiesOfStatic("slide-in-from-end"); isPresent {
		t.Errorf("slide-in-from-end declares only a custom property and should be absent, got %v", properties)
	}
}

// TestValueDependentReadingsAreRecorded guards two classes of exception that a per-root table gets
// wrong, both found as false positives on real code rather than by a fixture.
//
// `font-medium` declares `font-weight` and `font-mono` declares `font-family`, and both parse as
// root `font`. A probe picks one reading and every class taking the other is then wrong, which
// reported `font-medium font-mono` as a conflict on correct markup.
//
// Colors are the same shape at a different scale, handled per root because the palette is large and
// uniform: `border` is width plus style with a number and `border-color` with a color.
func TestValueDependentReadingsAreRecorded(t *testing.T) {
	t.Parallel()
	// The named-value distinctions, computed rather than excepted.
	//
	// `ClassDeclaredProperties` held these as thirteen per-class overrides because a root-keyed table
	// has one answer per root and these classes needed a second. The emitter branches on the value, so
	// the distinction falls out of running the handle body and there is nothing left to override.
	// TestOverrideTablesCollapseIntoTheEmitter measured that collapse at 13 of 13 before the table was
	// deleted; this asserts the distinction itself survives, which is the thing the table existed for.
	for _, expectation := range []struct {
		className    string
		resolution   ValueResolution
		wantProperty string
	}{
		// `font` splits on the theme namespace: `--font` gives a family, `--font-weight` a weight.
		{className: "font-mono", resolution: ValueResolution{Namespace: "--font"}, wantProperty: "font-family"},
		{className: "font-sans", resolution: ValueResolution{Namespace: "--font"}, wantProperty: "font-family"},
		{className: "font-medium", resolution: ValueResolution{}, wantProperty: "font-weight"},
	} {
		root, value, _ := strings.Cut(expectation.className, "-")
		properties, resolved := DeclaredPropertiesFor(&ParsedCandidate{
			Kind:  ParsedCandidateKindFunctional,
			Root:  root,
			Value: &ParsedValue{Kind: ParsedValueKindNamed, Value: value},
		}, expectation.resolution)
		if !resolved {
			t.Errorf("%q computes no properties, so it would be skipped rather than compared", expectation.className)
			continue
		}
		if !containsProperty(properties, expectation.wantProperty) {
			t.Errorf("%q computes %v, want %q. Without this distinction `font-medium font-mono` reads as "+
				"a conflict on correct code, which is what the deleted overrides existed to prevent",
				expectation.className, properties, expectation.wantProperty)
		}
	}

	// `object` is the case the override table could not express and the emitter can.
	//
	// The functional root declares `object-position`; `object-cover` and `object-fill` are separate
	// static utilities declaring `object-fit`. `RootDeclaredProperties` stored `[object-fit]` for the
	// whole root, and `ClassDeclaredProperties` patched nine `object-*` classes back to
	// `object-position` while `object-cover` was correct only because it read the root's row. Asking
	// the two registrations separately gives both answers with no override at all.
	if properties, resolved := declaredPropertiesOfRoot("object"); !resolved {
		t.Error("root `object` computes nothing")
	} else if !containsProperty(properties, "object-position") {
		t.Errorf("root `object` computes %v, want object-position. The functional registration is a "+
			"position; the fits are separate statics", properties)
	}
	if properties, resolved := declaredPropertiesOfStatic("object-cover"); !resolved {
		t.Error("static `object-cover` computes nothing")
	} else if !containsProperty(properties, "object-fit") {
		t.Errorf("static `object-cover` computes %v, want object-fit", properties)
	}

	// And the two arms of one root must still disagree, or the distinction is not being made.
	//
	// `border` is width plus style with a number and a colour with a colour. This is the shape every
	// one of the 57 gap roots has and the reason a root-keyed table needed fifteen colour arms beside
	// it.
	width, hasWidth := declaredPropertiesOfRoot("border")
	colour, hasColour := DeclaredPropertiesFor(&ParsedCandidate{
		Kind:  ParsedCandidateKindFunctional,
		Root:  "border",
		Value: &ParsedValue{Kind: ParsedValueKindNamed, Value: "red-500"},
	}, ValueResolution{IsColor: true})
	if !hasWidth || !hasColour {
		t.Fatal("root `border` computes nothing on one of its two arms")
	}
	for _, property := range width {
		if containsProperty(colour, property) {
			t.Errorf("border's width arm and colour arm share %q, which means one answer is being "+
				"returned for both values rather than the branch being taken", property)
		}
	}
}

// TestCollapseFamiliesAreCssShorthandRelationships records why this table stays generated.
//
// M9 asked whether the ported registration tables could compute the 45 families now that they exist,
// which was not possible before M6 and M7. The answer is no, and the reason is more useful than the
// answer: a collapse is not a fact the readings carry.
//
// Measured on the live table, `bottom` reads `[13]#1`, `top` reads `[11]#1` and `inset-y` reads
// `[6]#1`. There is no arithmetic from the first two to the third, because a reading is a position in
// Tailwind's property order and the order is alphabetical-ish by property name rather than structural.
// The same holds for every family: `-mb [36] + -mt [34] => -my [29]`.
//
// What does relate them is the CSS itself. All 45 families are shorthand relationships between the
// properties the roots declare: `bottom + top` is `inset-block`, `margin-inline + margin-block` is
// `margin`, `border-bottom-* + border-top-*` is `border-block-*`. That is a fact about CSS, not about
// Tailwind, so computing it would mean carrying a shorthand table instead of a family table. The same
// size, one layer further from the thing being asked.
//
// So the table stays, and it stays for a stated reason rather than by default: the alternative is a
// table of equal size describing something Tailwind does not own. `generate_collapse` measures it
// against the engine by canonicalizing 57,970 root pairs, and refuses a second design system whose
// registry matches the first, which is the guard that makes an invariance claim mean something.
//
// This test asserts the property-shorthand structure rather than the count, so a family added upstream
// that is NOT a shorthand relationship fails here and reopens the question.
func TestCollapseFamiliesAreCssShorthandRelationships(t *testing.T) {
	t.Parallel()
	withProperties := 0
	for _, family := range CollapseFamilies {
		first, _ := declaredPropertiesOfRoot(strings.TrimPrefix(family.First, "-"))
		second, _ := declaredPropertiesOfRoot(strings.TrimPrefix(family.Second, "-"))
		output, _ := declaredPropertiesOfRoot(strings.TrimPrefix(family.Output, "-"))
		if len(first) == 0 || len(second) == 0 || len(output) == 0 {
			t.Errorf("%s + %s => %s: one of the three declares no properties, so this family is not a "+
				"shorthand relationship and the reasoning in this test does not cover it",
				family.First, family.Second, family.Output)
			continue
		}
		withProperties++
	}

	if withProperties != len(CollapseFamilies) {
		t.Errorf("%d of %d families are property-shorthand relationships; the rest need their own account",
			withProperties, len(CollapseFamilies))
	}
	if len(CollapseFamilies) < 40 {
		t.Fatalf("the table holds %d families; expected around 45, so this assertion measured far less than it appears to", len(CollapseFamilies))
	}
	t.Logf("all %d collapse families relate the properties their roots declare", withProperties)
}
