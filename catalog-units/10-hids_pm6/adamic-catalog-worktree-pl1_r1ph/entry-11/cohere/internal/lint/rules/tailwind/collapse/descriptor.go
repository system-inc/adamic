// The descriptor model: a utility root's `{order, count}` reading as a function of its value.
//
// This is the data structure Phase 0 proved and the lookup that consumes it. The claim it rests on
// is that a functional utility's reading is decided by its value's inferred data type rather than by
// evaluating the utility, which is what lets `utilities.ts` (6,827 lines, most of it computing
// values `getPropertySort` never reads) collapse to a few hundred rows.
//
// The claim was measured before any of this was written, against the shipped engine on two
// repositories: 37,625 of 37,643 registry classes, every sampled modifier class, and 132,199 of
// 132,203 arbitrary probes over 327 roots and 527 value shapes. See
// `internal/lint/rules/tailwind/tools/generate_descriptors/README.md`, which is the specification this file implements. The
// numbers there are reproduced by running the extractor, not quoted from memory.
//
// # The descriptor is four things, and each correction cost a population of mispredictions
//
// The model as originally stated was `{typeList, readingByType}`, and it is right in outline and
// wrong in four ways that only a measurement finds. Every one of them is load-bearing here:
//
//  1. The type list is ordered, and the order is per-root. `InferDataType` returns the first match,
//     so `bg` listing `position` before `length` makes `bg-[3px]` a position reading `[254]` rather
//     than a length reading `[251]`. Fifty-eight probes turn on that single inversion. The order is
//     recovered by topological sort from observation in the extractor and carried here verbatim.
//     Sorting a TypeList, or building one from `AllDataTypes`, is a silent behavior change.
//
//  2. A bare value is resolved before it is inferred, in a fixed precedence: root literals, colour
//     keywords, theme namespaces longest-first, inference, then the `@none` reading. Theme before
//     inference is what `bold` proves: it is a `--font-weight` key that also satisfies
//     `family-name`, so inferring first reads `font-bold` as a font family. Eight registry classes
//     separate the two orderings. The last step is `@none` rather than the axis fallback, and the
//     two are different readings on 20 roots in this repository.
//
//  3. A modifier is a third axis with three states rather than a flag. `text-[3px]/50` emits a
//     `line-height` and `shadow-[3px]/50` emits an alpha, so a modifier moves the reading; `/none`
//     is a `--leading` key rather than an alpha, so it moves it differently. The modifier's value is
//     otherwise not consulted, which is what keeps the axis cheap: `/50`, `/[0.5]` and
//     `/[var(--a)]` are one bucket.
//
//  4. A type can discriminate on one axis and carry no information on another. `color` reads like
//     the fallback for `text-shadow` unmodified and is exactly what separates it under a modifier.
//     So a type's presence in TypeList is not evidence that it appears in every reading map.
//
// # What this file does not cover, and must not pretend to
//
// Two exceptions were measured, both understood, both countable, and both are answered here with
// "no descriptor answer" rather than a guess:
//
//   - 18 registry classes of the form `<animation-root>-translate-full`, from repository `@utility`
//     blocks whose arity is the count of surviving declarations. `50` is both an `integer` and a
//     `--percentage-*` key, so two declarations survive; `4` is only an integer, so one does. Arity
//     is therefore not a function of any single data type, which is the boundary of this model
//     rather than a defect in it. Answering them is the `@utility` evaluator's job.
//
//   - 4 sweep probes hitting an upstream shadow quirk on `[16/9]` values, where the handler splits
//     an arbitrary value on `/` looking for a colour slot and emits garbage CSS. No registry
//     contains them. Recorded, not modelled.
//
// A lookup that returned a plausible reading for these would be wrong in the one place the model
// knows it cannot answer, so `Lookup` reports `false` and lets the caller escalate.
//
// # Where a shortcut here returns the right answer for the wrong reason
//
// The generated table is small because readings are interned and because buckets equal to their
// axis fallback are dropped. The elision is correct for `ByType`, where a miss ends the lookup at
// the fallback, and wrong for `ByNamespace`, where a miss continues the bare-value precedence to
// inference instead. Applying it to both looked structural and symmetric while it was being
// written, and it silently changed 647 namespace buckets across 303 roots.
//
// It is worth stating plainly because of how it failed. Every one of those 647 still produced a
// well-formed reading, the table got 45 KB smaller, and the only symptom was that `font-bold` read
// `[284]#1` instead of `[288]#2`. Nothing but a comparison against the engine over the whole
// registry would have found it, which is why this package is tested that way rather than against
// its own expectations. See AxisReadings.ByNamespace.
package tailwind

import "strings"

// Reading is the `{order, count}` pair the class-order comparator sorts on.
//
// Order is the sorted property indices a utility declares, and Count is how many declarations it
// emitted. Both are needed: two utilities can declare the same properties and differ in arity, which
// is exactly what separates `slide-in-from-top-50` from `slide-in-from-top-4`.
//
// The zero Reading is not a valid reading. An absent reading is reported by the boolean every
// lookup returns, never by a sentinel value, because an empty order is itself meaningful: a class
// that declares nothing sorts last rather than first, and a control in Phase 0 got that backwards
// and reported 365,174 spurious mismatches while the model was fine.
type Reading struct {
	// Order holds property indices into the global property order, ascending.
	//
	// It is shared rather than copied. Readings are interned by the generator, so many descriptors
	// point at one Reading and the same backing array; callers must treat it as read-only. Nothing
	// in the port mutates a reading, and the alternative, copying on every lookup, would allocate on
	// the hot path of a rule that runs on every class attribute in the repository.
	Order []int
	// Count is the number of declarations the utility emitted.
	Count int
}

// Equal reports whether two readings are the same reading.
//
// Written out rather than deferring to `slices.Equal` plus a field compare so the two halves of the
// comparison stay visibly symmetric; a reading differing only in Count is the case the fixture
// corpus is thickest around.
func (reading Reading) Equal(other Reading) bool {
	if reading.Count != other.Count || len(reading.Order) != len(other.Order) {
		return false
	}
	for index, value := range reading.Order {
		if value != other.Order[index] {
			return false
		}
	}
	return true
}

// ModifierAxis is which of a descriptor's three parallel reading sets a class selects.
//
// Three states rather than a boolean, and the third is not an edge case: `/none` resolves through
// the theme as a `--leading` key while every other modifier is an alpha, so `text-sm/none` and
// `text-sm/50` read differently from each other and from `text-sm`.
type ModifierAxis int

const (
	// ModifierAbsent is a class written with no modifier at all.
	ModifierAbsent ModifierAxis = iota
	// ModifierAlpha is any modifier whose value is consumed as an alpha. `/50`, `/[0.5]` and
	// `/[var(--a)]` are one bucket, because what the modifier adds is a declaration and the
	// declaration exists whatever the alpha is.
	ModifierAlpha
	// ModifierThemed is a modifier whose value is a theme key rather than an alpha.
	ModifierThemed
)

// themedModifierValues is the set of modifier values that resolve through the theme.
//
// `none` is the only one in Tailwind 4.3.3: it is a `--leading` key, so it means something to `text`
// and nothing to `shadow`. Enumerated rather than derived, because deriving it would mean asking
// which namespaces a modifier consults, which is per-root, while the set of non-alpha modifier
// values is not.
var themedModifierValues = map[string]bool{"none": true}

// ModifierAxisFor classifies a modifier into its axis.
//
// A nil modifier is absent, which is the spelling the candidate parser uses and the reason this
// takes a pointer. A named modifier is themed only when it is in the set above; an arbitrary one
// never is, because `/[none]` is a bracketed value that happens to spell a theme key and the engine
// does not consult the theme for it.
func ModifierAxisFor(modifier *ParsedModifier) ModifierAxis {
	if modifier == nil {
		return ModifierAbsent
	}
	if modifier.Kind == ParsedModifierKindNamed && themedModifierValues[modifier.Value] {
		return ModifierThemed
	}
	return ModifierAlpha
}

// AxisReadings is one modifier axis's worth of a root's readings.
//
// The three axes hold the same shape, so they are one type used three times rather than nine fields
// on Descriptor.
type AxisReadings struct {
	// ByType maps a data type to its reading. A type in the root's TypeList and absent here reads
	// like Fallback, which is the fourth refinement above: presence in the list means the type
	// discriminates on some axis, not on this one.
	ByType map[DataType]Reading
	// ByNamespace maps a theme namespace to the reading a key in it produces. Keyed on `--color`,
	// `--font-weight` and so on, plus the two pseudo-namespaces below.
	//
	// Unlike ByType, entries here are kept even when they equal Fallback, and the difference is not
	// a matter of taste. A miss in ByType ends the lookup at the fallback, so an elided entry and an
	// absent one are the same answer. A miss in ByNamespace does not end anything: it continues the
	// bare-value precedence to inference and then to `@none`, so eliding an entry does not return
	// the fallback, it hands the value to a later step that answers differently.
	//
	// This was measured rather than reasoned about. Eliding it looked structural and correct, and it
	// silently changed 647 namespace buckets across 303 roots: `font-bold` is a `--font-weight` key
	// whose reading happens to equal `font`'s fallback, so eliding it dropped the value through to
	// inference, which types `bold` as a `family-name` and reads `[284]#1` where the engine reads
	// `[288]#2`. That is refinement 2 failing in exactly the way refinement 2 exists to prevent, and
	// it was invisible to every test that did not compare against the engine.
	ByNamespace map[string]Reading
	// Fallback is the reading for a value this axis has nothing more specific for. ByType omits
	// entries equal to it; ByNamespace does not, for the reason above.
	Fallback Reading
	// Empty is the reading for the root written with no value at all, such as `border`. Absent when
	// the root rejects that form, which is why it is a pointer: a root that reads `[]#0` for the
	// empty form and a root that refuses the empty form are different facts, and a zero Reading
	// cannot express both.
	Empty *Reading
}

// The two pseudo-namespaces, which are not theme namespaces and are keyed alongside them because
// they occupy the same step of the bare-value precedence.
const (
	// namespaceColorKeyword is the reading for Tailwind's built-in colour keywords. They are in no
	// theme namespace and infer as nothing, yet read as colours, so they have to be tested before
	// anything else can claim them.
	namespaceColorKeyword = "@colorKeyword"
	// namespaceNone is the reading for a bare value that matched no namespace and inferred as
	// nothing: bare numbers, fractions, and root-defined keywords. It is the last step of the bare
	// path rather than the axis Fallback, and the two differ on 20 roots in this repository.
	namespaceNone = "@none"
)

// colorKeywords is Tailwind's built-in colour keywords.
//
// Enumerated rather than derived because there is nothing to derive them from: they are literals in
// `utilities.ts`'s colour path, in no theme namespace, and `current` and `inherit` infer as nothing.
var colorKeywords = map[string]bool{"current": true, "inherit": true, "transparent": true}

// Descriptor is everything known about one functional utility root.
type Descriptor struct {
	// Root is the utility root, such as `bg` or `slide-in-from-top`.
	Root string
	// TypeList is the root's own type list, in the engine's order.
	//
	// The order is the first refinement above and it is carried, never derived. `InferDataType`
	// returns the first match, so re-sorting this list changes readings: `bg` puts `position` before
	// `length`, and the inverse makes `bg-[3px]` read `[251]` instead of `[254]`.
	TypeList []DataType
	// ByLiteral holds readings for values that resolve as neither a theme key nor an inferred type:
	// root-defined keywords whose reading differs from the no-namespace fallback. It applies only on
	// the unmodified axis, which is where the extractor measured it.
	ByLiteral map[string]Reading
	// Absent, Alpha and Themed are the three modifier axes.
	Absent AxisReadings
	Alpha  AxisReadings
	Themed AxisReadings
	// PerDeclaration marks a root whose arity is the count of surviving `@utility` declarations
	// rather than a function of a data type.
	//
	// These are the 18 known registry exceptions. A value can satisfy several resolution paths at
	// once, so no table row can answer them, and Lookup declines rather than returning the reading
	// the model would have guessed. Set by the generator from the repository's own `@utility`
	// blocks; answering these is task #4f04x54.
	PerDeclaration bool
}

// Table is a whole design system's descriptors.
//
// It is generated per repository and never checked in. `Namespaces` is keyed on this repository's
// own `@theme`, so a committed table would bake one repository's tokens into a file claiming to
// describe Tailwind, which is the bug this port exists to fix. Measured: generating against
// ~/Projects/ahra and www-connected-app produces byte-identical `roots` arrays but different
// namespace key sets, and a synthetic theme declaring `--color-weird-1` produces a `--color-weird`
// namespace that reads informatively on 20 roots. The structure is framework-invariant; which
// namespaces exist is not.
type Table struct {
	// TailwindVersion is the engine version the table was generated against, so a mismatch is a
	// loud failure rather than a subtly wrong sort.
	TailwindVersion string
	// Descriptors is every functional root, keyed by root.
	Descriptors map[string]*Descriptor
	// Statics is every static utility's constant reading. A static takes no value, so there is no
	// model, only a fact.
	Statics map[string]Reading
	// Namespaces is every theme namespace in the design system, longest first.
	//
	// The order is the precedence the engine applies and is part of the contract: `--drop-shadow`
	// must be tested before `--shadow`, or `drop-shadow-lg` resolves through the wrong namespace.
	Namespaces []string
	// KeysByNamespace is the set of keys in each namespace, which is what decides whether a bare
	// value is a theme key at all. Sets rather than slices because the lookup only ever asks for
	// membership, and the corpus has namespaces holding 558 keys.
	KeysByNamespace map[string]map[string]bool
	// theme is the design system's resolved theme, used by the ported registration tables.
	//
	// Unexported because it is an implementation detail of Lookup: the registration tables resolve a
	// candidate's value themselves, where the descriptor rows carry an already-resolved reading.
	theme *Theme
	// PropertyOrder maps a CSS property name to its index in Tailwind's global property order,
	// which is what an arbitrary property such as `[font:inherit]` needs and nothing else does.
	PropertyOrder map[string]int
}

// Lookup returns the reading for a parsed candidate, and whether the model can answer it.
//
// This is the whole descriptor model as executable code, and it is deliberately the same five steps
// the extractor's `predictReading` runs, in the same order, because that function is what was
// measured against the engine. A restructuring that is equivalent on the corpus and different in
// some untested corner would be a right answer for a wrong reason.
//
// It returns false rather than a guess in five situations, each of which the caller must escalate
// rather than trust: a nil or unparsed candidate, a static utility the table does not know, a root
// with no descriptor, a root whose arity is per-declaration, and a root that rejects the empty value
// form.
func (table *Table) Lookup(candidate *ParsedCandidate) (Reading, bool) {
	if candidate == nil {
		return Reading{}, false
	}

	// The ported registrations answer first, and the descriptor rows are what remains.
	//
	// M6 and M7 replaced 251 of the 315 base roots with tables built from upstream's own registration
	// shapes: a property name, the theme namespaces, a shared bare-value predicate. Those roots read
	// the same way for every value they accept, so a registration expresses them exactly.
	//
	// The 64 that remain partition their reading on the resolved type of the value: `border-[3px]` is
	// a width and `border-red-500` is a colour, and no registration shape carries both. That
	// partition is the descriptor model, so those rows are not leftovers, they are the part with no
	// simpler form.
	if reading, answered := table.frameworkReading(candidate); answered {
		return reading, true
	}

	switch candidate.Kind {
	case ParsedCandidateKindArbitrary:
		// Not a utility at all: a bare CSS declaration written as a class. Its reading is one
		// declaration whose property is the text before the colon, so there is nothing to look up.
		return table.ArbitraryPropertyReading(candidate.Property), true
	case ParsedCandidateKindStatic:
		reading, found := table.Statics[candidate.Root]
		return reading, found
	case ParsedCandidateKindFunctional:
	default:
		return Reading{}, false
	}

	descriptor, found := table.Descriptors[candidate.Root]
	if !found {
		return Reading{}, false
	}
	// The 18 known exceptions. See PerDeclaration: no row of this table can answer them, and a
	// reading returned here would be wrong in exactly the place the model knows its own boundary.
	if descriptor.PerDeclaration {
		return Reading{}, false
	}

	modifierAxis := ModifierAxisFor(candidate.Modifier)
	axis := descriptor.axisFor(modifierAxis)

	if candidate.Value == nil {
		if axis.Empty == nil {
			return Reading{}, false
		}
		return *axis.Empty, true
	}

	if candidate.Value.Kind == ParsedValueKindArbitrary {
		// An arbitrary value carries its own type when the author wrote one, as in
		// `bg-[color:var(--x)]`. The engine trusts that annotation rather than inferring, and the
		// parser has already removed it from Value, so inference below never sees it.
		if candidate.Value.DataType != "" {
			return axis.readingForType(DataType(candidate.Value.DataType)), true
		}
		inferred := InferDataType(candidate.Value.Value, descriptor.TypeList)
		if inferred == "" {
			return axis.Fallback, true
		}
		return axis.readingForType(inferred), true
	}

	return table.bareReading(descriptor, axis, candidate.Value.Value, modifierAxis), true
}

// bareReading is the bare-value precedence, and the order of these five steps is the second
// refinement above.
//
// Each step exists because the one before it left a measurable population mispredicted, and the
// riskiest single assumption available here is that "bare" means "theme lookup only": bare values
// are inferred too, which is what makes `via-25%` a percentage.
func (table *Table) bareReading(descriptor *Descriptor, axis AxisReadings, value string, modifierAxis ModifierAxis) Reading {
	// 1. Root-defined literals, measured only on the unmodified axis.
	if modifierAxis == ModifierAbsent {
		if reading, found := descriptor.ByLiteral[value]; found {
			return reading
		}
	}

	// 2. The colour keywords, before anything can claim them.
	if reading, found := axis.ByNamespace[namespaceColorKeyword]; found && colorKeywords[value] {
		return reading
	}

	// 3. The theme namespaces, longest first, and before inference. `bold` is a `--font-weight` key
	//    that also satisfies `family-name`, so inferring first reads `font-bold` as a font family.
	for _, namespace := range table.Namespaces {
		reading, hasNamespace := axis.ByNamespace[namespace]
		if !hasNamespace {
			continue
		}
		if !table.KeysByNamespace[namespace][value] {
			continue
		}
		return reading
	}

	// 4. Inference, against the root's own ordered type list.
	if inferred := InferDataType(value, descriptor.TypeList); inferred != "" {
		if reading, found := axis.ByType[inferred]; found {
			return reading
		}
	}

	// 5. Everything else: bare numbers, fractions, root-defined keywords that read like the
	//    no-namespace reading. This is distinct from the axis fallback and differs from it on 20
	//    roots in this repository.
	if reading, found := axis.ByNamespace[namespaceNone]; found {
		return reading
	}
	return axis.Fallback
}

// readingForType is the third and fourth refinements in one line: a type absent from this axis reads
// like the axis fallback, because presence in a root's TypeList means the type discriminates on
// *some* axis rather than on this one, and the generator drops every bucket equal to the fallback.
func (axis AxisReadings) readingForType(dataType DataType) Reading {
	if reading, found := axis.ByType[dataType]; found {
		return reading
	}
	return axis.Fallback
}

// axisFor selects one of the three parallel reading sets.
func (descriptor *Descriptor) axisFor(axis ModifierAxis) AxisReadings {
	switch axis {
	case ModifierAlpha:
		return descriptor.Alpha
	case ModifierThemed:
		return descriptor.Themed
	default:
		return descriptor.Absent
	}
}

// ArbitraryPropertyReading is the reading for a bare CSS declaration written as a class, such as
// `[font:inherit]`.
//
// It is not a utility and has no root and no descriptor: its reading is one declaration whose
// property is the text before the colon, so there is nothing to look up. It gets a branch rather
// than a table row. An unknown property yields an empty order, which sorts last rather than first.
func (table *Table) ArbitraryPropertyReading(property string) Reading {
	index, found := table.PropertyOrder[property]
	if !found {
		return Reading{Order: nil, Count: 1}
	}
	return Reading{Order: []int{index}, Count: 1}
}

// frameworkReading answers from the ported registration tables, and reports whether it could.
//
// Negative roots are keyed on their undashed name with the negation carried as a flag, matching how
// upstream registers `-m` as `m` with `supportsNegative`. A root in neither table is not an error: it
// is one of the 64 the descriptor rows answer.
func (table *Table) frameworkReading(candidate *ParsedCandidate) (Reading, bool) {
	if candidate.Kind != ParsedCandidateKindFunctional || table.theme == nil {
		return Reading{}, false
	}

	root, negative := candidate.Root, false
	if trimmed, cut := strings.CutPrefix(root, "-"); cut {
		root, negative = trimmed, true
	}

	if utility, known := FrameworkFunctionalUtilities[root]; known {
		if negative && !utility.SupportsNegative {
			return Reading{}, false
		}
		return utility.Reading(root, candidate, table.theme, negative)
	}
	if utility, known := FrameworkMultiDeclarationUtilities[root]; known {
		if negative && !utility.SupportsNegative {
			return Reading{}, false
		}
		return utility.ReadingFor(root, candidate, table.theme, negative)
	}
	return Reading{}, false
}

// IsColorKeyword reports whether a value is one of Tailwind's built-in colour keywords.
//
// `current`, `inherit` and `transparent` are literals in the colour path of `utilities.ts` rather
// than entries in any theme namespace, so a caller asking the theme cannot reach them and needs this
// beside it. Exported for `no-conflicting-classes`, which asks the live theme first and then here.
func IsColorKeyword(value string) bool {
	return colorKeywords[value]
}
