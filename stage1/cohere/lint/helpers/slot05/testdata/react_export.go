package react

import "github.com/microsoft/TypeScript/tsc/shim/ast"

func AdamicIdentifierNamed(node *ast.Node, name string) bool {
	return isIdentifierNamed(node, name)
}

func AdamicComponentBase(node *ast.Node) bool { return isComponentBase(node) }

func AdamicComponentBaseName(name string) bool { return isComponentBaseName(name) }
