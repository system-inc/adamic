package lower

import "github.com/microsoft/TypeScript/tsc/shim/ast"

// Singleton slots are live for qualified uses, not fields of a runtime object.
// Keep this a capability boundary until aliases can share those exact slots.
func (l *lowering) namespaceValueNotYet(node *ast.Node) error {
	name := sourceExpression(node)
	declaration := l.namespaceDeclaration(node)
	return l.notYet(node, "a namespace object used as a value ("+name+", declared at "+l.program.Where(declaration.Name())+"); no runtime container is emitted, so identity, receiver behavior, live export aliases and staged properties are not represented; use "+name+".member access, or pass fixed namespace functions and explicit state")
}
