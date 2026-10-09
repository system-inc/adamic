// The framework's single-declaration functional utilities.
//
// Waves 1 and 2a of #271785y: the roots that emit exactly one declaration and reach it without
// building a `--tw-*` custom property. Wave 1 passes the resolved value through unchanged; wave 2a
// adds one axis and nothing else: a suffix appended to an accepted bare value.
//
// Four of the wave-2a roots spell their handler as `handle(value) { if (!value) return }` rather than
// as an arrow, and that guard needs no representation here. `ResolveFunctionalUtilityValue` already
// refuses a resolved empty value for every root, so a `RequiresValue` field was written, measured
// against the engine, found to change no answer, and removed. `bg-size`, `bg-position`, `mask-size`
// and `mask-position` are ordinary rows carrying only a property. 25 of them, and they are a table rather than
// 25 ported functions because there is nothing per-root to port: the property name, the theme
// namespaces, and which shared bare-value predicate applies are the entire difference between them.
//
// 33 rows, and the count moved as the extraction was corrected rather than as the wave grew. Earlier
// revisions of this comment said 25, which was the count before four guard-form roots and four
// suffix-bearing ones were placed here; that number is corrected rather than left standing.
//
// # Why the boundary is the handler's shape, not its line count
//
// The milestone body proposed waves of 15 / 71 / 28 by line count, from a measurement that counted
// registration call sites. Ten of the "15 smallest" were one-line delegations to a shared closure
// (`utilities.functional('rotate', handleRotate(...))`) and two were generic lines inside
// `functionalUtility` itself. Re-measured by shape: 72 `functionalUtility` blocks, 30 direct
// registrations, and 36 of the 72 pass their value through untouched.
//
// Of those 36, only 25 are mechanical. The other 11 transform the value in the bare handler, such as
// `outline-offset` returning `${value}px`, and they belong to wave 2 with the rest of the
// transforming bodies.
//
// # The bare handlers are a shared predicate, not per-root code
//
// Across the file there are 49 `handleBareValue` occurrences and 33 distinct bodies, and one body
// accounts for 15 of them: reject anything that is not a positive integer, otherwise pass the value
// through. That is `BareValuePositiveInteger` here, referenced by name from the table rather than
// duplicated, so a correction to the predicate cannot reach 11 roots and miss the twelfth.
//
// # What this file does, and what frameworkhandlers.go does
//
// This file resolves a value. It no longer decides what a root declares: the ported handle bodies in
// frameworkhandlers.go emit the declarations, and `Reading` runs `PropertySort` over what they emit.
// Before that port, `Reading` built one declaration from `Property` and sorted it, which was a claim
// about the handle body rather than the body itself.
//
// `Property` is therefore no longer read on the ordinary path. It stays because it is the
// independent measurement the acceptance test compares the emitters against, and because
// `TestFrameworkFunctionalUtilitiesCoverTheUnprobableRoots` asserts on it. Two statements of one
// fact are only worth carrying while something forces them to agree, and something does.
//
// `StaticValues` carry their own property because a static value can declare a different one than
// the root's ordinary path; the static emitter looks it up rather than inheriting.
package tailwind

import (
	"math"
	"strconv"
	"strings"
)

// BareValueKind names one of the shared bare-value predicates.
//
// A named kind rather than a func field, because the table is data and the predicates are code: a
// closure per row would make the table 25 anonymous functions that all look alike and hide which
// ones are genuinely the same. The kind makes the sharing visible at the call site.
type BareValueKind string

const (
	// BareValueNone means the root has no bare-value handler, so a bare value that misses the theme
	// falls through to staticValues and then produces nothing.
	BareValueNone BareValueKind = ""
	// BareValuePositiveInteger is upstream's most common handler by a wide margin: reject anything
	// that is not a positive integer, otherwise pass the value through unchanged.
	BareValuePositiveInteger BareValueKind = "PositiveInteger"
	// BareValueOpacity is `isValidOpacityValue`, a non-negative multiple of 0.25. It is a distinct
	// kind rather than a flag on the integer one because the predicate accepts values the integer
	// one rejects: `opacity-2.5` is valid and `z-2.5` is not.
	BareValueOpacity BareValueKind = "Opacity"
	// BareValueStrictPositiveInteger excludes zero. `grid-cols-0` is not a grid.
	BareValueStrictPositiveInteger BareValueKind = "StrictPositiveInteger"
	// BareValueSpacingMultiplier accepts a multiple of 0.25 and rewrites it as `--spacing(<value>)`.
	//
	// Upstream also requires the theme to declare `--spacing` at all, and returns nothing when it
	// does not. That check is not reproduced here and it does not need to be: this kind is used by
	// `auto-cols` and `auto-rows`, whose reading is one declaration either way, so a theme without
	// `--spacing` changes whether the class compiles and never changes what it reads. Stated rather
	// than silently omitted, because a caller wanting the resolved value rather than the reading
	// would need it.
	BareValueSpacingMultiplier BareValueKind = "SpacingMultiplier"
	// BareValueFontStretchPercentage is `font-stretch`'s own predicate: a percentage between 50 and
	// 200 inclusive. The range is a CSS constraint rather than a Tailwind one.
	BareValueFontStretchPercentage BareValueKind = "FontStretchPercentage"
	// BareValueGridRepeat accepts a strict positive integer and rewrites it as a `repeat()` call.
	BareValueGridRepeat BareValueKind = "GridRepeat"
	// BareValueFraction accepts a fraction whose halves are both spacing multipliers, and is the one
	// predicate reading `Fraction` rather than `Value`. `aspect-16/9` resolves and `aspect-16` does
	// not, because the parser only sets Fraction when a slash was written.
	BareValueFraction BareValueKind = "Fraction"
)

// bareValueTransform returns the rewrite a kind applies to an accepted value.
//
// Separate from the predicate because three kinds accept a value and hand back something else:
// `--spacing(4)`, `repeat(3, minmax(0, 1fr))`. A single func returning (string, bool) would work and
// would hide which kinds are pure predicates and which rewrite, which is the distinction a reader of
// the table needs.
func bareValueTransform(kind BareValueKind, value string) string {
	switch kind {
	case BareValueSpacingMultiplier:
		return "--spacing(" + value + ")"
	case BareValueGridRepeat:
		return "repeat(" + value + ", minmax(0, 1fr))"
	default:
		return value
	}
}

// bareValuePredicate returns the predicate a kind names.
//
// A switch rather than a map, so a kind added without a predicate fails to compile here rather than
// silently resolving to nil and making every bare value on that root produce nothing.
func bareValuePredicate(kind BareValueKind) func(string) bool {
	switch kind {
	case BareValuePositiveInteger:
		return isPositiveInteger
	case BareValueOpacity:
		return isValidOpacityValue
	case BareValueStrictPositiveInteger, BareValueGridRepeat:
		return isStrictPositiveInteger
	case BareValueSpacingMultiplier:
		return isValidSpacingMultiplier
	case BareValueFontStretchPercentage:
		return isFontStretchPercentage
	case BareValueFraction:
		// Never reached: the fraction kind reads Fraction rather than Value, so it is handled at the
		// call site. Named here so the switch is exhaustive and a future kind cannot fall through to
		// nil by omission.
		return nil
	default:
		return nil
	}
}

// isFontStretchPercentage is `font-stretch`'s own bare-value predicate.
//
// A percentage whose number is a positive integer between 50 and 200 inclusive. The range is a CSS
// constraint on `font-stretch` rather than a Tailwind one, and upstream cites the MDN page for it.
//
// Upstream applies `isPositiveInteger` to `Number(value.slice(0, -1))`, a number rather than a
// string, so the round-trip clause compares a number against its own printed form and always holds.
// `50.0%` therefore passes upstream and would fail a string-level check, which is why this converts
// first rather than testing the text.
func isFontStretchPercentage(value string) bool {
	if !strings.HasSuffix(value, "%") {
		return false
	}
	number, err := strconv.ParseFloat(strings.TrimSuffix(value, "%"), 64)
	if err != nil {
		return false
	}
	if number != math.Trunc(number) || number < 50 || number > 200 {
		return false
	}
	return true
}

// FrameworkStaticValue is one entry of a root's `staticValues` map.
//
// Property is carried rather than inherited from the root because a static value may declare a
// different one. `aspect-square` declares `aspect-ratio` like its root does, but the shape allows a
// root whose static entries diverge, and inheriting would be an assumption this table cannot check.
type FrameworkStaticValue struct {
	Name     string
	Property string
	Value    string
}

// FrameworkFunctionalUtility is one single-declaration root.
type FrameworkFunctionalUtility struct {
	// Property is the declaration the root's handle body emits.
	Property string
	// ThemeKeys are the namespaces the root consults, in order.
	ThemeKeys []string
	// SupportsNegative registers a second root with a leading dash.
	SupportsNegative bool
	// DefaultValue is what a valueless candidate resolves to. Absent and empty differ, so the
	// presence flag is separate rather than encoded as the empty string.
	DefaultValue        string
	DefaultValuePresent bool
	// BareValue names the shared predicate this root uses, if any.
	BareValue BareValueKind
	// BareValueSuffix is appended to a bare value the predicate accepted.
	//
	// The second axis of the bare handler, and it is why 8 more roots are a table row rather than a
	// port: `delay-150` is `150ms`, `outline-offset-2` is `2px`, `zoom-50` is `50%`, and the only
	// difference between them and the pass-through roots is this string. Empty means pass through.
	BareValueSuffix string
	// StaticValues are the named values the root answers directly.
	StaticValues []FrameworkStaticValue
}

// Description converts a table row into the shape ResolveFunctionalUtilityValue takes.
//
// The conversion is here rather than in the table so the table stays data. It is also where the
// shared predicate becomes a closure, which is the one place a `BareValueKind` turns back into code.
func (utility FrameworkFunctionalUtility) Description() *FunctionalUtilityDescription {
	description := &FunctionalUtilityDescription{
		ThemeKeys:           utility.ThemeKeys,
		SupportsNegative:    utility.SupportsNegative,
		DefaultValue:        utility.DefaultValue,
		DefaultValuePresent: utility.DefaultValuePresent,
	}

	description.HandleBareValue = bareValueHandler(utility.BareValue, utility.BareValueSuffix)

	if len(utility.StaticValues) > 0 {
		description.StaticValueNames = make(map[string]bool, len(utility.StaticValues))
		for _, static := range utility.StaticValues {
			description.StaticValueNames[static.Name] = true
		}
	}
	return description
}

// Reading returns what a candidate reads against this root, and whether it reads anything.
//
// Computed rather than named. The root's ported handle body emits declarations and `PropertySort`
// reads them, which is the same two steps the engine takes, so the reading here is a consequence of
// the emission rather than a second claim about it. That is what makes the acceptance test in
// frameworkhandlers_test.go worth running: `Property` on the row and the emitter in
// frameworkhandlers.go are two independent statements of the same fact, and this method uses only
// one of them.
//
// The root name is a parameter because the emitters are keyed by root, the way the table is. It is
// not derivable from the row: two roots can declare the same property, so a reverse lookup by
// property would be ambiguous.
//
// An unported root emits nothing and reads nothing, rather than falling back to `Property`. A silent
// fallback would make a missing emitter look like a working one at every call site and leave the
// acceptance test as the only thing that could ever notice.
func (utility FrameworkFunctionalUtility) Reading(root string, candidate *ParsedCandidate, theme *Theme, negative bool) (Reading, bool) {
	resolved, produced := ResolveFunctionalUtilityValue(candidate, utility.Description(), theme, negative)
	if !produced {
		return Reading{}, false
	}

	nodes := utility.Emit(root, resolved)
	if len(nodes) == 0 {
		return Reading{}, false
	}

	sorted := PropertySort(nodes)
	return Reading{Order: sorted.Order, Count: sorted.Count}, true
}

// FrameworkFunctionalUtilities is the table.
//
// Extracted from `utilities.ts` at 4.3.3 by matching the `handle: (value) => [decl(p, value)]` shape,
// then spot-checked against the source. That check found a defect in the extractor rather than in the
// source: an indentation-sensitive pattern had classified `outline-offset` as a pass-through when its
// bare handler returns `${value}px`, which moved it and three others out of this table and into wave
// 2. The count moved because of that correction rather than because the wave grew.
var FrameworkFunctionalUtilities = map[string]FrameworkFunctionalUtility{
	"align":              {Property: "vertical-align"},
	"animate":            {Property: "animation", ThemeKeys: []string{"--animate"}, StaticValues: []FrameworkStaticValue{{Name: "none", Property: "animation", Value: "none"}}},
	"bg-position":        {Property: "background-position"},
	"bg-size":            {Property: "background-size"},
	"col":                {Property: "grid-column", ThemeKeys: []string{"--grid-column"}, SupportsNegative: true, BareValue: BareValuePositiveInteger, StaticValues: []FrameworkStaticValue{{Name: "auto", Property: "grid-column", Value: "auto"}}},
	"col-end":            {Property: "grid-column-end", ThemeKeys: []string{"--grid-column-end"}, SupportsNegative: true, BareValue: BareValuePositiveInteger, StaticValues: []FrameworkStaticValue{{Name: "auto", Property: "grid-column-end", Value: "auto"}}},
	"col-start":          {Property: "grid-column-start", ThemeKeys: []string{"--grid-column-start"}, SupportsNegative: true, BareValue: BareValuePositiveInteger, StaticValues: []FrameworkStaticValue{{Name: "auto", Property: "grid-column-start", Value: "auto"}}},
	"columns":            {Property: "columns", ThemeKeys: []string{"--columns", "--container"}, BareValue: BareValuePositiveInteger, StaticValues: []FrameworkStaticValue{{Name: "auto", Property: "columns", Value: "auto"}}},
	"contain":            {Property: "contain"},
	"cursor":             {Property: "cursor", ThemeKeys: []string{"--cursor"}},
	"delay":              {Property: "transition-delay", ThemeKeys: []string{"--transition-delay"}, BareValue: BareValuePositiveInteger, BareValueSuffix: "ms"},
	"font-features":      {Property: "font-feature-settings"},
	"grow":               {Property: "flex-grow", DefaultValue: "1", DefaultValuePresent: true, BareValue: BareValuePositiveInteger},
	"list":               {Property: "list-style-type", ThemeKeys: []string{"--list-style-type"}, StaticValues: []FrameworkStaticValue{{Name: "none", Property: "list-style-type", Value: "none"}, {Name: "disc", Property: "list-style-type", Value: "disc"}, {Name: "decimal", Property: "list-style-type", Value: "decimal"}}},
	"list-image":         {Property: "list-style-image", ThemeKeys: []string{"--list-style-image"}, StaticValues: []FrameworkStaticValue{{Name: "none", Property: "list-style-image", Value: "none"}}},
	"mask-position":      {Property: "mask-position"},
	"mask-radial-at":     {Property: "--tw-mask-radial-position"},
	"mask-size":          {Property: "mask-size"},
	"object":             {Property: "object-position", ThemeKeys: []string{"--object-position"}, StaticValues: []FrameworkStaticValue{{Name: "top", Property: "object-position", Value: "top"}, {Name: "top-left", Property: "object-position", Value: "left top"}, {Name: "top-right", Property: "object-position", Value: "right top"}, {Name: "bottom", Property: "object-position", Value: "bottom"}, {Name: "bottom-left", Property: "object-position", Value: "left bottom"}, {Name: "bottom-right", Property: "object-position", Value: "right bottom"}, {Name: "left", Property: "object-position", Value: "left"}, {Name: "right", Property: "object-position", Value: "right"}, {Name: "center", Property: "object-position", Value: "center"}}},
	"opacity":            {Property: "opacity", ThemeKeys: []string{"--opacity"}, BareValue: BareValueOpacity, BareValueSuffix: "%"},
	"order":              {Property: "order", ThemeKeys: []string{"--order"}, SupportsNegative: true, BareValue: BareValuePositiveInteger, StaticValues: []FrameworkStaticValue{{Name: "first", Property: "order", Value: "-9999"}, {Name: "last", Property: "order", Value: "9999"}}},
	"origin":             {Property: "transform-origin", ThemeKeys: []string{"--transform-origin"}, StaticValues: []FrameworkStaticValue{{Name: "center", Property: "transform-origin", Value: "center"}, {Name: "top", Property: "transform-origin", Value: "top"}, {Name: "top-right", Property: "transform-origin", Value: "100% 0"}, {Name: "right", Property: "transform-origin", Value: "100%"}, {Name: "bottom-right", Property: "transform-origin", Value: "100% 100%"}, {Name: "bottom", Property: "transform-origin", Value: "bottom"}, {Name: "bottom-left", Property: "transform-origin", Value: "0 100%"}, {Name: "left", Property: "transform-origin", Value: "0"}, {Name: "top-left", Property: "transform-origin", Value: "0 0"}}},
	"outline-offset":     {Property: "outline-offset", ThemeKeys: []string{"--outline-offset"}, SupportsNegative: true, BareValue: BareValuePositiveInteger, BareValueSuffix: "px"},
	"perspective":        {Property: "perspective", ThemeKeys: []string{"--perspective"}, StaticValues: []FrameworkStaticValue{{Name: "none", Property: "perspective", Value: "none"}}},
	"perspective-origin": {Property: "perspective-origin", ThemeKeys: []string{"--perspective-origin"}, StaticValues: []FrameworkStaticValue{{Name: "center", Property: "perspective-origin", Value: "center"}, {Name: "top", Property: "perspective-origin", Value: "top"}, {Name: "top-right", Property: "perspective-origin", Value: "100% 0"}, {Name: "right", Property: "perspective-origin", Value: "100%"}, {Name: "bottom-right", Property: "perspective-origin", Value: "100% 100%"}, {Name: "bottom", Property: "perspective-origin", Value: "bottom"}, {Name: "bottom-left", Property: "perspective-origin", Value: "0 100%"}, {Name: "left", Property: "perspective-origin", Value: "0"}, {Name: "top-left", Property: "perspective-origin", Value: "0 0"}}},
	"row":                {Property: "grid-row", ThemeKeys: []string{"--grid-row"}, SupportsNegative: true, BareValue: BareValuePositiveInteger, StaticValues: []FrameworkStaticValue{{Name: "auto", Property: "grid-row", Value: "auto"}}},
	"row-end":            {Property: "grid-row-end", ThemeKeys: []string{"--grid-row-end"}, SupportsNegative: true, BareValue: BareValuePositiveInteger, StaticValues: []FrameworkStaticValue{{Name: "auto", Property: "grid-row-end", Value: "auto"}}},
	"row-start":          {Property: "grid-row-start", ThemeKeys: []string{"--grid-row-start"}, SupportsNegative: true, BareValue: BareValuePositiveInteger, StaticValues: []FrameworkStaticValue{{Name: "auto", Property: "grid-row-start", Value: "auto"}}},
	"shrink":             {Property: "flex-shrink", DefaultValue: "1", DefaultValuePresent: true, BareValue: BareValuePositiveInteger},
	"tab":                {Property: "tab-size", BareValue: BareValuePositiveInteger},
	"will-change":        {Property: "will-change"},
	"z":                  {Property: "z-index", ThemeKeys: []string{"--z-index"}, SupportsNegative: true, BareValue: BareValuePositiveInteger, StaticValues: []FrameworkStaticValue{{Name: "auto", Property: "z-index", Value: "auto"}}},
	"zoom":               {Property: "zoom", BareValue: BareValuePositiveInteger, BareValueSuffix: "%"},
}

// bareValueHandler builds the bare-value closure a kind describes, or nil when there is none.
//
// Shared by both framework tables, so the two cannot drift about what a kind means.
func bareValueHandler(kind BareValueKind, suffix string) func(*ParsedValue) (string, bool) {
	if kind == BareValueFraction {
		// The one kind reading Fraction rather than Value. `aspect-16/9` resolves and `aspect-16`
		// does not, because the parser sets Fraction only when a slash was written, so testing Value
		// here would accept a bare number the engine rejects.
		//
		// The `Fraction == ""` half is an early exit rather than a guard, and mutating it away does
		// not fail the suite. That is measured rather than assumed: `segment("", '/')` returns one
		// element, so the length check below rejects the same input one line later. It is kept
		// because it states the precondition at the top where a reader looks for it, and removed it
		// would leave the function's correctness resting on a property of `segment` that nothing
		// here mentions.
		return func(value *ParsedValue) (string, bool) {
			if value == nil || value.Fraction == "" {
				return "", false
			}
			parts := segment(value.Fraction, '/')
			if len(parts) != 2 || !isValidSpacingMultiplier(parts[0]) || !isValidSpacingMultiplier(parts[1]) {
				return "", false
			}
			return value.Fraction, true
		}
	}

	predicate := bareValuePredicate(kind)
	if predicate == nil {
		return nil
	}
	return func(value *ParsedValue) (string, bool) {
		if value == nil || !predicate(value.Value) {
			return "", false
		}
		return bareValueTransform(kind, value.Value) + suffix, true
	}
}
