// The value-resolution descriptions of the framework roots registered as bare closures.
//
// frameworkgaphandlers.go holds what these 57 roots emit. This holds how they decide a value
// resolves at all, which is a different question and the one `no-unknown-classes` asks: `text-lg` is
// a class and `text-ss` is a typo, and nothing in a declaration list separates them.
//
// # Why these are not in the other two tables
//
// `FrameworkFunctionalUtilities` and `FrameworkMultiDeclarationUtilities` describe roots upstream
// registers through `functionalUtility`, which takes a declarative description: one namespace list,
// one bare-value handler, one default. 185 roots fit that and are fully answered by it.
//
// These 57 are registered as `utilities.functional(root, candidate => {...})`, a bare closure, and a
// closure can do things a description could not express until `Arms` existed: consult several
// namespace lists in sequence, infer a type instead of reading the theme, and accept a modifier on
// one path while refusing it on another.
//
// # Read at the tag, verified against the engine
//
// Every entry was read from `utilities.ts` at the v4.3.3 tag rather than inferred from the descriptor
// rows, which cannot supply it: a row records only the namespaces whose reading differs, so
// `outline`'s row carries `--color` and not `--outline-width`, and a description built from rows
// reports `outline-2` as unknown.
//
// Each family was then scored against `candidatesToCss` on a live design system, which is the engine
// answering whether a class compiles. The measurements are in gaproots_test.go, per family, and the
// pairs that carry them are the ones a careful reader gets wrong: `border-2/50` against
// `border-red-500/50`, `stroke-2/50` against both, `mask-x-from-1.5` against `mask-x-from-3.7`.
package tailwind

import "strings"

// colorArm is the `resolveThemeColor(candidate, theme, keys)` block these closures share.
func colorArm(keys ...string) FunctionalUtilityArm {
	return FunctionalUtilityArm{ThemeKeys: keys, IsColor: true}
}

// themeArm is a block that reads one namespace and stops there.
func themeArm(key string) FunctionalUtilityArm {
	return FunctionalUtilityArm{ThemeKeys: []string{key}}
}

// widthArm is a block that reads one namespace and falls back to a bare positive integer.
//
// The suffix and the modifier guard are per-root rather than per-shape, which is the detail this
// family is easiest to get wrong on. `border`, `outline` and `ring` append `px` and refuse a
// modifier; `stroke` does neither, so `stroke-2` resolves to `2` and `stroke-2/50` compiles where
// `border-2/50` does not.
func widthArm(key string, suffix string, refusesModifier bool) FunctionalUtilityArm {
	return FunctionalUtilityArm{
		ThemeKeys:       []string{key},
		BareValue:       BareValuePositiveInteger,
		BareValueSuffix: suffix,
		RefusesModifier: refusesModifier,
	}
}

// borderSideDescription is `borderSideUtility`, utilities.ts:2432, shared by eleven roots.
//
// One description rather than eleven, because upstream registers all eleven through one function and
// the sides differ only in which declarations they emit, which is frameworkgaphandlers.go's half.
func borderSideDescription() *FunctionalUtilityDescription {
	return &FunctionalUtilityDescription{
		DefaultValue:        "1px",
		DefaultValuePresent: true,
		Arms: []FunctionalUtilityArm{
			colorArm("--border-color", "--color"),
			widthArm("--border-width", "px", true),
		},
	}
}

// maskStopDescription is `maskStopUtility`, utilities.ts:3242, shared by eighteen roots.
//
// The second arm infers rather than reading the theme, and its `--spacing` entry is a precondition
// rather than a lookup: upstream refuses the class when the theme declares no `--spacing` at all.
func maskStopDescription() *FunctionalUtilityDescription {
	return &FunctionalUtilityDescription{
		Arms: []FunctionalUtilityArm{
			colorArm("--background-color", "--color"),
			{
				ThemeKeys:               []string{"--spacing"},
				InferTypes:              []DataType{DataTypeNumber, DataTypePercentage},
				PercentagePassesThrough: true,
				BareValue:               BareValueSpacingMultiplier,
				RefusesModifier:         true,
			},
		},
	}
}

// GapRootDescriptions is the resolution description of each root registered as a bare closure.
//
// Keyed on the undashed root, matching the emitters. A root absent here is one no description was
// read for, and a caller must decline rather than guess: answering "this value does not resolve" for
// a root nobody described would report working classes.
var GapRootDescriptions = map[string]*FunctionalUtilityDescription{
	// The eleven `borderSideUtility` registrations, utilities.ts:2505 through 2585.
	"border":    borderSideDescription(),
	"border-x":  borderSideDescription(),
	"border-y":  borderSideDescription(),
	"border-s":  borderSideDescription(),
	"border-e":  borderSideDescription(),
	"border-bs": borderSideDescription(),
	"border-be": borderSideDescription(),
	"border-t":  borderSideDescription(),
	"border-r":  borderSideDescription(),
	"border-b":  borderSideDescription(),
	"border-l":  borderSideDescription(),

	// The twelve `maskEdgeUtility` registrations and the six `maskStopUtility` ones, which share one
	// resolution body through `maskStopUtility`.
	"mask-x-from":      maskStopDescription(),
	"mask-x-to":        maskStopDescription(),
	"mask-y-from":      maskStopDescription(),
	"mask-y-to":        maskStopDescription(),
	"mask-t-from":      maskStopDescription(),
	"mask-t-to":        maskStopDescription(),
	"mask-r-from":      maskStopDescription(),
	"mask-r-to":        maskStopDescription(),
	"mask-b-from":      maskStopDescription(),
	"mask-b-to":        maskStopDescription(),
	"mask-l-from":      maskStopDescription(),
	"mask-l-to":        maskStopDescription(),
	"mask-linear-from": maskStopDescription(),
	"mask-linear-to":   maskStopDescription(),
	"mask-radial-from": maskStopDescription(),
	"mask-radial-to":   maskStopDescription(),
	"mask-conic-from":  maskStopDescription(),
	"mask-conic-to":    maskStopDescription(),

	// The colour-then-width closures. Each was read separately rather than assumed to match
	// `borderSideUtility`, and two of them do not: `decoration` puts its width arm first, and
	// `stroke` neither appends `px` nor refuses a modifier.
	"outline": {
		DefaultValue: "1px", DefaultValuePresent: true,
		Arms: []FunctionalUtilityArm{
			colorArm("--outline-color", "--color"),
			widthArm("--outline-width", "px", true),
		},
	},
	"ring": {
		DefaultValue: "1px", DefaultValuePresent: true,
		Arms: []FunctionalUtilityArm{
			colorArm("--ring-color", "--color"),
			widthArm("--ring-width", "px", true),
		},
	},
	"inset-ring": {
		DefaultValue: "1px", DefaultValuePresent: true,
		Arms: []FunctionalUtilityArm{
			colorArm("--inset-ring-color", "--color"),
			widthArm("--inset-ring-width", "px", true),
		},
	},
	"stroke": {
		Arms: []FunctionalUtilityArm{
			colorArm("--stroke", "--color"),
			widthArm("--stroke-width", "", false),
		},
	},
	// `decoration` puts its width arm first upstream and the order is reproduced rather than
	// normalised, though nothing here would catch it if it were wrong: measured by swapping the two
	// arms, every case still agrees, because no value is both a `--text-decoration-thickness` key
	// and a colour. The order is kept because it is what the source says and a future value that is
	// both would resolve on the wrong arm silently.
	"decoration": {
		Arms: []FunctionalUtilityArm{
			widthArm("--text-decoration-thickness", "px", true),
			colorArm("--text-decoration-color", "--color"),
		},
	},
	"fill": {
		Arms: []FunctionalUtilityArm{colorArm("--fill", "--color")},
	},

	// The colour-then-scale closures. A named value is a colour first and a size second, which is
	// what separates `text-red-500` from `text-lg` and `shadow-red-500` from `shadow-lg`.
	"text": {
		// A modifier on `text` is the line height, resolved through `--leading` or as a spacing
		// multiplier, so `text-[10px]/6` and `text-[10px]/relaxed` compile.
		AcceptsModifierOnArbitrary: true,
		Arms:                       []FunctionalUtilityArm{colorArm("--text-color", "--color"), themeArm("--text")},
	},
	"bg": {
		Arms: []FunctionalUtilityArm{colorArm("--background-color", "--color"), themeArm("--background-image")},
	},
	"font": {
		Arms: []FunctionalUtilityArm{themeArm("--font"), themeArm("--font-weight")},
	},

	// The shadow family, whose size arm reads its own namespace before the colour arm runs.
	// `none` and `inherit` are answered by the closure directly, which is what `StaticValueNames`
	// carries here: upstream spells them as `case` labels rather than as theme keys.
	"shadow": {
		DefaultValue: "var(--shadow)", DefaultValuePresent: true,
		Arms:             []FunctionalUtilityArm{themeArm("--shadow"), colorArm("--box-shadow-color", "--color")},
		StaticValueNames: map[string]bool{"none": true, "inherit": true},
	},
	"inset-shadow": {
		Arms:             []FunctionalUtilityArm{themeArm("--inset-shadow"), colorArm("--inset-shadow-color", "--color")},
		StaticValueNames: map[string]bool{"none": true, "inherit": true},
	},
	"text-shadow": {
		Arms:             []FunctionalUtilityArm{themeArm("--text-shadow"), colorArm("--text-shadow-color", "--color")},
		StaticValueNames: map[string]bool{"none": true, "inherit": true},
	},
	"drop-shadow": {
		DefaultValue: "var(--drop-shadow)", DefaultValuePresent: true,
		Arms:             []FunctionalUtilityArm{themeArm("--drop-shadow"), colorArm("--drop-shadow-color", "--color")},
		StaticValueNames: map[string]bool{"none": true},
	},

	// `duration` resolves through one namespace and appends `ms` to a bare integer, and answers
	// `initial` directly.
	"duration": {
		Arms:             []FunctionalUtilityArm{{ThemeKeys: []string{"--transition-duration"}, BareValue: BareValuePositiveInteger, BareValueSuffix: "ms"}},
		StaticValueNames: map[string]bool{"initial": true},
	},

	// The gradient roots, which carry a default and resolve an angle as a bare integer.
	"bg-conic": {
		DefaultValue: "in oklab", DefaultValuePresent: true,
		Arms: []FunctionalUtilityArm{{ThemeKeys: []string{"--conic"}, BareValue: BareValuePositiveInteger}},
	},
	"bg-radial": {DefaultValue: "in oklab", DefaultValuePresent: true},

	// The arbitrary-only closures. Each accepts a bracketed value or one named keyword and consults
	// no namespace at all, so a named value that is not the keyword resolves nowhere:
	// `filter-nonsense` and `mask-nonsense` compile to nothing while `filter-none` compiles.
	//
	// `mask` is the one that also refuses the bare root, because its body opens with
	// `if (!candidate.value) return` where `filter` and `transform` fall through to a default.
	"filter": {
		DefaultValue: "none", DefaultValuePresent: true,
		StaticValueNames: map[string]bool{"none": true},
	},
	"transform": {
		DefaultValue: "none", DefaultValuePresent: true,
		StaticValueNames: map[string]bool{"none": true},
	},
	"backdrop-filter": {
		DefaultValue: "none", DefaultValuePresent: true,
		StaticValueNames: map[string]bool{"none": true},
	},
	"mask": {
		StaticValueNames: map[string]bool{"none": true},
	},

	// `block` and `inline` are `staticUtility` registrations that the descriptor extractor probed as
	// functional roots, which `TestGapRootExemptionsAreExactlyTheStaticCollision` already records.
	// They take no value, so their description is empty and the emitter answers the static form.
	"block":  {},
	"inline": {},

	// `flex` is not valueless: utilities.ts:1189 takes a fraction whose halves are positive integers
	// and :1195 takes a bare positive integer, refusing a modifier on the latter. `flex-1` is 116 of
	// the corpus occurrences this rule sees, so treating the root as valueless reported it 116 times.
	"flex": {
		SupportsFractions: true,
		Arms: []FunctionalUtilityArm{
			{BareValue: BareValuePositiveInteger, RefusesModifier: true},
		},
	},

	// The negative spellings. Upstream registers each as a second root whose resolved value is
	// negated, and the description is the positive root's with the flag set, which is what
	// `SupportsNegative` means everywhere else in this port.
	"-scale":    {SupportsNegative: true, Arms: []FunctionalUtilityArm{{ThemeKeys: []string{"--scale"}, BareValue: BareValuePositiveInteger}}},
	"scale":     {Arms: []FunctionalUtilityArm{{ThemeKeys: []string{"--scale"}, BareValue: BareValuePositiveInteger}}},
	"-rotate":   {SupportsNegative: true, Arms: []FunctionalUtilityArm{{ThemeKeys: []string{"--rotate"}, BareValue: BareValuePositiveInteger, BareValueSuffix: "deg"}}},
	"rotate":    {Arms: []FunctionalUtilityArm{{ThemeKeys: []string{"--rotate"}, BareValue: BareValuePositiveInteger, BareValueSuffix: "deg"}}},
	"-bg-conic": {SupportsNegative: true, DefaultValue: "in oklab", DefaultValuePresent: true, Arms: []FunctionalUtilityArm{{ThemeKeys: []string{"--conic"}, BareValue: BareValuePositiveInteger}}},
}

// ClassValueResolvesIn reports whether a parsed class resolves to a value the engine would compile.
//
// The parser answers a different question and answers it correctly: it decides where a root ends and
// whether the variants read, which is structure. `text-ss` parses as root `text` with the named value
// `ss` and compiles to nothing, so a caller that stops at the parser cannot tell it from `text-lg`.
//
// # What this asks, and of whom
//
// Only a framework functional root is answered here. A repository `@utility` root is compiled by the
// evaluator, which knows the block; a static is a whole name rather than a root plus a value, so
// there is no value to resolve; an arbitrary property carries its own declaration. Each of those
// returns true, which is a decline rather than an endorsement: this function's job is to catch a
// value that resolves nowhere, and a root it does not describe has no such value to catch.
//
// Declining is the safe direction and it is the one this rule already takes. The gap it leaves is a
// class this port cannot answer going unreported, which is the state before this function existed.
// The gap it must never open is a working class reported, which is why an undescribed root returns
// true rather than false.
func ClassValueResolvesIn(candidate *ParsedCandidate, system *LoadedDesignSystem) bool {
	if candidate == nil || system == nil {
		return true
	}
	if candidate.Kind != ParsedCandidateKindFunctional {
		return true
	}

	root := candidate.Root
	negative := false
	if trimmed, cut := strings.CutPrefix(root, "-"); cut {
		root, negative = trimmed, true
	}

	// A root this repository declares is the evaluator's, and the evaluator is asked.
	//
	// It used to be trusted: a repository root answered true without compiling, on the reasoning
	// that declining is the safe direction. It hid a dead class. `content--0-4` reads as the
	// repository's `@utility content--*` with value `0-4`, which needs a `--color-content-0-4` the
	// theme never declares (it has `content-0` through `content-4` and `content--1` through
	// `content--3`), so it emits nothing and Prettier's Tailwind plugin ranks it null, while
	// no-unknown-classes reported nothing on the two Structure sites that carried it (#1hmzh0z).
	// The evaluator compiles the block, so its answer is the engine's, and a candidate it cannot
	// compile is one the engine cannot either.
	//
	// The comment this replaces cited `shadow--3` as a repository class. It is not: as the
	// repository's root it needs `--shadow-3`, which does not exist, and it renders through the
	// framework's `shadow` with value `-3`. That is a second candidate, which is why callers ask
	// whether any candidate resolves rather than only the first.
	//
	// `DeclaresFunctionalUtility` rather than `HasUtility`, which answers whether a root exists at
	// all and so says yes to every framework root. Measured with the wrong one: `text` reads as
	// repository-declared and every class goes to the evaluator.
	if system.DeclaresFunctionalUtility(candidate.Root) || system.DeclaresFunctionalUtility(root) {
		_, compiles := system.Utilities().Reading(candidate)
		return compiles
	}

	description := descriptionForRoot(root)
	if description == nil {
		return true
	}

	_, produced := ResolveFunctionalUtilityValue(candidate, description, system.Theme(), negative)
	return produced
}

// descriptionForRoot returns the resolution description of a framework functional root.
//
// The three slices partition the framework's functional roots and do not overlap, which
// TestEmitterSlicesDoNotOverlap holds, so the order here is not a precedence.
func descriptionForRoot(root string) *FunctionalUtilityDescription {
	if utility, known := FrameworkFunctionalUtilities[root]; known {
		return utility.Description()
	}
	if utility, known := FrameworkMultiDeclarationUtilities[root]; known {
		description := utility.Description()

		// A root that consults a colour namespace reads a modifier as the alpha, which upstream
		// composes onto an arbitrary value rather than refusing: `divide-[#abc]/10` compiles.
		// Detected from the namespaces rather than listed, so a root added upstream is covered.
		for _, key := range utility.ThemeKeys {
			if strings.HasSuffix(key, "-color") || key == "--color" {
				description.AcceptsModifierOnArbitrary = true
				break
			}
		}

		// The root-defined keywords, which `Description` does not carry because the reading path
		// answers them before resolution runs: `ReadingFor` checks `LiteralReadings` first and
		// returns, so a keyword never reaches the pipeline there.
		//
		// This caller has no reading to short-circuit and asks resolution directly, so the keywords
		// have to be visible to it or every one reads as an unresolvable value. Measured with them
		// absent: 335 corpus classes reported, `flex-1`, `rounded-full`, `transition-colors` and
		// `aspect-square` among them, every one a class the engine compiles.
		if len(utility.LiteralReadings) > 0 {
			if description.StaticValueNames == nil {
				description.StaticValueNames = make(map[string]bool, len(utility.LiteralReadings))
			}
			for literal := range utility.LiteralReadings {
				description.StaticValueNames[literal] = true
			}
		}
		return description
	}
	if description, known := GapRootDescriptions[root]; known {
		// Copied rather than mutated, because the map holds one description per root and several
		// roots share one: every `borderSideUtility` root returns the same pointer.
		resolved := *description
		resolved.AcceptsModifierOnArbitrary = gapRootAcceptsModifierOnArbitrary(description)
		return &resolved
	}
	return nil
}

// gapRootAcceptsModifierOnArbitrary is whether any of a gap root's arms resolves a colour.
//
// A colour arm reads a modifier as the alpha, and upstream's `asColor` composes it onto an arbitrary
// value rather than refusing: `bg-[#FD555E]/15` is a hex with fifteen percent alpha and compiles.
//
// Derived from the arms rather than set per root, so a root gaining a colour arm is covered without
// a second list to keep in step. Measured by running the linter over this repository with it absent:
// `bg-[#FD555E]/15`, `bg-[#FEA625]/15`, `bg-[#FEA625]/5` and `border-[#FEA625]/40` were reported as
// unknown, and all four compile.
func gapRootAcceptsModifierOnArbitrary(description *FunctionalUtilityDescription) bool {
	for index := range description.Arms {
		if description.Arms[index].IsColor {
			return true
		}
	}
	return false
}
