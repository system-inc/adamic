package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strings"
)

// Narrow contracts never serialize foreign declaration records or source bodies.
// Identity IDs belong to this Program; declarations retain compiler order.
func (p *Program) localSymbolFacts(source *ast.SourceFile, node *ast.Node, c *checker.Checker, out *fields, question string) (string, error) {
	symbol := c.GetSymbolAtLocation(node)
	switch question {
	case "local-symbol-details":
		out.yes(symbol != nil)
		if symbol != nil {
			out.number(p.symbolID(symbol))
			out.number(uint64(symbol.Flags))
			out.text(strings.ToValidUTF8(symbol.Name, "�"))
			out.number(uint64(len(symbol.Declarations)))
			var local []*ast.Node
			for _, declaration := range symbol.Declarations {
				if ast.GetSourceFileOfNode(declaration) == source {
					local = append(local, declaration)
				}
			}
			out.number(uint64(len(local)))
			for _, declaration := range local {
				out.text(strings.TrimPrefix(declaration.Kind.String(), "Kind"))
				out.number(uint64(declaration.Pos()))
				out.number(uint64(declaration.End()))
			}
		}
	case "symbol-identity":
		out.yes(symbol != nil)
		if symbol != nil {
			out.number(p.symbolID(symbol))
		}
	case "shorthand-value-identity":
		property := node
		if property.Kind == ast.KindIdentifier && property.Parent != nil {
			property = property.Parent
		}
		if property.Kind != ast.KindShorthandPropertyAssignment {
			return "", fmt.Errorf("shorthand-value-identity requires a shorthand property")
		}
		value := c.GetShorthandAssignmentValueSymbol(property)
		out.yes(value != nil)
		if value != nil {
			out.number(p.symbolID(value))
		}
	case "alias-target-identity", "skip-alias-identity":
		target := symbol
		if target != nil {
			if question == "skip-alias-identity" {
				target = checker.SkipAlias(target, c)
			} else if target.Flags&ast.SymbolFlagsAlias != 0 {
				target = c.GetAliasedSymbol(target)
			}
		}
		out.yes(target != nil)
		if target != nil {
			out.number(p.symbolID(target))
			out.number(uint64(target.Flags))
		}
	case "local-alias-declarations":
		if symbol != nil {
			symbol = checker.SkipAlias(symbol, c)
		}
		out.yes(symbol != nil)
		if symbol != nil {
			out.number(uint64(symbol.Flags))
			var local []*ast.Node
			for _, declaration := range symbol.Declarations {
				if ast.GetSourceFileOfNode(declaration) == source {
					local = append(local, declaration)
				}
			}
			out.number(uint64(len(local)))
			for _, declaration := range local {
				out.text(strings.TrimPrefix(declaration.Kind.String(), "Kind"))
				out.number(uint64(declaration.Pos()))
				out.number(uint64(declaration.End()))
			}
		}
	case "first-declaration-file":
		out.yes(symbol != nil)
		if symbol != nil {
			out.yes(len(symbol.Declarations) != 0)
			if len(symbol.Declarations) != 0 {
				file := ast.GetSourceFileOfNode(symbol.Declarations[0])
				if file == nil {
					return "", fmt.Errorf("first symbol declaration has no source")
				}
				out.yes(file.IsDeclarationFile)
			}
		}
	case "local-binding-declarations":
		if node.Kind != ast.KindIdentifier {
			return "", fmt.Errorf("local-binding-declarations requires an Identifier")
		}
		if node.Parent != nil && node.Parent.Kind == ast.KindShorthandPropertyAssignment {
			symbol = c.GetShorthandAssignmentValueSymbol(node.Parent)
		}
		out.yes(symbol != nil)
		if symbol != nil {
			var declarations []*ast.Node
			for _, declaration := range symbol.Declarations {
				if ast.GetSourceFileOfNode(declaration) == source {
					declarations = append(declarations, declaration)
				}
			}
			out.number(uint64(len(declarations)))
			for _, declaration := range declarations {
				out.text(strings.TrimPrefix(declaration.Kind.String(), "Kind"))
				out.number(uint64(declaration.Pos()))
				out.number(uint64(declaration.End()))
			}
		}
	default:
		return "", fmt.Errorf("unknown narrow symbol question")
	}
	return out.String(), nil
}
