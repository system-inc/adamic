package react

import "github.com/microsoft/TypeScript/tsc/shim/ast"

func AdamicBase(n *ast.Node) bool                    { return isComponentBase(n) }
func AdamicIdentifier(n *ast.Node, name string) bool { return isIdentifierNamed(n, name) }
func AdamicBaseName(name string) bool                { return isComponentBaseName(name) }
func AdamicCreateName(name string) bool              { return isCreateClassName(name) }
