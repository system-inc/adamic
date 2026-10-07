package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

func voidUnionParts(subject *checker.Type) []*checker.Type {
	if subject == nil {
		return nil
	}
	if subject.Flags()&checker.TypeFlagsUnion != 0 {
		return subject.Types()
	}
	return []*checker.Type{subject}
}
func voidSignatureReturns(c *checker.Checker, subject *checker.Type) []*checker.Type {
	var results []*checker.Type
	for _, part := range voidUnionParts(subject) {
		for _, signature := range c.GetSignaturesOfType(part, checker.SignatureKindCall) {
			results = append(results, checker.Checker_getReturnTypeOfSignature(c, signature))
		}
	}
	return results
}
func writeVoidReturnTypes(out *fields, subjects []*checker.Type) {
	out.number(uint64(len(subjects)))
	for _, subject := range subjects {
		out.number(uint64(subject.Flags()))
		parts := voidUnionParts(subject)
		out.number(uint64(len(parts)))
		for _, part := range parts {
			out.number(uint64(part.Flags()))
		}
	}
}

// voidPositionTypes exposes actual and contextual signatures and inherited member
// types. No decision about acceptable returns, messages or edits is made here.
func (p *Program) voidPositionTypes(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "void-position-types" {
		return "", fmt.Errorf("unexpected void-position-types suffix")
	}
	actual := c.GetTypeAtLocation(node)
	if actual != nil {
		actual = checker.Checker_getApparentType(c, actual)
	}
	writeVoidReturnTypes(out, voidSignatureReturns(c, actual))
	expected := checker.Checker_getContextualType(c, node, 0)
	if node.Kind == ast.KindMethodDeclaration && ast.IsObjectLiteralMethod(node) {
		expected = nil
		objectType := checker.Checker_getContextualType(c, node.Parent, 0)
		name := node.Name()
		if name != nil {
			switch name.Kind {
			case ast.KindIdentifier, ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindNoSubstitutionTemplateLiteral:
				for _, part := range voidUnionParts(objectType) {
					symbol := checker.Checker_getPropertyOfType(c, part, name.Text())
					if symbol != nil {
						expected = c.GetTypeOfSymbolAtLocation(symbol, node)
						if expected != nil {
							break
						}
					}
				}
			}
		}
	}
	writeVoidReturnTypes(out, voidSignatureReturns(c, expected))
	var inherited [][]*checker.Type
	if node.Kind == ast.KindPropertyDeclaration || node.Kind == ast.KindMethodDeclaration || node.Kind == ast.KindGetAccessor {
		name := node.Name()
		var heritage *ast.NodeList
		parent := node.Parent
		if parent != nil {
			if parent.Kind == ast.KindClassDeclaration {
				heritage = parent.AsClassDeclaration().HeritageClauses
			} else if parent.Kind == ast.KindClassExpression {
				heritage = parent.AsClassExpression().HeritageClauses
			}
		}
		if name != nil && heritage != nil {
			symbol := c.GetSymbolAtLocation(name)
			if symbol != nil {
				for _, clause := range heritage.Nodes {
					for _, base := range clause.AsHeritageClause().Types.Nodes {
						baseType := c.GetTypeAtLocation(base)
						if baseType == nil {
							continue
						}
						property := checker.Checker_getPropertyOfType(c, baseType, symbol.Name)
						if property != nil {
							inherited = append(inherited, voidSignatureReturns(c, c.GetTypeOfSymbolAtLocation(property, node)))
						}
					}
				}
			}
		}
	}
	out.number(uint64(len(inherited)))
	for _, group := range inherited {
		writeVoidReturnTypes(out, group)
	}
	return out.String(), nil
}
