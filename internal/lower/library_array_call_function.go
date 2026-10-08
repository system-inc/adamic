package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/adamic/internal/ir"
)

// Port GenericArrayShift's zero-length SetProperty from array-shift.tq. A fresh
// zero-argument function has immutable length zero, so this always throws. Only
// source without erased types is admitted, preserving V8's function error text.
func (l *lowering) libraryArrayFreshFunctionSource(node *ast.Node) (string, bool) {
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindFunctionExpression && node.Kind != ast.KindArrowFunction {
		return "", false
	}
	if node.Kind == ast.KindFunctionExpression && node.AsFunctionExpression().AsteriskToken != nil {
		return "", false
	}
	if len(node.Parameters()) != 0 || node.Type() != nil || len(node.TypeParameters()) != 0 || node.Body() == nil || node.Body().Kind != ast.KindBlock || len(node.Body().AsBlock().Statements.Nodes) != 0 {
		return "", false
	}
	source := ast.GetSourceFileOfNode(node)
	if libraryArrayFunctionSourceNeedsTransform(source.AsNode()) {
		return "", false
	}
	text := source.Text()[scanner.GetTokenPosOfNode(node, source, false):node.End()]
	// V8's error formatter abbreviates long source. Keep that formatting family
	// refused until the formatter itself is ported, rather than inventing text.
	if len(text) > 50 {
		return "", false
	}
	return text, true
}

func (l *lowering) libraryArrayFunctionShift(node *ast.Node, b *libraryArrayBuilder, source string) ir.Expression {
	b.body = append(b.body, b.typeError(node, "Cannot assign to read only property 'length' of function '"+source+"'")...)
	return b.finish("array_function_shift", ir.Undefined{})
}

// Transforming runtime TypeScript syntax can rewrite even an untyped function.
// Its exact error text is not proven by the source span in that module.
func libraryArrayFunctionSourceNeedsTransform(source *ast.Node) bool {
	needed := false
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		switch node.Kind {
		case ast.KindEnumDeclaration, ast.KindModuleDeclaration, ast.KindImportEqualsDeclaration:
			needed = true
		}
		if node.Kind == ast.KindParameter && ast.HasSyntacticModifier(node, ast.ModifierFlagsParameterPropertyModifier) {
			needed = true
		}
		return needed || node.ForEachChild(visit)
	}
	source.ForEachChild(visit)
	return needed
}
