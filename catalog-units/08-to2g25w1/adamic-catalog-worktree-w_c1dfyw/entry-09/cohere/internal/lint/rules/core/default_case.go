package core

import (
	"regexp"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/comments"
	"github.com/system-inc/cohere/internal/lint/ecmascript/text"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// DefaultCaseOptions configures which comment counts as an explicit "no default is intended".
//
// Upstream's schema is a single object with one string property, `commentPattern`. Our config layer
// unwraps the severity tuple before dispatch, so the decoder receives that object directly rather
// than upstream's one-element array.
type DefaultCaseOptions struct {
	// CommentPattern is a regular expression source. A comment after the last case whose trimmed
	// text matches it excuses the missing default. Absent means upstream's `^no default$`,
	// case-insensitive.
	CommentPattern string `json:"commentPattern"`
}

// defaultCaseDefaultCommentPattern is upstream's `DEFAULT_COMMENT_PATTERN`, `/^no default$/iu`.
//
// Anchored at both ends and case-insensitive, so `// no default` and `// NO DEFAULT` excuse the
// switch while `// no default case here` does not. The `u` flag has no Go equivalent and needs
// none: the pattern contains no escapes whose meaning changes under Unicode mode.
var defaultCaseDefaultCommentPattern = regexp.MustCompile(`(?i)^no default$`)

var messageDefaultCaseMissingDefaultCase = rule.Message{
	Id: "missingDefaultCase",
	Description: "This `switch` has no `default` case, so a value matching none of its cases falls " +
		"through silently and the statement does nothing. That is occasionally what the author " +
		"meant and is far more often an unhandled case that was added to the type later and never " +
		"added here. The two readings are indistinguishable from the code, which is the problem: a " +
		"reader cannot tell a deliberate omission from a forgotten branch. Add a `default`, even " +
		"one that throws, or write a `// no default` comment after the last case to say the " +
		"omission is intended.",
}

// DefaultCase requires a `switch` statement to carry a `default` case.
//
//	valid:   switch (a) { case 1: break; default: break; }
//	valid:   switch (a) { case 1: break; // no default
//	         }
//	valid:   switch (a) { }
//	invalid: switch (a) { case 1: break; }
//	invalid: switch (a) { case 1: break; // no default here
//	         }
//
// Ported from `default-case` in ESLint, read from the clone at `lib/rules/default-case.js`. One
// option (`commentPattern`), one message, no fixer: `meta` carries `type`, `defaultOptions`, `docs`,
// `schema` and `messages`, and the string `fixable` does not appear in the file.
//
// The whole 23-case corpus was extracted from upstream's own tester and replayed against the
// installed ESLint through the Linter API before any code was written.
//
// # An empty switch is skipped, and the reason is upstream's rather than ours
//
// `switch (a) {}` is clean, and upstream says why in a comment at the line: there is "no easy way to
// extract comments inside it now". That is a limitation of what ESLint's source code object can
// answer, not a judgment that an empty switch is fine. We could answer it here -- the comment
// scanner has no such restriction -- but reproducing the DECISION means reproducing the silence,
// because a port that reports on empty switches disagrees with the corpus and with every project
// currently relying on the rule. Recorded rather than silently improved.
//
// # Where the excusing comment has to be, and why the answer is not "inside the switch"
//
// Upstream reads `sourceCode.getCommentsAfter(lastCase)` and takes the LAST of them. So the comment
// must follow the final case clause, and when several comments follow it only the last one is
// consulted:
//
//	switch (a) { case 1: break; /* fallthrough */ // no default
//	}                                             <- excused, the last comment matches
//	switch (a) { case 1: break; // no default
//	             /* something else */ }           <- reports, the last comment does not match
//
// Both shapes are in the corpus. Anchoring on "any comment inside the switch" passes most of it and
// fails those two, which is why the position is computed rather than approximated.
//
// # Comments are read from the shelf rather than rescanned
//
// `comments.ForFile` computes the file's comment list once and shares it with every rule that asks.
// A rule scanning the file itself is invisible to that cache and pays the whole cost alone; the
// shelf's own doc comment records three rules that each rescanned the file through one well-factored
// helper before it existed.
var DefaultCase = rule.Rule{
	Name: "default-case",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		pattern := defaultCaseDefaultCommentPattern
		if resolved, isDefaultCaseOptions := rule.OptionsAs[DefaultCaseOptions](options); isDefaultCaseOptions &&
			resolved.CommentPattern != "" {
			// Upstream builds `new RegExp(options.commentPattern, "u")`, which THROWS on a bad
			// pattern and takes the lint run down naming the rule. A pattern Go's RE2 cannot compile
			// is kept as the default here instead: the rule keeps enforcing rather than crashing,
			// and a crash costs every OTHER rule its verdict on that file because the walk recovers
			// per file. Same drop `core/no-param-reassign` takes for the same reason.
			if compiled, err := regexp.Compile(resolved.CommentPattern); err == nil {
				pattern = compiled
			}
		}

		return rule.Listeners{
			ast.KindSwitchStatement: func(node *ast.Node) {
				statement := node.AsSwitchStatement()
				if statement == nil || statement.CaseBlock == nil {
					return
				}
				block := statement.CaseBlock.AsCaseBlock()
				if block == nil || block.Clauses == nil {
					return
				}
				clauses := block.Clauses.Nodes

				// An empty switch is skipped. Upstream's reason is a limitation of its own source
				// code object rather than a judgment; see the rule doc.
				if len(clauses) == 0 {
					return
				}

				for _, clause := range clauses {
					if clause.Kind == ast.KindDefaultClause {
						return
					}
				}

				if defaultCaseHasExcusingComment(ctx, clauses[len(clauses)-1], statement.CaseBlock, pattern) {
					return
				}

				ctx.ReportNode(node, messageDefaultCaseMissingDefaultCase)
			},
		}
	},
}

// defaultCaseHasExcusingComment answers upstream's `getCommentsAfter(lastCase)` plus the pattern test.
//
// Upstream takes the LAST comment after the final case clause, not any of them, so a matching
// comment followed by a non-matching one does not excuse the switch. Both orderings are in the
// corpus.
//
// The search window is bounded by the case block's closing brace rather than running to the end of
// the file, which is what keeps a `// no default` comment written AFTER the switch from excusing it.
// No corpus case writes that shape, so the bound is a decision this port makes rather than one it
// inherits, and it is the conservative direction: reporting where upstream might not is visible,
// while silently excusing is not.
func defaultCaseHasExcusingComment(
	ctx rule.Context,
	lastClause *ast.Node,
	caseBlock *ast.Node,
	pattern *regexp.Regexp,
) bool {
	searchStart := lastClause.End()
	searchEnd := caseBlock.End()
	if searchEnd <= searchStart {
		return false
	}

	// Read the shared list ONCE. `ForFile` is cached per file, so calling it inside the loop is
	// not a rescan, but it is a map lookup and a closure call per comment for no reason.
	fileComments := comments.ForFile(ctx)

	var lastComment *comments.Comment
	for index := range fileComments {
		comment := &fileComments[index]
		if comment.Range.Pos() < searchStart || comment.Range.End() > searchEnd {
			continue
		}
		lastComment = comment
	}
	if lastComment == nil {
		return false
	}

	return pattern.MatchString(defaultCaseCommentBody(lastComment))
}

// defaultCaseCommentBody strips a comment's delimiters and trims it, matching upstream's
// `comment.value.trim()`.
//
// ESLint's `comment.value` is the text BETWEEN the delimiters; the shelf's `Text` includes them,
// because what counts as content differs by rule. So the delimiters come off here rather than in the
// shelf. A block comment's closing delimiter is removed only when present, since an unterminated
// block comment at end of file has none.
func defaultCaseCommentBody(comment *comments.Comment) string {
	body := comment.Text
	if comment.IsBlock {
		body = strings.TrimPrefix(body, "/*")
		body = strings.TrimSuffix(body, "*/")
	} else {
		body = strings.TrimPrefix(body, "//")
	}
	return text.TrimWhitespace(body)
}

// DefaultDefaultCaseOptions is the unconfigured answer.
//
// An empty `CommentPattern` means upstream's `^no default$`, which is what
// `options.commentPattern ? new RegExp(...) : DEFAULT_COMMENT_PATTERN` resolves to.
func DefaultDefaultCaseOptions() DefaultCaseOptions {
	return DefaultCaseOptions{}
}

// DecodeDefaultCaseOptions reads this rule's configuration from the config layer.
//
// Hand-rolled rather than `rule.DecodeOptionsInto` so a bare `"error"` configuration, which arrives
// as empty input, resolves to the documented default instead of erroring.
func DecodeDefaultCaseOptions(raw []byte) (any, error) {
	options := DefaultDefaultCaseOptions()
	if len(raw) == 0 {
		return options, nil
	}
	if err := rule.UnmarshalOptions(raw, &options); err != nil {
		return options, err
	}
	return options, nil
}
