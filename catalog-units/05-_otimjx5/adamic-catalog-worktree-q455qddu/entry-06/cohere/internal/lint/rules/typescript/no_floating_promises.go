package typescript

import (
	"encoding/json"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// NoFloatingPromisesOptions is upstream's option struct, field-for-field.
type NoFloatingPromisesOptions struct {
	AllowForKnownSafeCalls          []type_checking.TypeOrValueSpecifier
	AllowForKnownSafeCallsInline    []string
	AllowForKnownSafePromises       []type_checking.TypeOrValueSpecifier
	AllowForKnownSafePromisesInline []string
	CheckThenables                  *bool
	IgnoreIIFE                      *bool
	IgnoreVoid                      *bool
}

var noFloatingPromisesMessageBase = "Promises must be awaited, end with a call to .catch, or end with a call to .then with a rejection handler."

var noFloatingPromisesMessageBaseVoid = "Promises must be awaited, end with a call to .catch, end with a call to .then with a rejection handler" +
	" or be explicitly marked as ignored with the `void` operator."

var noFloatingPromisesMessageRejectionHandler = "A rejection handler that is not a function will be ignored."

func buildFloatingMessage() rule.Message {
	return rule.Message{
		Id:          "floating",
		Description: noFloatingPromisesMessageBase,
	}
}

func buildFloatingFixAwaitMessage() rule.Message {
	return rule.Message{
		Id:          "floatingFixAwait",
		Description: "Add await operator.",
	}
}

func buildFloatingFixVoidMessage() rule.Message {
	return rule.Message{
		Id:          "floatingFixVoid",
		Description: "Add void operator to ignore.",
	}
}

func buildFloatingPromiseArrayMessage() rule.Message {
	return rule.Message{
		Id:          "floatingPromiseArray",
		Description: "An array of Promises may be unintentional. Consider handling the promises' fulfillment or rejection with Promise.all or similar.",
	}
}

func buildFloatingPromiseArrayVoidMessage() rule.Message {
	return rule.Message{
		Id: "floatingPromiseArrayVoid",
		Description: "An array of Promises may be unintentional. Consider handling the promises' fulfillment or rejection with Promise.all or similar," +
			" or explicitly marking the expression as ignored with the `void` operator.",
	}
}

func buildFloatingUselessRejectionHandlerMessage() rule.Message {
	return rule.Message{
		Id:          "floatingUselessRejectionHandler",
		Description: noFloatingPromisesMessageBase + " " + noFloatingPromisesMessageRejectionHandler,
	}
}

func buildFloatingUselessRejectionHandlerVoidMessage() rule.Message {
	return rule.Message{
		Id:          "floatingUselessRejectionHandlerVoid",
		Description: noFloatingPromisesMessageBaseVoid + " " + noFloatingPromisesMessageRejectionHandler,
	}
}

func buildFloatingVoidMessage() rule.Message {
	return rule.Message{
		Id:          "floatingVoid",
		Description: noFloatingPromisesMessageBaseVoid,
	}
}

// NoFloatingPromises flags a promise used as a statement and then abandoned: nothing awaits it,
// nothing catches it, and a rejection becomes an unhandled rejection nobody sees.
//
//	valid:   async function f() { await Promise.resolve('x'); }
//	valid:   Promise.resolve('x').catch(() => {});
//	valid:   Promise.resolve('x').then(() => {}, () => {});
//	valid:   void Promise.resolve('x');          // ignoreVoid, on by default
//	valid:   const held = Promise.resolve('x');
//	invalid: Promise.resolve('x');
//	invalid: Promise.resolve('x').then(() => {});          // no rejection handler
//	invalid: Promise.resolve('x').catch(null);             // handler is not a function
//	invalid: declare const all: Promise<number>[]; all;    // an array of promises
//
// # Absorbed from tsgolint, which is the source of record for this rule
//
// Provenance: tsgolint `internal/rules/no_floating_promises/no_floating_promises.go`, fetched from
// `main` and absorbed onto cohere's own rule interface here. The checker logic is upstream's
// unchanged; what moved is the interface it speaks. `rule.RuleContext` became `rule.Context`,
// `RuleListeners` became `Listeners`, `RuleMessage` became `Message`, `RuleSuggestion{FixesArr:}`
// became `Suggestion{Fixes:}`, and the four `RuleFix*` builders became their `ctx.` equivalents,
// which trim to the token exactly as upstream's do. A line-by-line diff against the fetched file
// shows those renames, the two declared flags below, and the nil guard, and nothing else: no
// predicate, no traversal, no message text.
//
// It reaches for `type_checking.UnionTypeParts`, `IsPromiseLike`, `GetCallSignatures`, `Some` and
// `TypeMatchesSomeSpecifier` because that is what upstream reaches for, and this note is why a
// reader finds those helpers in a file that otherwise looks native. tsgolint is not re-synced, so
// this file is now the only copy of the algorithm rather than a translation layer over a vendored
// one.
//
// oxc declares this rule `(tsgolint)` and carries no algorithm and no corpus, so the behavior
// oxlint exhibits IS tsgolint's: the release binary shells out to a `tsgolint` executable and
// refuses to run the rule when that binary is absent, which is what it does on this machine. That
// refusal is itself the proof, so oxlint cannot serve as ground truth here and tsgolint's own test
// file is the corpus instead: seventy-two clean cases, ninety-seven reporting a hundred and
// forty-six findings, and two hundred and thirty-two exact suggestion outputs.
//
// # The option surface is FIVE keys, and a captured inventory said none
//
// A rule catalog recorded `"options": "no"` for this rule. It has five, and both references
// agree on all of them: `allowForKnownSafeCalls`, `allowForKnownSafePromises`, `checkThenables`,
// `ignoreIIFE`, `ignoreVoid`. Counted from tsgolint's option struct and confirmed against the
// plugin's `meta.schema` rather than taken from the inventory.
//
// `ignoreVoid` defaults to TRUE, which is the load-bearing default: it decides which of two message
// families a finding lands in and whether the void suggestion is offered at all. An unconfigured
// run therefore reports `floatingVoid` rather than `floating`, and the decoder keeps all three
// booleans as pointers end to end so that absent stays distinguishable from an explicit false.
//
// The two allowlists take TYPE SPECIFIERS, which is why `typecheck/specifier.go` is
// load-bearing here. Each entry is either a bare string matched against the type's own name, or an
// object whose `from` is `file`, `lib` or `package`. Those are two different fields in the struct
// the rule reads, so `DecodeNoFloatingPromisesOptions` below is hand-written rather than a
// `rule.DecodeOptionsInto` binding.
//
// A `from: file` specifier carrying a `path` compares that path, resolved against the program's
// current directory, to the DECLARING file's absolute name. That makes one upstream case depend on
// what upstream named its fixture, and the test file says how that is reproduced.
//
// # What counts as handled, measured rather than reasoned about
//
// The anchor is `KindExpressionStatement`, so the question is only ever asked of a promise that is
// a statement by itself. That is what makes the whole rule tractable and it is also why several
// shapes are silent for a structural reason rather than a decision. Measured, holding the promise
// fixed and varying only the position:
//
//	await p;                    SILENT    the await expression returns early
//	void p;                     SILENT    under ignoreVoid, which is the default
//	p.catch(() => {});          SILENT    a catch with a callable handler
//	p.then(() => {}, () => {}); SILENT    a then with a callable SECOND argument
//	p.finally(() => {}).catch(() => {});  SILENT    finally is transparent, so `p` is re-asked
//	const x = p;                SILENT    not an expression statement at all
//	y = p;                      SILENT    an assignment expression returns early
//	sink(p);                    SILENT    an argument is not a statement
//	({ a: p });                 SILENT    an object literal is not promise-like
//	p;                          REPORTS   floatingVoid
//	p.then(() => {});           REPORTS   floatingVoid, no rejection handler present
//	p.catch();                  REPORTS   floatingVoid, the arity guard is >= 1
//	p.catch(null);              REPORTS   floatingUselessRejectionHandlerVoid
//	p.finally(() => {});        REPORTS   floatingVoid, finally alone handles nothing
//	p, p;                       REPORTS   both sides of a comma are asked
//	true ? p : p;               REPORTS   both branches of a ternary are asked
//	p || p;                     REPORTS   both sides of a logical operator are asked
//	[p];                        REPORTS   floatingPromiseArrayVoid
//
// `p.catch()` landing on `floatingVoid` rather than on the useless-handler message is worth naming,
// because the message text invites the opposite reading: the arity guard is `>= 1`, so a `catch`
// with no arguments never reaches the handler test and falls through to "unhandled" instead.
//
// # An array of promises reports ONCE, and the reason is structural
//
// Probed with a TRIPLE rather than a pair, reading every position rather than the count:
//
//	[a, b, c];      ONE finding, floatingPromiseArrayVoid, pointing at the whole statement
//	a; b; c;        THREE findings, one per statement, each pointing at its own statement
//	Promise.all([a, b, c]);   ONE finding, floatingVoid, on the CALL rather than the array
//
// There is no grouping decision here to get wrong. The listener fires per statement, so per-element
// reporting is not a close call the corpus happens not to cover, it is unreachable. The
// `Promise.all` row is the one that could mislead: it is not a promise array at all, because the
// array is consumed by the call and what floats is the call's own promise.
//
// The promise-array findings carry NO suggestions, which is the one report path that uses plain
// `ReportNode`. Neither `await` nor `void` is a correct repair for an array, so upstream offers
// nothing rather than offering something wrong.
//
// # Where the two references disagree, measured on both
//
// Fifty-three inputs were run through this rule and through `@typescript-eslint`'s version driven
// by ESLint's flat-config Linter API, one fresh directory per case so the plugin's program is
// rebuilt each time, with a negative control to prove a non-zero was data rather than a
// config-resolution error. Fifty-one agree. Both sides register exactly ONE listener
// (`ExpressionStatement`), carry the same eight message ids, and ship suggestions rather than
// fixes; there is no equivalent here of the fourth arm `await-thenable` has on one side only.
//
// The two disagreements:
//
//	enum E { Catch = 'catch' }
//	p[E.Catch](() => {});        tsgolint SILENT    @typescript-eslint REPORTS
//	(p.catch)(() => {});         tsgolint REPORTS   @typescript-eslint SILENT
//
// The first is deliberate on tsgolint's side and upstream says so at the line: its `TODO(port)`
// notes that swapping `getStaticMemberAccessValue` for `Checker_getAccessedPropertyName` is "an
// enhancement" and that it lacks tests. The checker resolves the enum member to `'catch'`, so the
// promise reads as handled; the plugin reads only literal keys and does not.
//
// The second is the paren axis, and it falls the OPPOSITE way from `no-implied-eval`, where
// tsgolint is the silent one. The cause is the same structural difference in both rules and the
// direction is reversed by where the parens sit. This rule runs `ast.SkipParentheses` over the
// statement's expression, so `(p.catch(() => {}))` is transparent and silent on both sides. It does
// not skip again during the descent, so a parenthesized CALLEE leaves `ast.IsAccessExpression`
// declining and the handler test never runs. ESTree has no parenthesized-expression node, so the
// plugin never sees the parens at all. Three positions, three answers, in one rule:
//
//	(p.catch(() => {}));    SILENT    parens around the statement expression, skipped
//	(p).catch(() => {});    SILENT    parens around the receiver, the checker sees through
//	(p.catch)(() => {});    REPORTS   parens around the callee, the access test declines
//
// tsgolint wins on both, because oxlint runs tsgolint and the differential harness compares against
// oxlint. The corpus writes no parenthesized form anywhere, so this whole axis is invisible to the
// imported fixtures and is pinned by measured cases in the test file instead.
//
// # The library pinning does NOT bite here, and that was measured
//
// The fixture tsconfig pins `lib: ["ES2022"]` with no way to raise it, which inverted three
// `await-thenable` cases and moved nothing on `no-implied-eval`. It is a per-rule question and the
// answer is not predictable from the rule's shape, so it was run rather than inherited: ten shapes,
// each once bare and once with a second file merging onto the global scope. NO VERDICT MOVED.
// `Promise` and `PromiseLike` are declared in the ES2015 library that ES2022 includes, so the
// globals this rule cares about resolve properly under the pinning. That is the opposite of
// `Disposable`, which resolves to the error type and reports flags of `any`.
//
// # The checker, the program, and the nil guard
//
// Every type question routes through `ctx.TypeChecker`, so the checker is required and the fixtures
// use the typed harness.
//
// `ProgramReads` is a measurement rather than an assumption: the body reaches `ctx.Program` at three
// sites, one of them on the core path. `type_checking.IsPromiseLike(ctx.Program, ...)` asks whether a
// declaration is the default library's `Promise`, which happens for every promise test this rule
// makes, and the two `TypeMatchesSomeSpecifier` calls read the compiler options and the default
// library too. Enum-key resolution reaches the declaring module through the checker, which a fixture
// pins across two files; the type fingerprint is what keeps a cached finding honest about that.
//
// The nil guard is the one addition to the body and it could not be written before. While this rule
// lived behind the tsgolint adapter the listener was upstream's and the adapter set
// `NeedsTypeChecker` unconditionally, so the nil case was unreachable and editing the listener
// would have turned a re-sync into a merge. Absorbing the rule removes both constraints.
//
// The guard matters more than crash-avoidance would suggest. On this shim `GetTypeAtLocation` and
// `GetSymbolAtLocation` return nil rather than panicking, so a checker-less run would not blow up:
// `isPromiseLike` would answer false for everything and the rule would go completely silent.
// Ninety-seven reporting fixtures would fail in a way that reads like a rule defect while
// seventy-two clean ones passed VACUOUSLY. A test pins the declaration and the decline together, so
// a later revert fails loudly.
//
// # Cost
//
// The anchor is every expression statement, which is common, and unlike `no-implied-eval` there is
// no cheap syntactic filter before the checker: the first thing the descent does on most inputs is
// ask for a type. `isPromiseArray` runs before `isPromiseLike` and walks union parts asking for
// apparent types, so a statement whose expression is not a promise still costs one type read. That
// is upstream's ordering and it is kept.
var NoFloatingPromises = rule.Rule{
	Name: "@typescript-eslint/no-floating-promises",

	NeedsTypeChecker: true,

	TypeReach: rule.TypeReachShapes,

	// Compiler options and the default library, through type_checking's builtin and specifier helpers.
	ProgramReads: rule.ReadsCompilerOptions | rule.ReadsDefaultLibrary,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		opts, ok := rule.OptionsAs[NoFloatingPromisesOptions](options)
		if !ok {
			opts = NoFloatingPromisesOptions{
				AllowForKnownSafeCalls:          []type_checking.TypeOrValueSpecifier{},
				AllowForKnownSafeCallsInline:    []string{},
				AllowForKnownSafePromises:       []type_checking.TypeOrValueSpecifier{},
				AllowForKnownSafePromisesInline: []string{},
			}
		}
		if opts.CheckThenables == nil {
			opts.CheckThenables = type_checking.Ref(false)
		}
		if opts.IgnoreIIFE == nil {
			opts.IgnoreIIFE = type_checking.Ref(false)
		}
		if opts.IgnoreVoid == nil {
			opts.IgnoreVoid = type_checking.Ref(true)
		}

		isHigherPrecedenceThanUnary := func(node *ast.Node) bool {
			operator := ast.KindUnknown
			if ast.IsBinaryExpression(node) {
				operator = node.AsBinaryExpression().OperatorToken.Kind
			}
			nodePrecedence := ast.GetOperatorPrecedence(node.Kind, operator, ast.OperatorPrecedenceFlagsNone)
			return nodePrecedence > ast.OperatorPrecedenceUnary
		}

		addAwait := func(
			expression *ast.Expression,
			node *ast.ExpressionStatement,
		) []rule.Fix {
			if ast.IsVoidExpression(expression) {
				voidTokenRange := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, expression.Pos())
				return []rule.Fix{rule.ReplaceRange(voidTokenRange, "await")}
			}
			if isHigherPrecedenceThanUnary(node.Expression) {
				return []rule.Fix{ctx.InsertBefore(&node.Node, "await ")}
			}
			return []rule.Fix{
				ctx.InsertBefore(&node.Node, "await ("),
				ctx.InsertAfter(expression, ")"),
			}
		}

		hasMatchingSignature := func(
			t *checker.Type,
			matcher func(signature *checker.Signature) bool,
		) bool {
			for part := range type_checking.UnionTypePartsSeq(t) {
				if type_checking.Some(type_checking.GetCallSignatures(ctx.TypeChecker, part), matcher) {
					return true
				}
			}

			return false
		}

		isFunctionParam := func(
			param *ast.Symbol,
			node *ast.Node,
		) bool {
			t := checker.Checker_getApparentType(ctx.TypeChecker, ctx.TypeChecker.GetTypeOfSymbolAtLocation(param, node))

			for part := range type_checking.UnionTypePartsSeq(t) {
				if len(type_checking.GetCallSignatures(ctx.TypeChecker, part)) != 0 {
					return true
				}
			}
			return false
		}

		isPromiseLike := func(node *ast.Node, t *checker.Type) bool {
			if t == nil {
				t = ctx.TypeChecker.GetTypeAtLocation(node)
			}

			// The highest priority is to allow anything allowlisted
			if type_checking.TypeMatchesSomeSpecifier(
				t,
				opts.AllowForKnownSafePromises,
				opts.AllowForKnownSafePromisesInline,
				ctx.Program,
			) {
				return false
			}

			// Otherwise, we always consider the built-in Promise to be Promise-like...
			typeParts := type_checking.UnionTypeParts(checker.Checker_getApparentType(ctx.TypeChecker, t))
			if type_checking.Some(typeParts, func(typePart *checker.Type) bool {
				return type_checking.IsPromiseLike(ctx.Program, ctx.TypeChecker, typePart)
			}) {
				return true
			}

			// ...and only check all Thenables if explicitly told to
			if !*opts.CheckThenables {
				return false
			}

			// Modified from tsutils.isThenable() to only consider thenables which can be
			// rejected/caught via a second parameter. Original source (MIT licensed):
			//
			//   https://github.com/ajafff/tsutils/blob/49d0d31050b44b81e918eae4fbaf1dfe7b7286af/util/type.ts#L95-L125
			for _, typePart := range typeParts {
				then := checker.Checker_getPropertyOfType(ctx.TypeChecker, typePart, "then")
				if then == nil {
					continue
				}

				thenType := ctx.TypeChecker.GetTypeOfSymbolAtLocation(then, node)
				if hasMatchingSignature(
					thenType,
					func(signature *checker.Signature) bool {
						params := checker.Signature_parameters(signature)
						return len(params) >= 2 && isFunctionParam(params[0], node) && isFunctionParam(params[1], node)
					}) {
					return true
				}
			}
			return false
		}

		isPromiseArray := func(node *ast.Node) bool {
			t := ctx.TypeChecker.GetTypeAtLocation(node)
			for typePart := range type_checking.UnionTypePartsSeq(t) {
				apparent := checker.Checker_getApparentType(ctx.TypeChecker, typePart)

				if checker.Checker_isArrayType(ctx.TypeChecker, apparent) {
					arrayType := checker.Checker_getTypeArguments(ctx.TypeChecker, apparent)[0]
					if isPromiseLike(node, arrayType) {
						return true
					}
				}

				if checker.IsTupleType(apparent) {
					for _, tupleElementType := range checker.Checker_getTypeArguments(ctx.TypeChecker, apparent) {
						if isPromiseLike(node, tupleElementType) {
							return true
						}
					}
				}
			}
			return false
		}

		isKnownSafePromiseReturn := func(node *ast.Node) bool {
			if !ast.IsCallExpression(node) {
				return false
			}

			t := ctx.TypeChecker.GetTypeAtLocation(node.AsCallExpression().Expression)

			return type_checking.TypeMatchesSomeSpecifier(
				t,
				opts.AllowForKnownSafeCalls,
				opts.AllowForKnownSafeCallsInline,
				ctx.Program,
			)
		}

		isAsyncIife := func(node *ast.ExpressionStatement) bool {
			if !ast.IsCallExpression(node.Expression) {
				return false
			}

			callee := ast.SkipParentheses(node.Expression.AsCallExpression().Expression)

			return ast.IsArrowFunction(callee) || ast.IsFunctionExpression(callee)
		}

		isValidRejectionHandler := func(rejectionHandler *ast.Node) bool {
			return len(type_checking.GetCallSignatures(ctx.TypeChecker, ctx.TypeChecker.GetTypeAtLocation(rejectionHandler))) > 0
		}

		var isUnhandledPromise func(
			node *ast.Node,
		) (
			bool, // isUnhandled
			bool, // nonFunctionHandler
			bool, // promiseArray
		)
		isUnhandledPromise = func(
			node *ast.Node,
		) (
			bool, // isUnhandled
			bool, // nonFunctionHandler
			bool, // promiseArray
		) {
			if ast.IsAssignmentExpression(node, false) {
				return false, false, false
			}

			// First, check expressions whose resulting types may not be promise-like
			if ast.IsCommaExpression(node) {
				expr := node.AsBinaryExpression()
				// Any child in a comma expression could return a potentially unhandled
				// promise, so we check them all regardless of whether the final returned
				// value is promise-like.
				isUnhandled, nonFunctionHandler, promiseArray := isUnhandledPromise(expr.Left)
				if isUnhandled {
					return isUnhandled, nonFunctionHandler, promiseArray
				}
				return isUnhandledPromise(expr.Right)
			}

			if !*opts.IgnoreVoid && ast.IsVoidExpression(node) {
				// Similarly, a `void` expression always returns undefined, so we need to
				// see what's inside it without checking the type of the overall expression.
				return isUnhandledPromise(node.Expression())
			}

			// Check the type. At this point it can't be unhandled if it isn't a promise
			// or array thereof.

			if isPromiseArray(node) {
				return true, false, true
			}

			// await expression addresses promises, but not promise arrays.
			if ast.IsAwaitExpression(node) {
				// you would think this wouldn't be strictly necessary, since we're
				// anyway checking the type of the expression, but, unfortunately TS
				// reports the result of `await (promise as Promise<number> & number)`
				// as `Promise<number> & number` instead of `number`.
				return false, false, false
			}

			if !isPromiseLike(node, nil) {
				return false, false, false
			}

			if ast.IsCallExpression(node) {
				// If the outer expression is a call, a `.catch()` or `.then()` with
				// rejection handler handles the promise.

				callExpr := node.AsCallExpression()
				callee := callExpr.Expression
				if ast.IsAccessExpression(callee) {
					// TODO(port): getStaticMemberAccessValue -> GetAccessedPropertyName is an
					// enhancement, we should probably add tests for it
					// const methodName = getStaticMemberAccessValue(callee, context);
					methodName, _ := checker.Checker_getAccessedPropertyName(ctx.TypeChecker, callee)

					if methodName == "catch" && len(callExpr.Arguments.Nodes) >= 1 {
						if isValidRejectionHandler(callExpr.Arguments.Nodes[0]) {
							return false, false, false
						}
						return true, true, false
					}
					if methodName == "then" && len(callExpr.Arguments.Nodes) >= 2 {
						if isValidRejectionHandler(callExpr.Arguments.Nodes[1]) {
							return false, false, false
						}
						return true, true, false
					}
					// `x.finally()` is transparent to resolution of the promise, so check `x`.
					// ("object" in this context is the `x` in `x.finally()`)
					if methodName == "finally" {
						return isUnhandledPromise(callee.Expression())
					}
				}

				// All other cases are unhandled.
				return true, false, false
			}

			if node.Kind == ast.KindConditionalExpression {
				expr := node.AsConditionalExpression()
				// We must be getting the promise-like value from one of the branches of the
				// ternary. Check them directly.
				isUnhandled, nonFunctionHandler, promiseArray := isUnhandledPromise(expr.WhenFalse)
				if isUnhandled {
					return isUnhandled, nonFunctionHandler, promiseArray
				}
				return isUnhandledPromise(expr.WhenTrue)
			}

			if ast.IsLogicalOrCoalescingBinaryExpression(node) {
				expr := node.AsBinaryExpression()
				isUnhandled, nonFunctionHandler, promiseArray := isUnhandledPromise(expr.Left)
				if isUnhandled {
					return isUnhandled, nonFunctionHandler, promiseArray
				}
				return isUnhandledPromise(expr.Right)
			}

			// Anything else is unhandled.
			return true, false, false
		}

		return rule.Listeners{
			ast.KindExpressionStatement: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}

				exprStatement := node.AsExpressionStatement()

				if *opts.IgnoreIIFE && isAsyncIife(exprStatement) {
					return
				}

				expression := ast.SkipParentheses(exprStatement.Expression)

				if isKnownSafePromiseReturn(expression) {
					return
				}

				isUnhandled, nonFunctionHandler, promiseArray := isUnhandledPromise(expression)

				if !isUnhandled {
					return
				}
				if promiseArray {
					var msg rule.Message
					if *opts.IgnoreVoid {
						msg = buildFloatingPromiseArrayVoidMessage()
					} else {
						msg = buildFloatingPromiseArrayMessage()
					}
					ctx.ReportNode(node, msg)
				} else if *opts.IgnoreVoid {
					var msg rule.Message
					if nonFunctionHandler {
						msg = buildFloatingUselessRejectionHandlerVoidMessage()
					} else {
						msg = buildFloatingVoidMessage()
					}

					ctx.ReportNodeWithSuggestions(node, msg, rule.Suggestion{
						Message: buildFloatingFixVoidMessage(),
						Fixes: (func() []rule.Fix {
							if isHigherPrecedenceThanUnary(exprStatement.Expression) {
								return []rule.Fix{ctx.InsertBefore(node, "void ")}
							}
							return []rule.Fix{
								ctx.InsertBefore(node, "void ("),
								ctx.InsertAfter(expression, ")"),
							}
						})(),
					}, rule.Suggestion{
						Message: buildFloatingFixAwaitMessage(),
						Fixes:   addAwait(expression, exprStatement),
					})
				} else {
					var msg rule.Message
					if nonFunctionHandler {
						msg = buildFloatingUselessRejectionHandlerMessage()
					} else {
						msg = buildFloatingMessage()
					}
					ctx.ReportNodeWithSuggestions(node, msg, rule.Suggestion{
						Message: buildFloatingFixAwaitMessage(),
						Fixes:   addAwait(expression, exprStatement),
					})
				}
			},
		}
	},
}

// noFloatingPromisesRawOptions is the wire shape, which is not the shape the rule reasons with.
//
// Three keys bind straight through and two do not. A `TypeOrValueSpecifier` on the wire is either
// a bare string (matched against the type's own name) or an object whose `from` is one of `file`,
// `lib`, `package`. In the struct the rule reads, the bare-string form lives in a separate
// `...Inline []string` field and `from` is a `uint8` enum, so neither can be expressed as a struct
// tag and `rule.DecodeOptionsInto` alone cannot produce it.
type noFloatingPromisesRawOptions struct {
	AllowForKnownSafeCalls    []noFloatingPromisesRawSpecifier `json:"allowForKnownSafeCalls"`
	AllowForKnownSafePromises []noFloatingPromisesRawSpecifier `json:"allowForKnownSafePromises"`
	CheckThenables            *bool                            `json:"checkThenables"`
	IgnoreIIFE                *bool                            `json:"ignoreIIFE"`
	IgnoreVoid                *bool                            `json:"ignoreVoid"`
}

// noFloatingPromisesRawSpecifier is one entry of either allowlist, in either of its two wire forms.
type noFloatingPromisesRawSpecifier struct {
	inline string

	From    string
	Name    []string
	Path    string
	Package string
}

// UnmarshalJSON accepts both wire forms: a bare string, or the object the schema describes.
//
// `name` is itself two shapes, a string or an array of strings, and it is required on the object
// form. An unrecognized `from` is left as the zero value rather than erroring, which matches how
// the rest of this config treats a value it cannot use: the specifier then matches nothing instead
// of failing the run.
func (s *noFloatingPromisesRawSpecifier) UnmarshalJSON(raw []byte) error {
	var asString string
	if err := json.Unmarshal(raw, &asString); err == nil {
		*s = noFloatingPromisesRawSpecifier{inline: asString}
		return nil
	}

	var object struct {
		From    string          `json:"from"`
		Name    json.RawMessage `json:"name"`
		Path    string          `json:"path"`
		Package string          `json:"package"`
	}
	if err := rule.UnmarshalOptions(raw, &object); err != nil {
		return err
	}

	*s = noFloatingPromisesRawSpecifier{From: object.From, Path: object.Path, Package: object.Package}

	var names []string
	if err := json.Unmarshal(object.Name, &names); err == nil {
		s.Name = names
		return nil
	}
	var name string
	if err := json.Unmarshal(object.Name, &name); err != nil {
		return err
	}
	s.Name = []string{name}
	return nil
}

// specifiersFrom splits one wire allowlist into the two fields the rule reads.
func specifiersFrom(raw []noFloatingPromisesRawSpecifier) ([]type_checking.TypeOrValueSpecifier, []string) {
	specifiers := []type_checking.TypeOrValueSpecifier{}
	inline := []string{}
	for _, entry := range raw {
		if entry.inline != "" {
			inline = append(inline, entry.inline)
			continue
		}

		specifier := type_checking.TypeOrValueSpecifier{Name: entry.Name, Path: entry.Path, Package: entry.Package}
		switch entry.From {
		case "file":
			specifier.From = type_checking.TypeOrValueSpecifierFromFile
		case "lib":
			specifier.From = type_checking.TypeOrValueSpecifierFromLib
		case "package":
			specifier.From = type_checking.TypeOrValueSpecifierFromPackage
		default:
			// An unrecognized `from` cannot be represented, and a specifier that matches nothing is
			// a better answer than refusing the whole configuration. `file` is the zero value, so
			// saying nothing here would silently mean `file`; skipping says nothing at all.
			continue
		}
		specifiers = append(specifiers, specifier)
	}
	return specifiers, inline
}

// DecodeNoFloatingPromisesOptions maps upstream's JSON onto the struct the rule reads.
//
// Hand-written rather than `rule.DecodeOptionsInto` because the two allowlists each arrive as one
// heterogeneous array and leave as two typed fields, and because `from` is a string on the wire and
// an integer enum in the struct. The three boolean keys stay pointers all the way through, so that
// "absent" is distinguishable from "false" and the rule's own defaulting is what fills them in.
// That matters most for `ignoreVoid`, whose default is TRUE: binding it to a plain bool would make
// an unconfigured rule behave as though it had been explicitly turned off.
func DecodeNoFloatingPromisesOptions(raw []byte) (any, error) {
	decoded, err := rule.DecodeOptionsInto[noFloatingPromisesRawOptions]()(raw)
	if err != nil {
		return NoFloatingPromisesOptions{}, err
	}

	wire, _ := decoded.(noFloatingPromisesRawOptions)

	options := NoFloatingPromisesOptions{
		CheckThenables: wire.CheckThenables,
		IgnoreIIFE:     wire.IgnoreIIFE,
		IgnoreVoid:     wire.IgnoreVoid,
	}
	options.AllowForKnownSafeCalls, options.AllowForKnownSafeCallsInline = specifiersFrom(wire.AllowForKnownSafeCalls)
	options.AllowForKnownSafePromises, options.AllowForKnownSafePromisesInline = specifiersFrom(wire.AllowForKnownSafePromises)

	return options, nil
}
