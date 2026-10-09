// Package directives is the grammar of suppression comments, and the short list of rules whose
// findings are about those comments.
//
// It lives on the shared shelf rather than inside `internal/lint/suppression` because two kinds of
// reader need the same answer to "is this comment a directive": the suppression index, which decides
// what a directive silences, and a rule that reports on directives themselves
// (`@eslint-community/eslint-comments/require-description`). Rule packages may only import the
// shelf, so the grammar sits here and the suppression index imports it, which keeps the two readers
// agreeing by construction rather than by a copied list.
package directives

import (
	"strings"
	"unicode/utf8"

	"github.com/system-inc/cohere/internal/lint/ecmascript/text"
)

// Directive spellings cohere honors, with identical grammar.
//
// `cohere-disable` is what new code writes. `eslint-disable` keeps working because 387 of them
// already exist across the codebase this tool gates, written by the linter this replaces, and
// rewriting that corpus is not worth what it would cost. Kirk decided this directly: both forms,
// all three shapes, no codemod, no migration, no flag day. `oxlint-disable` is honored on the same
// reasoning, since oxlint is the gate actually being replaced.
//
// `verify-disable` is honored too, and for a smaller reason than the other two: it is what this tool
// was called before it was called Cohere, and a directive already written under the old name should
// not start failing because the binary was renamed. New code writes `cohere-disable`.
//
// Honoring the old spellings is not deference to the old tools. It is refusing to make a mechanical
// rename the price of switching linters — the same reasoning that keeps a rule enabled by giving it
// a narrow escape hatch, applied one level up to the tool itself.
//
// The grammar does not vary by spelling, so nothing downstream of parsing knows which one it read.
// That is the property that keeps this from becoming several code paths that drift.
//
// Whitespace in this grammar is JavaScript's, because ESLint's is: it trims with
// `String.prototype.trim` and ends the directive word at `\s`. Go's `strings.TrimSpace` counts the
// next-line character (U+0085) as space and the byte order mark (U+FEFF) as not, and JavaScript does
// the reverse, so a Go trim read 15 of 19 comments holding one of them, a no-break space after the
// word included, differently from ESLint (#z4nssqs, found by @system_adamic's stream P2).
const (
	// DirectiveCohere is the spelling new code should use.
	DirectiveCohere = "cohere-disable"

	// DirectiveVerify is what this tool's directives were called before the rename to Cohere.
	DirectiveVerify = "verify-disable"

	// DirectiveEslint is the spelling the existing corpus uses. Permanent, not transitional.
	DirectiveEslint = "eslint-disable"

	// DirectiveOxlint is the spelling of the gate Cohere replaces.
	DirectiveOxlint = "oxlint-disable"
)

// disableDirectives is every honored disable spelling.
var disableDirectives = []string{DirectiveCohere, DirectiveVerify, DirectiveEslint, DirectiveOxlint}

// enableDirectives closes a block, one per disable spelling.
var enableDirectives = []string{"cohere-enable", "verify-enable", "eslint-enable", "oxlint-enable"}

// Scope is the span a disable directive covers.
type Scope int

const (
	// ScopeNextLine silences the line after the comment.
	ScopeNextLine Scope = iota

	// ScopeSameLine silences the line the comment sits on.
	ScopeSameLine

	// ScopeFile silences from the comment's own line to a matching enable, or to the end of the file.
	ScopeFile
)

// Disable is one parsed disable directive.
type Disable struct {
	Scope Scope

	// Rules are the rule names the comment named. Empty means blanket.
	Rules []string

	// Reason is the text after the first `--`, trimmed. Empty means the author did not say why.
	Reason string
}

// ParseDisable reads one comment's full source text, delimiters included, as a disable directive.
//
// The grammar, identical for every honored spelling:
//
//	<tool>-disable[-next-line|-line] [rule[, rule...]] [-- reason]
//
// Everything after the directive word is optional. No rules means blanket. No reason means the
// author did not say why, which is recorded rather than rejected.
func ParseDisable(commentText string) (Disable, bool) {
	body := text.TrimWhitespace(stripCommentMarkers(commentText))

	rest, found := splitDirective(body)
	if !found {
		return Disable{}, false
	}

	scope, rest, valid := splitScope(rest)
	if !valid {
		return Disable{}, false
	}
	if scope == ScopeFile && isLineComment(commentText) {
		return Disable{}, false
	}

	rules, reason := splitReason(rest)

	return Disable{
		Scope:  scope,
		Rules:  parseRuleNames(rules),
		Reason: reason,
	}, true
}

// ParseEnable reads one comment's full source text as an enable directive, returning the rule names
// it closes. No names closes every open block.
func ParseEnable(commentText string) (rules []string, found bool) {
	if isLineComment(commentText) {
		return nil, false
	}
	body := text.TrimWhitespace(stripCommentMarkers(commentText))

	for _, directive := range enableDirectives {
		if !strings.HasPrefix(body, directive) {
			continue
		}
		rest := body[len(directive):]
		// Same anchoring rule as a disable: the directive has to be the whole word.
		if !endsWord(rest) {
			continue
		}
		names, _ := splitReason(rest)
		return parseRuleNames(names), true
	}
	return nil, false
}

// The scope words of the directives honored here, in ESLint's vocabulary.
//
// Normalized to the `eslint-` spelling because that is the vocabulary upstream's
// `require-description` `ignore` option speaks, and because the grammar does not vary by spelling:
// `cohere-disable-next-line` and `eslint-disable-next-line` are the same directive to every reader.
const (
	ScopeWordDisable         = "eslint-disable"
	ScopeWordDisableLine     = "eslint-disable-line"
	ScopeWordDisableNextLine = "eslint-disable-next-line"
	ScopeWordEnable          = "eslint-enable"
)

// Recognize reports whether a comment is a directive the suppression index acts on, and which one.
//
// The input is one comment's full source text, delimiters included. The answer is the same parse
// the suppression index builds from, so a comment Recognize accepts is a comment that suppresses or
// closes a suppression in a real run, in any honored spelling.
func Recognize(commentText string) (scopeWord string, honored bool) {
	if parsed, isDisable := ParseDisable(commentText); isDisable {
		switch parsed.Scope {
		case ScopeNextLine:
			return ScopeWordDisableNextLine, true
		case ScopeSameLine:
			return ScopeWordDisableLine, true
		case ScopeFile:
			return ScopeWordDisable, true
		}
	}
	if _, isEnable := ParseEnable(commentText); isEnable {
		return ScopeWordEnable, true
	}
	return "", false
}

// splitDirective finds an honored directive word at the start of the comment body.
//
// Anchored at the start on purpose. A directive has to be the first thing in its comment to count,
// or every sentence that mentions one becomes a suppression. This codebase contains exactly such
// sentences, including one inside a lint rule's own documentation that reads "opt out explicitly
// with an eslint-disable."
func splitDirective(body string) (rest string, found bool) {
	for _, candidate := range disableDirectives {
		if strings.HasPrefix(body, candidate) {
			return body[len(candidate):], true
		}
	}
	return "", false
}

// splitScope reads the scope suffix and returns what follows it.
//
// The order of these cases is load-bearing and is the highest-consequence bug available in this
// file. `-next-line` must be tested before the bare form, or all 372 next-line directives in the
// corpus read as file-level disables — each one silencing every rule for the remainder of its file,
// with nothing in the output saying so. Go's switch does not fall through, but the cases are still
// written longest-first so the ordering is visible to a reader rather than incidental.
func splitScope(rest string) (scope Scope, remainder string, valid bool) {
	switch {
	case strings.HasPrefix(rest, "-next-line"):
		return ScopeNextLine, rest[len("-next-line"):], endsWord(rest[len("-next-line"):])
	case strings.HasPrefix(rest, "-line"):
		return ScopeSameLine, rest[len("-line"):], endsWord(rest[len("-line"):])
	case endsWord(rest):
		// The bare, file-level form.
		return ScopeFile, rest, true
	}

	// The directive is a prefix of a longer word — `eslint-disable-nonsense`, or `eslint-disabled`.
	// Prose that merely contains a directive word must not silence a rule.
	return ScopeFile, "", false
}

// endsWord reports whether the directive word just read is a whole word: the comment ends, or
// JavaScript whitespace follows, as ESLint's `(?:\s|$)` after the word has it. A no-break space or a
// byte order mark ends the word, and the next-line character does not.
//
// Every scope needs it, not only the bare form. ESLint 10.8.1 reads none of these as a directive,
// measured: `eslint-disable-next-line, no-console`, `eslint-disable-next-lineno-console`, and
// `eslint-disable-lines no-console`. Without the check the first named `no-console` and the last
// named a rule `s no-console`, each suppressing what ESLint reports.
func endsWord(rest string) bool {
	if rest == "" {
		return true
	}
	first, _ := utf8.DecodeRuneInString(rest)
	return text.IsWhitespace(first)
}

// splitReason separates the rule list from the ` -- reason` that may follow it.
//
// The separator is `--`, which is the convention the existing corpus uses and ESLint's own. A rule
// name cannot contain `--`, so the first occurrence wins.
func splitReason(rest string) (rules string, reason string) {
	if index := strings.Index(rest, "--"); index >= 0 {
		return text.TrimWhitespace(rest[:index]), text.TrimWhitespace(rest[index+2:])
	}
	return text.TrimWhitespace(rest), ""
}

// parseRuleNames splits a comma-separated rule list, dropping empties.
//
// The corpus writes multi-rule lines: `nexus/consistency-require-type-suffix,
// nexus/consistency-no-abbreviated-identifier` on one comment. Splitting on comma rather than
// whitespace matters because that is what the corpus and ESLint both do — a whitespace split reads
// two rules as one unknown name, which matches nothing and silences nothing while looking like it
// silenced something.
func parseRuleNames(list string) []string {
	if list == "" {
		return nil
	}
	names := []string{}
	for _, part := range strings.Split(list, ",") {
		if trimmed := text.TrimWhitespace(part); trimmed != "" {
			names = append(names, trimmed)
		}
	}
	return names
}

// isLineComment reports whether a comment is the `//` form.
//
// A file-scope disable and an enable are directives only in a block comment. That is ESLint's
// grammar, measured on 10.8.1: `// eslint-disable no-console` neither silences `no-console` nor is
// read as a directive, and `// eslint-enable` closes nothing, while the same text in `/* */` does
// both. Only `-line` and `-next-line` work as line comments.
//
// cohere honored the `//` form until Kirk's ruling of 2026-10-01, and the cost was prose becoming a
// directive: phi web's build scripts carry `// eslint-disable + generated banner keep the linter and
// future readers out.`, a sentence about a generated file, which cohere read as a file-wide disable.
// A sentence that silently disables a whole file is the gap cohere exists to close, and the parity
// doctrine is never worse than ESLint.
func isLineComment(commentText string) bool {
	return strings.HasPrefix(commentText, "//")
}

// stripCommentMarkers removes `//`, `/*`, `*/`, and any `*` continuation markers.
//
// A block comment spanning lines carries `*` at the start of each continuation line, and a
// suppression's reason often wraps. Joining those into one line before parsing is what lets a
// wrapped reason parse as a reason rather than as a truncated one. JSX `{/* ... */}` arrives here
// already unwrapped by the scanner, which returns the comment span rather than the braces.
func stripCommentMarkers(commentText string) string {
	if strings.HasPrefix(commentText, "//") {
		return commentText[2:]
	}

	inner := strings.TrimSuffix(strings.TrimPrefix(commentText, "/*"), "*/")
	if !strings.Contains(inner, "\n") {
		return inner
	}

	lines := strings.Split(inner, "\n")
	for index, line := range lines {
		lines[index] = strings.TrimPrefix(text.TrimWhitespace(line), "*")
	}
	return strings.Join(lines, " ")
}
