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
func (l *lowering) captureFreeSplit(expression *ast.Node, target *checker.Type) bool {
	call := ast.SkipParentheses(expression)
	if call.Kind != ast.KindCallExpression || !l.checker.IsArrayType(target) || l.checker.GetElementTypeOfArrayType(target) != l.checker.GetStringType() {
		return false
	}
	callee := ast.SkipParentheses(call.AsCallExpression().Expression)
	arguments := call.AsCallExpression().Arguments.Nodes
	if callee.Kind != ast.KindPropertyAccessExpression || callee.AsPropertyAccessExpression().QuestionDotToken != nil || callee.Name().Text() != "split" || len(arguments) == 0 || len(arguments) > 2 {
		return false
	}
	// The receiver test regexBuiltin dispatches a string's RegExp methods on.
	if of, _ := l.representation(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(callee.AsPropertyAccessExpression().Expression))); of != ir.String {
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
	facts, err := regex.CaptureFacts(pattern, flags)
	return err == nil && len(facts.Indices) == 1
}
