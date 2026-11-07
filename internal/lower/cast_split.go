package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	regex "github.com/system-inc/adamic/internal/regexp"
)

// captureFreeSplit proves text.split(/literal/) as string[] (#b5w3ycg, Adamic's and Compiler's ruling). The library
// types a RegExp split (string | undefined)[] (load/regexp_library.go), since ECMAScript's split pushes undefined for a
// capture that took no part in a match; captures are the only elements split ever pushes besides the pieces between
// matches, so a separator with no capturing group, named or unnamed, gives only strings. Adamic's own parser counts
// the groups (Indices holds the whole match at 0). Only a literal written at the call is read: a variable or
// new RegExp(...) is a value the proof would have to follow, and isn't followed. A readonly string[] target is
// admitted too, as a view that writes nothing needs the same proof as one that may write strings. Both arrays
// hold their elements as String, a missing one a null reference (ir.Maybe), so the value is the array itself.
//
// The proof leans on String.prototype.split dispatching to RegExp.prototype[Symbol.split] unchanged, which holds only
// while a program can't write a built-in prototype. No one rule names those writes; each way to one is refused:
// reading RegExp.prototype, String.prototype or a value's constructor is "inherited library member ... read as an own
// field" (property access lowering's guard in object.go, prototype.go's prototypeRead); Object.getPrototypeOf,
// setPrototypeOf, create and defineProperty are refused in library_object.go's objectCallArguments; and
// String.prototype.split read as a value is methodValueRefusal (refusals.go, ordinal 2900). A literal written at the
// call is a fresh RegExp, so no write to an instance reaches it.
func (lowering *lowering) captureFreeSplit(expression *ast.Node, target *checker.Type) bool {
	call := ast.SkipParentheses(expression)
	if call.Kind != ast.KindCallExpression ||
		!lowering.isArrayType(target) ||
		lowering.elementTypeOfArray(target) != lowering.checker.GetStringType() {
		return false
	}
	callee := ast.SkipParentheses(call.AsCallExpression().Expression)
	arguments := call.AsCallExpression().Arguments.Nodes
	if callee.Kind != ast.KindPropertyAccessExpression ||
		callee.AsPropertyAccessExpression().QuestionDotToken != nil ||
		callee.Name().Text() != "split" ||
		len(arguments) == 0 ||
		len(arguments) > 2 {
		return false
	}
	// The receiver test regexBuiltin dispatches a string's RegExp methods on.
	if of, _ := lowering.representation(
		lowering.checker.GetNonNullableType(
			lowering.checker.GetTypeAtLocation(callee.AsPropertyAccessExpression().Expression),
		),
	); of != ir.String {
		return false
	}
	separator := ast.SkipParentheses(arguments[0])
	if separator.Kind != ast.KindRegularExpressionLiteral {
		return false
	}
	pattern, flags, written := regexLiteralParts(separator)
	if !written {
		return false
	}
	facts, captureFactsError := regex.CaptureFacts(pattern, flags)
	return captureFactsError == nil && len(facts.Indices) == 1
}
