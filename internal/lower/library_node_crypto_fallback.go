package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// The absent-crypto fixture's declaration is type-only, not a native SHA binding.
// A reachable reference is still refused; only the proven absent arm is omitted.
func nodeCryptoAmbientHashDeclaration(node *ast.Node) bool {
	if node.Kind != ast.KindFunctionDeclaration || node.Body() != nil || node.Name() == nil || node.Name().Text() != "createSHA256Hash" || !ast.HasSyntacticModifier(node, ast.ModifierFlagsAmbient) {
		return false
	}
	if node.Type() == nil || node.Type().Kind != ast.KindStringKeyword || len(node.Parameters()) != 1 {
		return false
	}
	parameter := node.Parameters()[0].AsParameterDeclaration()
	return parameter.Type != nil && parameter.Type.Kind == ast.KindStringKeyword && parameter.QuestionToken == nil && parameter.DotDotDotToken == nil && parameter.Initializer == nil
}

func (l *lowering) nodeCryptoFallback(node *ast.Node) (ir.Expression, bool, error) {
	branch := node.AsConditionalExpression()
	absent := ast.SkipParentheses(branch.Condition)
	ambient := ast.SkipParentheses(branch.WhenTrue)
	if !ast.IsIdentifier(absent) || !ast.IsIdentifier(ambient) {
		return nil, false, nil
	}
	symbol := l.symbol(ambient)
	if symbol == nil || len(symbol.Declarations) != 1 || !nodeCryptoAmbientHashDeclaration(symbol.Declarations[0]) {
		return nil, false, nil
	}
	binding := l.symbol(absent)
	if binding == nil || len(binding.Declarations) != 1 || binding.Declarations[0].Kind != ast.KindVariableDeclaration {
		return nil, false, nil
	}
	declaration := binding.Declarations[0]
	if declaration.Parent == nil || declaration.Parent.Kind != ast.KindVariableDeclarationList || declaration.Parent.Flags&ast.NodeFlagsConst == 0 {
		return nil, false, nil
	}
	initial := declaration.AsVariableDeclaration().Initializer
	if initial == nil {
		return nil, false, nil
	}
	initial = ast.SkipParentheses(initial)
	// Literal false is the portable component witness; the real sys.ts fixture
	// uses the exact-undefined representation supplied by the compiler batch.
	isAbsent := initial.Kind == ast.KindFalseKeyword
	if ast.IsIdentifier(initial) && initial.Text() == "undefined" && l.checker.GetTypeAtLocation(initial).Flags()&checker.TypeFlagsUndefined != 0 {
		_, local := l.locals[l.symbol(initial)]
		isAbsent = !local
	}
	if !isAbsent {
		return nil, false, nil
	}
	condition, err := l.condition(branch.Condition)
	if err != nil {
		return nil, true, err
	}
	fallback, err := l.expression(branch.WhenFalse)
	if err != nil {
		return nil, true, err
	}
	// Keep the condition's read and readiness checks in source order. No read of
	// the erased ambient function is emitted, even though its type was checked.
	return ir.Conditional{Condition: condition, WhenTrue: fallback, WhenNot: fallback}, true, nil
}
