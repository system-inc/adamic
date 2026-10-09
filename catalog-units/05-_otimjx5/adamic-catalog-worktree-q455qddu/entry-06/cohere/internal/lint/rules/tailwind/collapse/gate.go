// Package tailwind decides which class names could possibly collapse, so that almost none of them
// have to be asked about.
//
// Tailwind's collapse logic — the thing that knows `px-4 py-4` is `p-4` and `w-8 h-8` is `size-8` —
// is a signature-equivalence search sitting on the whole Tailwind compiler. It is JavaScript, it is
// eight to ten thousand lines, and it changes every Tailwind minor. cohere is Go. That boundary is
// the one place in this tool where the answer is not "delete the boundary."
//
// This package is the part that does not have to cross it. Two classes can only merge when
// everything except their root already agrees: their variants, their importance, and their value.
// `px-4 py-4` becomes `p-4`, but `px-4 py-2` collapses to nothing and so does `px-4 sm:py-4`. So
// classes are bucketed by everything-but-the-root, and only classes landing in the same bucket are
// worth a real answer.
//
// Measured on the ahra corpus of 2,682 distinct className literals plus 14 planted violations, the
// gate takes 2,695 engine questions down to 735, losing no finding the ungated engine reports and
// inventing none. Planted violations rather than the tree alone, because the tree is very nearly
// canonical and an approach that catches nothing is fast and correct on it.
//
// # Why the boundary turned out not to matter
//
// This package was built to make a decision: whether to port the collapse logic to Go, keep it in a
// Node sidecar, or precompute a table. The measurement settled it, and not in the direction the
// question implied.
//
// Shipping the 469 surviving literals to a Node sidecar and reading the answers back costs 0.4 to
// 1.3ms of transport. The engine costs 3,300 to 4,100ms cold and roughly 500ms warm. The control
// that makes those numbers mean something is the same batch run in-process in Node, with no
// boundary at all: 3,834ms cold, 506 to 595ms warm. Statistically identical. So the crossing
// contributes about a millisecond to a ~3,800ms operation, and the engine costs what it costs
// whether Go is asking or Node is asking itself.
//
// The lever is warmup, not the boundary. A cold run averages 3,909ms and a warm batch on an
// already-started process averages 534ms, so keeping the process resident across runs is worth
// about 3,375ms while transport stays under a tenth of a percent. Porting eight to ten thousand
// lines of collapse logic to Go would buy back the millisecond and inherit a rewrite that changes
// every Tailwind minor.
//
// # Why the reduction is 3.7x and not the 20x on record
//
// An earlier measurement put the survivors near 300. The gate is not weaker; the corpus changed.
// Every survivor is a legitimate over-approximation, and two shapes dominate: 112 collisions from
// `items-center` and `justify-center`, which both declare `center`, and roughly 380 from this
// project's own `content--N`, `background--N`, and `border--N` design tokens. Those parse as
// functional candidates with named values, so `content--2` and `px-2` genuinely share the bucket
// `||2` and must be asked about. 820 of the 2,682 literals now carry such a token.
//
// Teaching the gate that project tokens cannot collapse with Tailwind spacing would recover most of
// the difference and is exactly the kind of under-approximation named below. It would need its own
// fixtures and its own measurement before it could be trusted.
//
// # The direction of error is not symmetric
//
// An over-approximation costs an engine call. An under-approximation loses a finding, silently, on
// a tree that looks clean because the rule stopped looking. Three separate attempts at a cheaper
// gate each lost real findings during the oxlint migration and each measured faster:
//
//   - Splitting the class on dashes to find its root reads `border-l` as root `border` with value
//     `l`, buckets it away from `border-r`, and stops reporting `border-x`.
//   - Giving every static utility its own bucket separates `overflow-x-hidden` from
//     `overflow-y-hidden` and stops reporting `overflow-hidden`.
//   - Requiring two classes to declare the same CSS property family cuts engine calls by an order
//     of magnitude and stops reporting `w-8 h-8` into `size-8`, because a shorthand can absorb
//     properties that share no name — `width` and `height` become `size`, `column-gap` and
//     `row-gap` become `gap`.
//
// Every one was caught by a fixture and none by review. So this package errs toward asking, always,
// and its tests assert the collapses it must not lose rather than the speed it achieves.
package tailwind

import (
	"sort"
	"strings"
)

// CandidateKind distinguishes the two shapes Tailwind parses a class into.
//
// A static utility is a whole name that maps to declarations (`flex`, `truncate`). A functional one
// has a root and optionally a value (`px-4` is root `px` value `4`; `border-l` is root `border-l`
// with no value at all). The distinction is not cosmetic: statics and functionals are bucketed by
// different keys, because what makes two of them mergeable is different.
type CandidateKind string

const (
	// CandidateKindStatic is a utility whose whole name is the identity: `flex`, `truncate`.
	CandidateKindStatic CandidateKind = "Static"
	// CandidateKindFunctional is a utility with a root and an optional value: `px-4`, `border-l`.
	CandidateKindFunctional CandidateKind = "Functional"
	// CandidateKindUnparsed is a class Tailwind could not read at all, including project-specific
	// names. It is never grouped with anything.
	CandidateKindUnparsed CandidateKind = "Unparsed"
)

// ValueKind separates a value the theme named from one the author wrote in brackets.
type ValueKind string

const (
	// ValueKindNone is a functional candidate with no value, such as `border-l`.
	ValueKindNone ValueKind = "None"
	// ValueKindNamed is a value the theme knows by name, such as the `4` in `px-4`.
	ValueKindNamed ValueKind = "Named"
	// ValueKindArbitrary is a bracketed value the author supplied, such as the `1` in `z-[1]`.
	ValueKindArbitrary ValueKind = "Arbitrary"
)

// Value is the part of a functional candidate after its root.
type Value struct {
	Kind ValueKind
	// Value is the literal text: `4` for `px-4`, `3px` for `px-[3px]`.
	Value string
	// Fraction is set when the class carried a slash, such as `1/2` in `w-1/2`. It participates in
	// the key because `w-1/2` and `w-1/3` must not share a bucket.
	Fraction string
}

// Variant is one variant applied to a candidate, such as the `sm` in `sm:px-4`.
//
// Variants are compared as a whole ordered list, because `sm:hover:px-4` and `hover:sm:px-4` are
// different selectors and must not be treated as one bucket.
type Variant struct {
	Kind string
	Root string
	// Value carries the payload of a variant that has one, such as the `[&>*]` of an arbitrary
	// variant. Two variants that differ only here are still different variants.
	Value string
}

// Candidate is one reading of a class name.
//
// One class can have several readings: Tailwind parses `border-b` as both root `border-b` with no
// value and root `border` with value `b`, and returns both. 185 of the 1,236 classes in the real
// corpus do this, and their readings produce *different* bucket keys, so which one is used is a
// real decision rather than a detail.
//
// It is the same ambiguity that lost `border-x` to a hand-written splitter, and the JavaScript gate
// this package ports resolves it by silently taking the first reading without recording why. That
// made it worth checking rather than inheriting: over all 4,375 candidate pairs in the corpus, of
// which 75 collapse, keying on the first reading loses none, and the order is stable across calls.
// So the choice is safe — but it was verified against the engine, not assumed from the source it
// was ported from. See DeclaredValues for the static case.
type Candidate struct {
	Kind      CandidateKind
	Root      string
	Value     Value
	Important bool
	Variants  []Variant
}

// Parser is what turns a class name into candidates.
//
// This is an interface rather than an implementation because parsing a class correctly requires the
// theme: whether `w-full` is a static utility or a functional one with value `full` depends on what
// the design system declares. Deciding that in Go without the theme is precisely the hand-written
// splitter that lost `border-x`. So the gate owns the bucketing logic, which is pure structure, and
// takes the parse as input from whatever can answer it correctly.
type Parser interface {
	// ParseCandidates returns every reading of a class, in Tailwind's own order. An empty result
	// means the class is unparseable, which puts it in a bucket of its own.
	ParseCandidates(className string) []Candidate
	// DeclaredValues returns the CSS values a class declares, which is how static utilities are
	// bucketed. `overflow-x-hidden` and `overflow-y-hidden` both declare `hidden` and can merge;
	// `overflow-x-hidden` and `overflow-y-auto` do not and cannot.
	//
	// Values rather than properties, and this is the load-bearing half: keying statics on their
	// declared properties separates the two `overflow` axes and loses `overflow-hidden`.
	DeclaredValues(className string) []string
}

// Gate buckets class names and answers which of them could possibly collapse.
//
// It memoizes per class name, so a class appearing in three hundred literals is parsed once. The
// memo is the reason the gate is cheap on a real tree rather than only on a fixture.
type Gate struct {
	parser Parser

	groupKeyByClassName map[string]string
	// singlesThatRewrite records classes an authority has confirmed rewrite on their own, so the
	// gate can report a literal as needing an answer without re-deriving why.
	singlesThatRewrite map[string]bool
}

// NewGate returns a gate that reads candidates through the given parser.
func NewGate(parser Parser) *Gate {
	return &Gate{
		parser:              parser,
		groupKeyByClassName: map[string]string{},
		singlesThatRewrite:  map[string]bool{},
	}
}

// unparsedKeyPrefix and staticKeyPrefix keep the three key namespaces from colliding.
//
// An unparseable class must never share a bucket with anything, including another unparseable one,
// so its key carries its own name. A static keyed on declared values must never collide with a
// functional keyed on an empty value, which is why the prefixes exist at all: `border-l` keys as
// `||` and a static declaring nothing would otherwise key the same way and drag them together.
const (
	unparsedKeyPrefix = "unparsed:"
	staticKeyPrefix   = "static:"
	uncompiledPrefix  = "uncompiled:"
)

// GroupKey returns the bucket key for a class: everything about it except its root.
//
// Two classes can only merge when this key matches. The key is derived from a real parse rather than
// from string surgery on the class name, which is the difference between reporting `border-x` and
// silently not.
func (g *Gate) GroupKey(className string) string {
	if cached, isCached := g.groupKeyByClassName[className]; isCached {
		return cached
	}

	key := g.computeGroupKey(className)
	g.groupKeyByClassName[className] = key
	return key
}

func (g *Gate) computeGroupKey(className string) string {
	candidates := g.parser.ParseCandidates(className)
	if len(candidates) == 0 {
		return unparsedKeyPrefix + className
	}

	// The first reading, deliberately. Tailwind returns readings in a stable order and the first is
	// the one that preserves every collapse the engine finds — measured over all 4,375 candidate
	// pairs in the real corpus, of which 75 collapse, with zero lost to this choice.
	candidate := candidates[0]

	switch candidate.Kind {
	case CandidateKindUnparsed:
		return unparsedKeyPrefix + className

	case CandidateKindStatic:
		// Statics bucket by what they declare, not by their name and not by their properties.
		return staticKeyPrefix + g.declaredValuesKey(className)

	default:
		return variantsKey(candidate.Variants) + "|" + importanceKey(candidate.Important) + "|" + valueKey(candidate.Value)
	}
}

// declaredValuesKey is the sorted, joined CSS values a static declares.
//
// A class that declares nothing gets a key carrying its own name rather than the empty string.
// Without that, every class the compiler could not read would collide under one key and send each
// other to the engine forever.
func (g *Gate) declaredValuesKey(className string) string {
	values := g.parser.DeclaredValues(className)
	if len(values) == 0 {
		return uncompiledPrefix + className
	}

	sorted := make([]string, len(values))
	copy(sorted, values)
	sort.Strings(sorted)
	return strings.Join(sorted, "|")
}

// variantsKey serializes the variant list in order.
//
// Order is preserved rather than sorted: `sm:hover:px-4` and `hover:sm:px-4` produce different CSS
// and must not share a bucket.
func variantsKey(variants []Variant) string {
	if len(variants) == 0 {
		return ""
	}

	parts := make([]string, 0, len(variants))
	for _, variant := range variants {
		parts = append(parts, variant.Kind+":"+variant.Root+":"+variant.Value)
	}
	return strings.Join(parts, "&")
}

func importanceKey(important bool) string {
	if important {
		return "!"
	}
	return ""
}

// valueKey renders a value so that named, arbitrary, and absent values never collide.
//
// The brackets around an arbitrary value are part of the key rather than decoration: they are also
// what CouldRewriteAlone reads to decide that `z-[1]` is worth asking about while `px-4` is not.
func valueKey(value Value) string {
	switch value.Kind {
	case ValueKindNone:
		return ""

	case ValueKindArbitrary:
		return "[" + value.Value + "]"

	default:
		if value.Fraction != "" {
			return value.Value + "/" + value.Fraction
		}
		return value.Value
	}
}

// CouldRewriteAlone reports whether a class might have a shorter spelling with no second class
// involved.
//
// Two shapes do, and both are visible without asking the engine. An arbitrary value can have a
// theme name: `z-[1]` is `z-1`, `w-[100%]` is `w-full`. A static can be renamed outright, which
// Tailwind keeps in a table: `order-none` is `order-0`, `break-words` is `wrap-break-word`.
//
// Functional candidates with a named value — `px-4`, `rounded-md`, `text-sm`, the overwhelming
// majority — never rewrite alone, because a named value is already the theme's own spelling. Those
// are the ones worth never asking about.
func (g *Gate) CouldRewriteAlone(className string) bool {
	key := g.GroupKey(className)

	if strings.HasPrefix(key, staticKeyPrefix) {
		return true
	}
	if strings.HasPrefix(key, unparsedKeyPrefix) {
		return false
	}
	return strings.HasSuffix(key, "]")
}

// RecordSingleRewrites tells the gate which classes an authority confirmed rewrite on their own.
//
// The gate can see that `z-[1]` is worth asking about; only the engine can say that it is actually
// `z-1`. Feeding the verdict back means a class is asked about once per process rather than once per
// literal that contains it.
func (g *Gate) RecordSingleRewrites(classNames []string) {
	for _, className := range classNames {
		g.singlesThatRewrite[className] = true
	}
}

// SingleClassesWorthChecking returns the classes here that should be asked about individually,
// excluding any already asked about.
//
// Returned rather than acted on, so the caller keeps its own record of what it has paid for and the
// gate stays free of the boundary it exists to avoid.
func (g *Gate) SingleClassesWorthChecking(classNames []string, alreadyChecked map[string]bool) []string {
	worthChecking := make([]string, 0, len(classNames))
	for _, className := range classNames {
		if alreadyChecked[className] {
			continue
		}
		if g.CouldRewriteAlone(className) {
			worthChecking = append(worthChecking, className)
		}
	}
	return worthChecking
}

// CannotCollapse reports that no two of these classes share a bucket and none rewrites alone.
//
// When this is true the caller may skip the engine entirely. When it is false the caller must ask:
// the gate claims only that a collapse is possible, never that one exists.
//
// The asymmetry is the whole design. A false here costs one engine call. A true here that should
// have been false loses a finding on a tree that then looks clean, which is the failure this package
// exists to prevent, so every doubtful case resolves to false.
func (g *Gate) CannotCollapse(classNames []string) bool {
	for _, className := range classNames {
		if g.singlesThatRewrite[className] {
			return false
		}
	}

	seenGroups := make(map[string]bool, len(classNames))
	for _, className := range classNames {
		key := g.GroupKey(className)
		if seenGroups[key] {
			return false
		}
		seenGroups[key] = true
	}
	return true
}

// NeedsEngine is the gate's whole verdict for one literal: must this set of classes be asked about?
//
// It runs the two tests in the order that matters. The bucket test is free and the singles test is
// not, because on the first class it touches, the engine builds signature tables for the entire
// utility registry and costs about 1.8 seconds. Running the free test first means a file holding no
// two classes in a shared bucket and no rewritable single never pays that at all.
//
// alreadyChecked carries singles the caller has resolved in this process; pass nil on a first call.
func (g *Gate) NeedsEngine(classNames []string, alreadyChecked map[string]bool) bool {
	if alreadyChecked == nil {
		alreadyChecked = map[string]bool{}
	}

	distinct := distinctPreservingOrder(classNames)
	if len(distinct) == 0 {
		return false
	}

	if g.CannotCollapse(distinct) && len(g.SingleClassesWorthChecking(distinct, alreadyChecked)) == 0 {
		return false
	}
	return true
}

// distinctPreservingOrder removes repeats without sorting.
//
// A repeated class is a different rule's finding, and passing it twice would confuse any
// leave-one-out attribution the caller does with the engine's answer. Order is preserved so that a
// reported set reads in the order the author wrote it.
func distinctPreservingOrder(classNames []string) []string {
	seen := make(map[string]bool, len(classNames))
	distinct := make([]string, 0, len(classNames))
	for _, className := range classNames {
		if className == "" || seen[className] {
			continue
		}
		seen[className] = true
		distinct = append(distinct, className)
	}
	return distinct
}
