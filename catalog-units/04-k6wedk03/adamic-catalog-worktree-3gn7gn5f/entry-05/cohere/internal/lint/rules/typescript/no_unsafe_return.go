package typescript

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// NoUnsafeReturn flags returning a value typed `any` from a function typed as something else.
//
//	valid:   function foo(): unknown { return JSON.parse('') as any; }
//	valid:   function foo(): any { return {} as any; }
//	valid:   function foo(): unknown[] { return [] as any[]; }
//	valid:   const foo = () => new Set<any>();
//	invalid: function foo() { return 1 as any; }
//	invalid: function foo() { return [] as any[]; }
//	invalid: async function foo() { return Promise.resolve({} as any); }
//	invalid: function foo(): Set<string> { return new Set<any>(); }
//
// A function that returns `any` hands its caller a value the type system has stopped checking, and
// the caller has no way to know: the signature reads `Set<string>` while the value inside is
// `Set<any>`. Every use downstream inherits the hole.
//
// # Three messages, and each one describes a different defect
//
// `unsafeReturn` is the plain case, and its slot names what the value actually was: `any`, `any[]`,
// `Promise<any>`, or the word `error` when the checker could not resolve the type at all.
// `unsafeReturnAssignment` is the subtler one, for a value that is not itself `any` but carries an
// `any` in a generic position, so `Set<any>` returned where `Set<string>` was promised. And
// `unsafeReturnThis` replaces the plain message when the value is `this` and `this` is implicitly
// `any`, since the fix there is a compiler option rather than a cast.
//
// # An explicit return annotation is a decision, and the rule respects it
//
// If the function declares a return type and the returned value's type IS that type, the rule
// declines even when the type is `any`. Upstream's comment says why: it is intentional, even if
// unsafe. The same test runs a second time through `getAwaitedType` for an async function, so
// `async function f(): Promise<any> { return 1 as any; }` is a decision rather than an accident.
//
// # `unknown` absorbs `any`, and that is the whole point of `unknown`
//
// Returning `any` where `unknown` was declared is safe, because `unknown` forces the caller to
// narrow before using it. The same holds for `any[]` into `unknown[]` and for `Promise<any>` into a
// function whose awaited return is `unknown`. Three of upstream's passing cases are exactly these,
// and a port without them reports the safest thing a caller can do with an `any`.
//
// # A function expression takes its return type from the receiver, not from itself
//
//	const foo: () => Set<string> = () => new Set<any>();
//
// The arrow's own type is `() => Set<any>`, so asking the function node produces the wrong answer
// and the rule stays silent on a real finding. Upstream asks for the CONTEXTUAL type instead for a
// function expression and an arrow, and falls back to the node's own type when there is none.
//
// # The `this` branch, and why no fixture in this package can produce its id
//
// Upstream runs this rule's whole corpus under `tsconfig.noImplicitThis.json`, and our harness pins
// `strict: true` with no way for a fixture to override it. Measured on the installed 8.67.0 build
// under both settings: the corpus's `return this` case reports TWICE either way, and only the id
// moves, from `unsafeReturnThis` under upstream's configuration to `unsafeReturn` under ours. The
// branch is ported and the fixtures assert the id our harness can actually observe, with the other
// recorded here rather than dropped.
//
// # Cost
//
// Two anchors. A return statement with no argument returns immediately, and a concise arrow body is
// a single kind test on the parent, so the checker is asked only where there is a value to judge.
var NoUnsafeReturn = rule.Rule{
	Name: "@typescript-eslint/no-unsafe-return",

	// Every judgment compares the returned value's type against the function's declared one.
	NeedsTypeChecker: true,

	// The compiler options.
	ProgramReads: rule.ReadsCompilerOptions,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindReturnStatement: func(node *ast.Node) {
				argument := node.AsReturnStatement().Expression
				if argument == nil {
					return
				}
				// The finding points at the whole return STATEMENT while the type comes from its
				// argument. Upstream passes both, and collapsing them would underline `1 as any`
				// where upstream underlines `return 1 as any;`.
				checkUnsafeReturn(ctx, argument, node)
			},
			ast.KindArrowFunction: func(node *ast.Node) {
				body := node.AsArrowFunction().Body
				if body == nil || body.Kind == ast.KindBlock {
					// A block body's returns are handled by the statement anchor above. Upstream's
					// selector says the same thing with `:not(BlockStatement).body`.
					return
				}
				checkUnsafeReturn(ctx, body, body)
			},
		}
	},
}

// checkUnsafeReturn is upstream's `checkReturn`, the single body both anchors share.
//
// `returnNode` is the value being returned and `reportingNode` is where the finding points. They
// are the same node for a concise arrow body and they differ for a return statement.
func checkUnsafeReturn(ctx rule.Context, returnNode *ast.Node, reportingNode *ast.Node) {
	if ctx.TypeChecker == nil {
		return
	}

	functionNode := type_checking.GetParentFunctionNode(returnNode)
	if functionNode == nil {
		// Upstream marks this unreachable with an istanbul directive. Kept as a guard rather than
		// a panic, because our parser recovers from broken source and can hand back a return
		// outside any function.
		return
	}

	returnNodeType := ctx.TypeChecker.GetTypeAtLocation(returnNode)
	if returnNodeType == nil {
		return
	}
	anyType := type_checking.DiscriminateAnyType(returnNodeType, ctx.TypeChecker, ctx.Program, returnNode)
	constrainedReturnNodeType := type_checking.GetConstrainedTypeAtLocation(ctx.TypeChecker, returnNode)

	functionType := unsafeReturnFunctionType(ctx, functionNode)
	if functionType == nil {
		return
	}
	// Upstream uses `tsutils.getCallSignaturesOfType`, which FLATTENS a union and an intersection.
	// The shelf's `GetCallSignatures` does not, and the difference is not academic: an arrow passed
	// to `foo(arg: null | (() => any))` has contextual type `(() => any) | null`, which reports zero
	// call signatures unflattened. That silently skipped the explicit-annotation check and turned
	// one of upstream's own passing cases into a finding. `CollectAllCallSignatures` is the shelf's
	// equivalent and its doc comment names the upstream function it mirrors.
	callSignatures := type_checking.CollectAllCallSignatures(ctx.TypeChecker, functionType)

	isAsync := unsafeReturnIsAsync(functionNode)

	// An explicit return annotation that matches what is actually returned is a decision the author
	// made, so the rule declines even when the type is `any`.
	if unsafeReturnTypeAnnotation(functionNode) != nil {
		for _, signature := range callSignatures {
			signatureReturnType := checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, signature)
			if signatureReturnType == nil {
				continue
			}
			if returnNodeType == signatureReturnType ||
				type_checking.IsTypeFlagSet(signatureReturnType,
					checker.TypeFlagsAny|checker.TypeFlagsUnknown) {
				return
			}
			if !isAsync {
				continue
			}
			// The same comparison one layer in, so an async function declaring `Promise<any>` is
			// recognised as the same decision.
			awaitedSignatureReturnType := checker.Checker_getAwaitedType(ctx.TypeChecker, signatureReturnType)
			awaitedReturnNodeType := checker.Checker_getAwaitedType(ctx.TypeChecker, returnNodeType)
			if awaitedReturnNodeType == awaitedSignatureReturnType ||
				(awaitedSignatureReturnType != nil &&
					type_checking.IsTypeFlagSet(awaitedSignatureReturnType,
						checker.TypeFlagsAny|checker.TypeFlagsUnknown)) {
				return
			}
		}
	}

	if anyType != type_checking.DiscriminatedAnyTypeSafe {
		// `unknown` is what a caller is supposed to receive when the value cannot be trusted, so
		// returning `any` into it is the safe move rather than the unsafe one.
		for _, signature := range callSignatures {
			functionReturnType := checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, signature)
			if functionReturnType == nil {
				continue
			}
			if anyType == type_checking.DiscriminatedAnyTypeAny &&
				type_checking.IsTypeUnknownType(functionReturnType) {
				return
			}
			if anyType == type_checking.DiscriminatedAnyTypeAnyArray &&
				type_checking.IsTypeUnknownArrayType(functionReturnType, ctx.TypeChecker) {
				return
			}
			awaitedType := checker.Checker_getAwaitedType(ctx.TypeChecker, functionReturnType)
			if awaitedType != nil && anyType == type_checking.DiscriminatedAnyTypePromiseAny &&
				type_checking.IsTypeUnknownType(awaitedType) {
				return
			}
		}

		// A `Promise<any>` returned from a SYNCHRONOUS function is not an unsafe return: the
		// promise is the value, and the `any` inside it is somebody else's problem at the await.
		if anyType == type_checking.DiscriminatedAnyTypePromiseAny && !isAsync {
			return
		}

		messageId := "unsafeReturn"
		if !type_checking.IsStrictCompilerOptionEnabled(
			ctx.Program.Options(), ctx.Program.Options().NoImplicitThis) {
			// See the doc comment: this arm is correct and its id is unreachable through this
			// package's fixtures, because the harness pins the option above the rule.
			thisExpression := type_checking.GetThisExpression(returnNode)
			if thisExpression != nil {
				thisType := type_checking.GetConstrainedTypeAtLocation(ctx.TypeChecker, thisExpression)
				if thisType != nil && type_checking.IsTypeAnyType(thisType) {
					messageId = "unsafeReturnThis"
				}
			}
		}

		ctx.ReportNode(reportingNode, unsafeReturnMessageFor(messageId,
			unsafeReturnTypeText(anyType, constrainedReturnNodeType)))
		return
	}

	// The value is not itself `any`, so the remaining question is whether it carries one in a
	// generic position. Upstream takes the FIRST call signature here rather than looping, which is
	// a narrower question than the loops above: those ask "does any signature excuse this", and
	// this one asks "what does this function actually return".
	//
	// Note which accessor upstream uses here: `functionType.getCallSignatures()`, the checker's own
	// method, rather than the flattening helper it used above. On a union that answers nothing,
	// which is why a union-typed receiver reaches this branch and declines rather than comparing
	// against an arbitrary member. Reproduced deliberately; using the flattened list here would
	// report inputs upstream is silent on.
	directSignatures := type_checking.GetCallSignatures(ctx.TypeChecker, functionType)
	if len(directSignatures) == 0 {
		return
	}
	functionReturnType := checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, directSignatures[0])
	if functionReturnType == nil {
		return
	}
	receiver, sender, unsafe := type_checking.IsUnsafeAssignment(returnNodeType, functionReturnType,
		ctx.TypeChecker, returnNode)
	if !unsafe {
		return
	}
	ctx.ReportNode(reportingNode, rule.Message{
		Id: "unsafeReturnAssignment",
		Description: "Unsafe return of type `" + ctx.TypeChecker.TypeToString(sender) +
			"` from function with return type `" + ctx.TypeChecker.TypeToString(receiver) + "`.",
	})
}

// unsafeReturnFunctionType answers the type whose call signatures describe what this function
// returns.
//
// A function expression and an arrow do not have their return type adjusted by the receiver, so
// asking the node itself gives `() => Set<any>` for `const foo: () => Set<string> = () => new
// Set<any>()` and the rule goes silent on a real finding. Upstream asks for the contextual type in
// exactly those two cases and falls back to the node's own type when there is none.
func unsafeReturnFunctionType(ctx rule.Context, functionNode *ast.Node) *checker.Type {
	if functionNode.Kind == ast.KindFunctionExpression || functionNode.Kind == ast.KindArrowFunction {
		if contextual := type_checking.GetContextualType(ctx.TypeChecker, functionNode); contextual != nil {
			return contextual
		}
	}
	return ctx.TypeChecker.GetTypeAtLocation(functionNode)
}

// unsafeReturnTypeAnnotation answers the function's declared return type, or nil when it has none.
//
// Upstream reads `functionTSNode.type`, which is one field across every function-like node. Our AST
// splits them by kind, so this is the same field reached four ways.
func unsafeReturnTypeAnnotation(functionNode *ast.Node) *ast.Node {
	switch functionNode.Kind {
	case ast.KindFunctionDeclaration:
		return functionNode.AsFunctionDeclaration().Type
	case ast.KindFunctionExpression:
		return functionNode.AsFunctionExpression().Type
	case ast.KindArrowFunction:
		return functionNode.AsArrowFunction().Type
	case ast.KindMethodDeclaration:
		return functionNode.AsMethodDeclaration().Type
	case ast.KindGetAccessor:
		return functionNode.AsGetAccessorDeclaration().Type
	case ast.KindSetAccessor:
		return functionNode.AsSetAccessorDeclaration().Type
	case ast.KindConstructor:
		return functionNode.AsConstructorDeclaration().Type
	}
	return nil
}

// unsafeReturnIsAsync answers whether the function carries an `async` modifier.
//
// Upstream reads `functionNode.async`, a boolean estree puts on every function node. Here it is a
// modifier, and the shelf already knows how to read it.
func unsafeReturnIsAsync(functionNode *ast.Node) bool {
	return ast.IsAsyncFunction(functionNode)
}

// unsafeReturnTypeText renders the `{{type}}` slot.
//
// Upstream fills it with one of four literal strings and never with a type name, so the slot has
// exactly four fillings. The backticks are upstream's message text rather than markup added here,
// and `error` is deliberately the one without them: it is a word about the checker's state rather
// than the name of a type anybody wrote.
func unsafeReturnTypeText(
	anyType type_checking.DiscriminatedAnyType,
	constrainedReturnNodeType *checker.Type,
) string {
	if type_checking.IsIntrinsicErrorType(constrainedReturnNodeType) {
		return "error"
	}
	switch anyType {
	case type_checking.DiscriminatedAnyTypeAny:
		return "`any`"
	case type_checking.DiscriminatedAnyTypePromiseAny:
		return "`Promise<any>`"
	}
	return "`any[]`"
}

// unsafeReturnMessageFor renders the two `{{type}}`-carrying messages.
//
// The `unsafeReturnThis` text is two lines joined by a newline, which is upstream's own shape: the
// second line is advice about the compiler option rather than a separate suggestion, because there
// is no edit to propose.
func unsafeReturnMessageFor(messageId string, typeText string) rule.Message {
	if messageId == "unsafeReturnThis" {
		return rule.Message{
			Id: messageId,
			Description: "Unsafe return of a value of type " + typeText +
				". `this` is typed as `any`.\nYou can try to fix this by turning on the " +
				"`noImplicitThis` compiler option, or adding a `this` parameter to the function.",
		}
	}
	return rule.Message{
		Id:          "unsafeReturn",
		Description: "Unsafe return of a value of type " + typeText + ".",
	}
}
