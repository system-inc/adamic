package core

import (
	"strings"

	esregexp "github.com/system-inc/cohere/internal/lint/ecmascript/regexp"
	"github.com/system-inc/cohere/internal/lint/ecmascript/text"
)

// noRestrictedImportsGlob is one compiled gitignore-style pattern from a `group` entry.
//
// This exists because upstream's `no-restricted-imports` matches a module specifier with the
// `ignore` npm package -- literally `ignore({allowRelativePaths: true, ignorecase: !caseSensitive})
// .add(group).ignores(importSource)` -- and there is no such matcher anywhere in this tree.
// `filepath.Match`, which `react/no-danger` uses in place of minimatch, is not a substitute: it has
// no negation, no `**`, and a bare `foo` under it does not match `import1/private/foo`, which is the
// single most common shape in upstream's corpus.
//
// # This is gitignore's grammar, not a glob's, and the differences all decide real cases
//
// Every rule below was measured against the installed `ignore` package over 864 (pattern, specifier)
// pairs built from upstream's own corpus, not read off its documentation. The table lives in
// `no_restricted_imports_matcher_test.go`.
//
//	foo             matches `foo`, `foo/bar` and `import1/private/foo`   -- a bare name is
//	                anchored at any segment boundary, in either direction
//	foo/bar         matches `foo/bar` only                               -- a pattern containing a
//	                slash anchors at the start
//	foo/*           matches `foo/bar`, not `foo/a/b`                     -- `*` stops at `/`
//	**              matches everything
//	**/my/relative-module   matches `../../my/relative-module`
//	!foo/bar        un-matches, and LAST match wins
//	\#foo           matches `#foo/bar`; an unescaped leading `#` would be a comment
//	FOO             matches `foo` unless caseSensitive is set
//
// The negation rule is the one that cannot be approximated. `["foo/*", "!foo/bar"]` matches
// `foo/baz` and not `foo/bar`, and reversing the entries changes the answer, so the patterns have to
// be evaluated in order with the last decision winning rather than as a set of alternatives.
type noRestrictedImportsGlob struct {
	// negated marks a `!` pattern, which un-matches rather than matches.
	negated bool

	// expression is the pattern translated to a regular expression, compiled through the same
	// engine every other user-supplied pattern in this tree goes through.
	expression *esregexp.RegExp
}

// noRestrictedImportsMatcher is a compiled `group`, evaluated in order.
type noRestrictedImportsMatcher struct {
	globs []noRestrictedImportsGlob
}

// noRestrictedImportsCompileGroup compiles a `group` array into a matcher.
//
// A pattern that cannot be translated is DROPPED rather than treated as matching nothing quietly or
// everything quietly, and the boolean says whether every pattern survived. The caller refuses the
// configuration when it did not, because a silently-dropped restriction is a rule that reports less
// than its author asked for while looking healthy, which is the failure this whole tree is written
// against.
func noRestrictedImportsCompileGroup(group []string, caseSensitive bool) (*noRestrictedImportsMatcher, bool) {
	matcher := &noRestrictedImportsMatcher{}
	for _, pattern := range group {
		compiled, ok := noRestrictedImportsCompileGlob(pattern, caseSensitive)
		if !ok {
			return nil, false
		}
		if compiled == nil {
			// A comment or an empty line. `ignore` skips these rather than failing, so a group
			// carrying one is a legal configuration that simply has one fewer rule in it.
			continue
		}
		matcher.globs = append(matcher.globs, *compiled)
	}
	return matcher, true
}

// Ignores answers `matcher.ignores(path)`.
//
// Every pattern is tested and the LAST one that matches decides, which is how a negation undoes an
// earlier match. Testing in order and returning on the first hit would make `["foo/*", "!foo/bar"]`
// report on `foo/bar`, which is the case upstream's corpus tests directly.
func (m *noRestrictedImportsMatcher) Ignores(path string) bool {
	if m == nil {
		return false
	}
	// Neither side is `./`-normalised, and that is a measured decision rather than an omission. The
	// natural assumption is that `ignore` strips the prefix so `./types` and `types` are one rule;
	// it does not, and every case below was measured against the installed package:
	//
	//	["types"]     against "./types"     true    -- absorbed by "any segment boundary"
	//	["foo/bar"]   against "./foo/bar"   FALSE   -- an anchored pattern anchors at the real start
	//	["foo/*"]     against "./foo/bar"   FALSE
	//	["./types"]   against "./types"     true    -- matches itself
	//	["./types"]   against "types"       FALSE   -- a pattern with a slash is anchored, and
	//	                                               `./types` has one
	//
	// So the prefix is handled entirely by the anchoring rule, in both directions. Trimming it from
	// the candidate made the second and third true, and trimming it from the pattern made the last
	// one true; each was a rule reporting on imports its author did not restrict. Both trims were
	// written here and both were removed after measurement.
	candidate := path

	ignored := false
	for _, glob := range m.globs {
		if glob.expression.Test(candidate) {
			ignored = !glob.negated
		}
	}
	return ignored
}

// noRestrictedImportsCompileGlob translates one gitignore pattern into a regular expression.
//
// Returns (nil, true) for a pattern that carries no rule -- a comment or an empty line -- and
// (nil, false) for one that could not be translated at all.
//
// The translation is deliberately written out rather than assembled from a glob library, because the
// three rules that decide the corpus are exactly the three a glob library does not have: where a
// pattern anchors, whether `*` crosses a separator, and what `**` does at each of the three
// positions it can occupy.
func noRestrictedImportsCompileGlob(pattern string, caseSensitive bool) (*noRestrictedImportsGlob, bool) {
	trimmed := pattern

	// `ignore` skips blank lines and comments. A `#` can be escaped to mean a literal one, which is
	// the `\#foo` case in the corpus, matching `#foo/bar`.
	if text.TrimWhitespace(trimmed) == "" {
		return nil, true
	}
	if strings.HasPrefix(trimmed, "#") {
		return nil, true
	}

	negated := false
	if strings.HasPrefix(trimmed, "!") {
		negated = true
		trimmed = trimmed[1:]
	}

	// A leading `\#` or `\!` needs no arm of its own. The per-segment translator below already
	// reads a backslash as "the next character is a literal", so `\#foo` becomes a literal `#foo`
	// there. An explicit strip was written here first and removed after a mutation deleting it
	// survived every fixture: it was dead code, and a dead branch in a matcher reads as a rule
	// somebody relies on.
	// A trailing slash restricts a gitignore pattern to directories. A module specifier is not a
	// directory, and the corpus has no such pattern; dropping the slash and matching as a prefix is
	// what `ignore` does with `allowRelativePaths`.
	trimmed = strings.TrimSuffix(trimmed, "/")
	if trimmed == "" {
		return nil, true
	}

	// Where the pattern anchors. A slash anywhere but at the end means the pattern is relative to
	// the root; without one it matches at any segment boundary. This is gitignore's own rule and it
	// is what makes `foo` match `import1/private/foo` while `foo/bar` matches only `foo/bar`.
	anchored := strings.Contains(trimmed, "/")
	trimmed = strings.TrimPrefix(trimmed, "/")

	var expression strings.Builder
	if anchored {
		expression.WriteString("^")
	} else {
		// Any segment boundary, which includes the start of the string and a leading `../`.
		expression.WriteString("^(?:.*/)?")
	}

	segments := strings.Split(trimmed, "/")
	for index, segment := range segments {
		if index > 0 {
			expression.WriteString("/")
		}
		if segment == "**" {
			// `**` spans any number of segments including none. Written to absorb the separator it
			// was about to be followed by, so `**/my/relative-module` matches `my/relative-module`
			// as well as `../../my/relative-module`.
			if index == len(segments)-1 {
				expression.WriteString(".*")
			} else {
				expression.WriteString("(?:.*/)?")
				// The separator for the next segment was just written into the group above.
				segments[index] = "**"
				if index+1 < len(segments) {
					rest, ok := noRestrictedImportsCompileSegments(segments[index+1:])
					if !ok {
						return nil, false
					}
					expression.WriteString(rest)
					break
				}
			}
			continue
		}
		translated, ok := noRestrictedImportsTranslateSegment(segment)
		if !ok {
			return nil, false
		}
		expression.WriteString(translated)
		// An EMPTY path segment never matches, even against a segment that is nothing but `*`.
		// `*` itself is zero-or-more -- `f*` matches bare `f`, measured -- so the constraint is on
		// the segment rather than on the wildcard, and it has to be written separately for that
		// reason. Without it `foo/*` matches `foo/` and `foo//bar`, which was 6 of 540 disagreements
		// with the installed `ignore` on an adversarial table.
		expression.WriteString("(?<=[^/])")
	}

	// A matched pattern also matches everything under it: `foo` matches `foo/bar`, and
	// `import1/private/*` matches `import1/private/bar`. gitignore expresses that as "a matched
	// directory ignores its contents".
	expression.WriteString("(?:/.*)?$")

	flags := "u"
	if !caseSensitive {
		flags = "iu"
	}
	compiled, err := esregexp.Compile(expression.String(), flags)
	if err != nil {
		return nil, false
	}
	return &noRestrictedImportsGlob{negated: negated, expression: compiled}, true
}

// noRestrictedImportsCompileSegments translates the remainder of a pattern after a `**`.
func noRestrictedImportsCompileSegments(segments []string) (string, bool) {
	var expression strings.Builder
	for index, segment := range segments {
		if index > 0 {
			expression.WriteString("/")
		}
		if segment == "**" {
			expression.WriteString(".*")
			continue
		}
		translated, ok := noRestrictedImportsTranslateSegment(segment)
		if !ok {
			return "", false
		}
		expression.WriteString(translated)
	}
	return expression.String(), true
}

// noRestrictedImportsTranslateSegment translates one path segment's wildcards.
//
// `*` matches within a segment and does not cross a separator, `?` matches one character, and a
// character class passes through. Everything else is escaped, which is what keeps `foo-bar-baz` from
// being read as a regular expression by a project that wrote a literal module name.
func noRestrictedImportsTranslateSegment(segment string) (string, bool) {
	var expression strings.Builder
	runes := []rune(segment)
	for index := 0; index < len(runes); index++ {
		character := runes[index]
		switch character {
		case '*':
			expression.WriteString("[^/]*")
		case '?':
			expression.WriteString("[^/]")
		case '[':
			closing := -1
			for scan := index + 1; scan < len(runes); scan++ {
				if runes[scan] == ']' && scan > index+1 {
					closing = scan
					break
				}
			}
			if closing == -1 {
				// An unclosed bracket is a literal one in gitignore.
				expression.WriteString(`\[`)
				continue
			}
			expression.WriteString(string(runes[index : closing+1]))
			index = closing
		case '\\':
			if index+1 < len(runes) {
				index++
				expression.WriteString(noRestrictedImportsEscapeRune(runes[index]))
				continue
			}
			expression.WriteString(`\\`)
		default:
			expression.WriteString(noRestrictedImportsEscapeRune(character))
		}
	}
	return expression.String(), true
}

// noRestrictedImportsEscapeRune escapes one character for use as a regular-expression literal.
func noRestrictedImportsEscapeRune(character rune) string {
	if strings.ContainsRune(`.+^$(){}|\[]*?/`, character) {
		return `\` + string(character)
	}
	return string(character)
}
