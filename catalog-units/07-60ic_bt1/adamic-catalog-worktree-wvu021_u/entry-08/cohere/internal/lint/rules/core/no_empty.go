package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// NoEmptyOptions tunes whether an empty `catch` is allowed.
type NoEmptyOptions struct {
	// AllowEmptyCatch permits `catch (error) {}` with nothing in it and no comment either.
	//
	// Off by default, matching ESLint. The rule's usual escape hatch is a comment saying why the
	// error is being dropped, and this option is for a codebase that has decided the empty catch is
	// idiomatic enough not to need one each time.
	AllowEmptyCatch bool `json:"allowEmptyCatch"`
}

var messageEmptyBlock = rule.Message{
	Id: "unexpectedBlock",
	Description: "This block is empty, which leaves a reader unable to tell a deliberate no-op from " +
		"a body that was deleted or never written. The distinction matters most in a `catch`, where " +
		"an empty block is the difference between swallowing an error on purpose and swallowing it by " +
		"accident. Write a comment saying why nothing happens here, or remove the block.",
}

var messageEmptySwitch = rule.Message{
	Id: "unexpectedSwitch",
	Description: "This switch has no clauses at all, so it evaluates its subject and does nothing " +
		"with the result. Either the arms were removed and the statement outlived them, or they were " +
		"never written. Add the clauses, or remove the switch and keep whatever the subject " +
		"expression was doing.",
}

// NoEmpty flags a block statement with nothing in it.
//
//	valid:   if (foo) { /* nothing to do */ }
//	valid:   function noop() {}
//	valid:   try { work(); } catch (error) { // reported elsewhere
//	         }
//	invalid: if (foo) {}
//	invalid: while (foo) {}
//	invalid: switch (foo) {}
//	invalid: try { work(); } catch (error) {}
//
// An empty block reads as unfinished. A reader cannot tell whether the body was deleted during a
// refactor, was never written, or is empty on purpose, and the three call for opposite responses.
// The escape hatch is a comment, which turns the ambiguity into a statement of intent, and it is why
// the rule ignores any block containing one.
//
// A function body is exempt and that is not an oversight: an empty function is a real thing to write.
// A no-op callback, a stub satisfying an interface, and a constructor that only declares parameter
// properties are all deliberate, and the block is the only way to write them. ESLint splits that
// concern into `no-empty-function`, which is a separate opinion a codebase can hold separately.
//
// An empty `switch` is included because it has the same defect through a different node: a switch
// with no clauses evaluates its subject and discards the result. ESLint reports the switch statement
// itself there rather than the case block, and so do we.
//
// The comment check reads the parser's own trivia rather than searching the source text for `//` or
// `/*`, which is a choice about what the code can be trusted to keep doing rather than a bug fix
// today. Both answers agree on every input reachable now: a block with no statements contains no
// string or regular expression to hide a marker in, and a switch's CaseBlock node was measured to
// span only its own braces rather than the discriminant that precedes it, so the text a substring
// search would read is `{}` and nothing else. The scanner is used anyway because it asks the
// question the rule means, at the same cost. A substring search is correct here only because of a
// span boundary nothing in the rule controls, and if a future node were scanned whose text did
// reach further, the approximation would start reporting nothing on code carrying `/*` inside a
// string while looking exactly as correct as it does now.
//
// No fix and no suggestion. Deleting an empty block changes control flow wherever the block is a
// statement body, and adding an explanatory comment means knowing the intent that was never written
// down. Both repairs are the author's.
var NoEmpty = rule.Rule{
	Name: "no-empty",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		settings := NoEmptyOptions{}
		if configured, isConfigured := rule.OptionsAs[NoEmptyOptions](options); isConfigured {
			settings = configured
		}

		return rule.Listeners{
			ast.KindBlock: func(node *ast.Node) {
				block := node.AsBlock()
				if block == nil {
					return
				}
				if block.Statements != nil && len(block.Statements.Nodes) > 0 {
					return
				}

				// A function's body is that function's own concern, which is where ESLint draws the
				// line and where `no-empty-function` picks it up.
				if node.Parent != nil && ast.IsFunctionLike(node.Parent) {
					return
				}

				if settings.AllowEmptyCatch && node.Parent != nil && node.Parent.Kind == ast.KindCatchClause {
					return
				}

				if blockContainsComment(ctx.SourceFile, node) {
					return
				}

				ctx.ReportNode(node, messageEmptyBlock)
			},

			ast.KindSwitchStatement: func(node *ast.Node) {
				switchStatement := node.AsSwitchStatement()
				if switchStatement == nil || switchStatement.CaseBlock == nil {
					return
				}
				caseBlock := switchStatement.CaseBlock.AsCaseBlock()
				if caseBlock != nil && caseBlock.Clauses != nil && len(caseBlock.Clauses.Nodes) > 0 {
					return
				}

				if blockContainsComment(ctx.SourceFile, switchStatement.CaseBlock) {
					return
				}

				ctx.ReportNode(node, messageEmptySwitch)
			},
		}
	},
}

// blockContainsComment says whether an otherwise-empty braced node holds a comment.
//
// The node has no children to walk, so there is exactly one place a comment can be: the leading
// trivia of the closing brace, which is the next token after the opening one. Scanning from
// node.End() minus the brace would find trailing comments outside the block instead, so the scan
// starts just past the opening brace and reads forward.
//
// This is the parser's own trivia rather than a search for comment markers in the text, which is what
// makes it immune to a `/*` living inside a string or a regular expression earlier in the node.
func blockContainsComment(sourceFile *ast.SourceFile, node *ast.Node) bool {
	if sourceFile == nil || node == nil {
		return false
	}
	sourceText := sourceFile.Text()

	openingBrace := findOpeningBrace(sourceFile, node)
	if openingBrace < 0 {
		return false
	}

	var factory ast.NodeFactory
	for range scanner.GetLeadingCommentRanges(&factory, sourceText, openingBrace+1) {
		return true
	}
	for range scanner.GetTrailingCommentRanges(&factory, sourceText, openingBrace+1) {
		return true
	}
	return false
}

// findOpeningBrace returns the offset of the brace that opens a node, or -1.
//
// The search starts at the node's first token rather than at Pos(), so a comment sitting in the
// node's leading trivia is never mistaken for the opening brace's position and never read as being
// inside the block.
func findOpeningBrace(sourceFile *ast.SourceFile, node *ast.Node) int {
	sourceText := sourceFile.Text()
	start := rule.TokenRange(sourceFile, node).Pos()
	end := node.End()
	if start < 0 || end > len(sourceText) || start >= end {
		return -1
	}
	for offset := start; offset < end; offset++ {
		if sourceText[offset] == '{' {
			return offset
		}
	}
	return -1
}
