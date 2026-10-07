package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strings"
)

// wave06Declarations projects compiler declarations. Native rules own judgments.
// Reached through the earlier wave-owned promised-return fallback.
func (p *Program) wave06Declarations(out *fields, c *checker.Checker, node *ast.Node, question string) error {
	parts := strings.Split(question, "\n")
	if len(parts) != 2 || parts[0] != "wave06-declarations" {
		return fmt.Errorf("invalid wave06 declaration question")
	}
	var declarations []*ast.Node
	var symbol *ast.Symbol
	switch parts[1] {
	case "call":
		if node.Kind != ast.KindCallExpression {
			return fmt.Errorf("call declaration requires CallExpression")
		}
		if signature := c.GetResolvedSignature(node); signature != nil && signature.Declaration() != nil {
			declarations = append(declarations, signature.Declaration())
		}
	case "symbol", "alias":
		symbol = c.GetSymbolAtLocation(node)
		if node.Parent != nil && node.Parent.Kind == ast.KindShorthandPropertyAssignment {
			symbol = c.GetShorthandAssignmentValueSymbol(node.Parent)
		}
		if parts[1] == "alias" && symbol != nil && symbol.Flags&ast.SymbolFlagsAlias != 0 {
			symbol = c.GetAliasedSymbol(symbol)
		}
		if symbol != nil {
			declarations = symbol.Declarations
		}
	default:
		return fmt.Errorf("invalid wave06 declaration mode")
	}
	out.number(p.symbolID(symbol))
	out.number(uint64(len(declarations)))
	for _, declaration := range declarations {
		source := ast.GetSourceFileOfNode(declaration)
		if source == nil {
			return fmt.Errorf("declaration has no source")
		}
		out.text(strings.TrimPrefix(declaration.Kind.String(), "Kind"))
		name := ""
		if wave06TextName(declaration.Name()) {
			name = declaration.Name().Text()
		}
		out.text(name)
		out.text(source.FileName())
		out.number(uint64(declaration.Pos()))
		out.number(uint64(declaration.End()))
		out.yes(source.IsDeclarationFile)
		out.yes(ast.IsExternalModule(source))
		var parents []*ast.Node
		for parent := declaration.Parent; parent != nil; parent = parent.Parent {
			parents = append(parents, parent)
		}
		out.number(uint64(len(parents)))
		for _, parent := range parents {
			out.text(strings.TrimPrefix(parent.Kind.String(), "Kind"))
			name = ""
			if wave06TextName(parent.Name()) {
				name = parent.Name().Text()
			}
			out.text(name)
			out.number(uint64(parent.Flags))
			out.yes(parent.Kind == ast.KindModuleDeclaration && ast.IsGlobalScopeAugmentation(parent))
		}
		body := ""
		if !source.IsDeclarationFile && parts[1] == "alias" {
			function := declaration
			if (declaration.Kind == ast.KindVariableDeclaration || declaration.Kind == ast.KindPropertyDeclaration) && declaration.Initializer() != nil {
				function = ast.SkipParentheses(declaration.Initializer())
			}
			if ast.IsFunctionLike(function) && function.Body() != nil {
				body = source.Text()[function.Body().Pos():function.Body().End()]
			}
		}
		out.text(body)
	}
	return nil
}

// wave06Question preserves the previous unsupported-mode error for other modes.
func (p *Program) wave06Question(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question == "wave06-jsx-bindings" {
		return p.wave06JSXBindings(out, c, node)
	}
	if strings.Split(question, "\n")[0] != "wave06-declarations" {
		return "", fmt.Errorf("unsupported checker question: %s", question)
	}
	if err := p.wave06Declarations(out, c, node, question); err != nil {
		return "", err
	}
	return out.String(), nil
}

// Binding patterns and computed names have no Node.Text projection.
func wave06TextName(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindIdentifier, ast.KindPrivateIdentifier, ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindNoSubstitutionTemplateLiteral:
		return true
	}
	return false
}
