package typescript

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/ecmascript/property"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// AwaitThenable flags `await` applied to a value that is not a Thenable, `for await...of` over a
// value that is not async iterable, `await using` on a value that is not async disposable, and a
// promise aggregator (`Promise.all` and friends) handed a value that is not a Thenable.
//
//	valid:   await Promise.resolve('value')
//	valid:   await (async () => true)()
//	valid:   let anyValue: any; await anyValue
//	valid:   async function wrapper<T>(value: T) { return await value; }
//	invalid: await 0
//	invalid: await (() => {})
//	invalid: for await (const value of yieldNumbers())
//	invalid: declare const d: Disposable; await using x = d
//	invalid: await Promise.all([fetchUser(), formatDate(now)])   // the second element already ran
//
// # Absorbed from tsgolint, which is the source of record for this rule
//
// Provenance: tsgolint `internal/rules/await_thenable/await_thenable.go`, vendored at commit
// `05b7fbc` and absorbed onto cohere's own rule interface here. It reaches for `NeedsToBeAwaited`,
// `GetWellKnownSymbolPropertyOfType` and `GetForStatementHeadLoc` because that is what upstream
// reaches for, and this note is why a reader finds those helpers in a file that otherwise looks
// native.
//
// tsgolint is not re-synced, so this file is now the only copy of the algorithm rather than a
// translation layer over a vendored one. The checker logic below is byte-identical to upstream's;
// what changed is the interface it speaks: `rule.RuleContext` became `rule.Context`,
// `RuleListeners` became `Listeners`, `rule.RuleMessage` became `rule.Message`, and each
// suggestion moved from `rule.RuleSuggestion{FixesArr: []rule.RuleFix{...}}` to
// `rule.Suggestion{Fixes: []rule.Fix{...}}`. The three fix constructors map one to one:
// `RuleFixRemoveRange` is `rule.RemoveRange`, and `RuleFixRemove(file, node)` is `ctx.RemoveNode`,
// which are the same expression under different names since `type_checking.TrimNodeTextRange` and
// `rule.TokenRange` are both `GetRangeOfTokenAtPosition(file, node.Pos()).WithEnd(node.End())`. No
// predicate, no tri-state and no traversal was touched.
//
// The reported SPANS are unchanged, and that is the part worth naming. While this rule was adapted,
// `upstream.Adapt` wrapped every node report in `rule.TokenRange(SourceFile, node)` because
// tsgolint's own runner does the same through `type_checking.TrimNodeTextRange`. Our native
// `ctx.ReportNodeWithSuggestions` applies exactly that trim itself, so absorbing the rule preserves
// the behavior rather than relying on the adapter to supply it. Passing `node.Loc` instead would
// include leading trivia, so an indented `await 0` would report the text "\n  await 0" and an
// `await using` initializer would report " disposable" with a leading space; that is the defect
// fixed in `8bdd70b`, and the span test in this package pins both shapes against it. The
// `for await...of` arm reports a computed head range rather than a node, so it goes through
// `ReportRangeWithSuggestions`, which trims nothing by design because the caller already said what
// it meant.
//
// oxc has no implementation. Its rule file declares `AwaitThenable(tsgolint)` and delegates, with
// fifty seven lines carrying documentation and no algorithm. So the behavior oxlint exhibits for
// this rule IS tsgolint's, literally: the release binary shells out to a `tsgolint` executable and
// refuses to run the rule when that binary is absent, which is what it does on this machine. All
// forty eight of tsgolint's own cases are replayed in the test file and all forty eight agree.
//
// # The aggregator arm, which tsgolint never had
//
// `@typescript-eslint` 8.67.0 carries a FOURTH listener that tsgolint has no counterpart for: a
// `CallExpression` arm reporting `invalidPromiseAggregatorInput` when `Promise.all`, `allSettled`,
// `race` or `any` receives a value that is not a Thenable. tsgolint registers exactly three
// listeners and predates the feature, so this file first reproduced its absence as silence, to keep
// a differential against oxlint meaningful.
//
// That decision is reversed. The parity doctrine (2026-10-01) is never worse than ESLint rule by
// rule, and the absence was measured costing a true positive: `modules/art/ArtGenerator.ts:210`
// hands `Promise.all` the result of `getDailyPartialContext()`, which is synchronous and returns a
// `string`, so that element is not "running in parallel", it already ran. ESLint reported it and
// cohere did not. The arm below is a port of the installed rule, and the reference corpus is now
// typescript-eslint's whole 122-case corpus rather than tsgolint's 48; see
// TestAwaitThenableUpstreamCorpus.
//
// Two judgments with different quantifiers, both upstream's. An array literal's element reports
// only when EVERY union part of its constrained type is never awaitable, so `number | Promise<number>`
// is fine. Any other argument reports when it is iterable and SOME element type (a tuple's members,
// an array-like's number index, an `Iterable<T>`'s first type argument) has a part that is never
// awaitable, so `Promise.all(values)` over `(number | Promise<number>)[]` reports while the same
// union written inline does not. The asymmetry is upstream's and is reproduced.
//
// # What decides a finding
//
// `NeedsToBeAwaited` returns a tri-state and only `Never` reports. It resolves a generic to its
// base constraint first, treats an unconstrained generic and `any` and `unknown` as `May`, and asks
// `IsThenableType` for the rest, which walks the union parts of the apparent type looking for a
// `then` property whose call signatures take a callback first parameter. So `await value` inside
// `wrapper<T>(value: T)` is silent while `wrapper<T extends number>(value: T)` reports, and the
// corpus pins eight cases around exactly that boundary. A port written from the rule file alone
// would miss the constraint resolution entirely, because it lives in the helper.
//
// The other two arms ask a different question: whether the type carries the well-known symbol
// `asyncIterator` or `asyncDispose`, looked up through `GetPropertyNameForKnownSymbolName`. Those
// are not `then` checks and do not go through the tri-state.
//
// # The checker, and the nil guard that now lives here
//
// Every listener reads `ctx.TypeChecker` unconditionally, so this needs the checker and its
// fixtures use `RunTyped`. While this rule was adapted, the standing
// `if ctx.TypeChecker == nil { return }` could not be written, because the listeners were
// upstream's and editing them would have turned a re-sync into a merge. Absorbing the rule removes
// that constraint, so the guard is now written where the advice always wanted it.
//
// It is placed in `Run` rather than repeated at the top of all three listeners, which declines the
// file once instead of three times per node and is the cheapest thing a rule can do. The guard is
// unreachable through registration, since `NeedsTypeChecker` is declared right above; it is here
// for the harness path, where a Context can be built by hand. A test in this package pins both the
// declaration and the decline, so a later revert fails loudly rather than going vacuously green.
//
// # Cost
//
// The anchor is what to price, not the checker. `KindAwaitExpression` and `KindForOfStatement` are
// uncommon nodes, and `KindVariableDeclarationList` is common but exits on the first line for
// anything that is not `await using`. Measured on the real tree, see the test file's note.
var AwaitThenable = rule.Rule{
	Name: "@typescript-eslint/await-thenable",

	// All four listeners read ctx.TypeChecker unconditionally, so the checker is required.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	// Compiler options and the default library, through type_checking's builtin and specifier helpers.
	ProgramReads: rule.ReadsCompilerOptions | rule.ReadsDefaultLibrary,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.TypeChecker == nil {
			return nil
		}

		return rule.Listeners{
			ast.KindAwaitExpression: func(node *ast.Node) {
				awaitArgument := node.AsAwaitExpression().Expression
				awaitArgumentType := ctx.TypeChecker.GetTypeAtLocation(awaitArgument)
				certainty := type_checking.NeedsToBeAwaited(ctx.TypeChecker, awaitArgument, awaitArgumentType)

				if certainty == type_checking.TypeAwaitableNever {
					ctx.ReportNodeWithSuggestions(node, buildAwaitMessage(), rule.Suggestion{
						Message: buildRemoveAwaitMessage(),
						Fixes: []rule.Fix{
							rule.RemoveRange(scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, node.Pos())),
						},
					})
				}
			},
			ast.KindCallExpression: func(node *ast.Node) {
				if !awaitThenableIsPromiseAggregatorCall(ctx, node) {
					return
				}
				arguments := node.Arguments()
				if len(arguments) == 0 {
					return
				}
				// ESTree drops parentheses, so upstream sees `Promise.all(([a]))` as an array
				// expression and reports each element without its parentheses.
				argument := ast.SkipParentheses(arguments[0])
				if argument.Kind == ast.KindSpreadElement {
					return
				}

				if ast.IsArrayLiteralExpression(argument) {
					for _, element := range argument.Elements() {
						if element.Kind == ast.KindOmittedExpression {
							continue
						}
						element = ast.SkipParentheses(element)
						elementType := type_checking.GetConstrainedTypeAtLocation(ctx.TypeChecker, element)
						if awaitThenableIsAlwaysNonAwaitable(ctx, element, elementType) {
							ctx.ReportNode(element, buildInvalidPromiseAggregatorInputMessage())
						}
					}
					return
				}

				argumentType := type_checking.GetConstrainedTypeAtLocation(ctx.TypeChecker, argument)
				if awaitThenableIsInvalidAggregatorInput(ctx, argument, argumentType) {
					ctx.ReportNode(argument, buildInvalidPromiseAggregatorInputMessage())
				}
			},
			ast.KindForOfStatement: func(node *ast.Node) {
				stmt := node.AsForInOrOfStatement()
				if stmt.AwaitModifier == nil {
					return
				}

				exprType := ctx.TypeChecker.GetTypeAtLocation(stmt.Expression)
				if type_checking.IsTypeAnyType(exprType) {
					return
				}

				for typePart := range type_checking.UnionTypePartsSeq(exprType) {
					if type_checking.GetWellKnownSymbolPropertyOfType(typePart, "asyncIterator", ctx.TypeChecker) != nil {
						return
					}
				}

				ctx.ReportRangeWithSuggestions(
					type_checking.GetForStatementHeadLoc(ctx.SourceFile, node),
					buildForAwaitOfNonAsyncIterableMessage(),
					// Note that this suggestion causes broken code for sync iterables
					// of promises, since the loop variable is not awaited.
					rule.Suggestion{
						Message: buildConvertToOrdinaryForMessage(),
						Fixes: []rule.Fix{
							ctx.RemoveNode(stmt.AwaitModifier),
						},
					},
				)
			},
			ast.KindVariableDeclarationList: func(node *ast.Node) {
				if !ast.IsVarAwaitUsing(node) {
					return
				}

				declaration := node.AsVariableDeclarationList()
			DeclaratorLoop:
				for _, declarator := range declaration.Declarations.Nodes {
					init := declarator.Initializer()
					if init == nil {
						continue
					}
					initType := ctx.TypeChecker.GetTypeAtLocation(init)
					if type_checking.IsTypeAnyType(initType) {
						continue
					}

					for typePart := range type_checking.UnionTypePartsSeq(initType) {
						if type_checking.GetWellKnownSymbolPropertyOfType(typePart, "asyncDispose", ctx.TypeChecker) != nil {
							continue DeclaratorLoop
						}
					}

					var suggestions []rule.Suggestion
					// let the user figure out what to do if there's
					// await using a = b, c = d, e = f;
					// it's rare and not worth the complexity to handle.
					if len(declaration.Declarations.Nodes) == 1 {
						suggestions = append(suggestions, rule.Suggestion{
							Message: buildRemoveAwaitMessage(),
							Fixes: []rule.Fix{
								rule.RemoveRange(scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, node.Pos())),
							},
						})
					}

					ctx.ReportNodeWithSuggestions(init, buildAwaitUsingOfNonAsyncDisposableMessage(), suggestions...)
				}
			},
		}
	},
}

// awaitThenablePromiseAggregators are the `PromiseConstructor` methods that take an iterable of
// promises, which is upstream's `PROMISE_CONSTRUCTOR_ARRAY_METHODS`.
var awaitThenablePromiseAggregators = map[string]bool{
	"all":        true,
	"allSettled": true,
	"race":       true,
	"any":        true,
}

// awaitThenableIsPromiseAggregatorCall is upstream's `isPromiseAggregatorMethod`: a member call
// whose static name is an aggregator and whose receiver's constrained type is the default
// library's `PromiseConstructor`, so `Promise.all` and a `typeof Promise` alias qualify and an
// object that merely has an `all` method does not.
//
// Upstream reads the name through `getStaticMemberAccessValue`, which also resolves a constant
// variable inside brackets (`const key = 'all'; Promise[key]`). That form is not followed here and
// stays silent; nothing in the corpus or the ahra tree writes it.
func awaitThenableIsPromiseAggregatorCall(ctx rule.Context, call *ast.Node) bool {
	callee := ast.SkipParentheses(call.Expression())
	var receiver *ast.Node
	var name string
	switch {
	case ast.IsPropertyAccessExpression(callee):
		access := callee.AsPropertyAccessExpression()
		text, ok := property.Name(access.Name(), property.Named)
		if !ok {
			return false
		}
		receiver, name = access.Expression, text
	case ast.IsElementAccessExpression(callee):
		access := callee.AsElementAccessExpression()
		text, ok := property.Name(ast.SkipParentheses(access.ArgumentExpression), property.Quoted|property.Templated)
		if !ok {
			return false
		}
		receiver, name = access.Expression, text
	default:
		return false
	}
	if !awaitThenablePromiseAggregators[name] {
		return false
	}
	receiverType := type_checking.GetConstrainedTypeAtLocation(ctx.TypeChecker, receiver)
	return type_checking.IsPromiseConstructorLike(ctx.Program, ctx.TypeChecker, receiverType)
}

// awaitThenableIsAlwaysNonAwaitable is upstream's `isAlwaysNonAwaitableType`: every union part is
// never awaitable.
func awaitThenableIsAlwaysNonAwaitable(ctx rule.Context, node *ast.Node, t *checker.Type) bool {
	for part := range type_checking.UnionTypePartsSeq(t) {
		if type_checking.NeedsToBeAwaited(ctx.TypeChecker, node, part) != type_checking.TypeAwaitableNever {
			return false
		}
	}
	return true
}

// awaitThenableContainsNonAwaitable is upstream's `containsNonAwaitableType`: some union part is
// never awaitable.
func awaitThenableContainsNonAwaitable(ctx rule.Context, node *ast.Node, t *checker.Type) bool {
	for part := range type_checking.UnionTypePartsSeq(t) {
		if type_checking.NeedsToBeAwaited(ctx.TypeChecker, node, part) == type_checking.TypeAwaitableNever {
			return true
		}
	}
	return false
}

// awaitThenableIsInvalidAggregatorInput is upstream's `isInvalidPromiseAggregatorInput` for an
// argument that is not an array literal.
//
// A non-iterable argument is already a type error, so it is left to the checker. An iterable one
// reports when any of its value types contains a part that is never awaitable.
func awaitThenableIsInvalidAggregatorInput(ctx rule.Context, node *ast.Node, t *checker.Type) bool {
	for part := range type_checking.UnionTypePartsSeq(t) {
		if type_checking.GetWellKnownSymbolPropertyOfType(part, "iterator", ctx.TypeChecker) == nil {
			return false
		}
	}
	for part := range type_checking.UnionTypePartsSeq(t) {
		for _, valueType := range awaitThenableValueTypesOfArrayLike(ctx, part) {
			if awaitThenableContainsNonAwaitable(ctx, node, valueType) {
				return true
			}
		}
	}
	return false
}

// awaitThenableValueTypesOfArrayLike is upstream's `getValueTypesOfArrayLike`: a tuple's members, an
// array-like's number index type, or an `Iterable<T>`'s first type argument.
func awaitThenableValueTypesOfArrayLike(ctx rule.Context, t *checker.Type) []*checker.Type {
	if checker.IsTupleType(t) {
		return ctx.TypeChecker.GetTypeArguments(t)
	}
	if ctx.TypeChecker.IsArrayLikeType(t) {
		if indexType := ctx.TypeChecker.GetNumberIndexType(t); indexType != nil {
			return []*checker.Type{indexType}
		}
		return nil
	}
	if checker.Type_objectFlags(t)&checker.ObjectFlagsReference != 0 {
		typeArguments := ctx.TypeChecker.GetTypeArguments(t)
		if len(typeArguments) > 1 {
			typeArguments = typeArguments[:1]
		}
		return typeArguments
	}
	return nil
}

func buildInvalidPromiseAggregatorInputMessage() rule.Message {
	return rule.Message{
		Id:          "invalidPromiseAggregatorInput",
		Description: "Unexpected iterable of non-Promise (non-\"Thenable\") values passed to promise aggregator.",
	}
}

func buildAwaitMessage() rule.Message {
	return rule.Message{
		Id:          "await",
		Description: "Unexpected `await` of a non-Promise (non-\"Thenable\") value.",
	}
}

func buildRemoveAwaitMessage() rule.Message {
	return rule.Message{
		Id:          "removeAwait",
		Description: "Remove unnecessary `await`.",
	}
}

func buildForAwaitOfNonAsyncIterableMessage() rule.Message {
	return rule.Message{
		Id:          "forAwaitOfNonAsyncIterable",
		Description: "Unexpected `for await...of` of a value that is not async iterable.",
	}
}

func buildConvertToOrdinaryForMessage() rule.Message {
	return rule.Message{
		Id:          "convertToOrdinaryFor",
		Description: "Convert to an ordinary `for...of` loop.",
	}
}

func buildAwaitUsingOfNonAsyncDisposableMessage() rule.Message {
	return rule.Message{
		Id:          "awaitUsingOfNonAsyncDisposable",
		Description: "Unexpected `await using` of a value that is not async disposable.",
	}
}
