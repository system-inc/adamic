package checker

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// classThisTypes returns compiler identities, not return-expression judgments.
func (p *Program) classThisTypes(out *fields, c *checker.Checker, node *ast.Node, question string) error {
	if question != "class-this-types" || (node.Kind != ast.KindClassDeclaration && node.Kind != ast.KindClassExpression) {
		return fmt.Errorf("class-this-types requires a class")
	}
	g := &graph{program: p, checker: c, seen: make(map[*checker.Type]bool)}
	t := c.GetTypeAtLocation(node)
	out.number(g.id(t))
	var receiver *checker.Type
	if t != nil && t.Flags()&checker.TypeFlagsObject != 0 && t.ObjectFlags()&checker.ObjectFlagsClassOrInterface != 0 {
		receiver = checker.InterfaceType_thisType(t.AsInterfaceType())
	}
	out.number(g.id(receiver))
	return nil
}
