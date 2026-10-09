package tailwind

import (
	"sort"
	"strings"
	"testing"
)

// gapRootResolutionCases is what the engine answers for each class, measured rather than reasoned.
//
// Every entry came from `designSystem.candidatesToCss([class])` on ahra's live design system, where
// `null` is the engine refusing to compile the class at all. That is the question `no-unknown-classes`
// needs answered and the one these descriptions exist to answer in Go.
//
// Chosen to reach each arm rather than to be large. For every family: a bare root, a theme hit, a
// colour hit, a colour keyword, a bare-integer fallback, a modifier on the colour arm, a modifier on
// the other arm, and a value that resolves nowhere.
var gapRootResolutionCases = map[string]bool{
	// borderSideUtility. `border-2/50` against `border-red-500/50` is the per-arm modifier rule:
	// the colour arm takes a modifier as its alpha and the width arm returns on one.
	"border": true, "border-2": true, "border-red-500": true, "border-inherit": true,
	"border-transparent": true, "border-current": true, "border-red-500/50": true,
	"border-x-2": true, "border-l-red-500": true,
	"border-2/50": false, "border-nonsense": false, "border-dark-4": false,

	// maskStopUtility. `1.5` against `3.7` is the spacing multiplier rather than any integer test,
	// and `50%` against `nonsense` is the inference gate, which returns rather than falling through.
	"mask-x-from-4": true, "mask-x-from-50%": true, "mask-x-from-red-500": true,
	"mask-x-from-inherit": true, "mask-x-from-red-500/50": true, "mask-t-to-8": true,
	"mask-linear-from-4": true, "mask-radial-to-red-500": true, "mask-conic-from-25%": true,
	"mask-x-from-1.5":  true,
	"mask-x-from-4/50": false, "mask-x-from-nonsense": false, "mask-x-from": false,
	"mask-x-from-3.7": false,

	// The colour-then-width closures. `stroke-2/50` compiling while `border-2/50` and `outline-2/50`
	// do not is the measurement that stopped these three sharing one description.
	"outline": true, "outline-2": true, "outline-red-500": true, "outline-inherit": true,
	"outline-red-500/50": true,
	"outline-2/50":       false, "outline-nonsense": false,
	"stroke-2": true, "stroke-red-500": true, "stroke-inherit": true, "stroke-red-500/50": true,
	"stroke-2/50": true,
	"stroke":      false, "stroke-nonsense": false,
	"decoration-2": true, "decoration-red-500": true, "decoration-inherit": true,
	"decoration-red-500/50": true,
	"decoration":            false, "decoration-2/50": false, "decoration-nonsense": false,
	"ring": true, "ring-2": true, "ring-red-500": true, "ring-inherit": true, "ring-red-500/50": true,
	"ring-2/50": false, "ring-nonsense": false,
	"inset-ring": true, "inset-ring-2": true, "inset-ring-red-500": true,
	"inset-ring-nonsense": false,
	"fill-red-500":        true, "fill-inherit": true,
	"fill": false, "fill-nonsense": false, "fill-2": false,

	// The colour-then-scale closures. `text-ss` and `text-dark-4` are the two classes this repository
	// actually writes that compile to nothing, and are why #31bbwty exists.
	"text-lg": true, "text-red-500": true, "text-inherit": true, "text-lg/6": true,
	"text-red-500/50": true, "text-lg/none": true, "text-lg/1.5": true,
	"text": false, "text-ss": false, "text-dark-4": false,
	"bg-red-500": true, "bg-inherit": true, "bg-red-500/50": true, "bg-conic": true,
	"bg-conic-30": true, "bg-radial": true,
	"bg": false, "bg-nonsense": false,
	"font-mono": true, "font-bold": true, "font-medium": true,
	"font": false, "font-nonsense": false,

	// The shadow family and the two single-namespace roots.
	"shadow": true, "shadow-lg": true, "shadow-red-500": true, "shadow-lg/50": true,
	"shadow-inherit":  true,
	"shadow-nonsense": false,
	"drop-shadow":     true, "drop-shadow-lg": true, "drop-shadow-red-500": true,
	"drop-shadow-nonsense": false,
	"text-shadow-lg":       true, "text-shadow-red-500": true,
	"text-shadow": false, "text-shadow-nonsense": false,
	"inset-shadow-sm": true, "inset-shadow-red-500": true,
	"inset-shadow": false, "inset-shadow-nonsense": false,
	"duration-300": true,
	"duration":     false, "duration-nonsense": false,

	// The arbitrary-only closures, which consult no namespace: the keyword compiles and any other
	// named value resolves nowhere.
	"filter": true, "filter-none": true, "transform": true,
	"backdrop-filter": true, "backdrop-filter-none": true,
	"filter-nonsense": false, "mask-nonsense": false, "mask": false,

	// `transform-none` and `mask-none` are absent for the same reason as `block`: the parser reads
	// each as its own static root rather than as the functional root plus a named value, so no
	// functional description is what answers them. `filter-none` and `backdrop-filter-none` are
	// present because those two do parse functionally, which is a difference between the four that
	// is worth having measured rather than assumed.

	// `block`, `inline` and `flex` are deliberately absent. The engine compiles all three, and it
	// compiles them as `staticUtility` registrations rather than through any functional path:
	// utilities.ts:956 and :958 register `display: block` and `display: inline` directly. The
	// descriptor extractor probed them as functional roots, which
	// TestGapRootExemptionsAreExactlyTheStaticCollision already records as the one place these two
	// registration kinds collide.
	//
	// So a functional description refusing them is correct rather than a gap, and a caller must ask
	// the static table before this one. Listing them here as `true` would make the description claim
	// something it does not do, and as `false` would claim the engine refuses them, and both are
	// worse than saying why they are not cases.

	// The negative spellings, which resolve exactly as their positive roots do.
	"scale-75": true, "-scale-75": true, "rotate-45": true, "-rotate-45": true,
	"-bg-conic-30": true,
}

// Every checked-in description answers what the engine answers, class for class.
//
// The two are independent claims: the descriptions were read from `utilities.ts` at the v4.3.3 tag
// and the expectations were measured through `candidatesToCss`, so an agreement is evidence rather
// than a restatement. That is the same discipline the three emitter slices were held to.
func TestGapRootDescriptionsResolveWhatTheEngineCompiles(t *testing.T) {
	t.Parallel()
	system, _ := liveTableFor(t, corpusRepositories[0].spelling)
	if system == nil {
		t.Skip("no design system loaded")
	}

	var agreed, unparsed, undescribed int
	var disagreements []string

	for className, engineCompiles := range gapRootResolutionCases {
		candidates := ParseCandidate(className, system)
		if len(candidates) == 0 {
			unparsed++
			t.Errorf("%s does not parse, so this case measures nothing", className)
			continue
		}

		candidate := &candidates[0]
		description, described := GapRootDescriptions[strings.TrimPrefix(candidate.Root, "-")]
		if !described {
			undescribed++
			t.Errorf("%s reads as root %q and no description covers it", className, candidate.Root)
			continue
		}

		_, produced := ResolveFunctionalUtilityValue(candidate, description, system.Theme(), false)
		if produced == engineCompiles {
			agreed++
			continue
		}
		disagreements = append(disagreements,
			className+": the engine "+compiledWord(engineCompiles)+" it and the description "+compiledWord(produced))
	}

	sort.Strings(disagreements)
	t.Logf("gap root resolution: %d cases, %d agreed, %d disagreed", len(gapRootResolutionCases), agreed, len(disagreements))

	for _, disagreement := range disagreements {
		t.Errorf("%s", disagreement)
	}
	if agreed == 0 {
		t.Fatal("no case was compared, so this test measured nothing")
	}
}

func compiledWord(compiles bool) string {
	if compiles {
		return "resolves"
	}
	return "refuses"
}

// Every root with an emitter has a description, and every description has an emitter.
//
// The two maps answer different halves of one root and drift is what leaves a class unanswerable: a
// root with an emitter and no description cannot be asked whether its value resolves, and a
// description with no emitter describes something this port does not otherwise know about.
func TestEveryGapRootIsDescribed(t *testing.T) {
	t.Parallel()
	var missingDescription, missingEmitter []string

	for root := range gapEmitters {
		if _, described := GapRootDescriptions[root]; !described {
			missingDescription = append(missingDescription, root)
		}
	}
	for root := range GapRootDescriptions {
		if _, emits := gapEmitters[root]; !emits {
			missingEmitter = append(missingEmitter, root)
		}
	}

	sort.Strings(missingDescription)
	sort.Strings(missingEmitter)
	t.Logf("gap roots: %d emitters, %d descriptions", len(gapEmitters), len(GapRootDescriptions))

	if len(missingDescription) > 0 {
		t.Errorf("%d gap roots emit and have no description: %s",
			len(missingDescription), strings.Join(missingDescription, ", "))
	}
	if len(missingEmitter) > 0 {
		t.Errorf("%d descriptions name a root with no emitter: %s",
			len(missingEmitter), strings.Join(missingEmitter, ", "))
	}
}

// An arbitrary colour carrying an alpha modifier resolves, on every root with a colour arm.
//
// `bg-[#FD555E]/15` is a hex with fifteen percent alpha and the engine compiles it. The check that
// decides this reads the arms, and a first version read only the multi-declaration table's theme
// keys, so it never reached a gap root: `bg`, `border`, `text` and their kin refused every arbitrary
// colour that carried a modifier.
//
// Found by building the linter and running it over this repository rather than by a test, which is
// why it is a test now: `bg-[#FD555E]/15`, `bg-[#FEA625]/15`, `bg-[#FEA625]/5` and
// `border-[#FEA625]/40` were reported as unknown classes in live markup.
//
// The population is every gap root with a colour arm rather than a list, so a root gaining one joins
// this test.
func TestArbitraryColorsWithAnAlphaResolveOnEveryColorRoot(t *testing.T) {
	t.Parallel()
	forEachEngineSystem(t, func(t *testing.T, system *LoadedDesignSystem) {

		var checked int
		for root, description := range GapRootDescriptions {
			if !gapRootAcceptsModifierOnArbitrary(description) {
				continue
			}

			className := root + "-[#FD555E]/15"
			candidates := ParseCandidate(className, system)
			if len(candidates) == 0 {
				continue
			}
			// Only where the root reads as itself; a spelling the parser splits differently is a
			// different root's case.
			if candidates[0].Root != root {
				continue
			}
			checked++

			if !ClassValueResolvesIn(&candidates[0], system) {
				t.Errorf("%s is an arbitrary colour with an alpha and does not resolve, which reports working markup", className)
			}
		}

		t.Logf("colour roots checked with an arbitrary value and an alpha: %d", checked)
		if checked == 0 {
			t.Fatal("no colour root was checked, so this test measured nothing")
		}
	})
}
