package checker

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Raw declaration ancestry and file facts. No producer or use classifications.
func writeContextNode(out *fields, node *ast.Node) {
	out.text(strings.TrimPrefix(node.Kind.String(), "Kind"))
	out.number(uint64(node.Pos()))
	out.number(uint64(node.End()))
	out.number(uint64(node.Flags))
	nameKind, name := "", ""
	if n := node.Name(); n != nil {
		nameKind = strings.TrimPrefix(n.Kind.String(), "Kind")
		switch n.Kind {
		case ast.KindIdentifier, ast.KindPrivateIdentifier, ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindNoSubstitutionTemplateLiteral:
			name = n.Text()
		}
	}
	out.text(nameKind)
	out.text(name)
	out.yes(node.Kind == ast.KindModuleDeclaration && ast.IsGlobalScopeAugmentation(node))
}
func writeDeclarationContext(out *fields, node *ast.Node) {
	source := ast.GetSourceFileOfNode(node)
	if source == nil {
		panic("declaration without source")
	}
	out.text(source.FileName())
	out.yes(source.IsDeclarationFile)
	out.yes(ast.IsExternalModule(source))
	writeContextNode(out, node)
	body := node.Body()
	out.yes(body != nil)
	if body != nil {
		out.text(strings.TrimPrefix(body.Kind.String(), "Kind"))
		out.number(uint64(body.Pos()))
		out.number(uint64(body.End()))
	}
	var parents []*ast.Node
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		parents = append(parents, parent)
	}
	out.number(uint64(len(parents)))
	for _, parent := range parents {
		writeContextNode(out, parent)
	}
}
func (p *Program) declarationContext(out *fields, c *checker.Checker, _ *ast.SourceFile, node *ast.Node, question string) error {
	split := strings.Split(question, "\n")
	if len(split) != 2 {
		return fmt.Errorf("declaration-context requires an operation")
	}
	var symbol *ast.Symbol
	var declarations []*ast.Node
	switch split[1] {
	case "signature":
		if node.Kind != ast.KindCallExpression && node.Kind != ast.KindNewExpression && node.Kind != ast.KindTaggedTemplateExpression {
			return fmt.Errorf("signature context requires a call")
		}
		if signature := c.GetResolvedSignature(node); signature != nil && signature.Declaration() != nil {
			declarations = []*ast.Node{signature.Declaration()}
		}
	case "node", "alias", "shorthand", "export":
		symbol = c.GetSymbolAtLocation(node)
		if split[1] == "alias" && symbol != nil && symbol.Flags&ast.SymbolFlagsAlias != 0 {
			symbol = c.GetAliasedSymbol(symbol)
		}
		if split[1] == "shorthand" {
			if node.Parent == nil || node.Parent.Kind != ast.KindShorthandPropertyAssignment {
				return fmt.Errorf("shorthand context requires a shorthand name")
			}
			symbol = c.GetShorthandAssignmentValueSymbol(node.Parent)
		}
		if split[1] == "export" {
			if node.Parent == nil || node.Parent.Kind != ast.KindExportSpecifier {
				return fmt.Errorf("export context requires an export name")
			}
			symbol = c.GetExportSpecifierLocalTargetSymbol(node.Parent)
		}
		if symbol != nil {
			declarations = symbol.Declarations
		}
	default:
		return fmt.Errorf("unknown declaration-context operation")
	}
	out.number(p.symbolID(symbol))
	flags := ast.SymbolFlags(0)
	if symbol != nil {
		flags = symbol.Flags
	}
	out.number(uint64(flags))
	out.number(uint64(len(declarations)))
	for _, declaration := range declarations {
		writeDeclarationContext(out, declaration)
	}
	return nil
}
func init() { additionalQuestions["declaration-context"] = (*Program).declarationContext }
