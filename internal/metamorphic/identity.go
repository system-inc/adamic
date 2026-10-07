package metamorphic

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// identitySites are values the program computes and hands on: an initializer, a returned value, an
// argument, an operand, an element, a property's value, a template's part. Each site passes the
// value through a generic identity function, or, every other site, through a closure called on the
// spot. Either way the value is computed where it was and arrives where it went, now as a parameter
// handed back, or as a closure's result: a borrow becomes a move, a temporary a returned value.
func identitySites(p *program) []site {
	name := p.fresh("metamorphicIdentity")
	header := declaration{name: name, text: "function " + name + "<T>(value: T): T {\n\treturn value;\n}"}
	var sites []site
	visit(p.file.AsNode(), func(node *ast.Node) bool {
		if ast.IsPartOfTypeNode(node) {
			return false
		}
		if !handedOn(node) || !computed(node) {
			return true
		}
		text := p.source(node)
		start := p.start(node)
		if len(sites)%2 == 0 {
			sites = append(sites, site{edits: []edit{{start, node.End(), name + "(" + text + ")"}}, header: []declaration{header}})
		} else {
			sites = append(sites, site{edits: []edit{{start, node.End(), "(() => " + text + ")()"}}})
		}
		return true
	})
	return sites
}

// handedOn says whether an expression is a value something receives as it is: not a callee, not a
// place written to, not a name.
func handedOn(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil || !ast.IsExpressionNode(node) || ast.IsAssignmentTarget(node) {
		return false
	}
	switch parent.Kind {
	case ast.KindVariableDeclaration:
		return parent.Initializer() == node
	case ast.KindReturnStatement:
		return true
	case ast.KindCallExpression, ast.KindNewExpression:
		return parent.Expression() != node
	case ast.KindBinaryExpression:
		operator := parent.AsBinaryExpression().OperatorToken.Kind
		if ast.IsAssignmentOperator(operator) {
			return parent.AsBinaryExpression().Right == node && operator == ast.KindEqualsToken
		}
		return operator != ast.KindCommaToken
	case ast.KindArrayLiteralExpression:
		return true
	case ast.KindPropertyAssignment:
		return parent.Initializer() == node
	case ast.KindTemplateSpan:
		return true
	}
	return false
}

// computed says whether an expression is worth passing through: a value something made, not a
// literal (whose type a function would widen) or a function (whose parameters would lose their
// contextual types).
func computed(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindIdentifier, ast.KindCallExpression, ast.KindNewExpression, ast.KindPropertyAccessExpression, ast.KindElementAccessExpression,
		ast.KindBinaryExpression, ast.KindTemplateExpression, ast.KindConditionalExpression, ast.KindParenthesizedExpression,
		ast.KindPrefixUnaryExpression, ast.KindTypeOfExpression:
		if node.Kind == ast.KindIdentifier && node.Text() == "undefined" {
			return false
		}
		return true
	}
	return false
}
