package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strings"
)

func (p *Program) valueDeclaration(out *fields, c *checker.Checker, node *ast.Node) error {
	if node.Kind != ast.KindIdentifier {
		return fmt.Errorf("value-declaration requires an Identifier")
	}
	symbol := c.GetSymbolAtLocation(node)
	out.yes(symbol != nil)
	if symbol == nil {
		return nil
	}
	out.number(uint64(symbol.Flags))
	d := symbol.ValueDeclaration
	if d == nil {
		out.number(0)
		return nil
	}
	file := ast.GetSourceFileOfNode(d)
	if file == nil {
		return fmt.Errorf("value declaration without source")
	}
	out.number(1)
	out.text(file.FileName())
	out.text(strings.TrimPrefix(d.Kind.String(), "Kind"))
	out.number(uint64(d.Pos()))
	out.number(uint64(d.End()))
	return nil
}
func init() { additionalQuestions["value-declaration"] = (*Program).valueDeclaration }
