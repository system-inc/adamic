// The candidate parser: the function that turns a class string into the structured readings the
// rest of the engine compiles.
//
// Ported from `src/candidate.ts` at Tailwind 4.3.3, read at the pinned tag rather than from the
// minified bundle, so a disagreement is a finding rather than version skew.
//
// # The one property that matters more than the rest
//
// `parseCandidate` is a generator, and a class can have more than one reading. `border-b` is both
// the functional root `border-b` with no value and the root `border` with the named value `b`, in
// that order, and the engine compiles them in order and keeps the first that produces CSS. 178 of
// the 1,229 classes this repository actually writes read more than one way, and 2,760 of the 5,056
// classes in the fixture do.
//
// So the readings are a sequence, never a set. A port that produces the same readings in a
// different order is not slightly wrong: it silently resolves `border-b` to a different utility,
// with different declared properties, and nothing anywhere raises an error. That failure looks
// exactly like correct behaviour from every direction except a differential test, which is why the
// fixture records candidates as an ordered array and the test compares them positionally.
//
// # Why this cannot be a table
//
// Where the root of a class ends is not a property of the string. `border-b` splits after `border-b`
// and after `border` because both exist as functional roots; `ring-offset-2` splits after
// `ring-offset` and after `ring` for the same reason. Which roots exist is a question about this
// repository's `@utility` blocks as much as about the framework, so the split is a function of the
// design system and is recomputed per repository rather than baked into a generated table. Three
// hand-written splitters in `internal/rules/tailwind` currently answer it three different ways
// against three different generated tables, and a bug where two of them disagreed about whether
// `ring-offset` exists is what put this port on the roadmap. Replacing them is a separate task;
// this file only has to be right.
//
// # Bytes, not runes
//
// Upstream indexes UTF-16 code units and every character it compares against is ASCII. This port
// indexes bytes, which agrees for the same reason cssparser.go's port does: every byte of a
// multi-byte UTF-8 sequence has its high bit set and therefore equals none of the ASCII constants,
// so multi-byte text is copied through untouched.
package tailwind

import "strings"

// ParsedCandidateKind distinguishes the three shapes a class can read as.
//
// Distinct from CandidateKind in gate.go, which is the lossy bucketing view the rules consume.
// This is the full upstream union, and the difference is real: gate.go has no arbitrary-property
// kind because `[color:red]` cannot merge with anything, while the compiler very much needs to know
// the property and the value.
type ParsedCandidateKind string

const (
	// ParsedCandidateKindArbitrary is a class that registers a declaration on the fly:
	// `[color:red]`.
	ParsedCandidateKindArbitrary ParsedCandidateKind = "arbitrary"
	// ParsedCandidateKindStatic is a whole-name utility: `underline`.
	ParsedCandidateKindStatic ParsedCandidateKind = "static"
	// ParsedCandidateKindFunctional is a root with an optional value: `bg-red-500`.
	ParsedCandidateKindFunctional ParsedCandidateKind = "functional"
)

// ParsedValueKind separates the two spellings of a functional candidate's value.
type ParsedValueKind string

const (
	// ParsedValueKindArbitrary is a bracketed value: the `#0088cc` of `bg-[#0088cc]`.
	ParsedValueKindArbitrary ParsedValueKind = "arbitrary"
	// ParsedValueKindNamed is a bare value the theme is expected to name: the `red-500` of
	// `bg-red-500`.
	ParsedValueKindNamed ParsedValueKind = "named"
)

// ParsedValue is a functional candidate's value.
//
// One struct rather than two, matching the upstream tagged union, because callers switch on Kind
// and the alternative is an interface whose two implementations differ by two fields. DataType is
// meaningful only when Kind is arbitrary and Fraction only when it is named; each is documented
// where it is set.
type ParsedValue struct {
	Kind ParsedValueKind
	// Value is the decoded text. For an arbitrary value that is post-decodeArbitraryValue, so
	// `w-[calc(100%-1rem)]` holds `calc(100% - 1rem)` rather than what the author typed.
	Value string
	// DataType is the explicit typehint of an arbitrary value: the `color` of
	// `bg-[color:var(--my-color)]`. Empty when there was none.
	DataType string
	// Fraction is the whole `value/modifier` text when a named value was followed by something that
	// might be a denominator rather than a modifier: `1/2` for `w-1/2`.
	//
	// It exists because the slash is genuinely ambiguous and the parser refuses to guess. `w-1/2`
	// is a fraction; `bg-red-500/50` is a modifier; they are spelled identically. So both readings
	// are carried and the utility matcher picks. Set only when a modifier segment was present and
	// that modifier parsed as named, because an arbitrary modifier cannot be half of a fraction.
	Fraction string
}

// ParsedModifierKind separates a bracketed modifier from a bare one.
type ParsedModifierKind string

const (
	// ParsedModifierKindArbitrary is `/[0.5]` or `/(--a)`.
	ParsedModifierKindArbitrary ParsedModifierKind = "arbitrary"
	// ParsedModifierKindNamed is `/50` or `/none`.
	ParsedModifierKindNamed ParsedModifierKind = "named"
)

// ParsedModifier is the part of a class after the slash.
//
// A pointer to one of these is how absence is spelled, because a modifier is a three-state axis and
// not a two-state one: absent, a theme key such as `/none`, or a value such as `/50` or `/[0.5]`.
// The descriptor work measured that all three move the reading differently, so collapsing absent
// into an empty value would erase a distinction the table depends on.
type ParsedModifier struct {
	Kind ParsedModifierKind
	// Value is the decoded text, so `/[var(--a)]` holds `var(--a)` and `/(--a)` holds the same,
	// because the shorthand is expanded during parsing rather than later.
	Value string
}

// ParsedVariantKind distinguishes the four variant shapes.
type ParsedVariantKind string

const (
	// ParsedVariantKindArbitrary is a variant written as a selector: `[&_p]`.
	ParsedVariantKindArbitrary ParsedVariantKind = "arbitrary"
	// ParsedVariantKindStatic is a variant that takes no argument: `hover`.
	ParsedVariantKindStatic ParsedVariantKind = "static"
	// ParsedVariantKindFunctional is a variant that takes a value: `aria-disabled`, `@lg`.
	ParsedVariantKindFunctional ParsedVariantKind = "functional"
	// ParsedVariantKindCompound is a variant that takes another variant: `group-hover`.
	ParsedVariantKindCompound ParsedVariantKind = "compound"
)

// ParsedVariantValue is a functional variant's argument.
type ParsedVariantValue struct {
	Kind ParsedValueKind
	// Value is the decoded text. A `(--a)` shorthand arrives here already wrapped as `var(--a)`.
	Value string
}

// ParsedVariant is one variant, which may contain another.
//
// Recursive through Variant, because `not-group-hover` is a compound wrapping a compound wrapping a
// static and the nesting is what the compiler walks. Fields are meaningful per kind and each says
// which.
type ParsedVariant struct {
	Kind ParsedVariantKind
	// Root is the registered variant name, for every kind except arbitrary.
	Root string
	// Selector is the decoded selector of an arbitrary variant, already wrapped in `&:is(…)` when
	// upstream wraps it.
	Selector string
	// Relative records that an arbitrary variant's selector began with a combinator, so it is
	// relative to the element rather than a standalone selector.
	Relative bool
	// Value is a functional variant's argument, nil when it has none.
	Value *ParsedVariantValue
	// Modifier is the part after the slash, nil when absent. A compound variant's modifier may have
	// been forwarded into its sub-variant instead; see parseVariant.
	Modifier *ParsedModifier
	// Variant is the nested variant of a compound, nil for every other kind.
	Variant *ParsedVariant
}

// ParsedCandidate is one reading of a class.
type ParsedCandidate struct {
	Kind ParsedCandidateKind
	// Root is the utility root, for static and functional candidates.
	Root string
	// Property is the declared property of an arbitrary candidate: the `color` of `[color:red]`.
	Property string
	// PropertyValue is the declared value of an arbitrary candidate, already decoded.
	//
	// Named apart from Value because an arbitrary candidate has no ParsedValue: it carries a
	// property and a value directly rather than a root and a value.
	PropertyValue string
	// Value is a functional candidate's value, nil when it has none, as in `border-b`.
	Value *ParsedValue
	// Modifier is the part after the slash, nil when absent.
	Modifier *ParsedModifier
	// Variants are the variants applied, in the order the compiler consumes them, which is the
	// reverse of how they were written. `sm:hover:flex` yields `hover` then `sm`.
	Variants []ParsedVariant
	// Important records a trailing `!` or a leading `!`.
	Important bool
	// Raw is the whole original class string, including variants and importance.
	Raw string
}

// UtilityKind is which registration a utility root was declared under.
type UtilityKind string

const (
	// UtilityKindStatic is a root registered as a whole name, which cannot take a value.
	UtilityKindStatic UtilityKind = "static"
	// UtilityKindFunctional is a root registered as taking a value.
	UtilityKindFunctional UtilityKind = "functional"
)

// VariantCompounds is the bitmask describing what kind of rules a variant produces or accepts.
//
// Upstream's `Compounds` enum. Two variants compound only when the parent accepts every kind of
// rule the child generates, which is a bitwise-and rather than an equality: a child that produces
// only at-rules compounds with a parent that accepts at-rules and style rules.
type VariantCompounds int

const (
	// VariantCompoundsNever means the variant cannot participate in compounding at all.
	VariantCompoundsNever VariantCompounds = 0
	// VariantCompoundsAtRules is `@media`, `@supports` and `@container`.
	VariantCompoundsAtRules VariantCompounds = 1 << 0
	// VariantCompoundsStyleRules is an ordinary selector.
	VariantCompoundsStyleRules VariantCompounds = 1 << 1
)

// DesignSystem is the four questions the candidate parser asks about the repository.
//
// An interface rather than a struct because the answers come from the theme, which is a separate
// component with its own owner, and because the differential test answers them from the tables the
// engine itself reported. Keeping the surface to four methods is deliberate: every one of them is a
// lookup the parser genuinely cannot avoid, and anything wider would let this file start depending
// on theme internals it has no business reading.
type DesignSystem interface {
	// Prefix is the configured utility prefix, empty when there is none. When set, every class must
	// begin with it as its first variant segment and the segment is then dropped.
	Prefix() string
	// HasUtility reports whether root is registered as a utility of the given kind.
	HasUtility(root string, kind UtilityKind) bool
	// HasVariant reports whether root is a registered variant.
	HasVariant(root string) bool
	// VariantKind returns the registration kind of a variant root. Only called for roots HasVariant
	// accepted.
	VariantKind(root string) ParsedVariantKind
	// VariantCompoundsWith reports whether the compound variant parent accepts the rules child
	// produces. child is a fully parsed variant, because an arbitrary child's compounding is
	// computed from its selector rather than looked up.
	VariantCompoundsWith(parent string, child ParsedVariant) bool
}

// isValidNamedValue ports `IS_VALID_NAMED_VALUE`, which is `/^[a-zA-Z0-9_.%-]+$/`.
//
// Written as a loop rather than a compiled regexp because it is called several times per class on
// every class in a file and a regexp match would allocate. The character class is closed and small,
// so the loop is the same predicate rather than an approximation of it. The `+` matters: an empty
// string does not match, which is what rejects `bg-` after the root is stripped.
func isValidNamedValue(value string) bool {
	if value == "" {
		return false
	}
	for index := 0; index < len(value); index++ {
		character := value[index]
		switch {
		case character >= 'a' && character <= 'z',
			character >= 'A' && character <= 'Z',
			character >= '0' && character <= '9',
			character == '_', character == '.', character == '%', character == '-':
		default:
			return false
		}
	}
	return true
}

// isBlank reports whether a string is empty or entirely whitespace.
//
// Upstream writes `value.length === 0 || value.trim().length === 0`, and JavaScript's `trim` strips
// the same set `\s` matches, which segment.go already models as isJavaScriptSpace. Reusing it keeps
// the two agreeing on a non-breaking space, which Go's strings.TrimSpace would also strip but which
// a hand-written ASCII test would not.
func isBlank(value string) bool {
	return strings.TrimFunc(value, isJavaScriptSpace) == ""
}

// ParseCandidate returns every reading of a class, in the engine's order.
//
// The order is the contract. See the file comment: the caller keeps the first reading that
// compiles, so returning the right readings in the wrong order is a silent wrong answer.
//
// An empty result means the class is not a utility. That is not an error condition: most strings in
// a `className` attribute that are not Tailwind classes reach here and read as nothing.
func ParseCandidate(input string, designSystem DesignSystem) []ParsedCandidate {
	// Upstream's generator returns early in several places, some spelled `return` and one spelled
	// `return null` inside a generator, which JavaScript treats identically. Every one of them
	// means "stop, keep what was already yielded". Collecting into a slice and returning it makes
	// that explicit; a Go iterator would reproduce the shape at the cost of making every early exit
	// a `return false` from a closure, which reads worse and buys nothing here because the caller
	// wants the whole list anyway.
	candidates := []ParsedCandidate{}

	rawVariants := adamicSegment(input, ':')

	// A prefix is a variant every utility must carry. When configured, a class that does not begin
	// with it reads as nothing at all rather than as an unprefixed utility.
	if prefix := designSystem.Prefix(); prefix != "" {
		if len(rawVariants) == 1 {
			return candidates
		}
		if rawVariants[0] != prefix {
			return candidates
		}
		rawVariants = rawVariants[1:]
	}

	// Safe unconditionally: segment always returns at least one part, even for an empty input.
	base := rawVariants[len(rawVariants)-1]
	rawVariants = rawVariants[:len(rawVariants)-1]

	// Variants are parsed right to left, so the list runs innermost first. `sm:hover:flex` yields
	// `hover` then `sm`, which is the order the compiler nests them in.
	parsedVariants := []ParsedVariant{}
	for index := len(rawVariants) - 1; index >= 0; index-- {
		parsedVariant := ParseVariant(rawVariants[index], designSystem)
		if parsedVariant == nil {
			return candidates
		}
		parsedVariants = append(parsedVariants, *parsedVariant)
	}

	important := false

	// Both spellings, and only one of them is checked when the other matched: `!flex!` is important
	// with the base `flex!`, not with `flex`, because the trailing test wins and consumes only one
	// character. That is upstream's `else if` and not a simplification.
	if len(base) > 0 && base[len(base)-1] == '!' {
		important = true
		base = base[:len(base)-1]
	} else if len(base) > 0 && base[0] == '!' {
		important = true
		base = base[1:]
	}

	// An exact static match is yielded first, before any functional splitting, so `flex` reads as
	// the static utility rather than as a root with no value. The `[` test keeps a class such as
	// `[color:red]` from matching a static root that happens to share its text.
	if designSystem.HasUtility(base, UtilityKindStatic) && !strings.Contains(base, "[") {
		candidates = append(candidates, ParsedCandidate{
			Kind:      ParsedCandidateKindStatic,
			Root:      base,
			Variants:  parsedVariants,
			Important: important,
			Raw:       input,
		})
	}

	// Split off the modifier. segment rather than a plain split, so the `/` inside `bg-[a/b]` does
	// not count.
	modifierParts := adamicSegment(base, '/')
	baseWithoutModifier := modifierParts[0]
	var modifierSegment *string
	if len(modifierParts) > 1 {
		modifierSegment = &modifierParts[1]
	}
	// More than one modifier is invalid: `bg-red-500/50/50`.
	//
	// Upstream destructures three elements and tests the third for truthiness, so an empty third
	// part does not trigger it. `bg-red-500/50/` has parts `["bg-red-500","50",""]` and the empty
	// string is falsy, so it is *not* rejected here. Testing `len(parts) > 2` instead would reject
	// it, which is a different answer on a real class.
	if len(modifierParts) > 2 && modifierParts[2] != "" {
		return candidates
	}

	var parsedModifier *ParsedModifier
	if modifierSegment != nil {
		parsedModifier = adamicModifier(*modifierSegment)
		// A modifier that was written but did not parse invalidates the whole class, rather than
		// being dropped: `bg-red-500/[]` is not `bg-red-500`.
		if parsedModifier == nil {
			return candidates
		}
	}

	// Arbitrary properties: `[color:red]`.
	if len(baseWithoutModifier) > 0 && baseWithoutModifier[0] == '[' {
		if baseWithoutModifier[len(baseWithoutModifier)-1] != ']' {
			return candidates
		}

		// The property may only begin with a lowercase letter or a dash, the latter for vendor
		// prefixes such as `-webkit-`. This is what stops `[0:red]` and `[Color:red]` from
		// registering a declaration.
		if len(baseWithoutModifier) < 2 {
			return candidates
		}
		leading := baseWithoutModifier[1]
		if leading != '-' && !(leading >= 'a' && leading <= 'z') {
			return candidates
		}

		inner := baseWithoutModifier[1 : len(baseWithoutModifier)-1]

		// The colon must exist and must be neither first nor last, so `[color]`, `[:red]` and
		// `[color:]` are all rejected.
		colonIndex := strings.IndexByte(inner, ':')
		if colonIndex == -1 || colonIndex == 0 || colonIndex == len(inner)-1 {
			return candidates
		}

		property := inner[:colonIndex]
		value := adamicDecode(inner[colonIndex+1:])

		if !adamicValidArbitrary(value) {
			return candidates
		}

		candidates = append(candidates, ParsedCandidate{
			Kind:          ParsedCandidateKindArbitrary,
			Property:      property,
			PropertyValue: value,
			Modifier:      parsedModifier,
			Variants:      parsedVariants,
			Important:     important,
			Raw:           input,
		})
		return candidates
	}

	// roots is the sequence of `{root, value}` splits to try, and its order is the order the
	// candidates come out in.
	type rootSplit struct {
		root string
		// value is nil for an exact match, where the root is the whole input and there is no value.
		value *string
	}
	var roots []rootSplit

	switch {
	// A base ending in `]` is a bracketed arbitrary value, so the root is everything before the
	// first `-[` and there is exactly one split to try.
	case len(baseWithoutModifier) > 0 && baseWithoutModifier[len(baseWithoutModifier)-1] == ']':
		bracketIndex := strings.Index(baseWithoutModifier, "-[")
		if bracketIndex == -1 {
			return candidates
		}
		root := baseWithoutModifier[:bracketIndex]
		if !designSystem.HasUtility(root, UtilityKindFunctional) {
			return candidates
		}
		// From the dash, not past it, so the value still carries its leading `[`.
		value := baseWithoutModifier[bracketIndex+1:]
		roots = []rootSplit{{root: root, value: &value}}

	// A base ending in `)` is the CSS-variable shorthand, `bg-(--my-var)`.
	case len(baseWithoutModifier) > 0 && baseWithoutModifier[len(baseWithoutModifier)-1] == ')':
		parenIndex := strings.Index(baseWithoutModifier, "-(")
		if parenIndex == -1 {
			return candidates
		}
		root := baseWithoutModifier[:parenIndex]
		if !designSystem.HasUtility(root, UtilityKindFunctional) {
			return candidates
		}
		value := baseWithoutModifier[parenIndex+2 : len(baseWithoutModifier)-1]

		// An optional typehint, as in `bg-(color:--my-var)`. Exactly two parts, so `a:b:--c` is not
		// a typehint plus a value and falls through to the `--` test, which rejects it.
		parts := adamicSegment(value, ':')
		dataType := ""
		if len(parts) == 2 {
			dataType = parts[0]
			value = parts[1]
		}

		// The shorthand always names a CSS variable, so it must begin with `--`.
		if len(value) < 2 || value[0] != '-' || value[1] != '-' {
			return candidates
		}
		if !adamicValidArbitrary(value) {
			return candidates
		}

		// Rewritten into the bracketed spelling, so the branch below handles both forms. The
		// rewrite is upstream's and is why `bg-(--a)` and `bg-[var(--a)]` produce identical
		// candidates rather than merely equivalent ones.
		var rewritten string
		if dataType == "" {
			rewritten = "[var(" + value + ")]"
		} else {
			rewritten = "[" + dataType + ":var(" + value + ")]"
		}
		roots = []rootSplit{{root: root, value: &rewritten}}

	// Everything else: try every split, longest root first.
	default:
		for _, found := range adamicFindRoots(false, baseWithoutModifier, func(root string) bool {
			return designSystem.HasUtility(root, UtilityKindFunctional)
		}) {
			split := rootSplit{root: found.root}
			if found.hasValue {
				value := found.value
				split.value = &value
			}
			roots = append(roots, split)
		}
	}

	for _, split := range roots {
		candidate := ParsedCandidate{
			Kind:      ParsedCandidateKindFunctional,
			Root:      split.root,
			Modifier:  parsedModifier,
			Variants:  parsedVariants,
			Important: important,
			Raw:       input,
		}

		if split.value == nil {
			candidates = append(candidates, candidate)
			continue
		}
		value := *split.value

		startArbitrary := strings.IndexByte(value, '[')
		if startArbitrary != -1 {
			// A value containing `[` must close it at the very end. This is a `return`, not a
			// `continue`: the whole class is abandoned rather than this one split, so a later split
			// that would have read the same text as a named value never happens.
			if value[len(value)-1] != ']' {
				return candidates
			}

			arbitraryValue := adamicDecode(value[startArbitrary+1 : len(value)-1])

			// Here it is a `continue`, so a different split may still succeed. The asymmetry with
			// the test just above is upstream's and is observable.
			if !adamicValidArbitrary(arbitraryValue) {
				continue
			}

			// An explicit typehint runs from the start to the first `:`, and only while every
			// character so far is a lowercase letter or a dash. Anything else ends the scan with no
			// typehint, which is what keeps `bg-[var(--a)]` from reading `var(--a` as one.
			typehint := ""
			hasTypehint := false
			for index := 0; index < len(arbitraryValue); index++ {
				code := arbitraryValue[index]
				if code == ':' {
					typehint = arbitraryValue[:index]
					hasTypehint = true
					arbitraryValue = arbitraryValue[index+1:]
					break
				}
				if code == '-' || (code >= 'a' && code <= 'z') {
					continue
				}
				break
			}

			if adamicBlank(arbitraryValue) {
				continue
			}
			// A colon at position zero yields an empty typehint, as in `bg-[:red]`, which is
			// rejected rather than treated as absent.
			if hasTypehint && typehint == "" {
				continue
			}

			candidate.Value = &ParsedValue{
				Kind:     ParsedValueKindArbitrary,
				DataType: typehint,
				Value:    arbitraryValue,
			}
		} else {
			// The slash is ambiguous and the parser refuses to resolve it. `w-1/2` is a fraction and
			// `bg-red-500/50` is a modifier, spelled identically, so both readings are carried and
			// the utility matcher decides. An arbitrary modifier cannot be a denominator, so it
			// suppresses the fraction rather than producing a nonsensical one.
			fraction := ""
			if modifierSegment != nil && !(parsedModifier != nil && parsedModifier.Kind == ParsedModifierKindArbitrary) {
				fraction = value + "/" + *modifierSegment
			}

			if !adamicValidNamed(value) {
				continue
			}

			candidate.Value = &ParsedValue{
				Kind:     ParsedValueKindNamed,
				Value:    value,
				Fraction: fraction,
			}
		}

		candidates = append(candidates, candidate)
	}

	return candidates
}

// parseModifier reads the part of a class after the slash.
//
// Returns nil for a modifier that was written but is invalid, which the callers treat as
// invalidating the whole class rather than as an absent modifier.
func parseModifier(modifier string) *ParsedModifier {
	if len(modifier) >= 2 && modifier[0] == '[' && modifier[len(modifier)-1] == ']' {
		arbitraryValue := decodeArbitraryValue(modifier[1 : len(modifier)-1])
		if !isValidArbitrary(arbitraryValue) {
			return nil
		}
		if isBlank(arbitraryValue) {
			return nil
		}
		return &ParsedModifier{Kind: ParsedModifierKindArbitrary, Value: arbitraryValue}
	}

	if len(modifier) >= 2 && modifier[0] == '(' && modifier[len(modifier)-1] == ')' {
		inner := modifier[1 : len(modifier)-1]

		// The shorthand always names a CSS variable.
		if len(inner) < 2 || inner[0] != '-' || inner[1] != '-' {
			return nil
		}
		if !isValidArbitrary(inner) {
			return nil
		}

		// Wrapped before decoding, not after, which matters because the wrapping introduces the
		// `var(` that makes decodeArbitraryValue take its AST path and preserve the underscores in
		// the variable name.
		arbitraryValue := decodeArbitraryValue("var(" + inner + ")")

		// Deliberately no blank test here. Upstream has one in the `[…]` branch and not in this
		// one, and it cannot fire anyway: the `--` requirement above already guarantees non-empty.
		return &ParsedModifier{Kind: ParsedModifierKindArbitrary, Value: arbitraryValue}
	}

	if !isValidNamedValue(modifier) {
		return nil
	}
	return &ParsedModifier{Kind: ParsedModifierKindNamed, Value: modifier}
}

// ParseVariant reads one variant segment, which may recurse for a compound.
//
// Returns nil when the segment is not a variant, which invalidates the whole class: a class with an
// unreadable variant is not a utility with fewer variants.
func ParseVariant(variant string, designSystem DesignSystem) *ParsedVariant {
	// Arbitrary variants: `[&_p]`.
	if len(variant) >= 2 && variant[0] == '[' && variant[len(variant)-1] == ']' {
		// A deprecated shape upstream still rejects: an at-rule that also carries a selector.
		if variant[1] == '@' && strings.Contains(variant, "&") {
			return nil
		}

		selector := adamicDecode(variant[1 : len(variant)-1])
		if !adamicValidArbitrary(selector) {
			return nil
		}
		if adamicBlank(selector) {
			return nil
		}

		relative := selector[0] == '>' || selector[0] == '+' || selector[0] == '~'

		// A selector with no `&` is wrapped so it always has one, unless it is relative or is an
		// at-rule. `[p]:flex` becomes `&:is(p)`.
		if !relative && selector[0] != '@' && !strings.Contains(selector, "&") {
			selector = "&:is(" + selector + ")"
		}

		return &ParsedVariant{
			Kind:     ParsedVariantKindArbitrary,
			Selector: selector,
			Relative: relative,
		}
	}

	// Static, functional and compound variants.
	modifierParts := adamicSegment(variant, '/')
	variantWithoutModifier := modifierParts[0]
	var modifier *string
	if len(modifierParts) > 1 {
		modifier = &modifierParts[1]
	}
	// The same truthiness test as the candidate side: an empty third part does not reject.
	if len(modifierParts) > 2 && modifierParts[2] != "" {
		return nil
	}

	for _, found := range adamicFindRoots(true, variantWithoutModifier, designSystem.HasVariant) {
		root := found.root
		var value *string
		if found.hasValue {
			held := found.value
			value = &held
		}

		switch designSystem.VariantKind(root) {
		case ParsedVariantKindStatic:
			// A static variant takes neither a value nor a modifier, and either one makes the whole
			// class invalid rather than sending the loop to the next root.
			if value != nil {
				return nil
			}
			if modifier != nil {
				return nil
			}
			return &ParsedVariant{Kind: ParsedVariantKindStatic, Root: root}

		case ParsedVariantKindFunctional:
			var parsedModifier *ParsedModifier
			if modifier != nil {
				parsedModifier = adamicModifier(*modifier)
				if parsedModifier == nil {
					return nil
				}
			}

			if value == nil {
				return &ParsedVariant{Kind: ParsedVariantKindFunctional, Root: root, Modifier: parsedModifier}
			}
			held := *value

			if len(held) > 0 && held[len(held)-1] == ']' {
				// A value that ends with `]` but does not start with one is not this root's; try
				// the next split rather than abandoning the class.
				if held[0] != '[' {
					continue
				}

				arbitraryValue := adamicDecode(held[1 : len(held)-1])
				if !adamicValidArbitrary(arbitraryValue) {
					return nil
				}
				if adamicBlank(arbitraryValue) {
					return nil
				}

				return &ParsedVariant{
					Kind:     ParsedVariantKindFunctional,
					Root:     root,
					Modifier: parsedModifier,
					Value:    &ParsedVariantValue{Kind: ParsedValueKindArbitrary, Value: arbitraryValue},
				}
			}

			if len(held) > 0 && held[len(held)-1] == ')' {
				if held[0] != '(' {
					continue
				}

				arbitraryValue := adamicDecode(held[1 : len(held)-1])
				if !adamicValidArbitrary(arbitraryValue) {
					return nil
				}
				if adamicBlank(arbitraryValue) {
					return nil
				}
				if len(arbitraryValue) < 2 || arbitraryValue[0] != '-' || arbitraryValue[1] != '-' {
					return nil
				}

				return &ParsedVariant{
					Kind:     ParsedVariantKindFunctional,
					Root:     root,
					Modifier: parsedModifier,
					Value:    &ParsedVariantValue{Kind: ParsedValueKindArbitrary, Value: "var(" + arbitraryValue + ")"},
				}
			}

			if !adamicValidNamed(held) {
				continue
			}

			return &ParsedVariant{
				Kind:     ParsedVariantKindFunctional,
				Root:     root,
				Modifier: parsedModifier,
				Value:    &ParsedVariantValue{Kind: ParsedValueKindNamed, Value: held},
			}

		case ParsedVariantKindCompound:
			if value == nil {
				return nil
			}
			held := *value
			forwardedModifier := modifier

			// Three compound roots forward their modifier into the sub-variant, so
			// `not-group-hover/name` reads as `not(group-hover/name)` rather than as a `not` with a
			// modifier. The forwarding is by name rather than by a property of the registration,
			// which is upstream's shape.
			if forwardedModifier != nil && (root == "not" || root == "has" || root == "in") {
				held = held + "/" + *forwardedModifier
				forwardedModifier = nil
			}

			subVariant := ParseVariant(held, designSystem)
			if subVariant == nil {
				return nil
			}

			// The two must be compatible: a parent that only wraps style rules cannot wrap a child
			// that emits a pseudo-element.
			if !designSystem.VariantCompoundsWith(root, *subVariant) {
				return nil
			}

			var parsedModifier *ParsedModifier
			if forwardedModifier != nil {
				parsedModifier = adamicModifier(*forwardedModifier)
				if parsedModifier == nil {
					return nil
				}
			}

			return &ParsedVariant{
				Kind:     ParsedVariantKindCompound,
				Root:     root,
				Modifier: parsedModifier,
				Variant:  subVariant,
			}
		}
	}

	return nil
}

// rootCandidate is one `{root, value}` split findRoots offers.
//
// hasValue rather than a nil string pointer because the distinction is between "the root is the
// whole input and there is no value" and "the value is the empty string", and the second never
// escapes findRoots: it stops the search instead.
type rootCandidate struct {
	root     string
	value    string
	hasValue bool
}

// findRoots yields every way the input splits into a registered root and a value, longest first.
//
// This is where the ambiguity comes from and the order is the whole point. For `border-b`, with both
// `border-b` and `border` registered, it yields `{border-b, -}` then `{border, b}`, and the caller
// keeps the first that compiles.
//
// Upstream is a generator and this returns a slice. That is not equivalent in general — a generator
// stops when the consumer stops — but it is equivalent here, because both consumers drain it: the
// candidate side loops over every root, and the variant side returns from inside the loop, which
// this port reproduces by returning from inside its own loop over the collected slice. The cost is
// that a variant returning on the first root still computed the rest, which is a slice of at most a
// handful of small strings.
func findRoots(input string, exists func(string) bool) []rootCandidate {
	roots := []rootCandidate{}

	// An exact match comes first, which is what makes `border-b` read as a valueless root before it
	// reads as `border` with a value.
	if exists(input) {
		roots = append(roots, rootCandidate{root: input})
	}

	// Then every prefix ending at a dash, longest first.
	index := strings.LastIndexByte(input, '-')
	for index > 0 {
		maybeRoot := input[:index]

		if exists(maybeRoot) {
			value := input[index+1:]

			// An empty value means the class ended in a dash, as in `bg-`. The whole search stops
			// rather than skipping this split, so no shorter root is tried either.
			if value == "" {
				break
			}

			// `@` followed by `-` is not a valid variant or utility, so `@-2xl` must not read as the
			// `@` root with the value `-2xl`. Stopping here rather than skipping is upstream's, and
			// it also prevents the `@` fallback below from firing.
			if maybeRoot == "@" && exists("@") && input[index] == '-' {
				break
			}

			roots = append(roots, rootCandidate{root: maybeRoot, value: value, hasValue: true})
		}

		index = strings.LastIndexByte(input[:index], '-')
	}

	// The `@` root is tried last, so `@max` of `@max-foo` matches before the bare `@` does.
	if len(input) > 0 && input[0] == '@' && exists("@") {
		roots = append(roots, rootCandidate{root: "@", value: input[1:], hasValue: true})
	}

	return roots
}

// compoundsForSelectors computes what kinds of rules a set of selectors produces.
//
// Ported from `compoundsForSelectors` in `src/variants.ts`. It is here rather than with the rest of
// the variant machinery because it is the one design-system answer the candidate parser cannot get
// from a table: a registered variant's compounding is looked up, but an arbitrary variant is
// written inline and its compounding has to be derived from the selector text at parse time.
//
// Three ways to reach Never, and they are the whole function. A non-conditional at-rule cannot be
// compounded because there is nowhere to nest the parent's selector. A pseudo-element cannot,
// because `group-hover:before:` would have to place a selector after `::before`, which does not
// mean anything. Everything else is a style rule.
//
// The at-rule test is `strings.HasPrefix`, matching upstream, which means `@mediaquery` also counts
// as `@media`. That is upstream's behaviour and is not tightened here; a selector spelled that way
// is not valid CSS anyway, so the only effect of "fixing" it would be to disagree with the engine.
func compoundsForSelectors(selectors []string) VariantCompounds {
	compounds := VariantCompoundsNever

	for _, selector := range selectors {
		if len(selector) > 0 && selector[0] == '@' {
			if !strings.HasPrefix(selector, "@media") &&
				!strings.HasPrefix(selector, "@supports") &&
				!strings.HasPrefix(selector, "@container") {
				return VariantCompoundsNever
			}
			compounds |= VariantCompoundsAtRules
			continue
		}

		if strings.Contains(selector, "::") {
			return VariantCompoundsNever
		}

		compounds |= VariantCompoundsStyleRules
	}

	return compounds
}

// GateCandidates converts full readings into the lossy shape gate.go buckets on.
//
// Two candidate models exist in this package on purpose. ParsedCandidate is the whole upstream
// structure, because the compiler needs the modifier, the typehint and the nested variant tree.
// gate.go's Candidate is a bucketing view: it asks only whether two classes could possibly merge,
// and it deliberately drops what cannot affect that answer.
//
// The conversion is where the two meet, and it is the seam `Parser.ParseCandidates` was declared
// against in gate.go before a parser existed to fill it. Order is preserved exactly, because
// gate.go's computeGroupKey takes candidates[0] and documents that choice as measured rather than
// assumed.
//
// An arbitrary candidate maps to CandidateKindUnparsed rather than growing a fourth kind. That is
// not a loss: gate.go gives an unparsed class a bucket key carrying its own name, so `[color:red]`
// never merges with anything, which is the correct answer for a class that declares a property
// directly.
func GateCandidates(parsed []ParsedCandidate) []Candidate {
	candidates := make([]Candidate, 0, len(parsed))

	for _, reading := range parsed {
		candidate := Candidate{
			Root:      reading.Root,
			Important: reading.Important,
		}

		switch reading.Kind {
		case ParsedCandidateKindStatic:
			candidate.Kind = CandidateKindStatic
		case ParsedCandidateKindFunctional:
			candidate.Kind = CandidateKindFunctional
		default:
			candidate.Kind = CandidateKindUnparsed
		}

		switch {
		case reading.Value == nil:
			candidate.Value = Value{Kind: ValueKindNone}
		case reading.Value.Kind == ParsedValueKindArbitrary:
			candidate.Value = Value{Kind: ValueKindArbitrary, Value: reading.Value.Value}
		default:
			candidate.Value = Value{
				Kind:     ValueKindNamed,
				Value:    reading.Value.Value,
				Fraction: reading.Value.Fraction,
			}
		}

		// The modifier is folded into the value key rather than dropped. Two classes differing only
		// by their modifier are not interchangeable — `bg-red-500/50` and `bg-red-500` produce
		// different colours — so losing it here would let them share a bucket.
		if reading.Modifier != nil {
			candidate.Value.Value += "/" + reading.Modifier.Value
		}

		candidate.Variants = make([]Variant, 0, len(reading.Variants))
		for _, variant := range reading.Variants {
			candidate.Variants = append(candidate.Variants, gateVariant(variant))
		}

		candidates = append(candidates, candidate)
	}

	return candidates
}

// gateVariant flattens one variant into the three fields gate.go compares.
//
// A compound variant's nested variant is rendered into Value rather than dropped, because
// `group-hover:` and `group-focus:` are different selectors and a flattening that lost the inner
// variant would bucket them together.
func gateVariant(variant ParsedVariant) Variant {
	flattened := Variant{
		Kind: string(variant.Kind),
		Root: variant.Root,
	}

	switch variant.Kind {
	case ParsedVariantKindArbitrary:
		flattened.Value = variant.Selector
	case ParsedVariantKindFunctional:
		if variant.Value != nil {
			flattened.Value = variant.Value.Value
		}
	case ParsedVariantKindCompound:
		if variant.Variant != nil {
			nested := gateVariant(*variant.Variant)
			flattened.Value = nested.Kind + ":" + nested.Root + ":" + nested.Value
		}
	}

	if variant.Modifier != nil {
		flattened.Value += "/" + variant.Modifier.Value
	}

	return flattened
}
