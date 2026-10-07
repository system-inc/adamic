package tailwind

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"io"
)

func adamicWave3ValueRows(values classValues, ids map[*ast.Node]int) [][]any {
	literals, templates := [][]any{}, [][]any{}
	for _, value := range values.literals {
		literals = append(literals, []any{ids[value.Node], value.Text, value.Range.Pos(), value.Range.End(), string(value.Origin), value.Edges.Leading, value.Edges.Trailing})
	}
	for _, value := range values.templates {
		templates = append(templates, []any{ids[value.node], string(value.origin), value.edges.Leading, value.edges.Trailing})
	}
	return [][]any{{literals}, {templates}}
}
func AdamicWave3Under(node *ast.Node, origin string, ids map[*ast.Node]int) [][]any {
	return adamicWave3ValueRows(classValuesUnder(node, ClassLiteralOrigin(origin)), ids)
}
func AdamicWave3Attribute(node *ast.Node, names map[string]bool, ids map[*ast.Node]int) [][]any {
	reader := &ClassLiteralReader{attributeNames: names}
	return adamicWave3ValueRows(reader.attributeValues(node), ids)
}
func AdamicWave3Print(output io.Writer, rows [][]any) {
	literals := rows[0][0].([][]any)
	templates := rows[1][0].([][]any)
	fmt.Fprintf(output, "%d %d\n", len(literals), len(templates))
	for _, v := range literals {
		fmt.Fprintf(output, "%v %v %v %v %v\n%v\n%v\n", v[0], v[2], v[3], v[5], v[6], v[1], v[4])
	}
	for _, v := range templates {
		fmt.Fprintf(output, "%v %v %v\n%v\n", v[0], v[2], v[3], v[1])
	}
}
