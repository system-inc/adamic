package wave08core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strconv"
)

// SymbolOriginsFields exposes declaration source-file flags without following aliases.
// Declaration order is preserved; lint policy belongs to the native rule.
func SymbolOriginsFields(c *checker.Checker, node *ast.Node) []string {
	out := []string{"1", "wave08-symbol-origins"}
	symbol := c.GetSymbolAtLocation(node)
	if symbol == nil {
		return append(out, "0")
	}
	out = append(out, strconv.Itoa(len(symbol.Declarations)))
	for _, declaration := range symbol.Declarations {
		file := ast.GetSourceFileOfNode(declaration)
		flag := "0"
		if file != nil && file.IsDeclarationFile {
			flag = "1"
		}
		out = append(out, flag)
	}
	return out
}
