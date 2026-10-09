package structure

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/comments"
	"github.com/system-inc/cohere/internal/lint/ecmascript/react"
	"github.com/system-inc/cohere/internal/lint/ecmascript/text"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// reactHookRequireEffectCommentText is the rule's message, whose wording lives in
// `policy/messages/react-hook-require-effect-comment.json`.
var reactHookRequireEffectCommentText = policy.MessageOf("structure/react-hook-require-effect-comment", "missingEffectComment")

// messageMissingEffectComment is the finding, rendered when it is reported so the text comes from
// the current catalog.
func messageMissingEffectComment() rule.Message {
	return rule.Message{Id: reactHookRequireEffectCommentText.Id, Description: reactHookRequireEffectCommentText.Render(nil)}
}

// effectCommentPrefix is what the first line of the comment must start with.
//
// A fixed prefix rather than "any comment", and the strictness is the point: it forces the sentence
// into the shape "Effect to <do something>", which is a claim about purpose. A free-form comment
// drifts into restating the code.
const effectCommentPrefix = "Effect to"

// ReactHookRequireEffectComment flags a React.useEffect with no explanatory comment above it.
//
//	valid:   // Effect to sync the title with the route
//	         React.useEffect(() => { ... }, [route])
//	invalid: React.useEffect(() => { ... }, [route])
//	invalid: // sync the title
//	         React.useEffect(() => { ... }, [route])
//
// Only `React.useEffect`, matching the original. A bare `useEffect(...)` is not reported, which
// looks like a gap and is not: `react-import-no-destructuring` already forbids importing the name,
// so in this codebase the bare form cannot exist without that rule firing first. Widening this one
// would produce two findings for one mistake.
//
// The comment is looked up on the enclosing statement rather than the call. Comments attach to
// statements, so asking at the call expression finds nothing for the ordinary
// `React.useEffect(...)` written as its own statement.
var ReactHookRequireEffectComment = rule.Rule{
	Name: "structure/react-hook-require-effect-comment",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if !FileContextFor(ctx.SourceFile.FileName().AsString()).IsReactFile {
			// An effect outside a React file is not a component's effect. Declining here also skips
			// the comment scan entirely on the majority of files.
			return nil
		}

		return rule.Listeners{
			ast.KindCallExpression: func(node *ast.Node) {
				callee := ast.SkipParentheses(node.AsCallExpression().Expression)
				if callee == nil || callee.Kind != ast.KindPropertyAccessExpression {
					return
				}
				if !react.IsNamespacedMember(callee, func(name string) bool { return name == "useEffect" }) {
					return
				}

				if !hasEffectCommentAbove(ctx, node) {
					ctx.ReportNode(callee, messageMissingEffectComment())
				}
			},
		}
	},
}

// hasEffectCommentAbove reports whether the run of comments above the call's statement starts with
// the required prefix.
//
// The topmost comment of a contiguous run is what is checked, not the nearest one. A multi-line
// explanation is several comment nodes and only its first line has to make the claim: requiring
// every line to start with "Effect to" would forbid explaining anything in a second sentence.
//
// The run comes from `comments.LeadingRunFor`, which reads the shared per-file scan. This rule
// scanned the file itself for every `React.useEffect` call before that existed, which is the cost
// the shared scan was built to remove: three rules calling a well-factored scan directly each
// rescanned the file, measured at 807ms, 619ms and 601ms while visiting one node per file.
func hasEffectCommentAbove(ctx rule.Context, call *ast.Node) bool {
	statement := enclosingStatement(call)
	if statement == nil {
		return false
	}
	run := comments.LeadingRunFor(ctx, statement)
	if len(run) == 0 {
		return false
	}
	return startsWithEffectPrefix(run[0])
}

// startsWithEffectPrefix reports whether a comment's first meaningful line makes the claim.
//
// A block comment's leading asterisks are stripped per line, so the JSDoc shape counts: the run of
// `*` characters is decoration rather than content, and requiring the prefix immediately after
// `/**` would reject the multi-line form everyone writes.
func startsWithEffectPrefix(comment comments.Comment) bool {
	body := comment.Text

	switch {
	case strings.HasPrefix(body, "//"):
		return strings.HasPrefix(text.TrimWhitespace(body[2:]), effectCommentPrefix)

	case strings.HasPrefix(body, "/*"):
		body = strings.TrimPrefix(body, "/*")
		body = strings.TrimSuffix(body, "*/")
		for _, line := range strings.Split(body, "\n") {
			line = text.TrimWhitespace(strings.TrimLeft(text.TrimWhitespace(line), "*"))
			if line == "" {
				continue
			}
			return strings.HasPrefix(line, effectCommentPrefix)
		}
	}

	return false
}

// enclosingStatement walks up to the statement the call belongs to.
//
// Comments attach to statements rather than to expressions, so the lookup has to happen there. The
// walk stops at whatever sits directly inside a block or a source file, which is the node a comment
// above the call would be attached to.
func enclosingStatement(node *ast.Node) *ast.Node {
	for current := node; current != nil; current = current.Parent {
		parent := current.Parent
		if parent == nil {
			return current
		}
		switch parent.Kind {
		case ast.KindBlock, ast.KindSourceFile, ast.KindModuleBlock, ast.KindCaseClause,
			ast.KindDefaultClause:
			return current
		}
	}
	return nil
}
