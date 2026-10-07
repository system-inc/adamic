package wave08next

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strconv"
	"strings"
)

// SymbolAncestryFields returns raw resolved declarations and syntax ancestors.
// Symbol aliases are followed, but no process/write/blocking verdict is returned.
func SymbolAncestryFields(c *checker.Checker, node *ast.Node) []string {
	flag := func(v bool) string {
		if v {
			return "1"
		}
		return "0"
	}
	symbol := c.GetSymbolAtLocation(node)
	if symbol != nil && symbol.Flags&ast.SymbolFlagsAlias != 0 {
		symbol = c.GetAliasedSymbol(symbol)
	}
	out := []string{"1", "wave08-symbol-ancestry"}
	if symbol == nil {
		return append(out, "0", "0")
	}
	out = append(out, strconv.FormatUint(uint64(symbol.Flags), 10), strconv.Itoa(len(symbol.Declarations)))
	for _, declaration := range symbol.Declarations {
		file := ast.GetSourceFileOfNode(declaration)
		out = append(out, file.FileName(), flag(file.IsDeclarationFile), flag(ast.IsExternalModule(file)), strings.TrimPrefix(declaration.Kind.String(), "Kind"), strconv.Itoa(declaration.Pos()), strconv.Itoa(declaration.End()))
		var ancestors []*ast.Node
		for parent := declaration.Parent; parent != nil; parent = parent.Parent {
			ancestors = append(ancestors, parent)
		}
		out = append(out, strconv.Itoa(len(ancestors)))
		for _, parent := range ancestors {
			name := ""
			if named := parent.Name(); named != nil && (named.Kind == ast.KindIdentifier || named.Kind == ast.KindStringLiteral || named.Kind == ast.KindNumericLiteral || named.Kind == ast.KindPrivateIdentifier) {
				name = named.Text()
			}
			out = append(out, strings.TrimPrefix(parent.Kind.String(), "Kind"), name, strconv.FormatUint(uint64(parent.Flags), 10), flag(ast.IsGlobalScopeAugmentation(parent)))
		}
	}
	return out
}
