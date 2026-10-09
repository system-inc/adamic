package typescript

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// NoUnsafeCall flags calling, constructing, or tagging a template with a value typed `any`.
//
//	valid:   function foo(x: () => void) { x(); }
//	valid:   new Map();
//	valid:   String.raw`foo`;
//	valid:   interface SurprisinglySafe extends Function { (): string }  declare const safe: SurprisinglySafe; safe();
//	invalid: function foo(x: any) { x(); }
//	invalid: function foo(x: any) { new x(); }
//	invalid: function foo(x: any) { x`foo`; }
//	invalid: const t: Function = () => {}; t();
//
// Calling an `any` turns off every check the arguments and the return value would otherwise get, so
// a mistake at that call site is invisible until it runs. The bare `Function` type is the same
// problem wearing a type: it accepts any arguments and returns `any`, so it documents "callable"
// and checks nothing.
//
// # Four call sites, eight messages, and the pairing is what carries the meaning
//
// The rule fires on a call's callee, a `new` expression's callee, and a tagged template's tag. Each
// site has a pair of messages: an `unsafe` one naming what the type was, and an `error` one for a
// type the checker could not resolve at all. `NotKnown` in an annotation is not `any` the author
// wrote, it is the checker's error type, and saying "unsafe call of an `any` typed value" about it
// would send the reader looking for an `any` that is not there. The distinguishing test is
// `IsIntrinsicErrorType`, which reads the intrinsic name rather than the flags, because the error
// type carries the `any` flag too and a flag test alone cannot separate them.
//
// # The `Function` branch, whose asymmetry is upstream's and is not obvious
//
// A type that is or extends `Function` is judged by which signatures it has, and construction and
// calling ask DIFFERENT questions. Upstream's own comment says it does not understand why, and it
// is reproduced rather than tidied:
//
//	any construct signature            safe to both call and construct
//	a call signature                   safe to CALL and to tag
//	a NON-void-returning call signature safe to CONSTRUCT
//
// So `interface CallGoodConstructBad extends Function { (): void }` is safe to call and unsafe to
// construct, while adding an overload returning `string` makes it safe to construct as well. Three
// of upstream's passing cases and one of its reporting ones turn on exactly this, and a port that
// asked the same question at both sites would pass most of the corpus and fail those four.
//
// `Function` here means the global one. Upstream resolves it through `isBuiltinSymbolLike`, which
// checks the symbol came from a default library file, so a locally declared `type Function = () =>
// void` is an ordinary function type and stays silent. The corpus writes that case in a block scope
// precisely to pin it.
//
// # The `this` branch, and why no fixture in this package can reach it
//
// When `noImplicitThis` is OFF, an unannotated `this` is implicitly `any`, and upstream swaps the
// message for one that says so and suggests turning the compiler option on. That branch is ported
// below and it is UNREACHABLE THROUGH THIS PACKAGE'S TESTS, because the harness writes its own
// tsconfig with `strict: true` and no override, and it writes it AFTER the setup hook runs, so
// nothing a fixture can do changes the option. Probed rather than assumed, with a control: a rule
// reading `ctx.Program.Options()` under both `RunTyped` and `RunTypedFilesWithSetup` reports
// `strict true, noImplicitThis isTrueOrUnknown true` identically, and the hook's write of a
// different tsconfig is discarded.
//
// Upstream's entire corpus for this rule runs under `tsconfig.noImplicitThis.json`, so its two
// `this` cases assert message ids our harness cannot produce. Both were measured on the installed
// 8.67.0 build under BOTH settings rather than dropped:
//
//	                                      noImplicitThis false     noImplicitThis true (ours)
//	object literal `this.methodB()`       unsafeCallThis x2        SILENT
//	`function f(this: NotKnown) { ... }`  errorCallThis x2         errorCall x2
//
// The fixtures below assert the right-hand column, which is what upstream itself produces under our
// configuration, so the imported cases are pinned at the verdict our harness can actually observe
// rather than weakened or deleted. The left-hand column is recorded here because the branch is real
// and a later harness that can set compiler options should assert it.
//
// # Cost
//
// Three anchors, all common, and each one asks the checker immediately. That is upstream's shape
// too, and there is no cheaper pre-test available: whether a callee is `any` is precisely the
// question, so nothing syntactic can decline first.
var NoUnsafeCall = rule.Rule{
	Name: "@typescript-eslint/no-unsafe-call",

	// Every finding is a type question about the callee. The checker is required.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	// Compiler options and the default library, through type_checking's builtin and specifier helpers.
	ProgramReads: rule.ReadsCompilerOptions | rule.ReadsDefaultLibrary,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		// checkCall is upstream's single body, shared by all three anchors.
		//
		// `subject` is the thing being called and `reportOn` is where the finding points. They are
		// the same node for a call and for a tagged template, and they differ for `new`, where
		// upstream reports the whole NewExpression while typing its callee.
		checkCall := func(subject *ast.Node, reportOn *ast.Node, unsafeId string, errorId string) {
			if ctx.TypeChecker == nil {
				return
			}

			subjectType := type_checking.GetConstrainedTypeAtLocation(ctx.TypeChecker, subject)
			if subjectType == nil {
				return
			}

			if type_checking.IsTypeAnyType(subjectType) {
				// The `this` swap, gated on the compiler option. See the doc comment: this arm is
				// correct and unreachable through this package's fixtures, because the harness
				// pins the option above the rule.
				if !type_checking.IsStrictCompilerOptionEnabled(
					ctx.Program.Options(), ctx.Program.Options().NoImplicitThis) {
					// A mutation replacing this call with `subject` SURVIVES the whole fixture set,
					// and it is not a blind spot: the enclosing gate is false under this harness's
					// tsconfig, so nothing reaches the line at all. Scored both ways to be sure of
					// the reason rather than assuming it. Mutating this line alone survives;
					// mutating it TOGETHER with the gate above fails 20 lines, which is the
					// control proving the line is load-bearing and only the gate is unreachable.
					// A single-site sweep structurally cannot see this.
					thisExpression := type_checking.GetThisExpression(subject)
					if thisExpression != nil {
						thisType := type_checking.GetConstrainedTypeAtLocation(ctx.TypeChecker, thisExpression)
						if thisType != nil && type_checking.IsTypeAnyType(thisType) {
							unsafeId = "unsafeCallThis"
							errorId = "errorCallThis"
						}
					}
				}

				// The error type carries the `any` flag, so the flags cannot separate them and the
				// intrinsic name does. Getting this backwards renames every unresolved-type finding
				// into one claiming an `any` the reader never wrote.
				if type_checking.IsIntrinsicErrorType(subjectType) {
					ctx.ReportNode(reportOn, unsafeCallMessageFor(errorId, ""))
					return
				}
				ctx.ReportNode(reportOn, unsafeCallMessageFor(unsafeId, "an `any`"))
				return
			}

			if !type_checking.IsBuiltinSymbolLike(ctx.Program, ctx.TypeChecker, subjectType, "Function") {
				return
			}

			// The asymmetry, reproduced from upstream including its own admission that the reason
			// is unclear. A construct signature makes every use safe; otherwise calling and tagging
			// want any call signature at all, while constructing wants one that returns something.
			if len(type_checking.GetConstructSignatures(ctx.TypeChecker, subjectType)) > 0 {
				return
			}

			callSignatures := type_checking.GetCallSignatures(ctx.TypeChecker, subjectType)
			if unsafeId == "unsafeNew" {
				for _, signature := range callSignatures {
					returnType := checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, signature)
					if returnType != nil && !type_checking.IsIntrinsicVoidType(returnType) {
						return
					}
				}
			} else if len(callSignatures) > 0 {
				return
			}

			// Note the id: the `Function` branch reports the UNSAFE message even for a type the
			// checker resolved perfectly well, because the type is known and the problem is what it
			// permits. There is no error-type path here at all.
			ctx.ReportNode(reportOn, unsafeCallMessageFor(unsafeId, "a `Function`"))
		}

		return rule.Listeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				if call.Expression == nil {
					return
				}
				// A dynamic `import(...)` is a call expression in OUR parser and is not one in
				// estree, which models it as its own ImportExpression, so upstream's
				// `CallExpression > *.callee` selector never sees it. Left alone here for that
				// reason rather than by preference.
				//
				// The consequence of dropping this is not subtle and is not a difference of taste:
				// probed, the callee is a bare `KindImportKeyword` whose type is the checker's
				// ERROR type, so it satisfies the any test AND the error-type test, and every
				// dynamic import in the tree would report "Unsafe call of a type that could not be
				// resolved". Measured on the installed build, `import('./foo')` and
				// `await import('./nowhere')` are both silent, and upstream's corpus carries the
				// first as a PASSING case, which is how this was found.
				if call.Expression.Kind == ast.KindImportKeyword {
					return
				}
				// Upstream's selector reports on the callee itself rather than the whole call, so
				// `x.a.b.c()` underlines `x.a.b.c`.
				checkCall(skipCalleeParentheses(call.Expression), skipCalleeParentheses(call.Expression),
					"unsafeCall", "errorCall")
			},
			ast.KindNewExpression: func(node *ast.Node) {
				newExpression := node.AsNewExpression()
				if newExpression.Expression == nil {
					return
				}
				// The one anchor where the typed node and the reported node differ: upstream types
				// `node.callee` and reports `node`, so the finding underlines `new x()` entire.
				checkCall(skipCalleeParentheses(newExpression.Expression), node, "unsafeNew", "errorNew")
			},
			ast.KindTaggedTemplateExpression: func(node *ast.Node) {
				tagged := node.AsTaggedTemplateExpression()
				if tagged.Tag == nil {
					return
				}
				checkCall(skipCalleeParentheses(tagged.Tag), skipCalleeParentheses(tagged.Tag),
					"unsafeTemplateTag", "errorTemplateTag")
			},
		}
	},
}

// unsafeCallMessageFor renders one of the rule's eight messages.
//
// Upstream interpolates `{{type}}` with the literal strings "an `any`" and "a `Function`", never
// with a type name, so the slot has exactly two fillings and the error-type messages have none. The
// backticks are part of upstream's message text rather than markup this port added.
//
// The `unsafeCallThis` text is two lines joined by a newline, which is upstream's own shape: the
// second line is advice about the compiler option rather than a separate suggestion, because there
// is no edit to propose.
func unsafeCallMessageFor(messageId string, typeText string) rule.Message {
	switch messageId {
	case "errorCall":
		return rule.Message{Id: messageId, Description: "Unsafe call of a type that could not be resolved."}
	case "errorCallThis":
		return rule.Message{Id: messageId, Description: "Unsafe call of a `this` type that could not be resolved."}
	case "errorNew":
		return rule.Message{Id: messageId, Description: "Unsafe construction of a type that could not be resolved."}
	case "errorTemplateTag":
		return rule.Message{Id: messageId, Description: "Unsafe use of a template tag whose type could not be resolved."}
	case "unsafeCallThis":
		return rule.Message{Id: messageId, Description: "Unsafe call of " + typeText + " typed value. `this` is typed as " +
			typeText + ".\nYou can try to fix this by turning on the `noImplicitThis` compiler option, " +
			"or adding a `this` parameter to the function."}
	case "unsafeNew":
		return rule.Message{Id: messageId, Description: "Unsafe construction of " + typeText + " typed value."}
	case "unsafeTemplateTag":
		return rule.Message{Id: messageId, Description: "Unsafe use of " + typeText + " typed template tag."}
	default:
		return rule.Message{Id: "unsafeCall", Description: "Unsafe call of " + typeText + " typed value."}
	}
}

// skipCalleeParentheses unwraps a parenthesized callee, tag, or construction target.
//
// estree has no parenthesized-expression node, so upstream's selectors bind to the inner expression
// and the parentheses never reach the rule. Our parser does produce one, and both the TYPE and the
// SPAN would differ without this: the checker answers the same type either way, but the finding
// would underline `(x)` where upstream underlines `x`.
//
// Measured on the installed 8.67.0 build rather than reasoned about, because the corpus writes no
// parenthesized form anywhere and so has no opinion:
//
//	(x)()      reports on `x`     columns 4-5
//	new (x)()  reports on `new (x)()`  the whole expression, as every `new` does
//	(x)`tag`   reports on `x`     columns 4-5
//	(x.a)()    reports on `x.a`   columns 4-7
//
// SkipParentheses dereferences its argument, so every caller nil-tests first.
func skipCalleeParentheses(callee *ast.Node) *ast.Node {
	return ast.SkipParentheses(callee)
}
