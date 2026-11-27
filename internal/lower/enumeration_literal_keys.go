package lower

import "github.com/microsoft/TypeScript/tsc/shim/ast"

const enumerationSpreadLimit = 32

// enumerationLiteralKeys returns the literals supplying the operand's own
// keys. Nil means the closed set of keys could not be proved.
func (l *lowering) enumerationLiteralKeys(operand *ast.Node, modules []*ast.SourceFile) map[*ast.Node]int {
	depths := make(map[*ast.Node]int)
	sources := make(map[*ast.Symbol]*ast.Node)
	var collect func(*ast.Node, int) bool
	collect = func(literal *ast.Node, depth int) bool {
		literal = unwrapEnumerationLiteral(literal)
		if literal == nil || literal.Kind != ast.KindObjectLiteralExpression || depth > enumerationSpreadLimit {
			return false
		}
		// A check at greater depth proves this path is within the bound too.
		if depths[literal] > depth {
			return true
		}
		for _, property := range literal.AsObjectLiteralExpression().Properties.Nodes {
			switch property.Kind {
			case ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment:
				name := property.Name()
				if name == nil || (name.Kind != ast.KindIdentifier && name.Kind != ast.KindStringLiteral && name.Kind != ast.KindNumericLiteral) {
					return false
				}
				if name.Text() == "__proto__" {
					return false
				}
			case ast.KindSpreadAssignment:
				source := unwrapEnumerationLiteral(property.AsSpreadAssignment().Expression)
				if source == nil || !ast.IsIdentifier(source) {
					return false
				}
				symbol := l.valueSymbol(source)
				if symbol == nil || len(symbol.Declarations) != 1 {
					return false
				}
				declaration := symbol.Declarations[0]
				if declaration.Kind != ast.KindVariableDeclaration || declaration.Parent.Flags&ast.NodeFlagsConst == 0 {
					return false
				}
				sources[symbol] = declaration.Name()
				if !collect(declaration.AsVariableDeclaration().Initializer, depth+1) {
					return false
				}
			default:
				return false
			}
		}
		depths[literal] = depth + 1
		return true
	}
	if !collect(operand, 0) {
		return nil
	}
	if len(sources) == 0 {
		return depths
	}
	// A const binding can still escape to a writer. Require every use of
	// each spread source to be the value copied by an object spread.
	var escapes ast.Visitor
	escapes = func(node *ast.Node) bool {
		if ast.IsIdentifier(node) {
			if declaration, ok := sources[l.valueSymbol(node)]; ok && node != declaration {
				parent := node.Parent
				for parent.Kind == ast.KindParenthesizedExpression ||
					parent.Kind == ast.KindAsExpression && ast.IsConstTypeReference(parent.AsAsExpression().Type) {
					parent = parent.Parent
				}
				if parent.Kind != ast.KindSpreadAssignment {
					return true
				}
			}
		}
		return node.ForEachChild(escapes)
	}
	for _, module := range modules {
		if module.AsNode().ForEachChild(escapes) {
			return nil
		}
	}
	return depths
}

func unwrapEnumerationLiteral(node *ast.Node) *ast.Node {
	for node != nil {
		node = ast.SkipParentheses(node)
		if node.Kind != ast.KindAsExpression || !ast.IsConstTypeReference(node.AsAsExpression().Type) {
			break
		}
		node = node.AsAsExpression().Expression
	}
	return node
}
