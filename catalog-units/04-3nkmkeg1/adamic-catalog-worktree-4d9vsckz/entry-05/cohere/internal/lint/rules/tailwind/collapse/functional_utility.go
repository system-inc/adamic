// The value-resolution pipeline every framework functional utility runs.
//
// Ported from `functionalUtility` and its inner `handleFunctionalUtility` in `src/utilities.ts` at
// Tailwind 4.3.3, read at the pinned tag rather than from the minified bundle, so a disagreement
// with the engine is a finding rather than version skew.
//
// # What this is and what it deliberately is not
//
// Upstream's `functionalUtility` does two things: it turns a candidate's value into a resolved CSS
// value, and it hands that value to the root's own `handle` body which emits declarations. This file
// is the first half only. The `handle` bodies are 114 separate functions and they are #271785y.
//
// The split is not arbitrary. The resolution above is one function shared by every functional root,
// so an inexact port corrupts every reading downstream at once, in the direction of a plausible
// answer rather than an error. The handle bodies fail one root at a time and are visible against the
// engine per root. Testing them together would let a resolution defect hide behind a handler that
// happens to emit the same properties for two different values.
//
// # The order is the contract
//
// Five branches, and their order decides the answer rather than merely organising it:
//
//  1. No value at all: `defaultValue` when the descriptor defines one, otherwise a bare theme lookup.
//     A modifier with no value produces nothing, which is why `bg-red-500/50` resolves and `bg/50`
//     does not.
//  2. An arbitrary value: taken as written, carrying its typehint if the author gave one. A modifier
//     again produces nothing.
//  3. A named value: the theme first, by `theme.resolve(fraction ?? value, themeKeys)`. Theme before
//     inference is refinement 2 from Phase 0 and it is load-bearing: `bold` is a `--font-weight` key
//     that also satisfies `family-name`, so inferring first reads `font-bold` as a font family.
//  4. Fractions, when the root supports them and the theme did not answer. `w-1/2` becomes
//     `calc(1 / 2 * 100%)` without `1/2` existing as a theme value.
//  5. Bare values, negative first when the candidate was negative, then the ordinary handler, then
//     `staticValues` as a last resort.
//
// # Two details that are easy to port wrongly and produce plausible answers
//
// A bare handler that returns a value without a `/` cancels the whole utility when a modifier is
// present. Upstream spells this `if (!value?.includes('/') && candidate.modifier) return`, and it is
// what stops `grow-2/3` from reading as `grow` with a stray modifier. Dropping the test leaves those
// classes resolving to something, which sorts, which is why it would not surface as an error.
//
// A negative candidate wraps its resolved value in `calc(<value> * -1)` and routes it through
// `addWhitespaceAroundMathOperators`. Emitting `calc(<value>*-1)` instead is the same CSS and a
// different string, and since this port compares readings rather than text it would pass every
// assertion here while disagreeing with the engine anywhere the text reaches a data-type inference.
package tailwind

import (
	"strconv"
	"strings"
)

// FunctionalUtilityDescription is upstream's `UtilityDescription`, minus the parts that emit.
//
// `handle`, `staticValues` bodies and the suggestion machinery are absent: this file resolves a
// value and never emits a declaration. `StaticValueNames` carries only the keys of `staticValues`,
// because the resolution pipeline needs to know whether a name is one, and what it expands to is the
// handle body's business.
type FunctionalUtilityDescription struct {
	// SupportsNegative registers a second root with a leading dash, and makes a resolved value be
	// negated rather than used as written.
	SupportsNegative bool
	// SupportsFractions lets `w-1/2` resolve without `1/2` being a theme key.
	SupportsFractions bool
	// ThemeKeys are the namespaces this root consults, in the order it consults them.
	ThemeKeys []string
	// DefaultValue is what a valueless candidate resolves to. Absent and empty are different: a root
	// with no default falls through to a bare theme lookup, so the presence flag is carried
	// separately rather than encoded as the empty string.
	DefaultValue        string
	DefaultValuePresent bool
	// HandleBareValue and HandleNegativeBareValue are the root's own bare-value readers. They are
	// modelled as predicates returning a value because the resolution pipeline only needs to know
	// what came back, not how the root computed it.
	HandleBareValue         func(value *ParsedValue) (string, bool)
	HandleNegativeBareValue func(value *ParsedValue) (string, bool)
	// StaticValueNames are the names `staticValues` defines, as a set.
	StaticValueNames map[string]bool
	// AcceptsModifierOnArbitrary lets an arbitrary value carry a modifier.
	//
	// False for every root registered through `functionalUtility`, where upstream's
	// `if (candidate.modifier) return` guards the arbitrary branch outright. True for the closures
	// that read a modifier themselves: a colour root takes it as the alpha, so `divide-[#abc]/10`
	// compiles, and `text` takes it as the line height, so `text-[10px]/6` and `text-[10px]/relaxed`
	// compile.
	//
	// Measured rather than reasoned: with the guard unconditional, those seven classes were the last
	// false positives `no-unknown-classes` produced on the corpus.
	AcceptsModifierOnArbitrary bool
	// Arms are the additional resolution paths a root tries, in the order it tries them.
	//
	// Empty for the 185 roots registered through `functionalUtility`, which consult one namespace
	// list and are fully described by the fields above. The 57 roots in frameworkgaphandlers.go are
	// registered as bare closures instead, and a closure can consult several namespace lists in
	// sequence with different bare-value behaviour on each.
	//
	// `borderSideUtility` is the shape, read at the v4.3.3 tag: a colour arm resolving through
	// `['--border-color', '--color']`, then a width arm resolving through `['--border-width']` with a
	// positive-integer fallback that appends `px`. `ThemeKeys` cannot express that, because the two
	// arms differ in more than their keys: the width arm refuses a modifier and the colour arm
	// accepts one.
	//
	// Tried after `ThemeKeys` rather than instead of it, so a root can have both and the 185 keep
	// resolving exactly as they did. A root with arms and no `ThemeKeys` starts at its first arm,
	// which is what every gap root does.
	Arms []FunctionalUtilityArm
}

// FunctionalUtilityArm is one resolution path inside a root's closure.
//
// Upstream a gap root's body is a sequence of blocks, each trying one way to resolve the value and
// returning if it hit. An arm is one of those blocks, and the sequence is what makes the root's
// answer depend on which one matched rather than on the value alone.
type FunctionalUtilityArm struct {
	// ThemeKeys are the namespaces this arm consults, in order.
	ThemeKeys []string
	// IsColor marks an arm that resolves through `resolveThemeColor` rather than `theme.resolve`.
	//
	// The difference is not decoration. A colour arm answers `inherit`, `transparent` and `current`
	// before consulting the theme at all, and it accepts a modifier, which is the alpha. A width arm
	// does neither: `border-red-500/50` resolves and `border-4/50` produces nothing.
	IsColor bool
	// BareValue is the arm's own bare-value handler, applied when the theme misses.
	//
	// `borderSideUtility`'s width arm accepts a positive integer and appends `px`, so `border-4`
	// resolves without `4` being a `--border-width` key. An arm with no handler ends at the theme.
	BareValue BareValueKind
	// BareValueSuffix is appended to whatever the bare handler returned, as `px` above.
	BareValueSuffix string
	// RefusesModifier marks an arm that returns nothing when the candidate carries a modifier.
	//
	// Upstream spells this as an early `if (candidate.modifier) return` inside the block, and it is
	// per-arm rather than per-root: the same class resolves or does not depending on which arm
	// claimed its value.
	RefusesModifier bool
	// InferTypes is the ordered type list this arm infers a bare value against, when it infers at
	// all rather than reading the theme.
	//
	// `maskStopUtility`'s position arm is the shape: it runs `inferDataType(value, ['number',
	// 'percentage'])` and returns nothing when neither matches, then treats the two differently. A
	// number goes through the spacing multiplier and a percentage passes through once its digits are
	// a positive integer, so `mask-x-from-1.5` resolves, `mask-x-from-3.7` does not, and
	// `mask-x-from-50%` resolves while `mask-x-from-nonsense` does not.
	//
	// An arm that infers never reads the theme for its value: upstream's block is a switch with a
	// `default: return`, so a value that infers as nothing ends the utility rather than falling
	// through to the next arm.
	//
	// Empty for an arm that reads the theme instead, which is every colour arm and every width arm.
	InferTypes []DataType
	// PercentagePassesThrough lets an inferred percentage resolve as written.
	//
	// Only meaningful beside `InferTypes`, and separate from the bare handler because the two cases
	// of that switch disagree about what a valid value is: the number case accepts a spacing
	// multiplier and the percentage case accepts any positive integer followed by `%`.
	PercentagePassesThrough bool
}

// ResolvedUtilityValue is what the pipeline produces for one candidate.
//
// `StaticValueName` is set when resolution ended at a `staticValues` entry rather than at a value,
// because that branch returns a declaration list directly upstream and never reaches `handle`. A
// caller emitting declarations must check it before using Value.
type ResolvedUtilityValue struct {
	Value           string
	DataType        string
	StaticValueName string
	IsStaticValue   bool
}

// ResolveFunctionalUtilityValue runs the pipeline for one candidate against one root's description.
//
// Returns false where upstream returns without producing a value, which is a class that generates no
// CSS at all rather than one that generates an empty declaration. The distinction matters to every
// caller in this port: a class with no rules has no reading, and scoring that as agreement is the
// failure mode the differential harness carries a five-state lattice to prevent.
func ResolveFunctionalUtilityValue(
	candidate *ParsedCandidate,
	description *FunctionalUtilityDescription,
	theme *Theme,
	negative bool,
) (ResolvedUtilityValue, bool) {
	if candidate == nil || description == nil || theme == nil {
		return ResolvedUtilityValue{}, false
	}

	resolved := ResolvedUtilityValue{}

	switch {
	case candidate.Value == nil:
		// A modifier with no value is not a utility. `bg/50` reaches here and produces nothing.
		if candidate.Modifier != nil {
			return ResolvedUtilityValue{}, false
		}
		if description.DefaultValuePresent {
			resolved.Value = description.DefaultValue
			break
		}
		// The bare theme lookup: `rounded` resolving through `--radius`. Upstream passes `null` as
		// the candidate value, which is why the present flag is false rather than the value empty.
		value, found := theme.Resolve("", false, description.ThemeKeys, 0)
		if !found {
			return ResolvedUtilityValue{}, false
		}
		resolved.Value = value

	case candidate.Value.Kind == ParsedValueKindArbitrary:
		// A modifier on an arbitrary value cancels the utility for the roots this pipeline was
		// written against, and does not for a root whose closure reads one. `divide-[#abc]/10` is an
		// alpha and `text-[10px]/6` is a line height, both of which upstream composes onto the
		// arbitrary value rather than refusing.
		if candidate.Modifier != nil && !description.AcceptsModifierOnArbitrary {
			return ResolvedUtilityValue{}, false
		}
		resolved.Value = candidate.Value.Value
		resolved.DataType = candidate.Value.DataType

	default:
		value, found := resolveNamedValue(candidate, description, theme, negative)
		if !found.produced {
			return ResolvedUtilityValue{}, false
		}
		if found.isStaticValue {
			return ResolvedUtilityValue{StaticValueName: value, IsStaticValue: true}, true
		}
		resolved.Value = value
	}

	if resolved.Value == "" && !resolved.IsStaticValue {
		return ResolvedUtilityValue{}, false
	}

	if negative {
		resolved.Value = addWhitespaceAroundMathOperators("calc(" + resolved.Value + " * -1)")
	}
	return resolved, true
}

// namedValueOutcome distinguishes the three ways the named branch can end.
//
// A bool pair rather than a sentinel string, because a static value name and a resolved CSS value
// are both strings and confusing them would emit a declaration whose value is a key.
type namedValueOutcome struct {
	produced      bool
	isStaticValue bool
}

// resolveNamedValue is branch 3 through 5, in upstream's order.
func resolveNamedValue(
	candidate *ParsedCandidate,
	description *FunctionalUtilityDescription,
	theme *Theme,
	negative bool,
) (string, namedValueOutcome) {
	// The theme, with the fraction spelling preferred when the parser produced one. `w-1/2` asks the
	// theme for `1/2` before it asks for `1`, so a theme declaring `--width-1\/2` wins over the
	// fraction arithmetic below.
	lookup := candidate.Value.Value
	if candidate.Value.Fraction != "" && !description.AcceptsModifierOnArbitrary {
		lookup = candidate.Value.Fraction
	}
	if value, found := theme.Resolve(lookup, true, description.ThemeKeys, 0); found {
		return value, namedValueOutcome{produced: true}
	}

	// Fractions, only when the root opted in and only for two positive integers. `w-1/2` becomes
	// `calc(1 / 2 * 100%)`; `w-a/b` and `w-1.5/2` produce nothing rather than falling through, which
	// is upstream returning rather than continuing.
	if description.SupportsFractions && candidate.Value.Fraction != "" {
		parts := segment(candidate.Value.Fraction, '/')
		if len(parts) != 2 || !isPositiveInteger(parts[0]) || !isPositiveInteger(parts[1]) {
			return "", namedValueOutcome{}
		}
		return "calc(" + parts[0] + " / " + parts[1] + " * 100%)", namedValueOutcome{produced: true}
	}

	// The negative bare handler, which returns directly rather than continuing to the ordinary one.
	if negative && description.HandleNegativeBareValue != nil {
		value, found := description.HandleNegativeBareValue(candidate.Value)
		if found && !strings.Contains(value, "/") && candidate.Modifier != nil {
			return "", namedValueOutcome{}
		}
		if found {
			return value, namedValueOutcome{produced: true}
		}
	}

	if description.HandleBareValue != nil {
		value, found := description.HandleBareValue(candidate.Value)
		if found {
			if !strings.Contains(value, "/") && candidate.Modifier != nil {
				return "", namedValueOutcome{}
			}
			return value, namedValueOutcome{produced: true}
		}
	}

	// `staticValues` last, and never for a negative candidate or one carrying a modifier.
	if !negative && candidate.Modifier == nil && description.StaticValueNames != nil {
		if description.StaticValueNames[candidate.Value.Value] {
			return candidate.Value.Value, namedValueOutcome{produced: true, isStaticValue: true}
		}
	}

	// The arms, in the order the closure tries them.
	//
	// After everything above rather than instead of it, so a root carrying no arms resolves exactly
	// as it did before this field existed and the 185 registered through `functionalUtility` cannot
	// move. A gap root carries no `ThemeKeys` and no bare handler of its own, so it falls straight
	// through to here.
	for index := range description.Arms {
		value, found := resolveArm(candidate, &description.Arms[index], theme)
		if found {
			return value, namedValueOutcome{produced: true}
		}
	}

	return "", namedValueOutcome{}
}

// resolveArm runs one arm of a gap root's closure against a named value.
//
// Returns false where upstream's block falls through to the next one, which is what makes the arms a
// sequence rather than a set: `border-red-500` is claimed by the colour arm and never reaches the
// width arm, and `border-4` is refused by the colour arm and resolved by the width one.
func resolveArm(candidate *ParsedCandidate, arm *FunctionalUtilityArm, theme *Theme) (string, bool) {
	if arm.RefusesModifier && candidate.Modifier != nil {
		return "", false
	}

	if arm.IsColor {
		return resolveArmColor(candidate, arm, theme)
	}

	// An inferring arm never reads the theme for its value: upstream's block is a switch with a
	// `default: return`, so a value that infers as nothing ends the utility rather than falling
	// through to whatever comes next.
	if len(arm.InferTypes) > 0 {
		return resolveArmByInferredType(candidate, arm, theme)
	}

	if value, found := theme.Resolve(candidate.Value.Value, true, arm.ThemeKeys, 0); found {
		return value, true
	}

	if arm.BareValue != BareValueNone {
		handler := bareValueHandler(arm.BareValue, arm.BareValueSuffix)
		if handler != nil {
			if value, found := handler(candidate.Value); found {
				return value, true
			}
		}
	}
	return "", false
}

// resolveArmByInferredType is an arm that dispatches on the value's inferred type.
//
// Upstream's `maskStopUtility` position block, read at utilities.ts:3285. The type list is the arm's
// own and the two arms of the switch accept different values, which is why the percentage case is a
// field rather than another `BareValueKind`: a bare kind is one predicate and this is two.
//
// A percentage is valid when the digits before the `%` are a positive integer, so `50%` resolves and
// `12.5%` does not. A number goes through the arm's bare handler, which for this family is the
// spacing multiplier, so `1.5` resolves and `3.7` does not.
func resolveArmByInferredType(candidate *ParsedCandidate, arm *FunctionalUtilityArm, theme *Theme) (string, bool) {
	inferred := InferDataType(candidate.Value.Value, arm.InferTypes)

	switch inferred {
	case DataTypePercentage:
		if !arm.PercentagePassesThrough {
			return "", false
		}
		if !isPositiveInteger(strings.TrimSuffix(candidate.Value.Value, "%")) {
			return "", false
		}
		return candidate.Value.Value, true

	case "":
		return "", false
	}

	// Upstream reads `--spacing` before applying the multiplier and returns nothing when the theme
	// declares none, so a theme without it refuses the class rather than resolving it.
	if len(arm.ThemeKeys) > 0 {
		if _, found := theme.Resolve("", false, arm.ThemeKeys, 0); !found {
			return "", false
		}
	}
	if arm.BareValue == BareValueNone {
		return "", false
	}
	handler := bareValueHandler(arm.BareValue, arm.BareValueSuffix)
	if handler == nil {
		return "", false
	}
	return handler(candidate.Value)
}

// resolveArmColor is upstream's `resolveThemeColor`, which is not `theme.resolve` with a flag.
//
// Three keywords are answered before the theme is consulted at all, and they are answered whatever
// the theme holds: a repository declaring `--color-inherit` does not change what `border-inherit`
// means. `current` resolves to `currentcolor` rather than to itself, which is the one place the
// returned string is not the written one.
//
// The modifier is the alpha and is deliberately not applied here. This pipeline answers whether a
// value resolves, and `asColor` composes the alpha onto a value that already resolved, so applying
// it would change the string without changing the answer.
func resolveArmColor(candidate *ParsedCandidate, arm *FunctionalUtilityArm, theme *Theme) (string, bool) {
	switch candidate.Value.Value {
	case "inherit":
		return "inherit", true
	case "transparent":
		return "transparent", true
	case "current":
		return "currentcolor", true
	}
	return theme.Resolve(candidate.Value.Value, true, arm.ThemeKeys, 0)
}

// isPositiveInteger is upstream's `isPositiveInteger` from `src/utils/infer-data-type.ts`.
//
// Upstream is `Number.isInteger(num) && num >= 0 && String(num) === String(value)`, and that last
// clause is the whole predicate: it rejects every spelling JavaScript would coerce to a number but
// not print back identically. A digits-only scan agrees with it on `007`, `1.0`, `1e3`, `+1`, `-1`,
// ` 1` and `0x10`, all measured, and disagrees in exactly one place.
//
// That place is precision. `9007199254740993` is past 2^53, so `Number(value)` rounds it to
// `9007199254740992`, `String(num)` prints the rounded form, and upstream returns false. A
// digits-only scan returns true. The round-trip below reproduces the rejection rather than
// approximating it, because a bare value that large is a value the engine refuses and this port
// would otherwise resolve.
func isPositiveInteger(input string) bool {
	if input == "" {
		return false
	}
	for index := 0; index < len(input); index++ {
		if input[index] < '0' || input[index] > '9' {
			return false
		}
	}
	// The float round-trip, matching `String(Number(value)) === String(value)`. Parsed as a float
	// rather than an integer on purpose: the question is what JavaScript's single numeric type does
	// to this string, and a Go int64 parse would accept values JavaScript cannot represent.
	parsed, err := strconv.ParseFloat(input, 64)
	if err != nil {
		return false
	}
	return strconv.FormatFloat(parsed, 'f', -1, 64) == input
}

// isStrictPositiveInteger is upstream's `isStrictPositiveInteger`: the same predicate, excluding zero.
func isStrictPositiveInteger(input string) bool {
	return isPositiveInteger(input) && input != "0"
}
