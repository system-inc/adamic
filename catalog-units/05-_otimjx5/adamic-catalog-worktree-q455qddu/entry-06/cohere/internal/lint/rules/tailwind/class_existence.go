// Live class existence: whether a class is one *this repository's* design system defines.
//
// This is `no-unknown-classes`'s half of the seam #3r6cxrb crosses. Until this file, existence was
// answered from `KnownRoots` and `KnownStatics` in collapse_table.go, two tables generated from one
// repository under a header reading `Source: Tailwind 4.3.3`. After it, the question is put to the
// design system in front of the linter.
//
// # The defect, measured on a design system that shares nothing with ours
//
// `internal/lint/rules/tailwind/tools/generate_descriptor_base/testdata/independent_theme.css` exists because the two corpus
// repositories cannot answer a framework-versus-repository question: they both vendor
// `libraries/structure`, so a fact they agree on may be a fact about the submodule. It declares
// `@utility synthetic-static` and `@utility synthetic-fn-*`, names no framework contains.
//
// Loaded as a design system and asked about its own classes, the shipped table answered:
//
//	synthetic-static      table says unknown    the repository declares it
//	synthetic-fn-small    table says unknown    the repository declares it
//	markdown-content      table says known      no repository here declares it
//
// The first two are the failure the task names: a repository's own `@utility` classes read as
// unknown, so the rule reports correct code on every repository except the one the table was
// generated from. The third is the same defect seen from the other side — 27 of `KnownStatics`'s
// 895 entries and 20 of `KnownRoots`'s 315 are names ahra or www-connected-app declared, so the
// table vouches for them everywhere.
//
// # Why ParseCandidate is the oracle and a table cannot be
//
// Where a class's root ends is not a property of its string. `border-b` reads both as the root
// `border-b` with no value and as `border` with the named value `b`, and which roots exist is a
// question about this repository's `@utility` blocks as much as about the framework. candidate.go's
// file comment states this as the reason the parser takes a design system rather than a table, and
// existence is the same question asked once: a class exists when the design system in front of us
// can read it as at least one candidate.
//
// That replaces the hand-written prefix walk this file's predecessor did over `KnownRoots`. The walk
// had to re-derive the value boundary itself — `remainder[0] == '-' || '/' || '[' || '('` — which is
// an approximation of what `parseCandidate` does properly, and the same approximation in a different
// guise is what lost `border-x` during the migration.
//
// # What does not change, deliberately
//
// The gap `TestKnownRootWithUnknownValueIsNotReported` pins stays open. `text-huge` parses as root
// `text` with the named value `huge` and compiles to nothing; the parser accepts it because the
// parser's job is structure rather than resolution. Closing it needs the theme's per-root value
// scales, and the error direction is the safe one: this rule under-reports rather than flagging
// working classes, which is how a rule survives contact with a codebase.
package tailwind

import (
	"regexp"

	tailwindengine "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse"
)

// alwaysKnownClasses are markers Tailwind treats specially rather than as utilities.
//
// `group` and `peer` generate no CSS of their own: they exist to be referenced by `group-hover:` and
// `peer-checked:` variants on other elements. `ParseCandidate` returns no readings for them for
// exactly that reason, so they must be recognised before it is asked or the two most common classes
// in any real codebase get reported. Upstream ignores them by explicit regular expression, and this
// is that expression.
var alwaysKnownClasses = regexp.MustCompile(`^(group|peer)(/\S*)?$`)

// arbitraryPropertyPattern matches a whole-property escape hatch such as `[font:inherit]`.
//
// These have no root to look up: the author wrote the CSS declaration directly. Tailwind accepts any
// syntactically valid one, so there is nothing to check against and reporting them would flag a
// deliberate feature.
var arbitraryPropertyPattern = regexp.MustCompile(`^\[[^\]]+:[^\]]*\]$`)

// classExistsIn reports whether a design system defines a class, by asking it to read one.
//
// The design system is a parameter rather than a package-level table because that is the whole
// change: two repositories declaring different `@utility` blocks must get different answers, and a
// table cannot give them.
//
// A nil system answers true for everything. That is not a convenience: the caller must decline
// before reaching here, and this is the failing-safe direction if one ever does not. A rule that
// cannot load a design system reporting every class in the tree as unknown would bury a real
// codebase in findings, which is strictly worse than the silence design_system.go's failing-safe
// note argues for.
func classExistsIn(className string, system *tailwindengine.LoadedDesignSystem) bool {
	if system == nil {
		return true
	}

	// Checked before anything splits the class, because `dissectClass` cuts at the last colon and an
	// arbitrary property carries one inside its own brackets: `[font:inherit]` would become a
	// variant `[font:` and a base `inherit]`. The brackets are the class, not a prefix.
	if arbitraryPropertyPattern.MatchString(className) {
		return true
	}

	_, base, _ := dissectClass(className)
	if base == "" {
		return true
	}

	if alwaysKnownClasses.MatchString(base) {
		return true
	}
	if arbitraryPropertyPattern.MatchString(base) {
		return true
	}

	// The whole class rather than its base, so the variants are parsed by the same design system
	// that answers for the utility. A variant this repository never declared makes the class
	// unreadable, which is correct: `notavariant:flex` compiles to nothing.
	// Any candidate that resolves, not only the first. `shadow--0` is the repository's `shadow--*`
	// first and the framework's `shadow` second, and only the second compiles.
	return anyCandidateResolves(tailwindengine.ParseCandidate(className, system), system)
}

// anyCandidateResolves reports whether some reading of a class resolves to a value the design system
// has, which is when the engine emits CSS for it.
func anyCandidateResolves(candidates []tailwindengine.ParsedCandidate, system *tailwindengine.LoadedDesignSystem) bool {
	for index := range candidates {
		if tailwindengine.ClassValueResolvesIn(&candidates[index], system) {
			return true
		}
	}
	return false
}
