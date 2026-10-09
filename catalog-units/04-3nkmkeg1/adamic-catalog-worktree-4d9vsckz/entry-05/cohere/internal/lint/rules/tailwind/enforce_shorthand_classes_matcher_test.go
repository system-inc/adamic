package tailwind

import (
	"math/rand"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

// The shorthand table used to be matched with one regular expression per longhand, run against every
// class for every rule of every family on every class literal. The matcher now reads each pattern as
// the plain text it is (see shorthandTablePattern). This file keeps the expression matcher as it
// shipped, renamed, as the oracle the new one has to agree with on every input, rather than trusting
// a by-hand argument that `^border-t-(.*)$` and a prefix test are the same thing.

// TestShorthandMatcherAgreesWithTheExpressions builds every table rule's longhands and perturbs one of
// them in each way a collapse has to agree on (value, variant, sign, marker, or a missing side), adds
// near misses and written-out classes, and requires the two matchers to return the same collapses in
// the same order. The draw is seeded, so a disagreement reproduces.
func TestShorthandMatcherAgreesWithTheExpressions(t *testing.T) {
	t.Parallel()
	random := rand.New(rand.NewSource(20261004))

	var roots []string
	for _, pattern := range shorthandPatternTable.patterns {
		roots = append(roots, pattern.text)
	}
	// Near misses: a prefix without its dash, a longer root that shares a prefix, the shorthands
	// themselves, and classes that hold no pattern at all.
	roots = append(roots, "w", "border-spacing-", "border-", "rounded-", "scroll-", "inset-", "justify-",
		"place-content-", "size-", "p-", "m-", "px-", "my-", "flex", "truncate", "text-ellipsis",
		"overflow-x-hidden", "space-", "gap-", "scale-3d")
	values := []string{"4", "2", "0", "[2px]", "", "md", "-1", "screen", "auto", "x-2"}
	variants := []string{"", "hover:", "sm:", "[&_button]:", "sm:hover:"}

	draw := func() string {
		root := roots[random.Intn(len(roots))]
		className := root
		if strings.HasSuffix(root, "-") {
			className += values[random.Intn(len(values))]
		}
		switch random.Intn(6) {
		case 0:
			className = "!" + className
		case 1:
			className += "!"
		}
		if random.Intn(5) == 0 {
			className = "-" + className
		}
		return variants[random.Intn(len(variants))] + className
	}

	// A decoration a collapse has to agree on: variants, sign, marker.
	decorate := func(className string, variant string, negative bool, marker int) string {
		switch marker {
		case 1:
			className = "!" + className
		case 2:
			className += "!"
		}
		if negative {
			className = "-" + className
		}
		return variant + className
	}

	collapsing := 0
	for iteration := 0; iteration < 60_000; iteration++ {
		var classNames []string
		if random.Intn(3) > 0 {
			// Every longhand of one table rule, agreeing on everything, then perturbed: one class's
			// value, variant, sign or marker changed, or one class dropped, or none.
			family := shorthandPatternTable.families[random.Intn(len(shorthandPatternTable.families))]
			tableRule := family[random.Intn(len(family))]
			value := values[random.Intn(len(values))]
			variant := variants[random.Intn(len(variants))]
			negative := random.Intn(4) == 0
			marker := random.Intn(4) % 3
			for _, index := range tableRule.longhands {
				className := shorthandPatternTable.patterns[index].text
				if shorthandPatternTable.patterns[index].captures {
					className += value
				}
				classNames = append(classNames, decorate(className, variant, negative, marker))
			}
			// The same rule again behind another variant, a third of the time, so two groups can
			// both complete it and the order groups are tried in decides which one is reported.
			if random.Intn(3) == 0 {
				otherVariant := variants[random.Intn(len(variants))]
				otherValue := values[random.Intn(len(values))]
				for _, index := range tableRule.longhands {
					className := shorthandPatternTable.patterns[index].text
					if shorthandPatternTable.patterns[index].captures {
						className += otherValue
					}
					classNames = append(classNames, decorate(className, otherVariant, negative, marker))
				}
			}
			perturbed := random.Intn(len(classNames))
			switch random.Intn(8) {
			case 0:
				classNames[perturbed] = classNames[perturbed] + "0"
			case 1:
				classNames[perturbed] = "focus:" + classNames[perturbed]
			case 2:
				classNames[perturbed] = "-" + classNames[perturbed]
			case 3:
				classNames[perturbed] = classNames[perturbed] + "!"
			case 4:
				classNames = append(classNames[:perturbed], classNames[perturbed+1:]...)
			}
			// Noise, and sometimes the shorthand itself.
			for count := random.Intn(3); count > 0; count-- {
				classNames = append(classNames, draw())
			}
			if random.Intn(6) == 0 {
				classNames = append(classNames, decorate(strings.ReplaceAll(tableRule.shorthands[0], "$1", value),
					variant, negative, marker))
			}
			random.Shuffle(len(classNames), func(left int, right int) {
				classNames[left], classNames[right] = classNames[right], classNames[left]
			})
		} else {
			classNames = make([]string, 1+random.Intn(6))
			for index := range classNames {
				classNames[index] = draw()
			}
		}
		if expectShorthandMatchersAgree(t, classNames) {
			collapsing++
		}
	}
	// Proof the draw reached the cases that matter: an agreement made only of empty answers would
	// pass against a matcher that never finds anything.
	if collapsing < 10_000 {
		t.Fatalf("only %d of 60,000 draws collapsed, too few to say the matchers agree on collapses", collapsing)
	}
}

// TestShorthandMatcherAgreesOnTheFixtureCorpus runs the two matchers over the class lists the rule's
// fixtures and the migration's corpus name, where a disagreement would be a real finding gained or
// lost.
func TestShorthandMatcherAgreesOnTheFixtureCorpus(t *testing.T) {
	t.Parallel()
	for _, classes := range []string{
		"ps-4 pe-4", "pt-2 pb-2", "ml-1 mr-1 mt-1 mb-1", "w-4 h-4", "top-0 right-0 bottom-0 left-0",
		"hover:ps-4 hover:pe-4", "-mt-2 -mb-2", "!pt-2 !pb-2", "pt-2! pb-2!", "ps-[2px] pe-[2px]",
		"[&_button]:border-s-0 [&_button]:border-e-0", "ps-4 pe-8", "ps-4 hover:pe-4", "-mt-2 mb-2",
		"!pt-2 pb-2", "ps-4", "px-4 py-4 p-4", "flex items-center", "ps-4 pe-4 mt-2 mb-2",
		"h-screen w-screen", "h-auto w-auto", "px-4 py-4", "w-8 h-8", "gap-x-2 gap-y-2",
		"border-l border-r", "rounded-tl-md rounded-tr-md", "overflow-x-hidden overflow-y-hidden",
		"sm:px-4 sm:py-4", "start-0 end-0", "-mt-1 -mb-1", "px-[3px] py-[3px]", "px-4 py-2", "w-8 h-4",
		"px-4 sm:py-4", "ml-2", "text-left", "space-x-2", "truncate",
		"overflow-hidden text-ellipsis whitespace-nowrap", "scale-x-50 scale-y-50 scale-z-50",
		"content-center justify-content-center", "border-spacing-x-1 border-spacing-y-1",
		"border-t-2 border-b-2 border-l-2 border-r-2", "w- h-", "w-4 w-4 h-4",
		"hover:ps-4 hover:pe-4 ps-2 pe-2", "ps-2 hover:ps-4 pe-2 hover:pe-4",
		// A class matching nothing still puts its variant first, which decides which group reports.
		"justify-x-2 [&_button]:items-2 [&_button]:justify-items-2 items-4 justify-items-4",
	} {
		expectShorthandMatchersAgree(t, strings.Fields(classes))
	}
}

// expectShorthandMatchersAgree fails when the two matchers answer one class list differently, and
// reports whether the list collapsed.
func expectShorthandMatchersAgree(t *testing.T, classNames []string) bool {
	t.Helper()
	expected := regexShorthandCollapses(classNames)
	actual := shorthandCollapses(classNames)
	if len(expected) == 0 && len(actual) == 0 {
		return false
	}
	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("the matchers disagree on %q:\n  expressions: %+v\n  plain text:  %+v", classNames, expected, actual)
	}
	return true
}

// The expression matcher as it shipped, with its names prefixed by regex.

// regexShorthandCollapses finds every collapse available in one class list, at most one per family.
func regexShorthandCollapses(classNames []string) []shorthandCollapse {
	dissected := make([]shorthandDissection, 0, len(classNames))
	for _, className := range classNames {
		dissected = append(dissected, dissectShorthandClass(className))
	}

	present := make(map[string]bool, len(classNames))
	for _, className := range classNames {
		present[className] = true
	}

	collapses := []shorthandCollapse{}
	for _, group := range shorthandGroups {
		if collapse, found := regexFirstCollapseInGroup(group, dissected, present); found {
			collapses = append(collapses, collapse)
		}
	}
	return collapses
}

// regexFirstCollapseInGroup returns the first rule in one family that matches, longest-first.
//
// Upstream sorts each family by pattern count descending and takes the first match, so a class list
// holding all four margin sides collapses to `m-1` rather than reporting the two-sided rules that
// also match.
func regexFirstCollapseInGroup(group []shorthandRule, dissected []shorthandDissection,
	present map[string]bool) (shorthandCollapse, bool) {
	ordered := make([]shorthandRule, len(group))
	copy(ordered, group)
	for outer := 1; outer < len(ordered); outer++ {
		candidate := ordered[outer]
		inner := outer - 1
		for inner >= 0 && len(ordered[inner].Longhands) < len(candidate.Longhands) {
			ordered[inner+1] = ordered[inner]
			inner--
		}
		ordered[inner+1] = candidate
	}

	for _, shorthand := range ordered {
		if collapse, found := regexMatchShorthandRule(shorthand, dissected, present); found {
			return collapse, true
		}
	}
	return shorthandCollapse{}, false
}

// regexMatchShorthandRule looks for one rule's longhands among classes sharing a variant prefix.
func regexMatchShorthandRule(shorthand shorthandRule, dissected []shorthandDissection,
	present map[string]bool) (shorthandCollapse, bool) {
	byVariants := map[string][]shorthandDissection{}
	order := []string{}
	for _, class := range dissected {
		if _, seen := byVariants[class.variants]; !seen {
			order = append(order, class.variants)
		}
		byVariants[class.variants] = append(byVariants[class.variants], class)
	}

	for _, variants := range order {
		collapse, found := regexMatchWithinVariantGroup(shorthand, byVariants[variants], present)
		if found {
			return collapse, true
		}
	}
	return shorthandCollapse{}, false
}

// regexMatchWithinVariantGroup matches one rule against classes that already share a variant prefix.
func regexMatchWithinVariantGroup(shorthand shorthandRule, classes []shorthandDissection,
	present map[string]bool) (shorthandCollapse, bool) {
	matched := make([]shorthandDissection, 0, len(shorthand.Longhands))
	var captured []string

	for _, pattern := range shorthand.Longhands {
		expression := regexShorthandPattern(pattern)
		for _, class := range classes {
			groups := expression.FindStringSubmatch(class.base)
			if groups == nil {
				continue
			}
			if captured == nil {
				captured = groups
			} else {
				if len(groups) != len(captured) {
					continue
				}
				agrees := true
				for index := 1; index < len(groups); index++ {
					if groups[index] != captured[index] {
						agrees = false
						break
					}
				}
				if !agrees {
					continue
				}
			}
			matched = append(matched, class)
			break
		}
	}

	if len(matched) != len(shorthand.Longhands) {
		return shorthandCollapse{}, false
	}

	// The sign and the marker have to agree across the whole group, not merely be present on one.
	// Upstream checks each separately and each one silently widens the rule if dropped.
	markerEnd := false
	for _, class := range matched {
		if class.markerEnd {
			markerEnd = true
			break
		}
	}
	important := markerEnd
	if !important {
		for _, class := range matched {
			if class.important {
				important = true
				break
			}
		}
	}
	negative := false
	for _, class := range matched {
		if class.negative {
			negative = true
			break
		}
	}
	for _, class := range matched {
		if class.important != important && class.markerEnd != important {
			return shorthandCollapse{}, false
		}
		if class.negative != negative {
			return shorthandCollapse{}, false
		}
	}

	variants := matched[0].variants
	longhands := make([]string, 0, len(matched))
	for _, class := range matched {
		longhands = append(longhands, class.className)
	}

	shorthands := make([]string, 0, len(shorthand.Shorthands))
	for _, substitute := range shorthand.Shorthands {
		base := substitute
		for index := len(captured) - 1; index >= 1; index-- {
			base = strings.ReplaceAll(base, "$"+regexItoa(index), captured[index])
		}
		shorthands = append(shorthands, buildShorthandClass(variants, base, negative, important, markerEnd))
	}

	// Already written alongside the longhands, so there is nothing to say.
	allPresent := true
	for _, className := range shorthands {
		if !present[className] {
			allPresent = false
			break
		}
	}
	if allPresent {
		return shorthandCollapse{}, false
	}

	return shorthandCollapse{longhands: longhands, shorthands: shorthands}, true
}

// regexShorthandPatterns is the compiled table, built once at init.
//
// Built eagerly rather than memoized on first use, and that is a correctness fix rather than a
// preference. A lazily-filled map was the first version and it crashed the whole run with "concurrent
// map read and map write": `Graph.Walk` dispatches files across goroutines, so every rule's listener
// body runs in parallel with itself on different files. It failed one run in five, which is the
// worst rate to have — often enough to be real, rare enough to look like someone else's flake.
//
// The table is a package-level constant of 48 rules, so there is nothing to defer: compiling it at
// init costs one pass over a fixed list and removes the shared mutable state entirely, which is a
// better answer than a mutex around a cache that would never miss twice.
var regexShorthandPatterns = compileRegexShorthandPatterns()

func compileRegexShorthandPatterns() map[string]*regexp.Regexp {
	compiled := map[string]*regexp.Regexp{}
	for _, group := range shorthandGroups {
		for _, shorthand := range group {
			for _, pattern := range shorthand.Longhands {
				if _, isCompiled := compiled[pattern]; isCompiled {
					continue
				}
				compiled[pattern] = regexp.MustCompile("^" + pattern + "$")
			}
		}
	}
	return compiled
}

// shorthandPattern returns one table entry's compiled expression.
func regexShorthandPattern(pattern string) *regexp.Regexp {
	return regexShorthandPatterns[pattern]
}

// regexItoa is strconv.Itoa for the single digit a substitution placeholder can hold.
func regexItoa(value int) string {
	return string(rune('0' + value))
}
