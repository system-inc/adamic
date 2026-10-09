package nexus

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/comments"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// consistencyNoSingleLineJsDocText is the rule's message, whose wording lives in
// `policy/messages/consistency-no-single-line-jsdoc.json`.
var consistencyNoSingleLineJsDocText = policy.MessageOf("nexus/consistency-no-single-line-jsdoc", "useSimpleComment")

func messageUseSimpleComment() rule.Message {
	return rule.Message{
		Id:          "useSimpleComment",
		Description: consistencyNoSingleLineJsDocText.Render(nil),
	}
}

// ConsistencyNoSingleLineJsDoc turns a JSDoc comment holding nothing but one line of prose into a
// line comment.
//
//	valid:   // Does the thing.
//	valid:   /** Does the thing. @returns nothing */
//	invalid: /** Does the thing. */
//
// The fix is safe in a way most comment rewrites are not: the text is preserved verbatim and only
// the delimiters change, so nothing a reader wrote is lost or reflowed.
//
// It is withheld in three cases. A comment whose description already holds a double-slash would
// produce a line comment that reads as commented-out code, and a multi-line JSDoc collapses several
// source lines into one, which moves every position after it in the file. And a comment with code
// after it on its line, `createQueryClient(/** options */);` or `[a, b /**, c */] = values`, would
// comment that code out: the fix did not parse on trpc (#kq9vtva). None of them is a repair a rule
// should make unattended.
var ConsistencyNoSingleLineJsDoc = rule.Rule{
	Name: "nexus/consistency-no-single-line-jsdoc",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		return rule.Listeners{
			ast.KindSourceFile: func(node *ast.Node) {
				for _, comment := range comments.ForFile(ctx) {
					if !comment.IsJsDoc() {
						continue
					}

					lines := comment.ContentLines()
					// One line, and no tag. A tag is the entire reason JSDoc exists, so a comment
					// carrying one is a JSDoc comment whatever its length.
					if len(lines) != 1 || strings.HasPrefix(lines[0], "@") {
						continue
					}

					description := lines[0]
					isSingleSourceLine := comment.StartLine == comment.EndLine
					if strings.Contains(description, "//") || !isSingleSourceLine ||
						consistencyNoSingleLineJsDocCodeFollows(ctx.SourceFile.Text(), comment.Range.End()) {
						ctx.ReportRange(comment.Range, messageUseSimpleComment())
						continue
					}

					ctx.Report(rule.Diagnostic{
						Range:      comment.Range,
						Message:    messageUseSimpleComment(),
						SourceFile: ctx.SourceFile,
						Fixes: []rule.Fix{
							rule.ReplaceRange(comment.Range, "// "+description),
						},
					})
				}
			},
		}
	},
}

// consistencyNoSingleLineJsDocCodeFollows reports whether anything but whitespace follows a comment on
// its own line, which a line comment in its place would swallow.
func consistencyNoSingleLineJsDocCodeFollows(source string, commentEnd int) bool {
	for _, character := range source[commentEnd:] {
		switch character {
		case '\n', '\r':
			return false
		case ' ', '\t':
			continue
		}
		return true
	}
	return false
}
