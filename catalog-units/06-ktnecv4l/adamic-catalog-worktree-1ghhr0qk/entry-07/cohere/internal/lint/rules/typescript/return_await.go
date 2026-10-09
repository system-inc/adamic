package typescript

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// ReturnAwaitMode is the rule's single option, a bare string rather than an object.
type ReturnAwaitMode string

const (
	// ReturnAwaitAlways requires every returned promise to be awaited.
	ReturnAwaitAlways ReturnAwaitMode = "always"
	// ReturnAwaitErrorHandlingCorrectnessOnly requires the await only where dropping it would
	// change which frame catches a rejection, and says nothing anywhere else.
	ReturnAwaitErrorHandlingCorrectnessOnly ReturnAwaitMode = "error-handling-correctness-only"
	// ReturnAwaitInTryCatch requires the await in error-handling contexts and forbids it elsewhere.
	ReturnAwaitInTryCatch ReturnAwaitMode = "in-try-catch"
	// ReturnAwaitNever forbids awaiting any returned promise.
	ReturnAwaitNever ReturnAwaitMode = "never"
)

// ReturnAwaitOptions is the rule's option surface.
type ReturnAwaitOptions struct {
	Mode ReturnAwaitMode
}

// DefaultReturnAwaitSettings is upstream's `defaultOptions`, which is NOT "always".
//
// Worth stating because the natural guess is wrong in a way that disagrees with the reference
// implementation on almost every input: under `always` every unawaited returned promise reports,
// while under the real default an unawaited return outside a try block is correct and an AWAITED
// one reports instead. A port that assumed `always` would invert the rule for ordinary code.
func DefaultReturnAwaitSettings() ReturnAwaitOptions {
	return ReturnAwaitOptions{Mode: ReturnAwaitInTryCatch}
}

// DecodeReturnAwaitOptions reads the rule's configuration.
//
// The option is a bare JSON string, not an object, so upstream's `["error"]` tuple becomes a lone
// string once cohere's config layer strips the severity. An unrecognised value keeps the default
// rather than silently selecting a mode nobody asked for.
func DecodeReturnAwaitOptions(raw []byte) (any, error) {
	decoded, err := rule.DecodeOptionsInto[string]()(raw)
	if err != nil {
		return DefaultReturnAwaitSettings(), err
	}

	wire, _ := decoded.(string)
	switch ReturnAwaitMode(wire) {
	case ReturnAwaitAlways, ReturnAwaitErrorHandlingCorrectnessOnly,
		ReturnAwaitInTryCatch, ReturnAwaitNever:
		return ReturnAwaitOptions{Mode: ReturnAwaitMode(wire)}, nil
	}
	return DefaultReturnAwaitSettings(), nil
}

// returnAwaitDisposition is what the rule wants for one returned expression in one context.
type returnAwaitDisposition int

const (
	returnAwaitWant returnAwaitDisposition = iota
	returnAwaitForbid
	returnAwaitDoNotCare
)

// dispositionsFor is upstream's `getConfiguration`, a two-column table per mode.
//
// The two columns are the whole rule. An "error handling context" is one where an exception thrown
// by the returned promise changes which frame catches it, and there the await is a correctness
// question. Everywhere else it is a style question, and the modes differ only in whether they have
// an opinion about the style half.
func dispositionsFor(mode ReturnAwaitMode) (errorHandling returnAwaitDisposition, ordinary returnAwaitDisposition) {
	switch mode {
	case ReturnAwaitAlways:
		return returnAwaitWant, returnAwaitWant
	case ReturnAwaitErrorHandlingCorrectnessOnly:
		return returnAwaitWant, returnAwaitDoNotCare
	case ReturnAwaitNever:
		return returnAwaitForbid, returnAwaitForbid
	default: // ReturnAwaitInTryCatch
		return returnAwaitWant, returnAwaitForbid
	}
}

// ReturnAwait enforces consistent awaiting of returned promises.
//
//	valid:   async function f() { return Promise.resolve(1); }              default mode
//	valid:   async function f() { try { return await p; } catch {} }
//	invalid: async function f() { return await Promise.resolve(1); }        disallowedPromiseAwait
//	invalid: async function f() { try { return p; } catch {} }              requiredPromiseAwait
//	invalid: async function f() { return await 1; }                         nonPromiseAwait
//
// # Why this is a correctness rule and not a style rule
//
// `return promise` inside a `try` hands the promise to the CALLER before it settles, so a rejection
// arrives after this function's frame is gone and the local `catch` never runs. `return await
// promise` keeps the frame alive until it settles, so the local `catch` sees the rejection. The two
// spellings differ in which handler runs, which is why the rule's default forbids the await in
// ordinary positions and requires it inside error handling.
//
// Everything below exists to answer two questions: is the returned expression thenable, and does
// this position affect error handling.
//
// # The thenable question goes to the checker, not to the syntax
//
// `NeedsToBeAwaited` gives three answers rather than two, and the middle one is load bearing:
// `Always` for a real thenable, `Never` for something that cannot be, and `May` for `any`,
// `unknown`, and an unconstrained type parameter. An awaited `May` is left alone entirely, because
// reporting it would tell someone to remove an await from a value that might well be a promise.
// Only a `Never` produces `nonPromiseAwait`.
//
// # The error-handling question is a tree walk, and it is recursive
//
// A return inside a `try` block always affects error handling, because a try is always followed by
// a catch or a finally. A return inside a `catch` affects it when the same statement has a
// `finally`, and otherwise the question repeats one level out. A return inside a `finally` always
// repeats one level out. That recursion is upstream's and is reproduced: a return in a catch with
// no finally, nested inside an outer try, is still an error-handling context by virtue of the outer
// one.
//
// Measured rather than assumed, because this is the half ports get wrong: our `TryStatement`
// exposes `TryBlock`, `CatchClause` and `FinallyBlock`, and comparing the walked child against each
// by identity reproduces upstream's three-way answer, including a return nested inside an `if`
// inside a `try`. A port that silently answers "not in a try" removes an await it must not remove,
// which is a behaviour change rather than a formatting one.
//
// # `using` declarations are the second error-handling context, and the trivia trap is here
//
// A `using` or `await using` declaration in scope means the disposal runs when the scope exits, so
// whether the promise settles inside that scope changes observable behaviour. Upstream asks ESLint's
// scope analysis; this port walks enclosing blocks up to the function boundary, which reaches the
// same answer because a `using` is a block-scoped declaration.
//
// The position test is the part worth reading twice. Upstream requires the declaration to end
// BEFORE the return begins, and `node.Pos()` in our tree includes leading trivia, so it starts
// where the previous statement ended. Comparing against it makes a `using` on the line directly
// above the return read as "not before", which is the single most common shape this feature has.
// Probed: three of the corpus's `using` cases answered wrongly until the comparison moved to
// `rule.TokenRange`, and the two negative controls (a `using` after the return, and one outside an
// inner function) answered correctly both ways, so the mistake was invisible from the negatives.
//
// # A fix or a suggestion, decided by whether the edit can change behaviour
//
// When the position does NOT affect error handling, adding or removing the await cannot change
// which handler runs, so upstream ships a fix and the engine applies it unattended. When it DOES
// affect error handling, the same edit is a judgment about control flow and upstream ships a
// suggestion instead. Collapsing the two would have the edit engine silently rewriting which frame
// catches a rejection. `nonPromiseAwait` is always a fix, because removing an await from a
// non-thenable cannot change control flow at all.
//
// # This port follows the CLONE's fixer, not the installed build's, and that is deliberate
//
// The clone is 8.68.0 and the installed build our differential compares against is 8.67.0. 8.68
// added a guard to the await removal: when the operand starts with `{` and the await is the whole
// body of an arrow function, the operand is wrapped in parentheses. Without it:
//
//	const test = async () => await { a: 1 };   ->   const test = async () => { a: 1 };
//
// Measured on the installed 8.67 build, which produces exactly that, and confirmed against the
// compiler: the rewritten body parses as a Block containing a labelled statement rather than as an
// object literal, so the function silently stops returning anything. The brief's standing rule is
// that the installed build is the oracle because the differential compares against it. That reason
// does not hold when the difference is a bug fix rather than a disagreement, since reproducing it
// would mean shipping a repair that corrupts source. The guard is ported and four of upstream's own
// cases record the corrected output.
//
// # Cost
//
// The anchor is a return statement plus the concise body of an async arrow, and the body exits
// before touching the checker unless the enclosing function is async and there is an argument.
var ReturnAwait = rule.Rule{
	Name: "@typescript-eslint/return-await",

	// Whether the returned expression is thenable decides every finding.
	NeedsTypeChecker: true,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		// NeedsTypeChecker governs registration; the harness builds a Context by hand and can hand
		// this rule a nil checker, which is how a sibling rule panicked. Declining the file once
		// here is cheaper than a guard in every listener.
		if ctx.TypeChecker == nil {
			return nil
		}

		settings, isSettings := rule.OptionsAs[ReturnAwaitOptions](options)
		if !isSettings {
			settings = DefaultReturnAwaitSettings()
		}
		errorHandlingDisposition, ordinaryDisposition := dispositionsFor(settings.Mode)

		// test is upstream's `test`, run once per possibly-returned expression.
		var test func(node *ast.Node)
		test = func(node *ast.Node) {
			isAwait := node.Kind == ast.KindAwaitExpression

			// Upstream types the await's OPERAND when there is one and the node itself otherwise,
			// then asks `needsToBeAwaited` about the outer node. The two arguments are different on
			// purpose: the type describes what is being returned, and the node supplies the
			// location the checker resolves against.
			subject := node
			if isAwait {
				subject = node.AsAwaitExpression().Expression
				if subject == nil {
					return
				}
			}

			subjectType := ctx.TypeChecker.GetTypeAtLocation(subject)
			if subjectType == nil {
				return
			}
			certainty := type_checking.NeedsToBeAwaited(ctx.TypeChecker, node, subjectType)

			if certainty != type_checking.TypeAwaitableAlways {
				if !isAwait {
					return
				}
				// `May` covers `any`, `unknown` and an unconstrained type parameter. Telling
				// someone to drop the await there would be advice about a value that might be a
				// promise, so upstream stays silent and so does this.
				if certainty == type_checking.TypeAwaitableMay {
					return
				}
				// Removing an await from something that cannot be a thenable cannot change control
				// flow, so this repair is always a fix rather than a suggestion.
				if fixes, canFix := removeAwaitFixes(ctx, node); canFix {
					ctx.ReportNodeWithFixes(node, returnAwaitMessage("nonPromiseAwait"), fixes...)
				} else {
					ctx.ReportNode(node, returnAwaitMessage("nonPromiseAwait"))
				}
				return
			}

			affectsErrorHandling := affectsExplicitErrorHandling(node) ||
				usingDeclarationInScope(ctx, node)

			disposition := ordinaryDisposition
			if affectsErrorHandling {
				disposition = errorHandlingDisposition
			}

			switch disposition {
			case returnAwaitDoNotCare:
				return
			case returnAwaitWant:
				if isAwait {
					return
				}
				insertion := insertAwaitFixes(ctx, node)
				if affectsErrorHandling {
					// The edit changes which frame catches a rejection, so a person chooses it.
					ctx.ReportNodeWithSuggestions(node, returnAwaitMessage("requiredPromiseAwait"),
						rule.Suggestion{
							Message: returnAwaitMessage("requiredPromiseAwaitSuggestion"),
							Fixes:   insertion,
						})
					return
				}
				ctx.ReportNodeWithFixes(node, returnAwaitMessage("requiredPromiseAwait"), insertion...)
			case returnAwaitForbid:
				if !isAwait {
					return
				}
				fixes, canFix := removeAwaitFixes(ctx, node)
				if !canFix {
					ctx.ReportNode(node, returnAwaitMessage("disallowedPromiseAwait"))
					return
				}
				if affectsErrorHandling {
					ctx.ReportNodeWithSuggestions(node, returnAwaitMessage("disallowedPromiseAwait"),
						rule.Suggestion{
							Message: returnAwaitMessage("disallowedPromiseAwaitSuggestion"),
							Fixes:   fixes,
						})
					return
				}
				ctx.ReportNodeWithFixes(node, returnAwaitMessage("disallowedPromiseAwait"), fixes...)
			}
		}

		// testPossiblyReturned is upstream's `findPossiblyReturnedNodes`: a conditional returns
		// either arm, so both are judged and either can report on its own.
		var testPossiblyReturned func(node *ast.Node)
		testPossiblyReturned = func(node *ast.Node) {
			if node == nil {
				return
			}

			// Our parser keeps parentheses and upstream's folds them away, so `(await x)` arrives
			// here as a ParenthesizedExpression that upstream never sees. Without this the arm is
			// never judged at all: measured, three of upstream's own reporting cases went silent,
			// including `const test = async () => (await { a: 1 })`.
			//
			// A loop rather than one step, because `((x))` nests. Not `ast.SkipParentheses`, which
			// dereferences its argument and would panic on the nil this function is handed for an
			// omitted return argument.
			for node.Kind == ast.KindParenthesizedExpression {
				inner := node.AsParenthesizedExpression().Expression
				if inner == nil {
					return
				}
				node = inner
			}

			if node.Kind == ast.KindConditionalExpression {
				conditional := node.AsConditionalExpression()
				// Upstream builds `[...findPossiblyReturnedNodes(alternate),
				// ...findPossiblyReturnedNodes(consequent)]` and then reports over that array, so
				// the ALTERNATE's findings are collected first. But the reports come out in source
				// order because ESLint sorts diagnostics by position before emitting them, and our
				// harness preserves emission order instead.
				//
				// Measured against the installed build: `(await foo()) ? bar() : baz()` reports
				// bar() first. So the consequent is visited first here, which reproduces the
				// OBSERVED order rather than upstream's array order.
				testPossiblyReturned(conditional.WhenTrue)
				testPossiblyReturned(conditional.WhenFalse)
				return
			}
			test(node)
		}

		return rule.Listeners{
			ast.KindReturnStatement: func(node *ast.Node) {
				enclosing := enclosingFunctionOfReturn(node)
				if enclosing == nil || !isAsyncFunction(enclosing) {
					return
				}
				testPossiblyReturned(node.AsReturnStatement().Expression)
			},
			ast.KindArrowFunction: func(node *ast.Node) {
				// The concise body of an async arrow is a return position with no return
				// statement, so it needs its own anchor. A block-bodied arrow is covered by the
				// return listener above.
				arrow := node.AsArrowFunction()
				if !isAsyncFunction(node) || arrow.Body == nil || arrow.Body.Kind == ast.KindBlock {
					return
				}
				testPossiblyReturned(arrow.Body)
			},
		}
	},
}

// isAsyncFunction answers whether a function-like node carries the async modifier.
func isAsyncFunction(node *ast.Node) bool {
	return node.ModifierFlags()&ast.ModifierFlagsAsync != 0
}

// enclosingFunctionOfReturn finds the function a return statement belongs to.
//
// Upstream keeps a stack pushed by its enter/exit handlers and reads the top; the walk here has no
// exit hook, so the same answer comes from walking up to the first function-like ancestor. A return
// is only legal inside a function, so a nil result means the parse recovered from bad source.
func enclosingFunctionOfReturn(node *ast.Node) *ast.Node {
	for ancestor := node.Parent; ancestor != nil; ancestor = ancestor.Parent {
		if ast.IsFunctionLike(ancestor) {
			return ancestor
		}
	}
	return nil
}

// affectsExplicitErrorHandling answers whether an exception from this position changes control flow.
//
// Upstream's recursion, reproduced. A `try` block always qualifies because a try is always followed
// by a catch or a finally. A `catch` qualifies when the same try statement has a finally, and
// otherwise the question repeats around the enclosing statement. A `finally` always repeats.
func affectsExplicitErrorHandling(node *ast.Node) bool {
	for current := node; current != nil; {
		tryStatement, part := containingTryStatement(current)
		if tryStatement == nil {
			return false
		}
		switch part {
		case tryPartTry:
			return true
		case tryPartCatch:
			if tryStatement.AsTryStatement().FinallyBlock != nil {
				return true
			}
			current = tryStatement
		case tryPartFinally:
			current = tryStatement
		default:
			return false
		}
	}
	return false
}

// tryPart names which of a try statement's three children contains the node.
type tryPart int

const (
	tryPartNone tryPart = iota
	tryPartTry
	tryPartCatch
	tryPartFinally
)

// containingTryStatement finds the nearest enclosing try statement, without crossing a function.
//
// The child being carried up is what identifies the part: upstream compares it against `tryBlock`,
// `catchClause` and `finallyBlock` by identity, and our AST exposes the same three fields, so the
// comparison ports directly. Probed across all three parts plus a return nested inside an `if`
// inside a try, with a control that answers "not in a try" for an ordinary return.
func containingTryStatement(node *ast.Node) (*ast.Node, tryPart) {
	child := node
	for ancestor := node.Parent; ancestor != nil && !ast.IsFunctionLike(ancestor); ancestor = ancestor.Parent {
		if ancestor.Kind == ast.KindTryStatement {
			statement := ancestor.AsTryStatement()
			switch child {
			case statement.TryBlock:
				return ancestor, tryPartTry
			case statement.CatchClause:
				return ancestor, tryPartCatch
			case statement.FinallyBlock:
				return ancestor, tryPartFinally
			}
			return ancestor, tryPartNone
		}
		child = ancestor
	}
	return nil, tryPartNone
}

// usingDeclarationInScope answers upstream's `affectsExplicitResourceManagement`.
//
// A `using` or `await using` declaration disposes its resource when the scope exits, so whether the
// returned promise settles inside that scope is observable. The declaration has to be positioned
// before the node, and the search stops at the function boundary because a `using` in an outer
// function disposes on that function's exit rather than this one's.
//
// The comparison uses the node's TOKEN start rather than `Pos()`: `Pos()` includes leading trivia
// and begins where the previous statement ended, which makes a `using` on the line directly above
// read as "not before". See the rule's doc comment for the measurement.
func usingDeclarationInScope(ctx rule.Context, node *ast.Node) bool {
	// TokenRange rather than Pos, and the difference is EQUIVALENT for every input this rule can
	// produce rather than merely untested. Stated because a mutation swapping them survives the
	// whole suite, and the honest reason is not "no fixture covers it".
	//
	// `Pos()` includes leading trivia, so for a RETURN STATEMENT it starts where the previous
	// statement ended, and comparing a `using` against that reads "not before". Measured: 97
	// against 100 on a return directly below a `using`. But this function is handed the returned
	// EXPRESSION, never the statement, and an expression's leading trivia can only reach back to
	// the `return` keyword, which always sits between it and any preceding statement. Measured on
	// the tightest shape available, a `using` and a return on ONE line: the using ends at 97 while
	// the expression's Pos is 104 and its TokenRange 105, both after it.
	//
	// So no input separates the two spellings here. TokenRange is kept because it states the
	// intent and because a later caller passing the statement would make the difference real.
	nodeStart := rule.TokenRange(ctx.SourceFile, node).Pos()

	for ancestor := node.Parent; ancestor != nil; ancestor = ancestor.Parent {
		for _, statement := range blockStatementsOf(ancestor) {
			if statement.Kind != ast.KindVariableStatement {
				continue
			}
			list := statement.AsVariableStatement().DeclarationList
			if list == nil {
				continue
			}
			// Test the Using bit ALONE. `NodeFlagsAwaitUsing` is not a separate bit: it is
			// defined as `NodeFlagsConst | NodeFlagsUsing`, because on one node those two are
			// otherwise mutually exclusive. So a mask of `Using|AwaitUsing` reduces to
			// `Const|Using` and matches every `const` declaration in the file.
			//
			// That is not a near miss. It made an ordinary `const res = await ...` above a try
			// block turn the whole function into an error-handling context, which silenced a
			// finding upstream reports. Two of upstream's own cases caught it, and the single
			// `using` fixtures all still passed, because they satisfy either spelling.
			if list.Flags&ast.NodeFlagsUsing == 0 {
				continue
			}
			if statement.End() < nodeStart {
				return true
			}
		}
		if ast.IsFunctionLike(ancestor) {
			return false
		}
	}
	return false
}

// blockStatementsOf reads the statement list of anything that holds one.
func blockStatementsOf(node *ast.Node) []*ast.Node {
	switch node.Kind {
	case ast.KindBlock:
		if list := node.AsBlock().Statements; list != nil {
			return list.Nodes
		}
	case ast.KindSourceFile:
		if list := node.AsSourceFile().Statements; list != nil {
			return list.Nodes
		}
	// One accessor covers both kinds: the parser gives a case clause and a default clause the same
	// CaseOrDefaultClause type, so `AsCaseClause` does not exist. Grepped rather than guessed after
	// the natural spelling failed to compile.
	case ast.KindCaseClause, ast.KindDefaultClause:
		if list := node.AsCaseOrDefaultClause().Statements; list != nil {
			return list.Nodes
		}
	}
	return nil
}

// removeAwaitFixes builds the repair that deletes an `await` keyword.
//
// The removal runs from the keyword's start to the start of whatever follows it, so the separating
// whitespace goes too and `await  foo()` does not become `  foo()`.
//
// The parenthesis guard is 8.68's and is the reason this port follows the clone rather than the
// installed 8.67 build. When the operand begins with `{` and the await is an arrow's whole body,
// removing the keyword leaves a body that parses as a Block rather than an object literal, so the
// function silently returns nothing. Confirmed with the compiler and against the running 8.67 rule,
// which produces exactly that corruption.
func removeAwaitFixes(ctx rule.Context, node *ast.Node) ([]rule.Fix, bool) {
	if node.Kind != ast.KindAwaitExpression {
		return nil, false
	}
	operand := node.AsAwaitExpression().Expression
	if operand == nil {
		return nil, false
	}

	awaitStart := rule.TokenRange(ctx.SourceFile, node).Pos()
	operandStart := rule.TokenRange(ctx.SourceFile, operand).Pos()
	if operandStart <= awaitStart {
		return nil, false
	}

	// Upstream removes to the start of the next token INCLUDING COMMENTS, not to the operand's
	// token start, so a comment between the keyword and the operand survives the repair:
	//
	//	await /* comment */ 1   ->   /* comment */ 1
	//
	// Removing to the operand's token start instead swallows it. Two of upstream's own cases assert
	// the comment is kept, and both failed before this.
	// The scan starts at the OPERAND's Pos rather than at the await keyword's start. `Pos()`
	// includes leading trivia, so for `await /* c */ 1` it sits just after the keyword, which is
	// where the comment actually is. Starting at the keyword instead makes the scanner read the
	// keyword's own text and find nothing: probed, it returned zero comments for a gap that
	// visibly contained one.
	removalEnd := operandStart
	for comment := range type_checking.GetCommentsInRange(ctx.SourceFile,
		core.NewTextRange(operand.Pos(), operandStart)) {
		if comment.Pos() > awaitStart && comment.Pos() < removalEnd {
			removalEnd = comment.Pos()
		}
	}

	fixes := []rule.Fix{rule.RemoveRange(core.NewTextRange(awaitStart, removalEnd))}

	if operandNeedsParenthesesAfterAwaitRemoval(ctx, node, operand) {
		fixes = append(fixes,
			rule.ReplaceRange(core.NewTextRange(operandStart, operandStart), "("),
			rule.ReplaceRange(core.NewTextRange(operand.End(), operand.End()), ")"))
	}
	return fixes, true
}

// operandNeedsParenthesesAfterAwaitRemoval answers 8.68's `shouldWrapInParentheses`.
//
// Two conditions together: the operand's first token is `{`, and the await expression is the whole
// concise body of an arrow function. Either alone is harmless, since `{` elsewhere in an expression
// position stays an object literal and an arrow body that is not an object literal cannot be
// reparsed as a block.
func operandNeedsParenthesesAfterAwaitRemoval(ctx rule.Context, node *ast.Node, operand *ast.Node) bool {
	// Upstream tests the operand's FIRST TOKEN, not the operand's kind, and the difference is a
	// real case rather than a nicety: `await { a: 1 }.a` is a property access whose first token is
	// still `{`, so dropping the keyword leaves an arrow body that parses as a block. A kind test
	// misses it, which one of upstream's own fix vectors caught.
	if !operandStartsWithBrace(ctx, operand) {
		return false
	}
	parent := node.Parent
	if parent == nil || parent.Kind != ast.KindArrowFunction {
		return false
	}
	return parent.AsArrowFunction().Body == node
}

// operandStartsWithBrace answers whether an expression's own text begins with `{`.
//
// Reading the source rather than walking down the leftmost child: the leftmost node of
// `{ a: 1 }.a` is the object literal, but so is the leftmost node of shapes that do not start with
// a brace at all once parentheses are involved, and the question upstream asks is purely textual.
func operandStartsWithBrace(ctx rule.Context, operand *ast.Node) bool {
	start := rule.TokenRange(ctx.SourceFile, operand).Pos()
	text := ctx.SourceFile.Text()
	return start >= 0 && start < len(text) && text[start] == '{'
}

// insertAwaitFixes builds the repair that adds an `await` keyword.
//
// An operand that binds more tightly than `await` takes the keyword and a space. Anything looser
// has to be parenthesised, or `await a || b` would parse as `(await a) || b` and change what is
// returned.
func insertAwaitFixes(ctx rule.Context, node *ast.Node) []rule.Fix {
	start := rule.TokenRange(ctx.SourceFile, node).Pos()
	if type_checking.IsHigherPrecedenceThanAwait(node) {
		return []rule.Fix{rule.ReplaceRange(core.NewTextRange(start, start), "await ")}
	}
	return []rule.Fix{
		rule.ReplaceRange(core.NewTextRange(start, start), "await ("),
		rule.ReplaceRange(core.NewTextRange(node.End(), node.End()), ")"),
	}
}

// returnAwaitMessage renders one of the rule's five messages.
func returnAwaitMessage(messageId string) rule.Message {
	switch messageId {
	case "disallowedPromiseAwait":
		return rule.Message{Id: messageId,
			Description: "Returning an awaited promise is not allowed in this context."}
	case "disallowedPromiseAwaitSuggestion":
		return rule.Message{Id: messageId,
			Description: "Remove `await` before the expression. Use caution as this may impact control flow."}
	case "nonPromiseAwait":
		return rule.Message{Id: messageId,
			Description: "Returning an awaited value that is not a promise is not allowed."}
	case "requiredPromiseAwaitSuggestion":
		return rule.Message{Id: messageId,
			Description: "Add `await` before the expression. Use caution as this may impact control flow."}
	default:
		return rule.Message{Id: "requiredPromiseAwait",
			Description: "Returning an awaited promise is required in this context."}
	}
}
