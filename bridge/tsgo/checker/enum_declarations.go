package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strings"
)

// enumDeclarations supplies both symbol declaration lists and each first member's
// initializer syntax and type flags. Reference selection belongs to Adamic.
func (p *Program) enumDeclarations(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "enum-declarations" || node.Kind != ast.KindEnumDeclaration {
		return "", fmt.Errorf("enum-declarations requires an enum declaration")
	}
	symbol := c.GetSymbolAtLocation(node.Name())
	for _, subject := range []*ast.Symbol{symbol, node.LocalSymbol()} {
		var declarations []*ast.Node
		if subject != nil {
			declarations = subject.Declarations
		}
		out.number(uint64(len(declarations)))
		for _, declaration := range declarations {
			out.text(strings.TrimPrefix(declaration.Kind.String(), "Kind"))
			out.text(ast.GetSourceFileOfNode(declaration).FileName())
			out.number(uint64(declaration.Pos()))
			out.number(uint64(declaration.End()))
			var members []*ast.Node
			if declaration.Kind == ast.KindEnumDeclaration && declaration.AsEnumDeclaration().Members != nil {
				members = declaration.AsEnumDeclaration().Members.Nodes
			}
			out.number(uint64(len(members)))
			initializerKind := ""
			var flags uint64
			if len(members) > 0 {
				initializer := members[0].AsEnumMember().Initializer
				if initializer != nil {
					initializer = ast.SkipParentheses(initializer)
				}
				if initializer != nil {
					initializerKind = strings.TrimPrefix(initializer.Kind.String(), "Kind")
					if t := c.GetTypeAtLocation(initializer); t != nil {
						flags = enumInitializerFlags(t)
					}
				}
			}
			out.text(initializerKind)
			out.number(flags)
		}
	}
	return out.String(), nil
}

// Aggregate flags retain union constituents without selecting a rule classification.
func enumInitializerFlags(t *checker.Type) uint64 {
	flags := uint64(t.Flags())
	if t.Flags()&checker.TypeFlagsUnion != 0 {
		for _, part := range t.Types() {
			flags |= enumInitializerFlags(part)
		}
	}
	return flags
}
