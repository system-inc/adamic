package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strings"
)

// bindingState retains value-declaration metadata and alias/shorthand identities.
func (p *Program) bindingState(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "binding-state" || node.Kind != ast.KindIdentifier {
		return "", fmt.Errorf("binding-state requires an Identifier without suffix")
	}
	symbol := c.GetSymbolAtLocation(node)
	if node.Parent != nil && node.Parent.Kind == ast.KindShorthandPropertyAssignment {
		if value := c.GetShorthandAssignmentValueSymbol(node.Parent); value != nil {
			symbol = value
		}
	}
	out.number(p.symbolID(symbol))
	flags := uint64(0)
	if symbol != nil {
		flags = uint64(symbol.Flags)
	}
	out.number(flags)
	resolved := symbol
	if resolved != nil {
		resolved = checker.SkipAlias(resolved, c)
	}
	out.number(p.symbolID(resolved))
	write := func(subject *ast.Symbol) error {
		var declaration *ast.Node
		if subject != nil {
			declaration = subject.ValueDeclaration
			if declaration == nil && len(subject.Declarations) > 0 {
				declaration = subject.Declarations[0]
			}
		}
		out.yes(declaration != nil)
		if declaration == nil {
			return nil
		}
		source := ast.GetSourceFileOfNode(declaration)
		if source == nil {
			return fmt.Errorf("binding declaration lacks source")
		}
		out.text(source.FileName().AsString())
		out.text(strings.TrimPrefix(declaration.Kind.String(), "Kind"))
		out.number(uint64(declaration.Pos()))
		out.number(uint64(declaration.End()))
		out.yes(ast.IsVarConst(declaration))
		parentKind := ""
		parentStart, parentEnd := uint64(0), uint64(0)
		if declaration.Parent != nil {
			parentKind = strings.TrimPrefix(declaration.Parent.Kind.String(), "Kind")
			parentStart = uint64(declaration.Parent.Pos())
			parentEnd = uint64(declaration.Parent.End())
		}
		out.text(parentKind)
		out.number(parentStart)
		out.number(parentEnd)
		return nil
	}
	if err := write(symbol); err != nil {
		return "", err
	}
	if err := write(resolved); err != nil {
		return "", err
	}
	return out.String(), nil
}
