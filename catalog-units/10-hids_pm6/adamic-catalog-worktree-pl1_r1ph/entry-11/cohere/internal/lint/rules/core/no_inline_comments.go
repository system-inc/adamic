package core

import (
	"regexp"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/comments"
	"github.com/system-inc/cohere/internal/lint/ecmascript/text"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// NoInlineCommentsOptions configures an escape hatch.
//
// Upstream's schema is a single object with one string, `ignorePattern`. Our config layer unwraps
// the severity tuple before dispatch, so the decoder receives that object directly rather than
// upstream's one-element array.
type NoInlineCommentsOptions struct {
	// IgnorePattern is a regular expression source, tested against the comment's INNER text. A
	// comment matching it is allowed inline, which is how a project keeps a convention like
	// `// webpackChunkName: ...` while banning the rest.
	IgnorePattern string `json:"ignorePattern"`
}

var messageNoInlineCommentsUnexpected = rule.Message{
	Id: "unexpectedInlineComment",
	Description: "This comment shares a line with code. A trailing comment is the first thing lost " +
		"when a line is wrapped, reordered, or moved, because it belongs to a position rather than " +
		"to a statement, and it pushes the line past whatever width the rest of the file keeps to. " +
		"Put it on its own line above the code it describes, where it survives the next edit and " +
		"where a reader looking for it does not have to scan to the right margin.",
}

// noInlineCommentsDirectivePattern is upstream's `ESLINT_DIRECTIVE_PATTERN`.
//
// `/^(?:eslint[- ]|(?:globals?|exported) )/u`, applied to a BLOCK comment's trimmed inner text. A
// LINE comment uses a different and simpler test upstream, `startsWith("eslint-")`, so the two are
// not interchangeable and both are reproduced below.
var noInlineCommentsDirectivePattern = regexp.MustCompile(`^(?:eslint[- ]|(?:globals?|exported) )`)

// NoInlineComments flags a comment that shares a line with code.
//
//	valid:   // A comment on its own line
//	valid:   /* A block comment on its own line */
//	valid:   var a = 1; // eslint-disable-line no-debugger
//	valid:   foo(); /* global foo */
//	invalid: var a = 1; // trailing
//	invalid: var a = 1; /* leading */ var b = 2;
//
// Ported from `no-inline-comments` in ESLint, read from the clone at
// `lib/rules/no-inline-comments.js`. One option, one message, no fixer.
//
// The whole 49-case corpus was extracted from upstream's own tester and replayed against the rule
// through the ESLint Linter API before any code was written.
//
// # The oracle disagreed with the corpus on two cases, and the corpus was right
//
// Driving this rule standalone through the Linter API reports on `var a = 1; // eslint-disable-line
// no-debugger`, which upstream's corpus lists as CLEAN. That is not a disagreement about the rule:
// ESLint's own machinery consumes directive comments before a standalone rule run sees them in the
// same way, while the rule ALSO carries its own `isDirectiveComment` guard. Both exist, and the
// guard is the half that belongs to this port.
//
// Reproducing it here means all three of those cases are expressible as fixtures rather than being
// recorded as harness limitations, which is the better outcome: `testing.Run` never consults
// `internal/lint/suppression`, so a case that depended on the suppression layer could not be pinned
// at all.
//
// # The directive test is TWO tests, and they are not the same
//
// Upstream's `isDirectiveComment` branches on comment type:
//
//	Line comment    comment.startsWith("eslint-")
//	Block comment   /^(?:eslint[- ]|(?:globals?|exported) )/
//
// So `// global foo` is NOT a directive and reports, while `/* global foo */` IS one and does not.
// Collapsing the two into one test is the obvious simplification and it changes verdicts in both
// directions. The corpus contains the block form; the line form is covered by a fixture written
// here.
//
// # What counts as "inline" is decided on the raw LINE text, not on the AST
//
// Upstream slices the comment's own start and end lines and trims what sits outside it. A comment is
// inline when either side has anything left. That is a text question rather than a syntactic one,
// which is why this rule reads the source lines directly rather than asking about enclosing nodes.
//
// # The JSX exception, and why it needs the node under the comment
//
// `<div>{/* comment */}</div>` is clean. Upstream detects it by testing whether the preamble is
// empty or exactly `{`, the postamble empty or exactly `}`, and the node at the comment's position
// is a `JSXEmptyExpression` -- an expression container holding nothing but the comment. Ours has no
// such node kind: an empty JSX expression is `KindJsxExpression` with a nil `Expression`, which is
// the same shape under a different name.
var NoInlineComments = rule.Rule{
	Name: "no-inline-comments",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		var ignorePattern *regexp.Regexp
		if resolved, isNoInlineCommentsOptions := rule.OptionsAs[NoInlineCommentsOptions](options); isNoInlineCommentsOptions &&
			resolved.IgnorePattern != "" {
			// A pattern Go's RE2 cannot compile is dropped rather than crashing the run, which
			// reports MORE rather than less. Upstream's `new RegExp` throws and takes every other
			// rule's verdict on that file with it, because the walk recovers per file.
			if compiled, err := regexp.Compile(resolved.IgnorePattern); err == nil {
				ignorePattern = compiled
			}
		}

		return rule.Listeners{
			// Upstream's listener is `Program`, which fires once and walks every comment. The
			// equivalent here is `KindSourceFile`, which fires before its children.
			ast.KindSourceFile: func(node *ast.Node) {
				sourceFile := ctx.SourceFile
				if sourceFile == nil {
					return
				}
				lines := strings.Split(sourceFile.Text(), "\n")

				// Read the shared list once. `ForFile` is cached per file, so calling it in the
				// loop header is not a rescan, but it is a map lookup and a closure call per
				// comment for no reason.
				fileComments := comments.ForFile(ctx)
				for index := range fileComments {
					comment := &fileComments[index]
					if noInlineCommentsIsInline(ctx, lines, comment, ignorePattern) {
						ctx.ReportRange(comment.Range, messageNoInlineCommentsUnexpected)
					}
				}
			},
		}
	},
}

// noInlineCommentsIsInline answers upstream's `testCodeAroundComment`, minus the report.
func noInlineCommentsIsInline(
	ctx rule.Context,
	lines []string,
	comment *comments.Comment,
	ignorePattern *regexp.Regexp,
) bool {
	if comment.StartLine < 0 || comment.StartLine >= len(lines) ||
		comment.EndLine < 0 || comment.EndLine >= len(lines) {
		return false
	}

	startLine := lines[comment.StartLine]
	endLine := lines[comment.EndLine]

	// The text before the comment on its first line, and after it on its last.
	preamble := ""
	if comment.StartColumn <= len(startLine) {
		preamble = text.TrimWhitespace(startLine[:comment.StartColumn])
	}
	// The shelf's Comment carries StartLine, StartColumn and EndLine but no EndColumn, so it is
	// derived: the comment's end offset minus the offset at which its last line begins. Computed
	// from the same line split the preamble uses, which keeps the two sides measured on one
	// coordinate system rather than mixing a byte offset with a column.
	postamble := ""
	endColumn := comment.Range.End() - noInlineCommentsLineStartOffset(lines, comment.EndLine)
	if endColumn >= 0 && endColumn <= len(endLine) {
		postamble = text.TrimWhitespace(endLine[endColumn:])
	}

	// Nothing on either side: the comment owns its lines, which is the whole point of the rule.
	if preamble == "" && postamble == "" {
		return false
	}

	body := noInlineCommentsBody(comment)

	// The user's escape hatch, tested against the inner text.
	if ignorePattern != nil && ignorePattern.MatchString(body) {
		return false
	}

	// The JSX exception. A comment alone inside an expression container is idiomatic React and the
	// braces around it are not "code" in the sense this rule means.
	if (preamble == "" || preamble == "{") && (postamble == "" || postamble == "}") {
		if noInlineCommentsIsInsideEmptyJsxExpression(ctx, comment) {
			return false
		}
	}

	// A directive comment is machinery rather than prose, and it has to sit on the line it governs.
	if noInlineCommentsIsDirective(comment, body) {
		return false
	}

	return true
}

// noInlineCommentsBody strips a comment's delimiters and trims, matching ESLint's `comment.value`
// passed through `.trim()`.
//
// ESLint's `value` is the text BETWEEN the delimiters; the shelf's `Text` includes them, because
// what counts as content differs by rule. A block comment's closing delimiter is removed only when
// present, since an unterminated block comment at end of file has none.
func noInlineCommentsBody(comment *comments.Comment) string {
	body := comment.Text
	if comment.IsBlock {
		body = strings.TrimPrefix(body, "/*")
		body = strings.TrimSuffix(body, "*/")
	} else {
		body = strings.TrimPrefix(body, "//")
	}
	return text.TrimWhitespace(body)
}

// noInlineCommentsIsDirective answers upstream's `astUtils.isDirectiveComment`.
//
// Two different tests by comment type, and the asymmetry is upstream's rather than an oversight
// here. A line comment must start with `eslint-`; a block comment matches a wider pattern that also
// admits `global`, `globals` and `exported`. So `// global foo` reports and `/* global foo */` does
// not.
func noInlineCommentsIsDirective(comment *comments.Comment, body string) bool {
	if comment.IsBlock {
		return noInlineCommentsDirectivePattern.MatchString(body)
	}
	return strings.HasPrefix(body, "eslint-")
}

// noInlineCommentsIsInsideEmptyJsxExpression answers upstream's `JSXEmptyExpression` test.
//
// Upstream asks `getNodeByRangeIndex` for the node at the comment's start and checks its type. There
// is no `JSXEmptyExpression` kind here: an empty JSX expression container is `KindJsxExpression`
// whose `Expression` is nil, which is the same shape spelled differently.
//
// Walked from the source file rather than looked up by index, because there is no range-index lookup
// on the shim. The walk descends only into nodes whose range contains the comment, so it is a
// descent rather than a scan.
func noInlineCommentsIsInsideEmptyJsxExpression(ctx rule.Context, comment *comments.Comment) bool {
	if ctx.SourceFile == nil {
		return false
	}
	position := comment.Range.Pos()

	var found bool
	var descend func(*ast.Node)
	descend = func(node *ast.Node) {
		if found || node == nil {
			return
		}
		if position < node.Pos() || position >= node.End() {
			return
		}
		if node.Kind == ast.KindJsxExpression {
			if expression := node.AsJsxExpression(); expression != nil && expression.Expression == nil {
				found = true
				return
			}
		}
		node.ForEachChild(func(child *ast.Node) bool {
			descend(child)
			return found
		})
	}
	descend(ctx.SourceFile.AsNode())

	return found
}

// DefaultNoInlineCommentsOptions is the unconfigured answer: no ignore pattern.
func DefaultNoInlineCommentsOptions() NoInlineCommentsOptions {
	return NoInlineCommentsOptions{}
}

// DecodeNoInlineCommentsOptions reads this rule's configuration from the config layer.
//
// Hand-rolled rather than `rule.DecodeOptionsInto` so a bare `"error"` configuration, which arrives
// as empty input, resolves to the documented default instead of erroring.
func DecodeNoInlineCommentsOptions(raw []byte) (any, error) {
	options := DefaultNoInlineCommentsOptions()
	if len(raw) == 0 {
		return options, nil
	}
	if err := rule.UnmarshalOptions(raw, &options); err != nil {
		return options, err
	}
	return options, nil
}

// noInlineCommentsLineStartOffset returns the byte offset at which a zero-based line begins.
//
// Derived from the same `strings.Split` the caller uses, so the two agree by construction. Each line
// contributes its own length plus one for the newline that split consumed; a file with no trailing
// newline is unaffected, because the last line's start is computed from the lines before it.
func noInlineCommentsLineStartOffset(lines []string, line int) int {
	offset := 0
	for index := 0; index < line && index < len(lines); index++ {
		offset += len(lines[index]) + 1
	}
	return offset
}
