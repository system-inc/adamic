package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Raw inherited property flags and declaration containers. Adamic decides exemptions.
func (p *Program) baseMemberFacts(out *fields, c *checker.Checker, node *ast.Node, question string) error {
	if question != "base-member-facts" || (node.Kind != ast.KindGetAccessor && node.Kind != ast.KindPropertyDeclaration) {
		return fmt.Errorf("base-member-facts requires a getter or property")
	}
	class := node.Parent
	if class == nil || (class.Kind != ast.KindClassDeclaration && class.Kind != ast.KindClassExpression) {
		return fmt.Errorf("base-member-facts requires a class member")
	}
	member := node.Symbol()
	symbol := class.Symbol()
	out.yes(member != nil && symbol != nil)
	if member == nil || symbol == nil {
		return nil
	}
	out.number(uint64(checker.GetDeclarationModifierFlagsFromSymbol(member)))
	var bases []*ast.Symbol
	for _, t := range c.GetBaseTypes(c.GetDeclaredTypeOfSymbol(symbol)) {
		if b := c.GetPropertyOfType(t, member.Name); b != nil {
			bases = append(bases, b)
		}
	}
	out.number(uint64(len(bases)))
	for _, b := range bases {
		out.number(uint64(b.Flags))
		out.number(uint64(b.CheckFlags))
		out.number(uint64(checker.GetDeclarationModifierFlagsFromSymbol(b)))
		out.number(uint64(len(b.Declarations)))
		for _, d := range b.Declarations {
			out.yes(d.Parent != nil && d.Parent.Kind == ast.KindInterfaceDeclaration)
		}
	}
	return nil
}
