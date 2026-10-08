package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strconv"
	"strings"
)

// The legacy no-suffix questions retain their wire grammar. A SourceFile request
// may append the local node's exact byte start, end and kind so the harness can
// enforce programReads through askFile without opening or parsing another file.
func (p *Program) inspectDeclarations(source *ast.SourceFile, node *ast.Node, c *checker.Checker, out *fields, mode, question string) (bool, error) {
	switch mode {
	case "node-symbol-details", "binding-declarations", "symbol-provenance":
	default:
		return false, nil
	}
	if question != mode {
		pieces := strings.Split(question, "\n")
		if len(pieces) != 4 || node != source.AsNode() {
			return true, fmt.Errorf("%s requires a SourceFile selector or no suffix", mode)
		}
		first, e1 := strconv.ParseUint(pieces[1], 10, 64)
		last, e2 := strconv.ParseUint(pieces[2], 10, 64)
		if e1 != nil || e2 != nil || strconv.FormatUint(first, 10) != pieces[1] || strconv.FormatUint(last, 10) != pieces[2] {
			return true, fmt.Errorf("invalid declaration node selector")
		}
		_, selected, err := p.exact(source.FileName().AsString(), first, last, pieces[3])
		if err != nil {
			return true, err
		}
		node = selected
	}
	if mode == "symbol-provenance" {
		return true, p.symbolProvenance(out, c, node, mode)
	}
	if mode == "node-symbol-details" {
		return true, p.declarationFacts(out, c, node, mode, mode)
	}
	if node.Kind != ast.KindIdentifier {
		return true, fmt.Errorf("binding-declarations requires an Identifier")
	}
	symbol := c.GetSymbolAtLocation(node)
	if node.Parent != nil && node.Parent.Kind == ast.KindShorthandPropertyAssignment {
		if value := c.GetShorthandAssignmentValueSymbol(node.Parent); value != nil {
			symbol = value
		}
	}
	out.yes(symbol != nil)
	if symbol != nil {
		out.number(uint64(symbol.Flags))
		out.number(uint64(len(symbol.Declarations)))
		for _, declaration := range symbol.Declarations {
			file := ast.GetSourceFileOfNode(declaration)
			if file == nil {
				return true, fmt.Errorf("binding declaration has no source")
			}
			out.text(file.FileName().AsString())
			out.text(strings.TrimPrefix(declaration.Kind.String(), "Kind"))
			out.number(uint64(declaration.Pos()))
			out.number(uint64(declaration.End()))
		}
	}
	return true, nil
}
