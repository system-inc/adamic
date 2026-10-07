package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strings"
)

// wave06JSXBindings projects raw binder/import/initializer syntax, never a lint decision.
func (p *Program) wave06JSXBindings(out *fields, c *checker.Checker, node *ast.Node) (string, error) {
	symbol := c.GetSymbolAtLocation(node)
	out.yes(symbol != nil)
	var declarations []*ast.Node
	if symbol != nil {
		declarations = symbol.Declarations
	}
	out.number(uint64(len(declarations)))
	kind := func(n *ast.Node) string {
		if n == nil {
			return ""
		}
		return strings.TrimPrefix(n.Kind.String(), "Kind")
	}
	text := func(n *ast.Node) string {
		if wave06TextName(n) {
			return n.Text()
		}
		return ""
	}
	first := func(n *ast.Node) (string, string) {
		if n == nil || n.Kind != ast.KindCallExpression {
			return "", ""
		}
		arguments := n.AsCallExpression().Arguments
		if arguments == nil || len(arguments.Nodes) == 0 {
			return "", ""
		}
		value := arguments.Nodes[0]
		return kind(value), text(value)
	}
	for _, d := range declarations {
		if d == nil {
			return "", fmt.Errorf("nil symbol declaration")
		}
		out.text(kind(d))
		importedKind, imported, module := "", "", ""
		if d.Kind == ast.KindImportSpecifier {
			name := d.AsImportSpecifier().PropertyName
			if name == nil {
				name = d.Name()
			}
			importedKind = kind(name)
			imported = text(name)
			for parent := d.Parent; parent != nil; parent = parent.Parent {
				if parent.Kind == ast.KindImportDeclaration {
					module = text(parent.AsImportDeclaration().ModuleSpecifier)
					break
				}
			}
		}
		out.text(importedKind)
		out.text(imported)
		out.text(module)
		objectBinding := d.Kind == ast.KindBindingElement && d.Parent != nil && d.Parent.Kind == ast.KindObjectBindingPattern
		variableParent := objectBinding && d.Parent.Parent != nil && d.Parent.Parent.Kind == ast.KindVariableDeclaration
		out.yes(objectBinding)
		out.yes(variableParent)
		var initializer *ast.Node
		if d.Kind == ast.KindVariableDeclaration {
			initializer = d.Initializer()
		} else if variableParent {
			initializer = d.Parent.Parent.Initializer()
		}
		out.text(kind(initializer))
		name, receiverName := "", ""
		var receiver *ast.Node
		if initializer != nil {
			switch initializer.Kind {
			case ast.KindIdentifier:
				name = text(initializer)
			case ast.KindPropertyAccessExpression:
				access := initializer.AsPropertyAccessExpression()
				name = text(access.Name())
				receiver = access.Expression
				receiverName = text(receiver)
				if receiver != nil && receiver.Kind == ast.KindCallExpression {
					receiverName = text(receiver.AsCallExpression().Expression)
				}
			case ast.KindCallExpression:
				name = text(initializer.AsCallExpression().Expression)
			}
		}
		out.text(name)
		out.text(kind(receiver))
		out.text(receiverName)
		firstKind, firstText := first(initializer)
		out.text(firstKind)
		out.text(firstText)
		receiverFirstKind, receiverFirstText := first(receiver)
		out.text(receiverFirstKind)
		out.text(receiverFirstText)
	}
	return out.String(), nil
}
