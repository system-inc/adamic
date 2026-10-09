// The live candidate split: a class broken into the pieces that decide whether it can merge, using
// this repository's own roots.
//
// This is `enforce-canonical-classes`'s half of the seam #3r6cxrb crosses. Until this file, the
// split walked `RootDeclaredProperties` looking for the longest prefix that was followed by a dash,
// and that map is generated from one repository. After it, the split is `ParseCandidate`'s, which
// asks the design system where a root ends.
//
// # Why the walk had to go, and it is not that it was slow
//
// `splitCandidate`'s own comment recorded the failure it had already survived three times: a
// dash-splitter reads `border-l` as root `border` with value `l`, buckets it away from `border-r`,
// and silently stops reporting `border-x`. The fix was to match the longest registered root instead,
// which is right in outline and still an approximation, because "registered" meant a table.
//
// candidate.go states the general form of the problem: where a class's root ends is not a property
// of its string. `border-b` reads both as the functional root `border-b` with no value and as
// `border` with the named value `b`, and which roots exist is a question about this repository's
// `@utility` blocks as much as about the framework. So the prefix walk answers correctly exactly
// when the table happens to describe the repository in front of it.
//
// Measured: 26 of `KnownRoots`'s 315 entries are roots the ahra repository declares in its own
// stylesheet, carried under a `Source: Tailwind 4.3.3` header. On a repository that declares
// different ones, the walk finds roots that do not exist there and misses the ones that do.
//
// # What the gate already promised, and what this keeps
//
// `internal/lint/rules/tailwind/collapse/gate.go:227` takes `candidates[0]` with a measurement attached: the first
// reading preserves every collapse the engine finds, over all 4,375 candidate pairs in the real
// corpus, of which 75 collapse, with zero lost to that choice. This file takes the same first
// reading for the same reason, so the property that was measured before the parser existed is the
// property the parser is now used under.
//
// # The one thing that does not move
//
// `CollapseFamilies` stays generated, and that is a measurement rather than an omission. A collapse
// family is a fact about which framework roots merge into a third when everything else agrees, and a
// repository's `@utility` block adds roots without creating merge relationships among framework
// ones. Verified by generating against both corpus repositories: 44 families each, zero on either
// side alone. That check is recorded on #3r6cxrb rather than inferred here.
package tailwind

import (
	"strings"

	tailwindengine "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse"
)

// splitCandidateIn breaks a class into prefix, root, value and importance, against a design system.
//
// The design system is a parameter rather than a package-level table because that is the whole
// change: two repositories declaring different `@utility` roots must split the same class
// differently, and a table cannot do that.
//
// Returns false for a class this design system cannot read at all, which is the honest answer and
// not a failure: `mergeOnce` skips it, so an unparseable class simply does not participate in a
// collapse. That is the same outcome the prefix walk produced when no root matched, reached by
// asking rather than by guessing.
func splitCandidateIn(
	className string,
	system *tailwindengine.LoadedDesignSystem,
) (candidateParts, bool) {
	if system == nil {
		return candidateParts{}, false
	}

	parsed := tailwindengine.ParseCandidate(className, system)
	if len(parsed) == 0 {
		return candidateParts{}, false
	}

	// The first reading, matching `gate.go`'s own choice and the measurement attached to it. A
	// later reading is a root the engine does not compile the class as, so bucketing by it would
	// group classes that never merge.
	candidate := parsed[0]

	// A static utility has no root-plus-value structure to merge on. `flex` and `items-center`
	// reach here and are correctly excluded: the collapse families are relationships between
	// functional roots, and a static has no value for the merge precondition to compare.
	if candidate.Kind != tailwindengine.ParsedCandidateKindFunctional {
		return candidateParts{}, false
	}

	// The prefix is rebuilt from the source text rather than from the parsed variants, because it is
	// used verbatim to reassemble the output class. `dissectClass` cuts at the last colon, which is
	// exactly the text `rebuildClass` needs to put back, and reprinting parsed variants would have
	// to reproduce their spelling — including arbitrary ones like `[&_svg]:` — from a structure that
	// deliberately decoded them.
	prefix, _, _ := dissectClass(className)

	return candidateParts{
		ClassName:   className,
		Prefix:      prefix,
		Root:        candidate.Root,
		Value:       candidateValueText(candidate),
		SourceValue: sourceValueAfterRoot(className, candidate.Root),
		Important:   candidate.Important,
	}, true
}

// sourceValueAfterRoot is the value exactly as the author spelled it, taken from the class text.
//
// The parsed value is decoded and the decoded form is not always writable: `grid-cols-[1fr_auto]`
// decodes to `1fr auto` and `ring-(--x)` to `[var(--x)]`. A rewrite printed from either would name a
// class nobody can write, so the text between the root and the end of the class is recovered here
// and `rebuildClass` prints that instead. See candidateParts.SourceValue.
//
// Empty when the source and the decoding coincide, which is every ordinary class and lets
// `printedValue` fall through to the decoded form for them.
func sourceValueAfterRoot(className string, root string) string {
	_, base, _ := dissectClass(className)
	if !strings.HasPrefix(base, root) {
		// The root came from the parse and the base from a string split, and the two disagree on a
		// class whose variants confused `dissectClass`. Reporting nothing here is the safe half:
		// `printedValue` falls back to the decoded value, which is right whenever it is writable.
		return ""
	}

	value := strings.TrimPrefix(base[len(root):], "-")
	if value == "" {
		return ""
	}
	return value
}

// candidateValueText is the value half of a functional candidate, spelled as it merges.
//
// The merge precondition compares this for equality between two classes and `rebuildClass` prints it
// back after the new root, so it has to carry everything the class said after its root and nothing
// else. Two classes that differ anywhere in here do not merge, and a class rebuilt from a value that
// dropped a piece is a different class from the one the author wrote.
//
// # The two pieces a value alone does not carry, both found by differential
//
// A first version returned `Value.Value` and was wrong twice, on 88 real corpus classes:
//
//   - **The modifier is a sibling field, not part of the value.** `bg-black/20` parses as root `bg`,
//     value `black`, modifier `20`. Returning the value alone made `bg-black/20` and `bg-black/60`
//     compare equal on the merge precondition, which is exactly the permissive direction
//     `enforce-canonical-classes`'s own comment warns reports correct code. 84 of the 88 were this.
//
//     A NAMED modifier is also mirrored into `Fraction`, so the fraction branch below happens to
//     cover it and the modifier branch is only reachable for an ARBITRARY one:
//     `bg-emerald-500/[0.07]` parses with `Fraction` empty and modifier `[0.07]`. That is worth
//     stating because it is what a mutation found — removing the modifier branch left every named
//     modifier correct and silently dropped the arbitrary ones, which is one corpus class and would
//     have looked like an unused branch to anyone reading it.
//
//   - **A fraction is carried beside the value rather than inside it.** `-translate-x-1/2` parses as
//     value `1` with `Fraction` set to `1/2`, because the slash is genuinely ambiguous and the
//     parser refuses to guess: `w-1/2` is a fraction and `bg-red-500/50` is a modifier, spelled
//     identically. `ParsedValue.Fraction`'s own comment says so. Returning `1` would rebuild
//     `-translate-x-1/2` as `-translate-1` and merge it with anything else valued `1`.
//
// The fraction is preferred when present because it is the whole `value/modifier` text, so taking it
// and the modifier both would print the slash twice.
//
// Removing the fraction branch is an EQUIVALENT mutation on this corpus and is left in anyway.
// Measured: over every corpus class carrying a fraction, reconstructing `value + "/" + modifier`
// produces the fraction text exactly, 0 differing of all of them, because a named modifier is
// mirrored into Fraction by the parser. The branch stays because the equivalence is the parser's
// current behaviour rather than a documented guarantee, and `ParsedValue.Fraction` is the field that
// says what the text was; reconstructing it from two other fields would be relying on a coincidence
// this comment had to measure to discover.
func candidateValueText(candidate tailwindengine.ParsedCandidate) string {
	if candidate.Value == nil {
		// A root with nothing after it, which is what `border-l` is. `rebuildClass` appends no dash
		// for an empty value, so this reassembles exactly.
		return ""
	}

	// The whole `value/modifier` text, which already includes the slash and everything after it.
	if candidate.Value.Fraction != "" {
		return candidate.Value.Fraction
	}

	value := candidate.Value.Value
	// Arbitrary and named spellings are kept apart by the brackets, because `w-4` and `w-[4]` are
	// different classes and an arbitrary value's decoded text can coincide with a named one's.
	if candidate.Value.Kind == tailwindengine.ParsedValueKindArbitrary {
		value = "[" + value + "]"
	}

	if candidate.Modifier == nil {
		return value
	}
	if candidate.Modifier.Kind == tailwindengine.ParsedModifierKindArbitrary {
		return value + "/[" + candidate.Modifier.Value + "]"
	}
	return value + "/" + candidate.Modifier.Value
}
