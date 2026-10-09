package nexus

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// correctnessNoImplicitReturnText is the rule's message, whose wording lives in
// `policy/messages/correctness-no-implicit-return.json`.
var correctnessNoImplicitReturnText = policy.MessageOf("nexus/correctness-no-implicit-return", "implicitReturn")

/*
 * CorrectnessNoImplicitReturn reports exactly what TypeScript's noImplicitReturns reports, TS7030, in a
 * project whose tsconfig leaves the flag off (#yj96emr).
 *
 *     invalid: function label(kind: Kind) { if(kind === 'a') return 'A'; }      // falls off the end
 *     valid:   function label(kind: 'a' | 'b') { switch(kind) { case 'a': return 'A'; case 'b': return 'B'; } }
 *     valid:   function parse(text: string): unknown { try { return read(text); } catch { process.exit(1); } }
 *
 * # Why this asks the checker rather than walking the function
 *
 * The flag's question is whether the end of a function is reachable, and TypeScript answers it from
 * control flow the checker has already built: an exhaustive switch over a union ends nothing, and a call
 * to a function returning never ends the path it is on. The ported consistent-return answers it
 * syntactically, so it reports both of those, and Kirk ruled on 2026-08-25 that neither needs an explicit
 * return. Measured in api: consistent-return 57 and its typed twin 55, against TypeScript's 17, and about
 * 38 of the difference are exactly those two shapes. So this rule calls the checker's own
 * functionHasImplicitReturn and its own return-type tests, in the order checkAllCodePathsInNonVoidFunction
 * ReturnOrThrow and checkReturnStatement apply them, and its findings are TypeScript's by construction.
 *
 * # Where it stands down
 *
 * Where the tsconfig turns noImplicitReturns on, the type check already reports every one of these, so
 * the rule declines and says so in --coverage rather than reporting each a second time. It exists for a
 * project that cannot turn the flag on yet: Base, until #tz23yx2.
 *
 * The other diagnostics that function raises (a function returning never that can end, a declared
 * return type with no return at all, a declared type that excludes undefined) are type errors whatever
 * the flag says, so the rule leaves them to the type check, as TypeScript does.
 */
var CorrectnessNoImplicitReturn = rule.Rule{
	Name: "nexus/correctness-no-implicit-return",

	// The reachability of a function's end and the type it returns are the checker's answers
	NeedsTypeChecker: true,
	// Whether noImplicitReturns is already on, and whether strictNullChecks is
	ProgramReads: rule.ReadsCompilerOptions,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.TypeChecker == nil || ctx.Program == nil {
			return nil
		}
		compilerOptions := ctx.Program.Options()
		if compilerOptions.NoImplicitReturns.IsTrue() {
			ctx.SkipCovered(rule.CoverTypeCheck, "noImplicitReturns is on, so the type check reports these itself")
			return nil
		}
		strictNullChecks := compilerOptions.GetStrictOptionValue(compilerOptions.StrictNullChecks)

		// isUndefinedVoidOrAny is the checker's own test of a return type that needs no return value
		isUndefinedVoidOrAny := func(returnType *checker.Type, functionFlags ast.FunctionFlags) bool {
			unwrapped := checker.Checker_unwrapReturnType(ctx.TypeChecker, returnType, functionFlags)
			return unwrapped != nil && (checker.Checker_maybeTypeOfKind(ctx.TypeChecker, unwrapped, checker.TypeFlagsVoid) ||
				checker.Type_flags(unwrapped)&(checker.TypeFlagsAny|checker.TypeFlagsUndefined) != 0)
		}

		report := func(node *ast.Node) {
			ctx.ReportRange(scanner.GetErrorRangeForNode(ctx.SourceFile, node),
				rule.Message{Id: "implicitReturn", Description: correctnessNoImplicitReturnText.Render(nil)})
		}

		// The end of the function: checkAllCodePathsInNonVoidFunctionReturnOrThrow's noImplicitReturns branch
		checkEnd := func(function *ast.Node) {
			body := function.Body()
			if ast.IsMethodSignatureDeclaration(function) || body == nil || ast.NodeIsMissing(body) || !ast.IsBlock(body) {
				return
			}
			functionFlags := ast.GetFunctionFlags(function)
			var declared *checker.Type
			if function.Kind == ast.KindGetAccessor {
				declared = checker.Checker_getTypeOfAccessors(ctx.TypeChecker, function.Symbol())
			} else {
				declared = checker.Checker_getReturnTypeFromAnnotation(ctx.TypeChecker, function)
			}
			var unwrapped *checker.Type
			if declared != nil {
				unwrapped = checker.Checker_unwrapReturnType(ctx.TypeChecker, declared, functionFlags)
			}
			if unwrapped != nil && (checker.Checker_maybeTypeOfKind(ctx.TypeChecker, unwrapped, checker.TypeFlagsVoid) ||
				checker.Type_flags(unwrapped)&(checker.TypeFlagsAny|checker.TypeFlagsUndefined) != 0) {
				return
			}
			if !checker.Checker_functionHasImplicitReturn(ctx.TypeChecker, function) {
				return
			}
			hasExplicitReturn := function.Flags&ast.NodeFlagsHasExplicitReturn != 0
			switch {
			// The three branches TypeScript reports whatever the flag says: type errors, not this rule's
			case unwrapped != nil && checker.Type_flags(unwrapped)&checker.TypeFlagsNever != 0:
				return
			case unwrapped != nil && !hasExplicitReturn:
				return
			case unwrapped != nil && strictNullChecks &&
				!checker.Checker_isTypeAssignableTo(ctx.TypeChecker, checker.Checker_undefinedType(ctx.TypeChecker), unwrapped):
				return
			case unwrapped == nil:
				// No annotation: with no return value anywhere the inferred type is void, and nothing is owed
				if !hasExplicitReturn {
					return
				}
				signature := checker.Checker_getSignatureFromDeclaration(ctx.TypeChecker, function)
				if isUndefinedVoidOrAny(checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, signature), functionFlags) {
					return
				}
			}
			errorNode := function.Type()
			if errorNode == nil {
				if data := function.FunctionLikeData(); data != nil && data.FullSignature != nil {
					errorNode = data.FullSignature
				}
			}
			if errorNode == nil {
				errorNode = function
			}
			report(errorNode)
		}

		return rule.Listeners{
			ast.KindFunctionDeclaration: checkEnd,
			ast.KindMethodDeclaration:   checkEnd,
			ast.KindGetAccessor:         checkEnd,
			ast.KindFunctionExpression:  checkEnd,
			ast.KindArrowFunction:       checkEnd,

			// A bare return: checkReturnStatement's noImplicitReturns branch, reached only without
			// strictNullChecks, because with it a bare return is checked as returning undefined
			ast.KindReturnStatement: func(node *ast.Node) {
				if strictNullChecks || node.Expression() != nil {
					return
				}
				container := ast.FindAncestor(node.Parent, ast.IsFunctionLikeOrClassStaticBlockDeclaration)
				if container == nil || ast.IsClassStaticBlockDeclaration(container) || ast.IsConstructorDeclaration(container) {
					return
				}
				signature := checker.Checker_getSignatureFromDeclaration(ctx.TypeChecker, container)
				returnType := checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, signature)
				if checker.Type_flags(returnType)&checker.TypeFlagsNever != 0 {
					return
				}
				if !isUndefinedVoidOrAny(returnType, ast.GetFunctionFlags(container)) {
					report(node)
				}
			},
		}
	},
}
