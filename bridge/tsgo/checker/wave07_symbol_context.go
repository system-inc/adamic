package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strings"
)

// wave07SymbolContext exposes declarations and their complete syntax ancestry.
// It makes no platform-symbol, library-member, read, or lint decision.
func (p *Program) wave07SymbolContext(c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "wave07-symbol-context" {
		return "", fmt.Errorf("unexpected symbol context suffix")
	}
	out := &fields{}
	out.number(1)
	out.text(question)
	symbol := c.GetSymbolAtLocation(node)
	out.number(p.symbolID(symbol))
	value := symbol
	if node.Parent != nil && node.Parent.Kind == ast.KindShorthandPropertyAssignment {
		if s := c.GetShorthandAssignmentValueSymbol(node.Parent); s != nil {
			value = s
		}
	}
	out.number(p.symbolID(value))
	alias := symbol
	if symbol != nil && symbol.Flags&ast.SymbolFlagsAlias != 0 {
		alias = c.GetAliasedSymbol(symbol)
	}
	out.number(p.symbolID(alias))
	for _, selected := range []*ast.Symbol{symbol, alias} {
		if selected == nil {
			out.number(0)
			out.number(0)
		} else {
			out.number(uint64(selected.Flags))
			out.number(uint64(len(selected.Declarations)))
			for _, declaration := range selected.Declarations {
				p.wave07ContextDeclaration(out, declaration)
			}
		}
	}
	var signature *checker.Signature
	if node.Kind == ast.KindCallExpression || node.Kind == ast.KindNewExpression || node.Kind == ast.KindTaggedTemplateExpression {
		signature = c.GetResolvedSignature(node)
	}
	out.yes(signature != nil)
	if signature != nil {
		returns := c.GetReturnTypeOfSignature(signature)
		if returns == nil {
			out.number(0)
		} else {
			out.number(uint64(returns.Flags()))
		}
		declaration := signature.Declaration()
		out.yes(declaration != nil)
		if declaration != nil {
			p.wave07ContextDeclaration(out, declaration)
		}
	}
	return out.String(), nil
}

func (p *Program) wave07ContextDeclaration(out *fields, node *ast.Node) {
	file := ast.GetSourceFileOfNode(node)
	out.text(file.FileName().AsString())
	out.text(strings.TrimPrefix(node.Kind.String(), "Kind"))
	out.number(uint64(node.Pos()))
	out.number(uint64(node.End()))
	out.yes(file.IsDeclarationFile)
	out.yes(p.Compiler.IsSourceFileDefaultLibrary(file.PathKey()))
	out.yes(ast.IsExternalModule(file))
	name := ""
	if wave07TextName(node.Name()) {
		name = node.Name().Text()
	}
	out.text(name)
	out.number(uint64(node.Flags))
	functionFlags := ast.FunctionFlags(0)
	var body *ast.Node
	if ast.IsFunctionLike(node) {
		functionFlags = ast.GetFunctionFlags(node)
		body = node.Body()
	}
	out.number(uint64(functionFlags))
	out.yes(body != nil)
	if body != nil {
		out.number(uint64(body.Pos()))
		out.number(uint64(body.End()))
		out.text(strings.TrimPrefix(body.Kind.String(), "Kind"))
		out.text(file.Text()[body.Pos():body.End()])
	}
	var ancestors []*ast.Node
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		ancestors = append(ancestors, parent)
	}
	out.number(uint64(len(ancestors)))
	for _, parent := range ancestors {
		out.text(strings.TrimPrefix(parent.Kind.String(), "Kind"))
		name := ""
		if wave07TextName(parent.Name()) {
			name = parent.Name().Text()
		}
		out.text(name)
		nameKind := ""
		if parent.Name() != nil {
			nameKind = strings.TrimPrefix(parent.Name().Kind.String(), "Kind")
		}
		out.text(nameKind)
		out.number(uint64(parent.Flags))
		out.yes(ast.IsGlobalScopeAugmentation(parent))
	}
}

func wave07TextName(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindIdentifier, ast.KindPrivateIdentifier, ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindNoSubstitutionTemplateLiteral:
		return true
	}
	return false
}
