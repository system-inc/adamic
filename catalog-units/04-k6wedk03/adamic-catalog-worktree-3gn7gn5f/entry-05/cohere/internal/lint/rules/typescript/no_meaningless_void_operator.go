package typescript

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// NoMeaninglessVoidOperatorOptions is the rule's option surface.
type NoMeaninglessVoidOperatorOptions struct {
	// CheckNever extends the rule to an argument of type `never`, as a SUGGESTION rather than a fix.
	//
	// It defaults to FALSE, so unlike the sibling rule in this package the generic decoder happens
	// to produce the right zero value. The decoder below is still hand-rolled, because relying on
	// that coincidence means the next person to add a default-true key inherits a decoder that
	// silently does the wrong thing.
	CheckNever bool
}

// DefaultNoMeaninglessVoidOperatorSettings is upstream's `defaultOptions`.
func DefaultNoMeaninglessVoidOperatorSettings() NoMeaninglessVoidOperatorOptions {
	return NoMeaninglessVoidOperatorOptions{CheckNever: false}
}

// noMeaninglessVoidOperatorRawOptions is the wire shape, with a pointer so an absent key stays
// distinguishable from an explicit false.
type noMeaninglessVoidOperatorRawOptions struct {
	CheckNever *bool `json:"checkNever"`
}

// DecodeNoMeaninglessVoidOperatorOptions reads the rule's configuration.
//
// cohere's config layer strips ESLint's `[severity, options]` tuple before dispatch, so what
// arrives is the bare object rather than upstream's one-element array.
func DecodeNoMeaninglessVoidOperatorOptions(raw []byte) (any, error) {
	decoded, err := rule.DecodeOptionsInto[noMeaninglessVoidOperatorRawOptions]()(raw)
	if err != nil {
		return DefaultNoMeaninglessVoidOperatorSettings(), err
	}

	wire, _ := decoded.(noMeaninglessVoidOperatorRawOptions)
	options := DefaultNoMeaninglessVoidOperatorSettings()
	if wire.CheckNever != nil {
		options.CheckNever = *wire.CheckNever
	}
	return options, nil
}

// NoMeaninglessVoidOperator flags `void` that discards nothing: applied to something other than a
// call, or to a call that already returns void or undefined.
//
//	valid:   declare function bar(): number; void bar();  the value discarded is a call's number
//	valid:   const a = void 0;                            the undefined idiom
//	valid:   void (x = 1);                                an assignment, so the expression is undefined
//	valid:   declare const p: Promise<void>; void p;      a thenable, left to no-floating-promises
//	invalid: declare let x: number; void x;               not a call, so nothing returned is discarded
//	invalid: function foo() {} void foo();                foo() is already void
//	invalid: void (() => {})();
//
// `void` exists to say "this expression returns something and I am deliberately throwing it away".
// Applied to an expression that already evaluates to `void` or `undefined` it says nothing, so the
// reader is left looking for a discarded value that was never there.
//
// # A non-call discards nothing a caller produced (8.71)
//
// typescript-eslint 8.71 added `meaninglessVoidOnNonCall`. `void value;` evaluates a name and throws
// it away, which does nothing, and it is usually a leftover: a variable kept "used" for a linter, or
// a read meant to force something that the React Compiler then erases. Structure's Time.tsx had
// exactly that, a `void tickCount` the compiler dropped, so the clock never ticked (#bab6yg0).
//
// Before deciding, the argument is unwrapped the way upstream unwraps it: through parentheses (which
// ESTree drops), `as`, `!`, `satisfies`, `<T>`, and to the last expression of a comma sequence. An
// instantiation such as `make<string>` is not unwrapped, so it is a non-call. Three non-calls are
// left alone: the literal `0` (`void 0` is the undefined idiom), an assignment (`() => void (x = 1)`
// returns undefined on purpose), and a thenable, because no-floating-promises asks for `void p`.
// The repair removes the operator only where the void is a whole statement, since anywhere else
// the expression's value changes from undefined to the argument's.
//
// The type test below then runs only on a call, which is also new in 8.71: `void x` with `x: void`
// is now a non-call finding rather than a "used on void" one.
//
// # Two branches, and the SAME repair is a fix in one and a suggestion in the other
//
// That distinction is the most important thing in this rule and it is easy to flatten.
//
// When every constituent of the argument's type is `void` or `undefined`, the repair is a FIX:
// removing the operator cannot change what the program does, because the expression already
// evaluated to the value `void` would have produced.
//
// When `checkNever` is on and every constituent is `void`, `undefined` or `never`, the repair is a
// SUGGESTION. `never` means control does not return from that expression at all, so removing the
// operator is a judgment about code the type system says is unreachable, and upstream declines to
// make it unattended. Shipping that as a fix would have the edit engine rewriting it silently.
//
// The two tests are nested rather than exclusive, and the order matters: `void | never` under
// `checkNever` satisfies the FIRST predicate, because a union's `every` is over its constituents and
// `never` is absorbed out of a union by the checker before the rule sees it. Measured on the
// installed build: `declare const x: void | never; void x;` reports with a fix, not a suggestion.
//
// # The message interpolates the type, and the corpus never checks it
//
// `{{type}}` is `checker.typeToString(argType)`, so the same rule reports "on void", "on undefined",
// "on never" and "on void | undefined" for different inputs. Upstream's corpus carries five cases
// and not one of them asserts the `data` field, so nothing there can see the interpolation at all.
// Every fixture in this package asserts the whole rendered string, because a message-id assertion
// cannot see a format slot that is filled with the wrong thing.
//
// # What the fix removes, including a comment
//
// Upstream removes the range from the start of the first token to the start of the SECOND, which is
// the `void` keyword plus everything between it and the argument. That is whitespace usually, and
// it is not always:
//
//	void  foo()          removes "void  "        two spaces
//	void /* c */ foo()   removes "void /* c */ " the comment goes with it
//	void(foo())          removes "void"          no separator to take
//
// The comment case is upstream's behavior rather than an accident of this port, measured directly,
// and it is reproduced. A repair that kept the comment would be a difference the differential
// harness can see, and the comment is attached to an operator that is being deleted anyway.
//
// # Cost
//
// `KindVoidExpression` is a rare anchor and the checker is consulted only when one is found.
var NoMeaninglessVoidOperator = rule.Rule{
	Name: "@typescript-eslint/no-meaningless-void-operator",

	// The argument's type decides every finding, so the checker is required.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, isSettings := rule.OptionsAs[NoMeaninglessVoidOperatorOptions](options)
		if !isSettings {
			settings = DefaultNoMeaninglessVoidOperatorSettings()
		}

		// everyConstituentHasFlag is upstream's `unionConstituents(...).every(isTypeFlagSet(...))`.
		// A non-union is its own single constituent, so this reads the same for both shapes.
		everyConstituentHasFlag := func(argumentType *checker.Type, flags checker.TypeFlags) bool {
			for constituent := range type_checking.UnionTypePartsSeq(argumentType) {
				if !type_checking.IsTypeFlagSet(constituent, flags) {
					return false
				}
			}
			return true
		}

		return rule.Listeners{
			ast.KindVoidExpression: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}

				argument := node.AsVoidExpression().Expression
				if argument == nil {
					return
				}

				argumentType := ctx.TypeChecker.GetTypeAtLocation(argument)
				if argumentType == nil {
					return
				}

				// The removal range: from the `void` keyword's own start to the argument's start,
				// which is upstream's "first token start to second token start".
				//
				// `rule.TokenRange` supplies the argument's start rather than `Pos()`, because
				// `Pos()` includes leading trivia and would put the range's END inside the
				// whitespace, leaving a stray fragment behind. That is the fix hazard the span
				// hazard is usually described as, arriving through the other side of the range.
				voidKeyword := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, node.Pos())
				removal := core.NewTextRange(voidKeyword.Pos(), rule.TokenRange(ctx.SourceFile, argument).Pos())

				inner := unwrapVoidArgument(argument)
				if !isDiscardedCall(inner) {
					if inner.Kind == ast.KindNumericLiteral && inner.Text() == "0" {
						// `void 0` is the undefined idiom. The parser renders a literal's text as
						// its canonical decimal, so `0x0` and `0.0` arrive as "0" too, as upstream's
						// `value === 0` reads them. `0n` and `-0` are not this literal and report.
						return
					}
					if ast.IsAssignmentExpression(inner, false) {
						// `void (x = value)` makes the expression undefined on purpose, compound
						// and logical assignment included.
						return
					}
					if type_checking.IsThenableType(ctx.TypeChecker, argument, argumentType) {
						// Left to no-floating-promises, which asks for exactly this.
						return
					}
					if isWholeStatement(node) {
						ctx.ReportNodeWithFixes(node, buildMeaninglessVoidOnNonCallMessage(),
							rule.RemoveRange(removal))
						return
					}
					// Anywhere else the void's undefined is the value, so removing it would change
					// what the expression evaluates to.
					ctx.ReportNode(node, buildMeaninglessVoidOnNonCallMessage())
					return
				}

				// Upstream interpolates `checker.typeToString(argType)`, which is this method
				// rather than the shelf's `GetTypeName`.
				//
				// Swapping the two is an EQUIVALENT mutation here, and the sweep says so: no input
				// this rule reports on can distinguish them. `GetTypeName` differs only by
				// collapsing string-like types to the word "string", and it delegates to
				// `TypeToString` for everything else, while this rule reports only when every
				// constituent is void, undefined or never. Measured over the six argument types
				// that reach a report, including the union: both spellings agree on all six.
				//
				// `TypeToString` is kept because it is what upstream calls and because the
				// equivalence is a property of which types get here rather than of the two
				// functions. A later widening of the type test would make the difference real, and
				// `GetTypeName` would then start naming a type the reader never wrote.
				typeName := ctx.TypeChecker.TypeToString(argumentType)

				if everyConstituentHasFlag(argumentType, checker.TypeFlagsVoid|checker.TypeFlagsUndefined) {
					// Removing the operator cannot change behavior here, so it is applied
					// unattended.
					ctx.ReportNodeWithFixes(node, buildMeaninglessVoidOperatorMessage(typeName),
						rule.RemoveRange(removal))
					return
				}

				if settings.CheckNever && everyConstituentHasFlag(argumentType,
					checker.TypeFlagsVoid|checker.TypeFlagsUndefined|checker.TypeFlagsNever) {
					// `never` means this expression does not return, so the same edit becomes a
					// judgment rather than a rewrite and a person chooses it.
					ctx.ReportNodeWithSuggestions(node, buildMeaninglessVoidOperatorMessage(typeName),
						rule.Suggestion{
							Message: buildRemoveVoidMessage(),
							Fixes:   []rule.Fix{rule.RemoveRange(removal)},
						})
				}
			},
		}
	},
}

// unwrapVoidArgument is upstream's `unwrapVoidArgument`, which looks through what does not change
// which expression is being discarded.
//
// ESTree drops parentheses and keeps the rest as nodes, so parentheses are one more arm here. A comma
// sequence is a binary expression in our tree, nested to the left, so its last expression is the
// right operand, and a nested sequence unwraps again on the next turn. An instantiation expression
// is deliberately absent, as it is upstream: `make<string>` names a function without calling it.
func unwrapVoidArgument(node *ast.Node) *ast.Node {
	current := node
	for {
		switch current.Kind {
		case ast.KindParenthesizedExpression, ast.KindAsExpression, ast.KindNonNullExpression,
			ast.KindSatisfiesExpression, ast.KindTypeAssertionExpression:
			current = current.Expression()
		case ast.KindBinaryExpression:
			binary := current.AsBinaryExpression()
			if binary.OperatorToken.Kind != ast.KindCommaToken {
				return current
			}
			current = binary.Right
		default:
			return current
		}
	}
}

// isDiscardedCall answers upstream's `inner.type === CallExpression`.
//
// `import(...)` is a CallExpression in our tree and an ImportExpression in ESTree, so it is a non-call
// here as it is upstream. Either way it is a thenable and stays silent, but the branch it takes is
// upstream's.
func isDiscardedCall(node *ast.Node) bool {
	return node.Kind == ast.KindCallExpression && !ast.IsImportCall(node)
}

// isWholeStatement answers upstream's `node.parent.type === ExpressionStatement`, looking through
// the parentheses ESTree drops, so `(void x);` is a statement as it is upstream.
func isWholeStatement(node *ast.Node) bool {
	parent := ast.WalkUpParenthesizedExpressions(node.Parent)
	return parent != nil && parent.Kind == ast.KindExpressionStatement
}

func buildMeaninglessVoidOnNonCallMessage() rule.Message {
	return rule.Message{
		Id:          "meaninglessVoidOnNonCall",
		Description: "void operator is useless here; it should only discard a call's return value",
	}
}

func buildMeaninglessVoidOperatorMessage(typeName string) rule.Message {
	return rule.Message{
		Id: "meaninglessVoidOperator",
		Description: "void operator shouldn't be used on " + typeName +
			"; it should convey that a return value is being ignored",
	}
}

func buildRemoveVoidMessage() rule.Message {
	return rule.Message{
		Id:          "removeVoid",
		Description: "Remove 'void'",
	}
}
