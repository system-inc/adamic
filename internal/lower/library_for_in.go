package lower

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// forIn snapshots runtime keys and skips keys removed before their turn.
// Prototype mutation remains refused, so supported objects have no changing prototype chain.
func (l *lowering) forIn(node *ast.Node) ([]ir.Statement, error) {
	statement := node.AsForInOrOfStatement()
	object, err := l.expression(statement.Expression)
	if err != nil {
		return nil, err
	}
	if object.Type() != ir.Object && object.Type() != ir.Array && object.Type() != ir.Union {
		return nil, l.notYet(statement.Expression, "for...in over a value without an enumerable runtime representation")
	}
	// Hold the original receiver even when the body reassigns its source binding.
	receiver := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "for_in_receiver", Type: object.Type(), Function: l.functionIndex})
	read := ir.Read{Local: receiver, Of: object.Type()}
	loop := ir.ForOf{Iterable: ir.ObjectKeys{Object: read, Enumeration: true}, Element: ir.String}
	initializer := statement.Initializer
	var assignment ir.Statement
	if initializer.Kind == ast.KindVariableDeclarationList {
		if initializer.Flags&ast.NodeFlagsBlockScoped == 0 {
			return nil, &Refused{Where: l.program.Where(initializer), What: "var", Fix: "use const or let"}
		}
		declarations := initializer.AsVariableDeclarationList().Declarations.Nodes
		if len(declarations) != 1 || !ast.IsIdentifier(declarations[0].Name()) {
			return nil, l.notYet(initializer, "a for...in binding that is not one plain name")
		}
		loop.Local, err = l.declareLocal(declarations[0].Name())
	} else {
		target := ast.SkipParentheses(initializer)
		local, found := l.local(target)
		if !ast.IsIdentifier(target) || !found || l.result.Locals[local].Type != ir.String {
			return nil, l.notYet(initializer, "a for...in assignment that is not a string variable")
		}
		loop.Local = len(l.result.Locals)
		l.result.Locals = append(l.result.Locals, ir.Local{Name: "key", Type: ir.String, Function: l.functionIndex})
		assignment = ir.Assign{Local: local, Value: ir.Read{Local: loop.Local, Of: ir.String}, Checked: l.checked(local)}
	}
	if err != nil {
		return nil, err
	}
	body, err := l.statement(statement.Statement)
	if err != nil {
		return nil, err
	}
	if assignment != nil {
		loop.Body = append(loop.Body, assignment)
	}
	loop.Body = append(loop.Body, body...)
	loop.Body = []ir.Statement{ir.If{Condition: ir.ForInOwn{Object: read, Key: ir.Read{Local: loop.Local, Of: ir.String}}, Then: loop.Body}}
	return []ir.Statement{ir.Declare{Local: receiver, Value: object}, loop}, nil
}

func (l *lowering) plainEnumerableObject(node *ast.Node, visiting map[*ast.Symbol]bool) bool {
	node = ast.SkipParentheses(node)
	// Enum objects have a complete fixed shape, including numeric reverse keys.
	if l.enumObject(node) != nil {
		return true
	}
	if node.Kind == ast.KindAsExpression {
		return l.plainEnumerableObject(node.AsAsExpression().Expression, visiting)
	}
	if node.Kind == ast.KindConditionalExpression {
		branch := node.AsConditionalExpression()
		return l.plainEnumerableObject(branch.WhenTrue, visiting) && l.plainEnumerableObject(branch.WhenFalse, visiting)
	}
	if node.Kind == ast.KindObjectLiteralExpression {
		for _, property := range node.AsObjectLiteralExpression().Properties.Nodes {
			switch property.Kind {
			case ast.KindSpreadAssignment:
				source := property.AsSpreadAssignment().Expression
				for _, field := range l.checker.GetPropertiesOfType(l.checker.GetTypeAtLocation(source)) {
					if field.Flags&ast.SymbolFlagsOptional != 0 {
						return false
					}
				}
				if !l.plainEnumerableObject(source, visiting) {
					return false
				}
			case ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment:
				name := property.Name()
				if name == nil || strings.ContainsRune(name.Text(), 0) || (property.Kind == ast.KindPropertyAssignment && name.Text() == "__proto__") {
					return false
				}
			default:
				return false
			}
		}
		return true
	}
	if !ast.IsIdentifier(node) {
		return false
	}
	// Optional structural fields can be absent in the runtime shape; writing them would add keys.
	for _, field := range l.checker.GetPropertiesOfType(l.checker.GetTypeAtLocation(node)) {
		if field.Flags&ast.SymbolFlagsOptional != 0 {
			return false
		}
	}
	symbol := l.symbol(node)
	if symbol == nil || visiting[symbol] || len(symbol.Declarations) != 1 {
		return false
	}
	declaration := symbol.Declarations[0]
	if declaration.Kind != ast.KindVariableDeclaration || declaration.AsVariableDeclaration().Initializer == nil {
		return false
	}
	visiting[symbol] = true
	defer delete(visiting, symbol)
	if !l.plainEnumerableObject(declaration.AsVariableDeclaration().Initializer, visiting) {
		return false
	}
	// Every assignment must have another proved plain-object origin. An alias may be assigned in
	// another module, but imported bindings are immutable and writes into object fields keep shape.
	safe := true
	var visit ast.Visitor
	visit = func(inner *ast.Node) bool {
		if !safe {
			return true
		}
		if inner.Kind == ast.KindBinaryExpression {
			binary := inner.AsBinaryExpression()
			if ast.IsIdentifier(ast.SkipParentheses(binary.Left)) && l.symbol(ast.SkipParentheses(binary.Left)) == symbol && binary.OperatorToken.Kind == ast.KindEqualsToken {
				safe = l.plainEnumerableObject(binary.Right, visiting)
			}
			if binary.OperatorToken.Kind == ast.KindEqualsToken && !ast.IsIdentifier(ast.SkipParentheses(binary.Left)) {
				var target ast.Visitor
				target = func(part *ast.Node) bool {
					if ast.IsIdentifier(part) && l.symbol(part) == symbol {
						safe = false
					}
					return part.ForEachChild(target)
				}
				if binary.Left.Kind == ast.KindArrayLiteralExpression || binary.Left.Kind == ast.KindObjectLiteralExpression {
					binary.Left.ForEachChild(target)
				}
			}
		}
		return inner.ForEachChild(visit)
	}
	ast.GetSourceFileOfNode(declaration).AsNode().ForEachChild(visit)
	return safe
}
