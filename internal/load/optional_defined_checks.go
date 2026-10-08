package load

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"strings"
)

// A failed strict optional relation can propagate undefined into a later use.
// Admit only the simple undefined-to-required diagnostic, never nested failures.
func (p *Program) acceptOptionalDefined(site OptionSite) bool {
	if len(site.Options) != 1 || site.Options[0] != "exactOptionalPropertyTypes" || (site.Code != 2322 && site.Code != 2345) {
		return false
	}
	lines := strings.Split(site.Message, "\n")
	if len(lines) != 2 || !strings.Contains(lines[0], " | undefined' is not assignable") || !strings.HasPrefix(strings.TrimSpace(lines[1]), "Type 'undefined' is not assignable") {
		return false
	}
	for _, file := range p.compiler.GetSourceFiles() {
		if file.FileName().AsString() != site.File {
			continue
		}
		var found *ast.Node
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			switch node.Kind {
			case ast.KindIdentifier:
				if site.Position == scanner.GetTokenPosOfNode(node, file, false) {
					found = node
				}
			case ast.KindReturnStatement:
				if site.Position == scanner.GetTokenPosOfNode(node, file, false) {
					found = node.AsReturnStatement().Expression
				}
			case ast.KindPropertyAssignment:
				if site.Position >= node.Name().Pos() && site.Position < node.Name().End() {
					found = node.AsPropertyAssignment().Initializer
				}
			case ast.KindBinaryExpression:
				binary := node.AsBinaryExpression()
				if binary.OperatorToken.Kind == ast.KindEqualsToken && site.Position >= binary.Left.Pos() && site.Position < binary.Left.End() {
					found = binary.Right
				}
			}
			if found != nil {
				return true
			}
			return node.ForEachChild(visit)
		}
		file.AsNode().ForEachChild(visit)
		if found != nil {
			if p.optionalViews == nil {
				p.optionalViews = map[string]OptionalViewContract{}
			}
			if p.checkedOptions == nil {
				p.checkedOptions = map[string]bool{}
			}
			p.optionalViews[p.Where(ast.SkipParentheses(found))] = OptionalViewContract{Site: site, Defined: true}
			return true
		}
	}
	return false
}
