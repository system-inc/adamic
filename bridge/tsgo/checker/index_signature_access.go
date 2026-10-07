package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Raw compiler option, property presence and index key flags. Adamic decides exemptions.
func (p *Program) indexSignatureAccess(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "index-signature-access" {
		return "", fmt.Errorf("unexpected index-signature-access suffix")
	}
	out.yes(p.Compiler.Options().NoPropertyAccessFromIndexSignature.IsTrue())
	if node.Kind == ast.KindSourceFile {
		return out.String(), nil
	}
	if node.Kind != ast.KindElementAccessExpression {
		return "", fmt.Errorf("index-signature-access requires a SourceFile or ElementAccessExpression")
	}
	access := node.AsElementAccessExpression()
	if access.Expression == nil || access.ArgumentExpression == nil {
		return "", fmt.Errorf("incomplete element access")
	}
	key := access.ArgumentExpression
	for key.Kind == ast.KindParenthesizedExpression {
		key = key.AsParenthesizedExpression().Expression
	}
	subject := c.GetNonNullableType(c.GetTypeAtLocation(access.Expression))
	property := c.GetSymbolAtLocation(key)
	if property == nil && key.Kind == ast.KindStringLiteral {
		for _, candidate := range c.GetPropertiesOfType(subject) {
			if candidate.Name == key.Text() {
				property = candidate
				break
			}
		}
	}
	out.yes(property != nil)
	infos := c.GetIndexInfosOfType(subject)
	out.number(uint64(len(infos)))
	for _, info := range infos {
		out.number(uint64(info.KeyType().Flags()))
	}
	return out.String(), nil
}
