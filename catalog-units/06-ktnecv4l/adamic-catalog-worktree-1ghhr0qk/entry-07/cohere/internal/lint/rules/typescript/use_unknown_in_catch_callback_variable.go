package typescript

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// UseUnknownInCatchCallbackVariable flags a rejection callback whose error parameter is typed as
// anything other than `unknown`.
//
//	valid:   Promise.resolve().catch((err: unknown) => { throw err; })
//	valid:   Promise.resolve().catch(() => { throw new Error(); })
//	valid:   declare const handler: (x: any) => void; Promise.reject().catch(handler)
//	invalid: Promise.resolve().catch((err: Error) => { throw err; })
//	invalid: Promise.resolve().catch(err => { throw err; })
//	invalid: Promise.resolve().then(() => {}, error => {})
//
// A rejection value is whatever was thrown, and anything can be thrown, so a callback annotating it
// as `Error` is asserting something the language does not guarantee. `unknown` forces the body to
// narrow before it reads a property, which is the whole point.
//
// # What decides a finding
//
// Two gates before the callback is even looked at. The callee must be a static member access naming
// `catch` or `then`, and the RECEIVER's type must be thenable. The name alone is not enough: the
// corpus writes `declare const notAPromise: { catch: (f: Function) => void }` precisely so that a
// port keying on the name reports it and fails.
//
// Which argument is checked depends on the method: `catch` takes the handler first, `then` takes it
// second. Upstream carries that as a table of two rows and this reproduces it as one, because the
// `append` column is the only other thing that varies and it is derivable from the method.
//
// The verdict on the callback itself comes from the CHECKER rather than from the annotation's text.
// `isFlaggableHandlerType` walks the union parts of the argument's type, and for each call signature
// asks whether the first parameter's type is intrinsic `unknown`. That is why `type U = unknown` is
// clean while `any` reports: the alias resolves to the intrinsic and the flag test sees through it.
// Measured on the installed rule, both directions.
//
// A signature with NO first parameter is skipped rather than reported, so `catch(() => {})` is
// clean. A union part with no call signatures is skipped too, which is what makes
// `(() => void) | 2` clean.
//
// # Why this walks the argument instead of testing it
//
// The argument can be an expression that RESOLVES to several function literals, and upstream reports
// each one separately. `collectFlaggedNodes` recurses through a logical expression into both sides,
// through a conditional into both branches, and through a sequence into its LAST element only.
// Measured: `catch((() => {}, (err: Error) => {}))` reports and `catch(((err: Error) => {}, () => {}))`
// is clean, which pins the last-element rule rather than leaving it read off the source.
//
// Anything else terminates the walk, and that is a deliberate narrowness rather than an oversight.
// A bare identifier bound to a badly-typed handler is CLEAN upstream, with a comment and an issue
// link in the corpus saying so: the rule only wants function literals written at the call site,
// because those are the ones the author can fix here. A type assertion is likewise not a literal and
// is clean.
//
// # Parentheses, which are two decisions here rather than one
//
// ESTree has no parenthesis node, so upstream's walk never mentions parens and gets paren
// transparency for free. Our AST makes them real nodes, so every place upstream sees through them
// this must skip explicitly, and there are three: the argument itself, each recursion into a
// sub-expression, and the callee. Measured against the installed rule, all three are transparent
// there: `catch((err => {}))`, `(p.catch)(err => {})`, and the corpus's own
// `(condition && (err => {})) || (err => {})` all report.
//
// The skips are NOT free correctness. `ast.SkipParentheses` dereferences its argument, so each call
// is guarded by the caller having a non-nil node, and `useUnknownSkipParentheses` exists to make
// that a property of the helper rather than of every call site.
//
// # Where the finding points, which is never the callback
//
// Upstream reports `node: argument` and then unconditionally overrides it from
// `refineReportIfPossible`, whose switch covers every shape a first parameter can take. So in
// practice the span is always the FIRST PARAMETER, and the callback span is unreachable. Verified
// against the installed rule across identifier, annotated, optional, defaulted, rest, array-pattern
// and object-pattern parameters: all seven report the parameter.
//
// This port therefore reports the parameter directly and has no callback-span arm to be wrong about.
// The one place upstream's fallthrough could fire is a first parameter that is none of those kinds,
// and TypeScript has no such shape.
//
// # The `this` parameter, an upstream defect reproduced deliberately
//
// A `this` parameter is not a real parameter: the CHECKER's signature omits it, while both ESTree's
// `params` and our `Parameters()` include it. Upstream reads the checker to DECIDE (so it judges the
// first real parameter) and reads the syntax to REPORT (so it points at `this`). The two disagree,
// and the suggestion it offers would write `function (this: unknown, err: Error)`, which is wrong
// code.
//
// Measured on the installed rule, with controls that pin the mechanism rather than the symptom:
//
//	p.catch(function (this: W, err: Error) {})    reports, span "this: W", suggests ": unknown" over ": W"
//	p.catch(function (this: W, err: unknown) {})  CLEAN, which proves the decision reads `err`
//	p.catch(function (this: W) {})                CLEAN, which proves `this` is no checker parameter
//
// Our AST agrees with ESTree here (probed: `Parameters()` has two entries, the checker signature has
// one), so reproducing this costs nothing and diverging would cost fidelity. It is reproduced, and
// the fixture pinning it says it is a defect so the next reader does not helpfully correct it.
//
// # The suggestions, and why there is no fix
//
// Upstream declares `hasSuggestions` and no `fixable`, so every repair here needs a human. That is
// right: rewriting a callback's error type changes what the body is allowed to do with it, and the
// body will usually stop compiling until the author narrows. The engine must not do that unattended.
//
// The annotation-replacement span is the one piece of arithmetic that is not a node. Our type node
// starts AFTER the colon and carries leading trivia, so replacing it would produce `err:: unknown`
// or eat a comment. The span upstream replaces is the colon through the type, which here is
// "just past the name, or just past the `?` if there is one" to the type's end. Measured against the
// installed rule's own byte ranges: `(err?: string)` replaces [29,37] and `(err: any /* c1 */ = 2)`
// replaces [28,33], and both are reproduced by that arithmetic.
var UseUnknownInCatchCallbackVariable = rule.Rule{
	Name: "@typescript-eslint/use-unknown-in-catch-callback-variable",

	// Every verdict comes from the checker: whether the receiver is thenable, and whether the
	// handler's first parameter is intrinsic `unknown`. Neither is answerable from syntax.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		// Declines the file once rather than per node. Unreachable through registration, since
		// NeedsTypeChecker is declared above; this covers the harness path, where a Context can be
		// built by hand. Without it every listener below would read a nil checker and go silent,
		// which is the vacuous-green failure rather than a visible one.
		if ctx.TypeChecker == nil {
			return nil
		}

		// isFlaggableHandlerType answers upstream's predicate of the same name: does this type have
		// any call signature whose first parameter is not `unknown`.
		//
		// The two `continue`s are the whole discrimination and neither is defensive. A union part
		// with no call signatures is not this rule's problem, which is what keeps
		// `(() => void) | 2` clean. A signature with no parameters cannot have a badly typed catch
		// variable, which is what keeps `catch(() => {})` clean.
		//
		// The UNION WALK is inert here and is kept for fidelity rather than for effect. Mutating it
		// to read only the first part survives the whole corpus, and that verdict is correct rather
		// than a fixture gap: this is only ever called on a node `collectFlaggedCallbacks` has
		// already narrowed to an arrow or a function expression, and a function LITERAL's own type
		// is its own signature. Probed over five shapes, printing the part count directly: a plain
		// arrow, both arms of a conditional, an `as` assertion to a union of two signatures, an
		// angle-bracket assertion to the same union, and a double assertion to an overloaded
		// interface all answer `unionParts=1, isUnion=false`. An assertion does not change what
		// `GetTypeAtLocation` says about the literal underneath it.
		//
		// The enumeration that verdict rests on is one call site, the arrow-and-function-expression
		// arm of `collectFlaggedCallbacks` below. If a caller is ever added that passes a type read
		// from something other than a literal, such as a bare identifier's type, this walk becomes
		// live again and the verdict here is void.
		//
		// Upstream's structure is kept because the cost is one loop over a one-element slice and
		// removing it would make this rule and its reference diverge in shape for no gain, which is
		// the harder thing to re-derive later.
		isFlaggableHandlerType := func(handlerType *checker.Type) bool {
			if handlerType == nil {
				return false
			}
			for unionPart := range type_checking.UnionTypePartsSeq(handlerType) {
				for _, callSignature := range type_checking.GetCallSignatures(ctx.TypeChecker, unionPart) {
					parameters := checker.Signature_parameters(callSignature)
					if len(parameters) == 0 {
						continue
					}
					firstParameter := parameters[0]
					firstParameterType := checker.Checker_getTypeOfSymbol(ctx.TypeChecker, firstParameter)
					if firstParameterType == nil {
						continue
					}

					// A rest parameter holds the whole argument list, so the type to judge is its
					// ELEMENT type rather than the container. A rest parameter that is neither an
					// array nor a tuple cannot be spread from a well-typed call at all, and
					// upstream flags it outright rather than trying to look inside.
					if declaration := firstParameter.ValueDeclaration; declaration != nil &&
						type_checking.IsRestParameterDeclaration(declaration) {
						switch {
						case checker.Checker_isArrayType(ctx.TypeChecker, firstParameterType),
							checker.IsTupleType(firstParameterType):
							typeArguments := checker.Checker_getTypeArguments(ctx.TypeChecker, firstParameterType)
							if len(typeArguments) == 0 {
								// `...args: []` has no element type to judge, so there is no badly
								// typed catch variable here. Upstream indexes [0] unguarded and
								// would read undefined, which its `isIntrinsicUnknownType` then
								// answers false for, reaching the same verdict by a route that
								// would panic in Go.
								continue
							}
							firstParameterType = typeArguments[0]
						default:
							return true
						}
					}

					if !type_checking.IsTypeUnknownType(firstParameterType) {
						return true
					}
				}
			}
			return false
		}

		// collectFlaggedCallbacks walks an argument expression down to the function literals it can
		// evaluate to, and returns the ones whose handler type is flaggable.
		//
		// Recursive rather than a listener because the shape is an expression tree the rule has to
		// descend on its own terms: only the last element of a sequence counts, and both sides of a
		// logical or a conditional do.
		var collectFlaggedCallbacks func(node *ast.Node, into []*ast.Node) []*ast.Node
		collectFlaggedCallbacks = func(node *ast.Node, into []*ast.Node) []*ast.Node {
			node = useUnknownSkipParentheses(node)
			if node == nil {
				return into
			}

			switch {
			case ast.IsBinaryExpression(node):
				binary := node.AsBinaryExpression()
				switch binary.OperatorToken.Kind {
				case ast.KindAmpersandAmpersandToken,
					ast.KindBarBarToken,
					ast.KindQuestionQuestionToken:
					// ESTree's LogicalExpression. Either side can be the handler that runs.
					into = collectFlaggedCallbacks(binary.Left, into)
					return collectFlaggedCallbacks(binary.Right, into)
				case ast.KindCommaToken:
					// ESTree's SequenceExpression, which our parser spells as a comma operator and
					// nests to the LEFT, so the last element is the rightmost operand. Only that
					// one is the value of the expression. Measured: a flaggable callback in the
					// left operand is clean and in the right operand reports.
					return collectFlaggedCallbacks(binary.Right, into)
				}
				return into

			case ast.IsConditionalExpression(node):
				conditional := node.AsConditionalExpression()
				into = collectFlaggedCallbacks(conditional.WhenTrue, into)
				return collectFlaggedCallbacks(conditional.WhenFalse, into)

			case ast.IsArrowFunction(node), ast.IsFunctionExpression(node):
				if isFlaggableHandlerType(ctx.TypeChecker.GetTypeAtLocation(node)) {
					return append(into, node)
				}
				return into
			}

			// Everything else terminates the walk. A bare identifier, a call, or a type assertion
			// is not a literal written here, and upstream declines all three deliberately.
			return into
		}

		return rule.Listeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()

				callee := useUnknownSkipParentheses(call.Expression)
				if callee == nil {
					return
				}

				receiver, methodName := useUnknownStaticMemberAccess(ctx.TypeChecker, callee)
				if receiver == nil {
					return
				}

				// `catch` takes the handler first; `then` takes it second, after the fulfillment
				// callback. The message's `append` slot is the only other thing that varies.
				var argumentIndexToCheck int
				var messageAppend string
				switch methodName {
				case "catch":
					argumentIndexToCheck, messageAppend = 0, ""
				case "then":
					argumentIndexToCheck, messageAppend = 1, " rejection"
				default:
					return
				}

				arguments := call.Arguments.Nodes
				if len(arguments) < argumentIndexToCheck+1 {
					return
				}

				// A spread at or before the position under test means the rule cannot tell which
				// value lands there, so it declines the whole call. Fourteen of the corpus's clean
				// cases are exactly this shape, several of them with a badly typed handler inside
				// the spread tuple, which is what makes this a real discrimination rather than a
				// convenience.
				for _, argument := range arguments[:argumentIndexToCheck+1] {
					if argument.Kind == ast.KindSpreadElement {
						return
					}
				}

				// The receiver has to be thenable. Without this, any object with a `catch` method
				// reports, and the corpus writes that object deliberately.
				receiverType := ctx.TypeChecker.GetTypeAtLocation(receiver)
				if !type_checking.IsThenableType(ctx.TypeChecker, callee, receiverType) {
					return
				}

				for _, callback := range collectFlaggedCallbacks(arguments[argumentIndexToCheck], nil) {
					useUnknownReport(ctx, callback, methodName, messageAppend)
				}
			},
		}
	},
}

// useUnknownSkipParentheses is ast.SkipParentheses with a nil answer for a nil argument.
//
// The bare helper DEREFERENCES its argument, so a nil reaching it is a panic, and a panic costs
// every rule its verdict on the whole file rather than costing this rule one finding.
//
// The honest status of this guard is DEFENSIVE rather than load-bearing, and it is written down that
// way because the tempting sentence, that error recovery reaches it, is one I could not measure.
// Probed over five malformed shapes including `p[](err => {})`, `p[]`, `p?.[](x)` and `()(err => {})`,
// counting nils at both sites: zero nil callees and zero nil argument expressions. Our parser
// SYNTHESIZES a missing element-access key as an identifier rather than leaving the field absent, so
// `p[]` presents as `KindIdentifier` and not as nil.
//
// So no fixture in this package can prove this guard matters, and removing it leaves the malformed-
// input test green. That was checked rather than assumed: the guard was deleted, the crash test was
// re-run, and it still passed. It is kept anyway, because the cost is one comparison, the parser's
// recovery behaviour is not a contract this rule should depend on, and the failure it prevents is
// the one that is silent in the summary line and expensive on the real tree. What is not claimed is
// that anything here has been shown to reach it.
func useUnknownSkipParentheses(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	return ast.SkipParentheses(node)
}

// useUnknownStaticMemberAccess reads a property access, returning the receiver and the property
// name, for both the dotted and the computed spellings.
//
// The computed form is not incidental. The corpus's second INVALID case is
// `let method = 'catch'; Promise.resolve()[method]((error: Error) => {})` and its second VALID case
// is `let x = Math.random() ? 'ca' + 'tch' : 'catch'; Promise.resolve()[x](...)`. Both are computed
// accesses through a binding, and upstream separates them, so a port testing only for a string
// literal is silent on the first and a port accepting any identifier reports the second.
//
// Upstream answers this with ESLint's `getStaticValue`, a constant folder over the SCOPE. We answer
// it with the checker, through `getAccessedPropertyName`, which is the same question asked of the
// type: a key whose type is a single string literal names one property.
//
// The two agree on three of the four spellings and disagree on one, measured on the installed rule
// against a probe of our own checker over the same four inputs:
//
//	const method = 'catch'; p[method](...)     reports upstream, resolves here      agree
//	let x = cond ? 'ca'+'tch' : 'catch'        clean upstream, unresolved here      agree
//	declare const method: string               clean upstream, unresolved here      agree
//	let method = 'catch'; p[method](...)       reports upstream, UNRESOLVED here    DIVERGE
//
// The divergence is a property of `let` rather than of the approach: a `let` initialised with a
// string literal WIDENS to `string`, so no literal type survives for the checker to read, while
// ESLint's folder sees the single assignment in the scope and folds it. Reproducing it needs a
// constant-folding pass over bindings, which is substantially larger than this whole rule.
//
// So upstream invalid case 1 is silent here, deliberately, and its fixture records that rather than
// being deleted or quietly relaxed. This is the same divergence already recorded on
// prefer-promise-reject-errors and only-throw-error, which share the upstream helper, but this rule
// reaches FURTHER than those two do: they accept a string literal only, so they lose the `const`
// spelling as well, and the checker recovers it here at no cost.
func useUnknownStaticMemberAccess(typeChecker *checker.Checker, callee *ast.Node) (receiver *ast.Node, propertyName string) {
	switch {
	case ast.IsPropertyAccessExpression(callee):
		name := callee.AsPropertyAccessExpression().Name()
		if name == nil || name.Kind != ast.KindIdentifier {
			return nil, ""
		}
		return callee.AsPropertyAccessExpression().Expression, name.Text()

	case ast.IsElementAccessExpression(callee):
		// The literal spelling is answered directly rather than through the checker, because
		// `p['catch']` needs no type at all and the corpus writes it as its own invalid case.
		argumentExpression := useUnknownSkipParentheses(callee.AsElementAccessExpression().ArgumentExpression)
		if argumentExpression != nil && ast.IsStringLiteralLike(argumentExpression) {
			return callee.AsElementAccessExpression().Expression, argumentExpression.Text()
		}
		if name, resolved := checker.Checker_getAccessedPropertyName(typeChecker, callee); resolved {
			return callee.AsElementAccessExpression().Expression, name
		}
		return nil, ""
	}

	return nil, ""
}

// useUnknownReport reports one flagged callback, pointing at its first parameter and offering the
// repair that parameter's shape allows.
//
// The span is the parameter's own text including its annotation and any default, which is
// rule.TokenRange of the parameter node. Our parser folds ESTree's AssignmentPattern into the
// parameter's Initializer, so upstream's outer/inner split collapses: the outer is the parameter and
// the inner is its Name.
func useUnknownReport(ctx rule.Context, callback *ast.Node, methodName string, messageAppend string) {
	parameters := callback.Parameters()
	if len(parameters) == 0 {
		// Unreachable through the rule: isFlaggableHandlerType only answers true for a signature
		// that HAS a first parameter, and a callback with no syntactic parameters has none in its
		// signature either. Kept because it is the difference between declining and indexing past
		// the end of a slice, and the callers that could reach it are enumerated in one place
		// directly above rather than spread across the file.
		return
	}
	parameter := parameters[0]
	parameterDeclaration := parameter.AsParameterDeclaration()
	name := parameterDeclaration.Name()
	if name == nil {
		return
	}

	message := buildUseUnknownMessage(methodName, messageAppend)

	switch {
	case name.Kind == ast.KindArrayBindingPattern && parameterDeclaration.DotDotDotToken == nil:
		// Destructuring an array asserts the thrown value is iterable, which nothing guarantees, so
		// there is no annotation to repair and upstream offers none.
		ctx.ReportNode(parameter, buildUseUnknownArrayDestructuringMessage(methodName, messageAppend))
		return

	case name.Kind == ast.KindObjectBindingPattern && parameterDeclaration.DotDotDotToken == nil:
		// Same, for shape and nullability.
		ctx.ReportNode(parameter, buildUseUnknownObjectDestructuringMessage(methodName, messageAppend))
		return

	case parameterDeclaration.DotDotDotToken != nil:
		// A rest parameter's correct annotation is the one-element TUPLE `[unknown]` rather than
		// `unknown`, because the parameter holds the argument list and not the error.
		if parameterDeclaration.Type == nil {
			ctx.ReportNodeWithSuggestions(parameter, message, rule.Suggestion{
				Message: buildAddUnknownRestTypeAnnotationMessage(),
				Fixes:   []rule.Fix{useUnknownInsertAnnotation(parameterDeclaration, ": [unknown]")},
			})
			return
		}
		ctx.ReportNodeWithSuggestions(parameter, message, rule.Suggestion{
			Message: buildWrongRestTypeAnnotationMessage(),
			Fixes:   []rule.Fix{rule.ReplaceRange(useUnknownAnnotationRange(parameterDeclaration), ": [unknown]")},
		})
		return
	}

	// A plain identifier, which is the common case.
	if parameterDeclaration.Type == nil {
		// A parenless arrow has nowhere to put an annotation, so the repair has to add the
		// parentheses too. `err => {}` becomes `(err: unknown) => {}` rather than `err: unknown => {}`,
		// which does not parse.
		if ast.IsArrowFunction(callback) && type_checking.IsParenlessArrowFunction(callback) {
			ctx.ReportNodeWithSuggestions(parameter, message, rule.Suggestion{
				Message: buildAddUnknownTypeAnnotationMessage(),
				Fixes: []rule.Fix{
					ctx.InsertBefore(name, "("),
					useUnknownInsertAnnotation(parameterDeclaration, ": unknown)"),
				},
			})
			return
		}
		ctx.ReportNodeWithSuggestions(parameter, message, rule.Suggestion{
			Message: buildAddUnknownTypeAnnotationMessage(),
			Fixes:   []rule.Fix{useUnknownInsertAnnotation(parameterDeclaration, ": unknown")},
		})
		return
	}

	ctx.ReportNodeWithSuggestions(parameter, message, rule.Suggestion{
		Message: buildWrongTypeAnnotationMessage(),
		Fixes:   []rule.Fix{rule.ReplaceRange(useUnknownAnnotationRange(parameterDeclaration), ": unknown")},
	})
}

// useUnknownAnnotationRange is the span a replacement annotation overwrites: the colon through the
// end of the type.
//
// It is arithmetic rather than a node because our type node begins after the colon and carries the
// whitespace before itself as leading trivia, so replacing the node alone would leave the colon and
// write `err:: unknown`. Starting from the end of the name (or the end of the `?`, which sits
// between the name and the colon) reaches back over the colon exactly once.
//
// The END is the type node's own end rather than the parameter's, which is what leaves a default and
// any trailing comment untouched: upstream turns `(err: any /* c1 */ = 2)` into
// `(err: unknown /* c1 */ = 2)`, and taking the parameter's end would swallow both.
// useUnknownInsertAnnotation inserts a new annotation at the point where one would be written.
//
// That point is after the `?` when the parameter is optional, not after the name. `(err?)` has to
// become `(err?: unknown)`, and inserting after the name produces `(err: unknown?)`, which is not
// valid syntax at all.
//
// Caught by upstream's own output for invalid case 10 rather than by reading, and it is exactly the
// defect the brief says an id assertion cannot see: the finding, its message id and its span were
// all correct while the repair wrote a broken file. The two insertion sites and the two replacement
// sites now share one definition of where an annotation goes, so they cannot disagree again.
func useUnknownInsertAnnotation(parameterDeclaration *ast.ParameterDeclaration, text string) rule.Fix {
	at := parameterDeclaration.Name().End()
	if parameterDeclaration.QuestionToken != nil {
		at = parameterDeclaration.QuestionToken.End()
	}
	return rule.ReplaceRange(core.NewTextRange(at, at), text)
}

func useUnknownAnnotationRange(parameterDeclaration *ast.ParameterDeclaration) core.TextRange {
	start := parameterDeclaration.Name().End()
	if parameterDeclaration.QuestionToken != nil {
		start = parameterDeclaration.QuestionToken.End()
	}
	return core.NewTextRange(start, parameterDeclaration.Type.End())
}

func buildUseUnknownMessage(methodName string, messageAppend string) rule.Message {
	return rule.Message{
		Id:          "useUnknown",
		Description: "Prefer the safe `: unknown` for a `" + methodName + "`" + messageAppend + " callback variable.",
	}
}

func buildUseUnknownArrayDestructuringMessage(methodName string, messageAppend string) rule.Message {
	return rule.Message{
		Id: "useUnknownArrayDestructuringPattern",
		Description: "Prefer the safe `: unknown` for a `" + methodName + "`" + messageAppend +
			" callback variable. The thrown error may not be iterable.",
	}
}

func buildUseUnknownObjectDestructuringMessage(methodName string, messageAppend string) rule.Message {
	return rule.Message{
		Id: "useUnknownObjectDestructuringPattern",
		Description: "Prefer the safe `: unknown` for a `" + methodName + "`" + messageAppend +
			" callback variable. The thrown error may be nullable, or may not have the expected shape.",
	}
}

func buildAddUnknownTypeAnnotationMessage() rule.Message {
	return rule.Message{
		Id:          "addUnknownTypeAnnotationSuggestion",
		Description: "Add an explicit `: unknown` type annotation to the rejection callback variable.",
	}
}

func buildWrongTypeAnnotationMessage() rule.Message {
	return rule.Message{
		Id:          "wrongTypeAnnotationSuggestion",
		Description: "Change existing type annotation to `: unknown`.",
	}
}

func buildAddUnknownRestTypeAnnotationMessage() rule.Message {
	return rule.Message{
		Id:          "addUnknownRestTypeAnnotationSuggestion",
		Description: "Add an explicit `: [unknown]` type annotation to the rejection callback rest variable.",
	}
}

func buildWrongRestTypeAnnotationMessage() rule.Message {
	return rule.Message{
		Id:          "wrongRestTypeAnnotationSuggestion",
		Description: "Change existing type annotation to `: [unknown]`.",
	}
}
