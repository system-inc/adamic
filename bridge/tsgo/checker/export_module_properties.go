package checker

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Raw compiler facts only. The native consumer decides findings and edits.
func (p *Program) exportModuleProperties(c *checker.Checker, node *ast.Node, question string) (string, error) {
	const mode = "export-module-properties"
	if question != mode {
		return "", fmt.Errorf("unexpected export facts suffix")
	}
	out := &fields{}
	out.number(1)
	out.text(mode)
	if node.Kind != ast.KindStringLiteral {
		return "", fmt.Errorf("export-module-properties requires a module string")
	}
	symbol := c.GetSymbolAtLocation(node)
	out.yes(symbol != nil)
	if symbol != nil {
		t := c.GetTypeOfSymbol(symbol)
		out.yes(t != nil)
		if t != nil {
			properties := checker.Checker_getPropertiesOfType(c, t)
			out.number(uint64(len(properties)))
			for _, property := range properties {
				out.text(strings.ToValidUTF8(property.Name, "�"))
				out.yes(checker.Checker_getPropertyOfType(c, t, property.Name) != nil)
			}
		}
	}
	return out.String(), nil
}
