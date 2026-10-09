package nexus

import (
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/ecmascript/comments"
	"github.com/system-inc/cohere/internal/lint/ecmascript/text"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// Four or fewer double-slash lines read fine as a stack; the block delimiters would be noise.
const defaultMaximumLineCount = 4

// A directive only works as a line comment. Folding one into a block silently disables it, and the
// rule it was suppressing then fires somewhere else, which reads as an unrelated regression. This
// covers eslint's own directives plus the neighbours that share the shape.
var directivePrefixes = []string{
	"eslint-", "@ts-", "prettier-ignore", "istanbul ", "c8 ", "v8 ",
	"#__PURE__", "global ", "globals ", "exported ",
}

const consistencyNoLongLineCommentId = "longLineComment"

// consistencyNoLongLineCommentText is the rule's message, whose wording lives in
// `policy/messages/consistency-no-long-line-comment.json`.
var consistencyNoLongLineCommentText = policy.MessageOf("nexus/consistency-no-long-line-comment", consistencyNoLongLineCommentId)

// ConsistencyNoLongLineComment requires a block comment for a long run of line comments.
//
//	valid:   four consecutive // lines
//	invalid: five consecutive // lines
//
// The fix preserves each line's text verbatim and changes only the delimiters, which is what makes
// it safe to apply unattended. It is withheld when the run contains a star-slash sequence, since
// that would close the block early and change what the rest of the run means.
var ConsistencyNoLongLineComment = rule.Rule{
	Name: "nexus/consistency-no-long-line-comment",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		maximumLineCount := defaultMaximumLineCount
		if settings, hasSettings := rule.OptionsAs[ConsistencyNoLongLineCommentOptions](options); hasSettings && settings.MaximumLineCount > 0 {
			maximumLineCount = settings.MaximumLineCount
		}

		return rule.Listeners{
			ast.KindSourceFile: func(node *ast.Node) {
				sourceLines := strings.Split(ctx.SourceFile.Text(), "\n")

				var run []comments.Comment
				flush := func() {
					reportLineCommentRun(ctx, run, maximumLineCount)
					run = nil
				}

				for _, comment := range comments.ForFile(ctx) {
					if !isFoldableLineComment(comment, sourceLines) {
						flush()
						continue
					}
					if len(run) > 0 && !isConsecutiveComment(run[len(run)-1], comment) {
						flush()
					}
					run = append(run, comment)
				}
				flush()
			},
		}
	},
}

// ConsistencyNoLongLineCommentOptions lets a project pick its own threshold.
type ConsistencyNoLongLineCommentOptions struct {
	// MaximumLineCount is the longest run of line comments left alone. Zero means the default.
	MaximumLineCount int `json:"maximumLineCount"`
}

// isFoldableLineComment reports whether a comment can join a run at all.
func isFoldableLineComment(comment comments.Comment, sourceLines []string) bool {
	if comment.IsBlock {
		return false
	}
	if isDirectiveComment(comment) {
		return false
	}
	// Only a comment alone on its line can be folded. One trailing real code shares the line, so
	// replacing it with a block would swallow the code before it.
	if comment.StartLine < 0 || comment.StartLine >= len(sourceLines) {
		return false
	}
	line := sourceLines[comment.StartLine]
	if comment.StartColumn > len(line) {
		return false
	}
	return text.TrimWhitespace(line[:comment.StartColumn]) == ""
}

// isDirectiveComment reports whether a comment is a pragma rather than prose.
func isDirectiveComment(comment comments.Comment) bool {
	body := text.TrimWhitespace(strings.TrimPrefix(comment.Text, "//"))
	for _, prefix := range directivePrefixes {
		if strings.HasPrefix(body, prefix) {
			return true
		}
	}
	return false
}

// isConsecutiveComment reports whether two comments stack.
//
// Adjacent lines are not enough: the columns must match too, so a trailing comment sitting at the
// end of a code line never joins the passage above it.
func isConsecutiveComment(previous comments.Comment, current comments.Comment) bool {
	if previous.EndLine+1 != current.StartLine {
		return false
	}
	return previous.StartColumn == current.StartColumn
}

// reportLineCommentRun reports a run that is too long, with a fix when one is safe.
func reportLineCommentRun(ctx rule.Context, run []comments.Comment, maximumLineCount int) {
	if len(run) <= maximumLineCount {
		return
	}

	first, last := run[0], run[len(run)-1]
	runRange := core.NewTextRange(first.Range.Pos(), last.Range.End())

	message := rule.Message{
		Id:          consistencyNoLongLineCommentId,
		Description: consistencyNoLongLineCommentText.Render(map[string]string{"lineCount": strconv.Itoa(len(run))}),
	}

	// A star-slash anywhere in the text would close the block early, so report without a fix and
	// let a human decide how to phrase it.
	for _, comment := range run {
		if strings.Contains(comment.Text, "*/") {
			ctx.ReportRange(runRange, message)
			return
		}
	}

	ctx.Report(rule.Diagnostic{
		Range:      runRange,
		Message:    message,
		SourceFile: ctx.SourceFile,
		Fixes:      []rule.Fix{rule.ReplaceRange(runRange, blockCommentFromRun(run))},
	})
}

// blockCommentFromRun rewrites a run as a block, keeping each line's text and the run's indentation.
//
// The double-slash and its single following space are stripped so the prose lands flush against the
// star gutter, which is the only reflowing the fix does.
func blockCommentFromRun(run []comments.Comment) string {
	indent := strings.Repeat(" ", run[0].StartColumn)

	var builder strings.Builder
	builder.WriteString("/*\n")
	for _, comment := range run {
		body := strings.TrimPrefix(comment.Text, "//")
		body = strings.TrimPrefix(body, " ")
		if text.TrimWhitespace(body) == "" {
			builder.WriteString(indent + " *\n")
			continue
		}
		builder.WriteString(indent + " * " + body + "\n")
	}
	builder.WriteString(indent + " */")
	return builder.String()
}
