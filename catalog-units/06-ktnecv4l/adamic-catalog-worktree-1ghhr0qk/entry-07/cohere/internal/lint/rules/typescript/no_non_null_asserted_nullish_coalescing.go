package typescript

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/lint/ecmascript/reference"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// NoNonNullAssertedNullishCoalescing flags `x!` on the left of `??`.
//
//	valid:   foo ?? bar;
//	valid:   foo ?? bar!;                      the assertion is on the RIGHT, which says nothing
//	valid:   (foo ?? bar)!;                    the assertion wraps the whole expression
//	valid:   let x: string; x! ?? '';          never assigned, so the assertion may be load-bearing
//	invalid: foo! ?? bar;
//	invalid: foo.bazz! ?? bar;
//	invalid: let x = foo(); x! ?? '';
//
// The two operators contradict each other. `??` exists to handle a left side that may be null or
// undefined; `!` asserts it is neither. Writing both says "this cannot be nullish, and here is what
// to do when it is", so one of the two is wrong and the reader cannot tell which.
//
// # The repair is a SUGGESTION, and upstream explains why in a comment worth preserving
//
// Removing `!` can change the resulting TYPE, so it is offered rather than applied:
//
//	function test(x?: string): string {
//	  const bar = x! ?? false;   the analysis gives bar the type string
//	  //          x  ?? false;   the analysis gives bar the type string | false
//	  return bar;
//	}
//
// The nullish coalesce takes the right operand into its result only when the left CAN be nullish, so
// deleting the assertion can widen the type and break a return that used to compile. A fixer would
// do that unattended. This ships the same single suggestion upstream ships.
//
// # An identifier is treated differently from every other left side, and the reason is scope
//
// For a non-identifier left side, the rule reports unconditionally: `foo.bazz!`, `foo()!` and
// `foo!.bazz!` are always contradictions.
//
// For a bare identifier, upstream asks whether the variable has been ASSIGNED before this point,
// and stays silent when it has not:
//
//	let x: string; x! ?? '';               silent, x has no value yet
//	let x!: string; x! ?? '';              reports, the definite assertion counts as an assignment
//	let x = foo(); x! ?? '';               reports, the initializer counts
//	let x: string; x = foo(); x! ?? '';    reports, the write is above
//	let x: string; x! ?? ''; x = foo();    silent, the write is BELOW
//
// The reasoning is that a declared-but-unassigned variable is genuinely `undefined` at that point,
// so the assertion is doing something and the pair is not obviously a mistake. Six of upstream's
// eighteen passing cases and seven of its fifteen failing ones ride entirely on this test, so it is
// most of the rule rather than an edge.
//
// # Where upstream gets that from, and where this port gets it instead
//
// Upstream reads ESLint's scope manager: `findVariable`, then `variable.references` for a write
// whose identifier ends before this node, and `variable.defs` for a declarator that is `definite`
// or has an `init`.
//
// This tree has no scope manager, and it does not need one for this question. The same two halves
// are available from the checker and the syntax tree:
//
//	the definition half   the declaration node itself carries `Initializer` and `ExclamationToken`,
//	                      which are exactly `init != null` and `definite`
//	the reference half    walk the file for identifiers resolving to the SAME SYMBOL, and ask the
//	                      shelf's `reference.WritesToBinding` whether each one is a write
//
// This is the technique `no-unassigned-vars` and `no-class-assign` already use for the same
// question, and it is fidelity to what the rule DECIDES rather than to how upstream obtains it.
//
// Symbols are compared by identity rather than by indexing `Declarations[0]`. A symbol can carry
// several declarations, and taking the first would answer about the wrong one for a merged name;
// comparing the symbols themselves asks the question the rule actually has, which is "is this the
// same binding", not "which declaration is it".
//
// # What the suggestion removes, and why it is found by scanning rather than by arithmetic
//
// The edit removes the `!` token. Its position is not `node.End() - 1`: the assertion's own text can
// end in trivia, and the corpus has a case with two spaces between the name and the mark, whose
// repair must delete the mark and leave the spaces. Upstream finds the token with
// `getLastToken(node, isNonNullAssertionPunctuator)` for the same reason. Here the token is located
// by scanning from the inner expression's end, which finds the same byte.
//
// # Cost
//
// The anchor is a binary expression, which is common, and the operator test rejects everything but
// `??` on the second line. The file walk runs only for a bare identifier under a non-null assertion
// on the left of a nullish coalesce, which is rare enough that the corpus is most of its occurrences.
var NoNonNullAssertedNullishCoalescing = rule.Rule{
	Name: "@typescript-eslint/no-non-null-asserted-nullish-coalescing",

	// Resolving the identifier to its binding is a checker question. Upstream asks its scope
	// manager instead, so seeing no `getParserServices` upstream is not evidence this can be
	// answered without types here.
	NeedsTypeChecker: true,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				binary := node.AsBinaryExpression()
				if binary.OperatorToken == nil ||
					binary.OperatorToken.Kind != ast.KindQuestionQuestionToken {
					return
				}

				// Parentheses are skipped, and this line was written the other way first on an
				// argument that sounded right and was false. Upstream's selector is
				// `> TSNonNullExpression.left`, a direct child relation, and the reasoned
				// conclusion was that a parenthesized left side cannot match it. estree has no
				// parenthesized node at all, so the child relation sees straight through one and
				// `(foo!) ?? bar` REPORTS. Measured against the installed build, after the same
				// mistake in a sibling rule in this batch produced a false negative no imported
				// fixture could see, since the corpus writes no parenthesized left side.
				//
				// SkipParentheses dereferences its argument, so the nil test comes first.
				if binary.Left == nil {
					return
				}
				assertion := ast.SkipParentheses(binary.Left)
				if !ast.IsNonNullExpression(assertion) {
					return
				}

				inner := assertion.AsNonNullExpression().Expression
				if inner == nil {
					return
				}

				// A bare identifier is the only shape that gets the assignment test. Everything
				// else is an unconditional contradiction.
				if ast.IsIdentifier(inner) && !hasAssignmentBeforeNonNullAssertion(ctx, inner, assertion) {
					return
				}

				exclamation, foundExclamation := nonNullExclamationRange(ctx, assertion, inner)
				if !foundExclamation {
					// The token has to be there for the node to have parsed as an assertion, so
					// this cannot be reached from any source that produced this node. Reporting
					// without the repair rather than reporting a repair over a guessed range,
					// because a suggestion anchored on the wrong bytes is worse than none.
					ctx.ReportNode(assertion, buildNoNonNullAssertedNullishCoalescingMessage())
					return
				}

				ctx.ReportNodeWithSuggestions(assertion,
					buildNoNonNullAssertedNullishCoalescingMessage(),
					rule.Suggestion{
						Message: buildSuggestRemovingNonNullMessage(),
						Fixes:   []rule.Fix{rule.RemoveRange(exclamation)},
					})
			},
		}
	},
}

// hasAssignmentBeforeNonNullAssertion answers upstream's `hasAssignmentBeforeNode`.
//
// Two halves, either of which is enough, matching upstream's `||`:
//
//	a WRITE to the binding whose identifier ends before this node ends
//	a DECLARATION with an initializer or a definite-assignment mark, ending before this node ends
//
// The comparison is on END positions in both halves, which is upstream's `range[1] < node.range[1]`
// rather than a start comparison. That distinction is visible in the corpus: a write BELOW the
// assertion leaves the case clean, so the direction is load-bearing.
//
// An identifier the checker cannot resolve answers TRUE, so the rule reports. That is upstream's
// behavior rather than a choice here: `findVariable` returns null for an unresolvable name, the
// `variable && ...` guard fails, and the early return is skipped. Every one of upstream's failing
// cases with a bare `foo!` relies on it, since `foo` is declared nowhere in them.
func hasAssignmentBeforeNonNullAssertion(ctx rule.Context, identifier *ast.Node, assertion *ast.Node) bool {
	if ctx.TypeChecker == nil {
		return false
	}

	symbol := ctx.TypeChecker.GetSymbolAtLocation(identifier)
	if symbol == nil {
		return true
	}

	for _, declaration := range symbol.Declarations {
		if !ast.IsVariableDeclaration(declaration) {
			continue
		}
		variableDeclaration := declaration.AsVariableDeclaration()
		definedWithValue := variableDeclaration.Initializer != nil ||
			variableDeclaration.ExclamationToken != nil
		if definedWithValue && declaration.End() < assertion.End() {
			return true
		}
	}

	sourceFile := ast.GetSourceFileOfNode(assertion)
	if sourceFile == nil {
		return false
	}

	// The whole file rather than the enclosing scope, matching the sibling rules that ask this
	// question: a write can sit inside a hoisted function above the declaration or nested
	// arbitrarily deep, and anchoring the match on the symbol makes its position irrelevant except
	// for the ordering test the rule actually cares about.
	name := identifier.Text()
	found := false
	var visit func(*ast.Node)
	visit = func(current *ast.Node) {
		if current == nil || found {
			return
		}
		// The text comparison is a pre-filter rather than a discrimination, since symbol identity
		// already implies it. It is here because it is far cheaper than a checker call and this
		// walk visits every identifier in the file.
		if current.Kind == ast.KindIdentifier && current.Text() == name &&
			current.End() < assertion.End() &&
			reference.WritesToBinding(current) &&
			ctx.TypeChecker.GetSymbolAtLocation(current) == symbol {
			found = true
			return
		}
		current.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return false
		})
	}
	visit(sourceFile.AsNode())

	return found
}

// nonNullExclamationRange locates the `!` token the suggestion removes.
//
// Scanning forward from the inner expression's end finds the mark itself rather than assuming it is
// the assertion's last byte. Upstream ships a case with two spaces between the name and the mark:
// the assertion's whole text is then the name, two spaces and the mark, so deleting a trailing byte
// happens to work while deleting one at `node.End()-1` on a spaced or commented form does not.
func nonNullExclamationRange(ctx rule.Context, assertion *ast.Node, inner *ast.Node) (core.TextRange, bool) {
	position := rule.TokenRange(ctx.SourceFile, inner).End()
	for position < assertion.End() {
		tokenRange := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, position)
		if tokenRange.End() <= tokenRange.Pos() {
			return core.TextRange{}, false
		}
		if ctx.SourceFile.Text()[tokenRange.Pos():tokenRange.End()] == "!" {
			return tokenRange, true
		}
		position = tokenRange.End()
	}
	return core.TextRange{}, false
}

func buildNoNonNullAssertedNullishCoalescingMessage() rule.Message {
	return rule.Message{
		Id: "noNonNullAssertedNullishCoalescing",
		Description: "The nullish coalescing operator is designed to handle undefined and null - " +
			"using a non-null assertion is not needed.",
	}
}

func buildSuggestRemovingNonNullMessage() rule.Message {
	return rule.Message{
		Id:          "suggestRemovingNonNull",
		Description: "Remove the non-null assertion.",
	}
}
