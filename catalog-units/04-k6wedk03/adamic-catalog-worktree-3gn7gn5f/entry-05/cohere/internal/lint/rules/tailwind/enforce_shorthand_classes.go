package tailwind

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	tailwindengine "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse"
)

// messageShorthandClasses names the longhands and the class that replaces them.
func messageShorthandClasses(longhands string, shorthands string) rule.Message {
	return rule.Message{
		Id: "shorthandClasses",
		Description: "\"" + longhands + "\" is the longhand spelling of \"" + shorthands + "\", " +
			"which sets the same declarations in one class. The pair carries no information the " +
			"single class does not, and it costs the next reader a moment working out whether the " +
			"two halves were meant to move together or drifted apart. Write it as \"" +
			shorthands + "\".",
	}
}

// EnforceShorthandClassesOptions lets a project name the surfaces that carry class strings.
type EnforceShorthandClassesOptions struct {
	TailwindLocationOptions
	TailwindClassLiteralOptions
}

// shorthandRule is one collapse: a set of longhand patterns and the classes they become.
type shorthandRule struct {
	Longhands  []string
	Shorthands []string
}

// shorthandGroups is upstream's table, transcribed rather than reasoned about.
//
// Extracted mechanically from `enforce-shorthand-classes.ts` at 4.7.0 rather than retyped, because
// it is 48 rules and a transcription error would be a silent miss on one utility family rather than
// a failure anywhere.
//
// The two levels are load-bearing. The outer level groups rules that describe the same property
// family, and only one rule per family may fire, so `px-4 py-4 p-4` does not report twice. Within a
// family the rules are tried longest-first, which is what makes `ml-1 mr-1 mt-1 mb-1` collapse
// straight to `m-1` rather than to `mx-1 my-1` and then again.
var shorthandGroups = [][]shorthandRule{
	{
		{Longhands: []string{"w-(.*)", "h-(.*)"}, Shorthands: []string{"size-$1"}},
	},
	{
		{Longhands: []string{"ml-(.*)", "mr-(.*)", "mt-(.*)", "mb-(.*)"}, Shorthands: []string{"m-$1"}},
		{Longhands: []string{"mx-(.*)", "my-(.*)"}, Shorthands: []string{"m-$1"}},
		{Longhands: []string{"ms-(.*)", "me-(.*)"}, Shorthands: []string{"mx-$1"}},
		{Longhands: []string{"ml-(.*)", "mr-(.*)"}, Shorthands: []string{"mx-$1"}},
		{Longhands: []string{"mt-(.*)", "mb-(.*)"}, Shorthands: []string{"my-$1"}},
	},
	{
		{Longhands: []string{"pl-(.*)", "pr-(.*)", "pt-(.*)", "pb-(.*)"}, Shorthands: []string{"p-$1"}},
		{Longhands: []string{"px-(.*)", "py-(.*)"}, Shorthands: []string{"p-$1"}},
		{Longhands: []string{"ps-(.*)", "pe-(.*)"}, Shorthands: []string{"px-$1"}},
		{Longhands: []string{"pl-(.*)", "pr-(.*)"}, Shorthands: []string{"px-$1"}},
		{Longhands: []string{"pt-(.*)", "pb-(.*)"}, Shorthands: []string{"py-$1"}},
	},
	{
		{Longhands: []string{"border-t-(.*)", "border-b-(.*)", "border-l-(.*)", "border-r-(.*)"}, Shorthands: []string{"border-$1"}},
		{Longhands: []string{"border-x-(.*)", "border-y-(.*)"}, Shorthands: []string{"border-$1"}},
		{Longhands: []string{"border-s-(.*)", "border-e-(.*)"}, Shorthands: []string{"border-x-$1"}},
		{Longhands: []string{"border-l-(.*)", "border-r-(.*)"}, Shorthands: []string{"border-x-$1"}},
		{Longhands: []string{"border-t-(.*)", "border-b-(.*)"}, Shorthands: []string{"border-y-$1"}},
	},
	{
		{Longhands: []string{"border-spacing-x-(.*)", "border-spacing-y-(.*)"}, Shorthands: []string{"border-spacing-$1"}},
	},
	{
		{Longhands: []string{"rounded-tl-(.*)", "rounded-tr-(.*)", "rounded-bl-(.*)", "rounded-br-(.*)"}, Shorthands: []string{"rounded-$1"}},
		{Longhands: []string{"rounded-t-(.*)", "rounded-b-(.*)"}, Shorthands: []string{"rounded-$1"}},
		{Longhands: []string{"rounded-l-(.*)", "rounded-r-(.*)"}, Shorthands: []string{"rounded-$1"}},
		{Longhands: []string{"rounded-tl-(.*)", "rounded-tr-(.*)"}, Shorthands: []string{"rounded-t-$1"}},
		{Longhands: []string{"rounded-bl-(.*)", "rounded-br-(.*)"}, Shorthands: []string{"rounded-b-$1"}},
		{Longhands: []string{"rounded-tl-(.*)", "rounded-bl-(.*)"}, Shorthands: []string{"rounded-l-$1"}},
		{Longhands: []string{"rounded-tr-(.*)", "rounded-br-(.*)"}, Shorthands: []string{"rounded-r-$1"}},
	},
	{
		{Longhands: []string{"scroll-mt-(.*)", "scroll-mb-(.*)", "scroll-ml-(.*)", "scroll-mr-(.*)"}, Shorthands: []string{"scroll-m-$1"}},
		{Longhands: []string{"scroll-mx-(.*)", "scroll-my-(.*)"}, Shorthands: []string{"scroll-m-$1"}},
		{Longhands: []string{"scroll-ms-(.*)", "scroll-me-(.*)"}, Shorthands: []string{"scroll-mx-$1"}},
		{Longhands: []string{"scroll-ml-(.*)", "scroll-mr-(.*)"}, Shorthands: []string{"scroll-mx-$1"}},
		{Longhands: []string{"scroll-mt-(.*)", "scroll-mb-(.*)"}, Shorthands: []string{"scroll-my-$1"}},
	},
	{
		{Longhands: []string{"scroll-pt-(.*)", "scroll-pb-(.*)", "scroll-pl-(.*)", "scroll-pr-(.*)"}, Shorthands: []string{"scroll-p-$1"}},
		{Longhands: []string{"scroll-px-(.*)", "scroll-py-(.*)"}, Shorthands: []string{"scroll-p-$1"}},
		{Longhands: []string{"scroll-pl-(.*)", "scroll-pr-(.*)"}, Shorthands: []string{"scroll-px-$1"}},
		{Longhands: []string{"scroll-ps-(.*)", "scroll-pe-(.*)"}, Shorthands: []string{"scroll-px-$1"}},
		{Longhands: []string{"scroll-pt-(.*)", "scroll-pb-(.*)"}, Shorthands: []string{"scroll-py-$1"}},
	},
	{
		{Longhands: []string{"top-(.*)", "right-(.*)", "bottom-(.*)", "left-(.*)"}, Shorthands: []string{"inset-$1"}},
		{Longhands: []string{"inset-x-(.*)", "inset-y-(.*)"}, Shorthands: []string{"inset-$1"}},
	},
	{
		{Longhands: []string{"divide-x-(.*)", "divide-y-(.*)"}, Shorthands: []string{"divide-$1"}},
	},
	{
		{Longhands: []string{"space-x-(.*)", "space-y-(.*)"}, Shorthands: []string{"space-$1"}},
	},
	{
		{Longhands: []string{"gap-x-(.*)", "gap-y-(.*)"}, Shorthands: []string{"gap-$1"}},
	},
	{
		{Longhands: []string{"translate-x-(.*)", "translate-y-(.*)"}, Shorthands: []string{"translate-$1"}},
	},
	{
		{Longhands: []string{"rotate-x-(.*)", "rotate-y-(.*)"}, Shorthands: []string{"rotate-$1"}},
	},
	{
		{Longhands: []string{"skew-x-(.*)", "skew-y-(.*)"}, Shorthands: []string{"skew-$1"}},
	},
	{
		{Longhands: []string{"scale-x-(.*)", "scale-y-(.*)", "scale-z-(.*)"}, Shorthands: []string{"scale-$1", "scale-3d"}},
		{Longhands: []string{"scale-x-(.*)", "scale-y-(.*)"}, Shorthands: []string{"scale-$1"}},
	},
	{
		{Longhands: []string{"content-(.*)", "justify-content-(.*)"}, Shorthands: []string{"place-content-$1"}},
		{Longhands: []string{"items-(.*)", "justify-items-(.*)"}, Shorthands: []string{"place-items-$1"}},
		{Longhands: []string{"self-(.*)", "justify-self-(.*)"}, Shorthands: []string{"place-self-$1"}},
	},
	{
		{Longhands: []string{"overflow-hidden", "text-ellipsis", "whitespace-nowrap"}, Shorthands: []string{"truncate"}},
	},
}

// EnforceShorthandClasses reports longhand class pairs that a single class already expresses.
//
//	valid:   <div className="px-4" />
//	invalid: <div className="ps-4 pe-4" />
//	invalid: <div className="ml-1 mr-1 mt-1 mb-1" />
//
// # What this is for
//
// `ps-4 pe-4` and `px-4` emit the same declarations, so nothing surfaces the difference. The cost
// is that the pair is ambiguous evidence: it reads either as a deliberate choice to set the two
// sides independently, or as two edits that met in the middle, and the next person to change one
// side cannot tell which. The single class says the two move together.
//
// # Everything about a group has to agree, not just the value
//
// The value is the obvious half and the rest is what a reader would miss. Upstream requires the
// captured values to match (`ps-4 pe-8` is not a pair), and also the variants (`ps-4 hover:pe-4` is
// not), the negative sign (`-mt-2 mb-2` is not), and the important marker (`!pt-2 pb-2` is not).
// Each of those was measured against upstream rather than assumed, and each one silently widens the
// rule if dropped.
//
// The rebuilt class carries the variants, the sign and the marker of the group it replaces, so
// `hover:ps-4 hover:pe-4` becomes `hover:px-4` and `-mt-2 -mb-2` becomes `-my-2`.
//
// # Membership is checked, as upstream checks it
//
// Upstream asks the design system whether the shorthand it is about to suggest exists, and declines
// when it does not. This rule skipped that on the reasoning that a missing shorthand needs a theme
// that removed a framework utility, a shape neither corpus repository has, and that
// `no-unknown-classes` would catch the suggestion if one ever appeared.
//
// The table is patterns, and a pattern can build a class Tailwind never had. `w-screen h-screen`
// matches `w-(.*) h-(.*)` and suggests `size-screen`, which does not exist: `w-screen` is `100vw`
// and `h-screen` is `100vh`, so no one class sets both. Found when nested literals reached two of
// them in ahra (#1hmzh0z), and the author following the finding would have deleted a working pair
// for a class that does nothing. `no-unknown-classes` did catch it, one edit too late.
//
// So the design system is asked when there is one. Without a program (a syntactic fixture) or
// without a Tailwind entry point there is nothing to ask, and the rule reports as it did.
//
// Not fixable, for the reason its siblings are not: a class literal here wraps across lines with
// indentation that carries intent, and a fixer would be the first thing to reflow it.
//
// # Declining the fixer removes a finding, not just a fix
//
// Upstream reports every collapse twice on a real tree: once as the longhand pair, and once as
// "Unnecessary whitespace" at the gap its own fixer leaves when it replaces the second class with an
// empty string. Measured on Kirk's tree, where its four findings are the same two sites. This rule
// reports two, and the difference is entirely that second class of finding, which describes the
// state upstream's fix passes through rather than anything in the source a reader wrote.
var EnforceShorthandClasses = rule.Rule{
	Name: "better-tailwindcss/enforce-shorthand-classes",
	// The design system: its stylesheets, read through the recording file system (DesignSystemFS).
	ProgramReads: rule.ReadsCompilerOptions | rule.ReadsDesignSystem,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		configured, isConfigured := rule.OptionsAs[EnforceShorthandClassesOptions](options)

		// nil when there is no design system to ask, and classExistsIn answers true for nil.
		var system *tailwindengine.LoadedDesignSystem
		if ctx.Program != nil {
			if designSystem := DesignSystemForProgramAt(ctx.Program, configured.Location()); designSystem.Err == nil {
				system = designSystem.System
			}
		}

		surfaces := DefaultClassLiteralSurfaces()
		if isConfigured {
			surfaces = configured.ClassLiteralSurfaces()
		}

		reader := surfaces.ReaderFor(ctx.FileCache)

		report := func(node *ast.Node) {
			for _, literal := range reader.ClassLiteralsIn(node) {
				for _, collapse := range shorthandCollapses(SplitClasses(literal.Text)) {
					if !shorthandsExistIn(collapse.shorthands, system) {
						continue
					}
					ctx.ReportRange(literal.Range, messageShorthandClasses(
						strings.Join(collapse.longhands, " "),
						strings.Join(collapse.shorthands, " "),
					))
				}
			}
		}

		listeners := rule.Listeners{}
		for _, kind := range ListenerKinds() {
			listeners[kind] = report
		}
		return listeners
	},
}

// shorthandsExistIn reports whether every class a collapse would suggest is one the design system has.
func shorthandsExistIn(shorthands []string, system *tailwindengine.LoadedDesignSystem) bool {
	for _, shorthand := range shorthands {
		if !classExistsIn(shorthand, system) {
			return false
		}
	}
	return true
}

// shorthandCollapse is one reportable pair: the classes written and the class that replaces them.
type shorthandCollapse struct {
	longhands  []string
	shorthands []string
}

// shorthandDissection is one class split into the parts a collapse has to match on.
type shorthandDissection struct {
	className string
	variants  string
	base      string
	negative  bool
	important bool
	markerEnd bool
	// matches is every table pattern the base matches, with what each captured.
	matches []shorthandMatch
}

// shorthandMatch is one table pattern a class's base matches, and the value its `(.*)` captured.
type shorthandMatch struct {
	pattern int
	capture string
}

// match reports whether this class matches one pattern, and what it captured.
func (d shorthandDissection) match(pattern int) (string, bool) {
	for _, candidate := range d.matches {
		if candidate.pattern == pattern {
			return candidate.capture, true
		}
	}
	return "", false
}

// shorthandPatternSet is a set of pattern indexes, one bit each.
type shorthandPatternSet [2]uint64

func (s *shorthandPatternSet) add(pattern int) {
	s[pattern/64] |= 1 << (pattern % 64)
}

// holds reports whether every pattern in needed is in this set.
func (s shorthandPatternSet) holds(needed shorthandPatternSet) bool {
	return s[0]&needed[0] == needed[0] && s[1]&needed[1] == needed[1]
}

// shorthandTablePattern is one longhand pattern of the table, read as the plain text it is.
//
// Every longhand in upstream's table is either a class written out (`overflow-hidden`) or a literal
// prefix ending in a dash followed by one `(.*)` (`border-t-(.*)`), and anchored at both ends that is
// a prefix test and a slice, so the table is matched without running a regular expression. A pattern
// of any other shape panics at init, so a table edit that brings in real regular-expression syntax
// fails on the first run rather than matching less than it says.
type shorthandTablePattern struct {
	text     string
	captures bool
}

// shorthandTableRule is one rule of the table with its longhands as pattern indexes.
type shorthandTableRule struct {
	longhands  []int
	shorthands []string
	// needs is the set of its longhands, so a variant group missing any one of them is passed over
	// without matching.
	needs shorthandPatternSet
}

// shorthandTable is the table read once at init: every family's rules in the order they are tried,
// every distinct pattern, and an index from a pattern's text to its indexes.
type shorthandTable struct {
	families [][]shorthandTableRule
	patterns []shorthandTablePattern
	// prefixes maps a capturing pattern's prefix, its trailing dash included, to its pattern index,
	// and exact maps a written-out class to its own.
	prefixes map[string]int
	exact    map[string]int
}

// shorthandPatternTable is built eagerly at init, not on first use. The walk runs every rule's
// listener in parallel across files, and a lazily filled map was once "concurrent map read and map
// write" one run in five; a table read whole at init has no shared mutable state at all.
var shorthandPatternTable = buildShorthandTable(shorthandGroups)

func buildShorthandTable(groups [][]shorthandRule) shorthandTable {
	table := shorthandTable{prefixes: map[string]int{}, exact: map[string]int{}}
	indexOf := map[string]int{}
	for _, group := range groups {
		// Upstream sorts each family by pattern count descending and takes the first match, so a class
		// list holding all four margin sides collapses to `m-1` rather than reporting the two-sided
		// rules that also match. The sort is stable, so rules of equal length keep the table's order.
		ordered := make([]shorthandRule, len(group))
		copy(ordered, group)
		sort.SliceStable(ordered, func(left int, right int) bool {
			return len(ordered[left].Longhands) > len(ordered[right].Longhands)
		})

		family := make([]shorthandTableRule, 0, len(ordered))
		for _, shorthand := range ordered {
			tableRule := shorthandTableRule{shorthands: shorthand.Shorthands}
			for _, longhand := range shorthand.Longhands {
				index, isIndexed := indexOf[longhand]
				if !isIndexed {
					index = len(table.patterns)
					indexOf[longhand] = index
					pattern := readShorthandPattern(longhand)
					table.patterns = append(table.patterns, pattern)
					if pattern.captures {
						table.prefixes[pattern.text] = index
					} else {
						table.exact[pattern.text] = index
					}
				}
				tableRule.longhands = append(tableRule.longhands, index)
				tableRule.needs.add(index)
			}
			family = append(family, tableRule)
		}
		table.families = append(table.families, family)
	}
	if len(table.patterns) > len(shorthandPatternSet{})*64 {
		panic(fmt.Sprintf("enforce-shorthand-classes: %d table patterns outgrow the %d-bit pattern set",
			len(table.patterns), len(shorthandPatternSet{})*64))
	}
	return table
}

// readShorthandPattern reads one longhand as a written-out class or as a dash-ended prefix and `(.*)`.
func readShorthandPattern(longhand string) shorthandTablePattern {
	if prefix, isCapturing := strings.CutSuffix(longhand, "(.*)"); isCapturing &&
		strings.HasSuffix(prefix, "-") && regexp.QuoteMeta(prefix) == prefix {
		return shorthandTablePattern{text: prefix, captures: true}
	}
	if regexp.QuoteMeta(longhand) == longhand {
		return shorthandTablePattern{text: longhand}
	}
	panic("enforce-shorthand-classes: the longhand " + strconv.Quote(longhand) +
		" is neither a written-out class nor a dash-ended prefix followed by (.*), so it cannot be read as plain text")
}

// shorthandMatchesOf is every pattern a base matches, as `^pattern$` would.
//
// A capturing pattern's prefix ends in a dash, so the only prefixes of the base worth looking up end
// at one of its dashes. `(.*)` takes the rest, which may be empty: `w-` matches `w-(.*)` with nothing
// captured, exactly as the expression does. Class names come from splitting on whitespace, so the one
// character `.` refuses, a newline, never reaches here.
func shorthandMatchesOf(base string) []shorthandMatch {
	var matches []shorthandMatch
	if index, isExact := shorthandPatternTable.exact[base]; isExact {
		matches = append(matches, shorthandMatch{pattern: index})
	}
	for position := 0; position < len(base); position++ {
		if base[position] != '-' {
			continue
		}
		if index, isPrefix := shorthandPatternTable.prefixes[base[:position+1]]; isPrefix {
			matches = append(matches, shorthandMatch{pattern: index, capture: base[position+1:]})
		}
	}
	return matches
}

// shorthandVariantGroup is the classes behind one variant prefix that match any pattern, in source
// order, and the patterns they match between them.
type shorthandVariantGroup struct {
	classes []shorthandDissection
	matched shorthandPatternSet
}

// shorthandCollapses finds every collapse available in one class list, at most one per family.
//
// Groups by variant prefix once per list rather than once per rule, in order of each prefix's first
// appearance, and keeps in each group only the classes that match some pattern: a class matching none
// can never be one of a rule's longhands. It still places its prefix in that order, though, and the
// order decides which group reports when two complete one rule: in
// `justify-x-2 [&_button]:items-2 [&_button]:justify-items-2 items-4 justify-items-4` the bare prefix
// comes first because of a class that matches nothing, and upstream reports `place-items-4`. Grouping
// only the matching classes reported `[&_button]:place-items-2` instead; the differential against the
// expressions found it. Most class lists hold no candidate at all and return before a rule is tried.
func shorthandCollapses(classNames []string) []shorthandCollapse {
	var groups []*shorthandVariantGroup
	groupIndex := map[string]int{}
	candidates := false
	for _, className := range classNames {
		class := dissectShorthandClass(className)
		index, isGrouped := groupIndex[class.variants]
		if !isGrouped {
			index = len(groups)
			groupIndex[class.variants] = index
			groups = append(groups, &shorthandVariantGroup{})
		}
		class.matches = shorthandMatchesOf(class.base)
		if len(class.matches) == 0 {
			continue
		}
		candidates = true
		group := groups[index]
		group.classes = append(group.classes, class)
		for _, match := range class.matches {
			group.matched.add(match.pattern)
		}
	}
	if !candidates {
		return nil
	}

	present := make(map[string]bool, len(classNames))
	for _, className := range classNames {
		present[className] = true
	}

	collapses := []shorthandCollapse{}
	for _, family := range shorthandPatternTable.families {
		if collapse, found := firstCollapseInFamily(family, groups, present); found {
			collapses = append(collapses, collapse)
		}
	}
	return collapses
}

// firstCollapseInFamily returns the first rule in one family that matches, longest-first, and for
// that rule the first variant group it matches in.
func firstCollapseInFamily(family []shorthandTableRule, groups []*shorthandVariantGroup,
	present map[string]bool) (shorthandCollapse, bool) {
	for index := range family {
		for _, group := range groups {
			if !group.matched.holds(family[index].needs) {
				continue
			}
			if collapse, found := matchWithinVariantGroup(&family[index], group.classes, present); found {
				return collapse, true
			}
		}
	}
	return shorthandCollapse{}, false
}

// matchWithinVariantGroup matches one rule against classes that already share a variant prefix.
func matchWithinVariantGroup(shorthand *shorthandTableRule, classes []shorthandDissection,
	present map[string]bool) (shorthandCollapse, bool) {
	matched := make([]shorthandDissection, 0, len(shorthand.longhands))
	// The first class matched fixes the value every other longhand must capture too. A written-out
	// longhand captures nothing, and the expression read it as a match of a different shape, which
	// never agreed with a capturing one; no rule mixes the two, and this keeps that answer if one does.
	captured, capturedCaptures, hasCaptured := "", false, false

	for _, pattern := range shorthand.longhands {
		captures := shorthandPatternTable.patterns[pattern].captures
		for _, class := range classes {
			capture, isMatch := class.match(pattern)
			if !isMatch {
				continue
			}
			if !hasCaptured {
				captured, capturedCaptures, hasCaptured = capture, captures, true
			} else if captures != capturedCaptures || capture != captured {
				continue
			}
			matched = append(matched, class)
			break
		}
	}

	if len(matched) != len(shorthand.longhands) {
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

	shorthands := make([]string, 0, len(shorthand.shorthands))
	for _, substitute := range shorthand.shorthands {
		base := substitute
		if capturedCaptures {
			base = strings.ReplaceAll(base, "$1", captured)
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

// buildShorthandClass reassembles a class from the parts its group agreed on.
func buildShorthandClass(variants string, base string, negative bool, important bool, markerEnd bool) string {
	className := base
	if negative {
		className = "-" + className
	}
	if important {
		if markerEnd {
			className += "!"
		} else {
			className = "!" + className
		}
	}
	if variants != "" {
		className = variants + ":" + className
	}
	return className
}

// dissectShorthandClass splits a class into the parts a collapse matches on.
func dissectShorthandClass(className string) shorthandDissection {
	prefix, utility, hasVariants := splitVariantPrefix(className)
	if !hasVariants {
		utility = className
		prefix = ""
	}

	negative := strings.HasPrefix(utility, "-")
	base := strings.TrimPrefix(utility, "-")
	important := strings.HasPrefix(base, "!")
	base = strings.TrimPrefix(base, "!")
	markerEnd := strings.HasSuffix(base, "!")
	base = strings.TrimSuffix(base, "!")

	return shorthandDissection{
		className: className,
		variants:  prefix,
		base:      base,
		negative:  negative,
		important: important || markerEnd,
		markerEnd: markerEnd,
	}
}
