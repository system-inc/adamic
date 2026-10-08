package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// Record the meaning of NULL in physical reference fields before erasing the type.
func (l *lowering) jsonLiteralStorage(node *ast.Node, value ir.Expression) ir.Expression {
	literal, ok := value.(ir.ObjectLiteral)
	if !ok || ast.SkipParentheses(node).Kind != ast.KindObjectLiteralExpression {
		return value
	}
	node = ast.SkipParentheses(node)
	literal.JSONNull = make([]bool, len(literal.Fields))
	own := l.checker.GetTypeAtLocation(node)
	contextual := l.checker.GetContextualType(node, checker.ContextFlagsNone)
	for index, field := range literal.Fields {
		if !field.Value.Type().IsReference() || field.Value.Type() == ir.Union {
			continue
		}
		property := l.checker.GetPropertyOfType(own, field.Name)
		if property == nil {
			continue
		}
		t := l.checker.GetTypeOfSymbol(property)
		if contextual != nil {
			if c := l.checker.GetPropertyOfType(contextual, field.Name); c != nil && l.declaredField(node, field.Name) == field.Value.Type() {
				t = l.checker.GetTypeOfSymbol(c)
			}
		}
		literal.JSONNull[index] = l.includesNull(t)
	}
	return literal
}

// Actual allocation layouts supply fields hidden by structural views. Until callbacks,
// accessors and host internal slots supply the same metadata, refuse their possible
// origins at compile time rather than infer behavior from the declared object type.
func (l *lowering) jsonUnionOrigins(node *ast.Node) error {
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return err
	}
	var gap error
	var visit ast.Visitor
	visit = func(n *ast.Node) bool {
		if gap != nil {
			return true
		}
		switch n.Kind {
		case ast.KindClassDeclaration, ast.KindClassExpression:
			gap = l.notYet(node, "JSON.stringify container unions in a program with class internal slots")
			return true
		case ast.KindNewExpression:
			callee := n.AsNewExpression().Expression
			if !l.isLibraryGlobal(callee, "Array") && !l.isLibraryGlobal(callee, "Map") && !l.isLibraryGlobal(callee, "Set") {
				gap = l.notYet(node, "JSON.stringify container unions with constructor internal slots")
				return true
			}
		case ast.KindCallExpression:
			callee := ast.SkipParentheses(n.AsCallExpression().Expression)
			if symbol := l.checker.GetSymbolAtLocation(callee); symbol != nil && len(symbol.Declarations) > 0 {
				declared := ast.GetSourceFileOfNode(symbol.Declarations[0])
				if load.IsLibrary(declared) || load.IsPrelude(declared) {
					held, _ := l.representation(l.checker.GetTypeAtLocation(n))
					if held == ir.Object || held == ir.Array {
						allowed := false
						if held == ir.Array && callee.Kind == ast.KindPropertyAccessExpression {
							switch callee.Name().Text() {
							case "map", "filter", "slice", "concat", "reverse", "sort", "fill", "splice", "split":
								allowed = true
							case "from":
								allowed = l.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "Array")
							}
						}
						if held == ir.Array && ast.IsIdentifier(callee) && callee.Text() == "programArguments" {
							allowed = true
						}
						if !allowed {
							gap = l.notYet(node, "JSON.stringify container unions with opaque library allocation metadata")
							return true
						}
					}
				}
			}
			if callee.Kind == ast.KindPropertyAccessExpression && l.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "Object") {
				switch callee.Name().Text() {
				case "defineProperty", "defineProperties", "create", "assign", "fromEntries":
					gap = l.notYet(node, "JSON.stringify container unions with dynamically defined property metadata")
					return true
				}
			}
		case ast.KindObjectLiteralExpression:
			for _, p := range n.AsObjectLiteralExpression().Properties.Nodes {
				if p.Kind != ast.KindPropertyAssignment && p.Kind != ast.KindShorthandPropertyAssignment {
					gap = l.notYet(node, "JSON.stringify container unions with accessor, method or spread origins")
					return true
				}
				if p.Name() != nil && (p.Name().Text() == "toJSON" || p.Name().Kind == ast.KindComputedPropertyName) {
					gap = l.notYet(node, "JSON.stringify container unions with toJSON or computed-key origins")
					return true
				}
			}
		}
		return n.ForEachChild(visit)
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
		if gap != nil {
			return gap
		}
	}
	return nil
}
