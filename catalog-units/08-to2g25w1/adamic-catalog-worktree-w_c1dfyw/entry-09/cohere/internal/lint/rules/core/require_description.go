package core

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/comments"
	"github.com/system-inc/cohere/internal/lint/ecmascript/directives"
	"github.com/system-inc/cohere/internal/lint/ecmascript/text"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// RequireDescriptionIgnorable is upstream's `ignore` enum, verbatim from `meta.schema`.
//
// `eslint-env` stays although ESLint 10 no longer honors the directive: upstream's DIRECTIVE_PATTERN
// still matches it and the rule still reports it. Measured on the installed 10.8.1 driving the cloned
// rule, `/* eslint-env node */` produces this rule's finding beside ESLint's own fatal "no longer
// supported" message, and `/* eslint-env -- d */` produces only the fatal one.
var RequireDescriptionIgnorable = []string{
	"eslint",
	"eslint-disable",
	"eslint-disable-line",
	"eslint-disable-next-line",
	"eslint-enable",
	"eslint-env",
	"exported",
	"global",
	"globals",
}

// RequireDescriptionOptions is upstream's single option object.
//
// `Ignore` names directive kinds to skip, in upstream's vocabulary. A cohere spelling is ignored by
// the kind it shares with its `eslint-` twin, so `ignore: ["eslint-disable-next-line"]` also skips
// `cohere-disable-next-line`. ESLint never sees a `cohere-` comment at all, so this extends the key's
// meaning only over comments the other engine cannot read.
//
// `AdditionalDirectives` is upstream's escape hatch for other tools' comment directives (`c8`,
// `istanbul`), and is ported as upstream wrote it.
type RequireDescriptionOptions struct {
	Ignore               []string `json:"ignore"`
	AdditionalDirectives []string `json:"additionalDirectives"`
}

// DecodeRequireDescriptionOptions reads the option object, refusing what upstream's schema refuses.
//
// The schema says `additionalProperties: false`, an `ignore` enum, and `uniqueItems`. There is no
// schema layer here, so each is checked by hand: a misspelled key or kind would otherwise decode
// silently to "ignore nothing", which reports more than the author asked for and says nothing.
// A bare severity hands this no bytes, which is the zero options and not an error.
func DecodeRequireDescriptionOptions(raw []byte) (any, error) {
	options := RequireDescriptionOptions{}
	if len(raw) == 0 {
		return options, nil
	}
	if err := rule.UnmarshalOptions(raw, &options); err != nil {
		return options, fmt.Errorf("%s: %w", RequireDescription.Name, err)
	}
	seen := map[string]bool{}
	for _, kind := range options.Ignore {
		known := false
		for _, ignorable := range RequireDescriptionIgnorable {
			if kind == ignorable {
				known = true
				break
			}
		}
		if !known {
			return options, fmt.Errorf("%s: ignore names %q, wanted one of %s",
				RequireDescription.Name, kind, strings.Join(RequireDescriptionIgnorable, ", "))
		}
		if seen[kind] {
			return options, fmt.Errorf("%s: ignore names %q twice", RequireDescription.Name, kind)
		}
		seen[kind] = true
	}
	return options, nil
}

// messageRequireDescriptionMissing is upstream's `missingDescription`, with the repair spelled out.
var messageRequireDescriptionMissing = rule.Message{
	Id: "missingDescription",
	Description: "Unexpected undescribed directive comment. Include descriptions to explain why the " +
		"comment is necessary: write the reason after ` -- ` at the end of the directive, as in " +
		"`// eslint-disable-next-line some-rule -- why this line is the exception`. A suppression " +
		"that does not say why cannot be told apart from one nobody remembers adding.",
}

// RequireDescription reports a directive comment that carries no ` -- reason`.
//
//	valid:   // eslint-disable-next-line eqeqeq -- comparing against a DOM attribute
//	valid:   /* eslint-disable -- generated file */
//	valid:   // cohere-disable-next-line no-alert -- the kiosk build has no other channel
//	invalid: // eslint-disable-line
//	invalid: /* eslint-disable-next-line eqeqeq -- */          a separator with nothing after it
//	invalid: /* global _ */
//	invalid: // cohere-disable-next-line no-alert
//
// Ported from `@eslint-community/eslint-comments/require-description`, reading the clone of
// eslint-community/eslint-plugin-eslint-comments at de11b0e (`lib/rules/require-description.js`,
// `lib/internal/get-all-directive-comments.js`, `lib/internal/utils.js`) and measuring every verdict
// in the test file against that rule driven through the installed ESLint 10.8.1 Linter API. The
// plugin is not installed in the ahra tree, so the oracle loads the cloned rule directly.
//
// # Two parsers, one verdict per comment
//
// A comment is a directive if upstream's parser says so, or, failing that, if this tree's
// suppression parser says cohere acts on it. The first arm is upstream verbatim and decides every
// case in upstream's corpus. The second covers what upstream cannot know about:
//
//   - cohere's own spellings, `cohere-`, `verify-` and `oxlint-` with every scope suffix and their
//     `-enable` twins, read by `directives.Recognize`, the same parse the suppression index builds
//     from, so the two agree on what a directive is by construction rather than by a copied list;
//   - `eslint-` comments ESLint ignores and cohere honors: `// eslint-disable` and `// eslint-enable`
//     as line comments, a multi-line `/* eslint-disable-line ... */`, and a block whose continuation
//     lines carry `*`. Upstream reports none of them because in ESLint none of them do anything.
//     In cohere each one suppresses, so it is a suppression without a reason, which is exactly what
//     Kirk asked this rule to refuse. This is the one place the engines diverge, and it diverges only
//     where the engines' directive grammars already do.
//
// # A reason is ESLint's grammar, in both arms
//
// The description is what follows the first whitespace-dashes-whitespace (`/\s-{2,}\s/u`), trimmed,
// and it has to be non-empty. So `-- ` with nothing after it is no reason, and `foo--bar` is not a
// separator at all, even for a `cohere-` spelling whose own suppression parser reads `bar` as a
// reason. Measured: `// eslint-disable-next-line foo--bar`, `... foo --` and `... foo -- ` all report;
// `... foo -- -- ` and `... foo -- --bar` do not, since the second piece is `--` and `--bar`.
//
// # The span is the comment, and the comment cannot silence its own finding
//
// Upstream reports from column -1 of the comment's first line to the comment's end. The -1 is not a
// decision about where to point. It exists so ESLint's directive matcher, which starts every
// same-line and next-line directive at column 0, can never cover the finding, and a bare
// `x; // eslint-disable-line` cannot hide its own report. cohere's suppression is line-based and has
// no column to hide behind, so the same decision is made by registering through
// `directives.RegisterSubject`; the measurements are on `coversAFindingAboutADirective` in the
// suppression package. The end of the span is upstream's exactly; the start is
// the comment's own first byte rather than the code before it.
//
// # Not fixable
//
// `meta.fixable` is null upstream. Only the author knows the reason.
var RequireDescription = rule.Rule{
	// Upstream's plugin-qualified name, so the same config key means the same thing in ESLint and
	// in cohere.
	Name: "@eslint-community/eslint-comments/require-description",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}
		// A bare severity decodes to the zero options, and a harness that bypasses the decoder hands
		// nil. Both mean upstream's defaults: ignore nothing, no additional directives.
		configured, _ := rule.OptionsAs[RequireDescriptionOptions](options)

		return rule.Listeners{
			ast.KindSourceFile: func(node *ast.Node) {
				reportRequireDescription(ctx, configured)
			},
		}
	},
}

// reportRequireDescription is upstream's `create` body, run once per file.
func reportRequireDescription(ctx rule.Context, options RequireDescriptionOptions) {
	ignored := map[string]bool{}
	for _, kind := range options.Ignore {
		ignored[kind] = true
	}
	additional := requireDescriptionAdditionalPattern(options.AdditionalDirectives)

	for _, comment := range comments.ForFile(ctx) {
		interior, readable := requireDescriptionInterior(comment)
		if !readable {
			continue
		}

		kind, isDirective := "", false
		if requireDescriptionCouldBeUpstream(interior, options.AdditionalDirectives) {
			kind, isDirective = requireDescriptionUpstreamKind(comment, interior, options.AdditionalDirectives, additional)
		}
		if !isDirective {
			kind, isDirective = directives.Recognize(comment.Text)
		}
		if !isDirective || ignored[kind] {
			continue
		}
		if requireDescriptionHasDescription(interior) {
			continue
		}
		ctx.ReportRange(comment.Range, messageRequireDescriptionMissing)
	}
}

// requireDescriptionInterior is ESLint's `comment.value`: the text between the delimiters.
//
// An unterminated block comment has no interior ESLint would ever see, since the file fails to
// parse there, so it is declined rather than sliced into something shorter than it is.
func requireDescriptionInterior(comment comments.Comment) (string, bool) {
	if !comment.IsBlock {
		return comment.Text[2:], true
	}
	if len(comment.Text) < 4 || !strings.HasSuffix(comment.Text, "*/") {
		return "", false
	}
	return comment.Text[2 : len(comment.Text)-2], true
}

// jsWhitespace is JavaScript's `\s` (WhiteSpace plus LineTerminator), as a character class.
//
// Go's `\s` is ASCII and omits vertical tab, no-break space and the rest, so a directive separated
// from its rules by a no-break space would parse differently from upstream with Go's class.
const jsWhitespace = `[\t\n\x0B\f\r \x{00A0}\x{1680}\x{2000}-\x{200A}\x{2028}\x{2029}\x{202F}\x{205F}\x{3000}\x{FEFF}]`

// requireDescriptionSeparator is upstream's `divideDirectiveComment` split, `/\s-{2,}\s/u`.
var requireDescriptionSeparator = regexp.MustCompile(jsWhitespace + `-{2,}` + jsWhitespace)

// requireDescriptionDirective is upstream's DIRECTIVE_PATTERN, with JavaScript's `\s`.
var requireDescriptionDirective = regexp.MustCompile(
	`^(eslint(?:-env|-enable|-disable(?:(?:-next)?-line)?)?|exported|globals?)(?:` + jsWhitespace + `|$)`)

// requireDescriptionLineKind is upstream's LINE_COMMENT_PATTERN: the only built-in kinds a line
// comment can carry. Every other built-in directive is block-only in ESLint.
var requireDescriptionLineKind = regexp.MustCompile(`^eslint-disable-(next-)?line$`)

// requireDescriptionAdditionalPattern is upstream's per-call additional-directive regex, or nil.
func requireDescriptionAdditionalPattern(additional []string) *regexp.Regexp {
	if len(additional) == 0 {
		return nil
	}
	quoted := make([]string, len(additional))
	for index, directive := range additional {
		quoted[index] = regexp.QuoteMeta(directive)
	}
	return regexp.MustCompile(`^(` + strings.Join(quoted, "|") + `)(?:` + jsWhitespace + `|$)`)
}

// requireDescriptionCouldBeUpstream reports whether upstream's parser could read a directive here at all,
// so the split and the patterns run only on comments that open with a directive word.
//
// Exact rather than a heuristic. Both patterns are anchored at the start of the text before the first
// separator, trimmed, and that piece is a prefix of the interior, so a directive's word begins the
// interior once its leading whitespace is gone. A comment opening any other way cannot match, and on a
// real tree that is nearly every comment: before this, each one paid a regexp split, which allocates,
// and a submatch search (#fcac58b).
func requireDescriptionCouldBeUpstream(interior string, additional []string) bool {
	head := strings.TrimLeftFunc(interior, text.IsWhitespace)
	if strings.HasPrefix(head, "eslint") || strings.HasPrefix(head, "exported") ||
		strings.HasPrefix(head, "global") {
		return true
	}
	for _, directive := range additional {
		if strings.HasPrefix(head, directive) {
			return true
		}
	}
	return false
}

// requireDescriptionUpstreamKind is upstream's `parseDirectiveComment`, returning the kind.
func requireDescriptionUpstreamKind(
	comment comments.Comment, interior string, additional []string, additionalPattern *regexp.Regexp,
) (string, bool) {
	directiveText := text.TrimWhitespace(requireDescriptionSeparator.Split(interior, 2)[0])

	kind := ""
	if match := requireDescriptionDirective.FindStringSubmatch(directiveText); match != nil {
		kind = match[1]
	} else if additionalPattern != nil {
		match := additionalPattern.FindStringSubmatch(directiveText)
		if match == nil {
			return "", false
		}
		kind = match[1]
	} else {
		return "", false
	}

	lineCommentSupported := requireDescriptionLineKind.MatchString(kind)
	for _, directive := range additional {
		if directive == kind {
			lineCommentSupported = true
		}
	}
	if !comment.IsBlock && !lineCommentSupported {
		return "", false
	}

	// A disable-line comment spanning lines is not a directive to ESLint. Measured:
	// `/* eslint-disable-line\n */` reports nothing upstream.
	//
	// Kept for fidelity, and invisible in this rule's verdict: cohere's suppression index honors the
	// same comment as a same-line directive, so the second arm in reportRequireDescription answers
	// with the same kind and the same finding. A mutation sweep scores removing this as a survivor
	// for that reason, not because the fixture is missing.
	if kind == "eslint-disable-line" && comment.StartLine != comment.EndLine {
		return "", false
	}
	return kind, true
}

// requireDescriptionHasDescription is upstream's `!directiveComment.description` test, inverted.
//
// `split` with a regex keeps every piece, and upstream reads only the second, so `a -- b -- c`
// describes itself as `b`. Splitting into at most three pieces keeps that second piece exactly as
// JavaScript would cut it, without the rest of the comment folded into it.
func requireDescriptionHasDescription(interior string) bool {
	pieces := requireDescriptionSeparator.Split(interior, 3)
	return len(pieces) > 1 && text.TrimWhitespace(pieces[1]) != ""
}
