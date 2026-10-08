package load

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
)

// Admit only a fresh literal's own data properties. Spread, accessor and nested
// function contracts need their own checks and remain diagnostics.
func (p *Program) acceptOptionalLiteral(site OptionSite) bool {
	if len(site.Options) != 1 || site.Options[0] != "exactOptionalPropertyTypes" {
		return false
	}
	switch site.Code {
	case 2375, 2379, 2322, 2345:
	default:
		return false
	}
	for _, file := range p.compiler.GetSourceFiles() {
		if file.FileName().AsString() != site.File {
			continue
		}
		var found *ast.Node
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			var candidate *ast.Node
			switch node.Kind {
			case ast.KindObjectLiteralExpression:
				if site.Position == scanner.GetTokenPosOfNode(node, file, false) {
					candidate = node
				}
			case ast.KindVariableDeclaration:
				if node.Name() != nil && site.Position >= node.Name().Pos() && site.Position < node.Name().End() {
					candidate = node.AsVariableDeclaration().Initializer
				}
			case ast.KindBinaryExpression:
				binary := node.AsBinaryExpression()
				if binary.OperatorToken.Kind == ast.KindEqualsToken && site.Position >= binary.Left.Pos() && site.Position < binary.Left.End() {
					candidate = binary.Right
				}
			case ast.KindReturnStatement:
				if site.Position == scanner.GetTokenPosOfNode(node, file, false) {
					candidate = node.AsReturnStatement().Expression
				}
			case ast.KindPropertyAssignment:
				if site.Position >= node.Name().Pos() && site.Position < node.Name().End() {
					candidate = node.AsPropertyAssignment().Initializer
				}
			}
			if candidate != nil {
				candidate = ast.SkipParentheses(candidate)
				if candidate.Kind == ast.KindObjectLiteralExpression && plainDataLiteral(candidate) {
					found = candidate
					return true
				}
			}
			return node.ForEachChild(visit)
		}
		file.AsNode().ForEachChild(visit)
		if found != nil {
			if p.optionalLiterals == nil {
				p.optionalLiterals = map[string][]OptionSite{}
			}
			if p.checkedOptions == nil {
				p.checkedOptions = map[string]bool{}
			}
			where := p.Where(found)
			p.optionalLiterals[where] = append(p.optionalLiterals[where], site)
			return true
		}
	}
	return false
}

func plainDataLiteral(node *ast.Node) bool {
	for _, field := range node.AsObjectLiteralExpression().Properties.Nodes {
		if field.Kind != ast.KindPropertyAssignment && field.Kind != ast.KindShorthandPropertyAssignment {
			return false
		}
		if field.Name() == nil || (field.Name().Kind != ast.KindIdentifier && field.Name().Kind != ast.KindStringLiteral && field.Name().Kind != ast.KindNumericLiteral) {
			return false
		}
		if field.Kind == ast.KindPropertyAssignment && field.Name().Text() == "__proto__" {
			return false
		}
	}
	return len(node.AsObjectLiteralExpression().Properties.Nodes) > 0
}

func (p *Program) OptionalLiteralSite(node *ast.Node) bool {
	return len(p.optionalLiterals[p.Where(node)]) > 0
}
func (p *Program) RecordOptionalLiteral(node *ast.Node) { p.checkedOptions[p.Where(node)] = true }
