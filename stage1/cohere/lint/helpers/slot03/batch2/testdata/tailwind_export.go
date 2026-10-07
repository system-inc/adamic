package tailwind

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

func AdamicHoleEdges(before, after string, first, last, leading, trailing bool) (bool, bool) {
	edges := holeEdges(before, after, first, last, classValueEdges{Leading: leading, Trailing: trailing})
	return edges.Leading, edges.Trailing
}

var adamicRoute string

type AdamicValues struct {
	Literals  []string `json:"literals"`
	Templates []string `json:"templates"`
}

func adamicValues(values classValues) AdamicValues {
	out := AdamicValues{Literals: []string{}, Templates: []string{}}
	for _, literal := range values.literals {
		out.Literals = append(out.Literals, fmt.Sprintf("%d,%d,%s,%t,%t,%q", literal.Range.Pos(), literal.Range.End(), literal.Origin, literal.Edges.Leading, literal.Edges.Trailing, literal.Text))
	}
	for _, template := range values.templates {
		out.Templates = append(out.Templates, fmt.Sprintf("%d,%d,%s,%t,%t", template.node.Pos(), template.node.End(), template.origin, template.edges.Leading, template.edges.Trailing))
	}
	return out
}
func AdamicDispatch(node *ast.Node, settings ClassLiteralSettings) (AdamicValues, string, AdamicValues, AdamicValues, AdamicValues) {
	reader := NewClassLiteralReader(settings)
	adamicRoute = ""
	values := adamicValues(reader.readClassValues(node))
	route := adamicRoute
	attribute, callee, variable := classValues{}, classValues{}, classValues{}
	if node != nil {
		switch node.Kind {
		case ast.KindJsxAttribute:
			attribute = reader.attributeValues(node)
		case ast.KindCallExpression:
			callee = reader.calleeValues(node)
		case ast.KindVariableDeclaration:
			variable = reader.variableValues(node)
		}
	}
	return values, route, adamicValues(attribute), adamicValues(callee), adamicValues(variable)
}
