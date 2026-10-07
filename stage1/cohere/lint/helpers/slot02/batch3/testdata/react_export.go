package react

import "github.com/microsoft/TypeScript/tsc/shim/ast"

var AdamicSlot02IdentifierCalls int
var AdamicSlot02IdentifierNode *ast.Node
var AdamicSlot02IdentifierWanted string

func AdamicSlot02IdentifierNamed(node *ast.Node, name string) bool {
	AdamicSlot02IdentifierCalls++
	AdamicSlot02IdentifierNode = node
	AdamicSlot02IdentifierWanted = name
	return isIdentifierNamed(node, name)
}
func AdamicSlot02CreateClassName(name string) bool      { return isCreateClassName(name) }
func AdamicSlot02IsReactIdentifier(node *ast.Node) bool { return isIdentifierNamed(node, "React") }
func AdamicSlot02ResetIdentifierProbe() {
	AdamicSlot02IdentifierCalls = 0
	AdamicSlot02IdentifierNode = nil
	AdamicSlot02IdentifierWanted = ""
}
