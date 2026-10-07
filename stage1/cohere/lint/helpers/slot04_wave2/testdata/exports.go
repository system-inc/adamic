package tailwind

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"strings"
)

func AdamicLiteral(node *ast.Node) ClassLiteral {
	return classLiteralFrom(node, ClassLiteralOriginAttribute)
}
func AdamicHole(before, after string, first, last, leading, trailing bool) (bool, bool) {
	edges := holeEdges(before, after, first, last, classValueEdges{Leading: leading, Trailing: trailing})
	return edges.Leading, edges.Trailing
}
func adamicPrint(label string, values classValues, ids map[*ast.Node]int) string {
	var result strings.Builder
	result.WriteString(label)
	for _, v := range values.literals {
		index, ok := ids[v.Node]
		if !ok {
			panic("literal absent from arena")
		}
		fmt.Fprintf(&result, " L%d:%s:%t:%t:%d:%d:", index, v.Origin, v.Edges.Leading, v.Edges.Trailing, v.Range.Pos(), v.Range.End())
		for _, b := range []byte(v.Text) {
			fmt.Fprintf(&result, "%d,", b)
		}
	}
	for _, v := range values.templates {
		index, ok := ids[v.node]
		if !ok {
			panic("template absent from arena")
		}
		fmt.Fprintf(&result, " T%d:%s:%t:%t", index, v.origin, v.edges.Leading, v.edges.Trailing)
	}
	return result.String()
}
func AdamicOutputs(node *ast.Node, origin string, ids map[*ast.Node]int) []string {
	result := []string{}
	seed := ClassLiteral{Text: "", Origin: "seed", Edges: classValueEdges{Trailing: true}}
	// nil has ID -1 in normal input; this distinct synthetic node marks preexisting output.
	marker := &ast.Node{}
	ids[marker] = -2
	seed.Node = marker
	for mask := 0; mask < 4; mask++ {
		values := classValues{literals: []ClassLiteral{seed}}
		// The seed range in the adapter is -1,-1, not Go's zero range.
		collectClassValues(node, ClassLiteralOrigin(origin), classValueEdges{Leading: mask >= 2, Trailing: mask%2 == 1}, &values)
		line := adamicPrint("collect", values, ids)
		line = strings.Replace(line, "L-2:seed:false:true:0:0:", "L-2:seed:false:true:-1:-1:", 1)
		result = append(result, line)
	}
	values := classValuesUnder(node, ClassLiteralOrigin(origin))
	result = append(result, adamicPrint("under", values, ids))
	values.literals = append(values.literals, seed)
	result = append(result, adamicPrint("fresh", classValuesUnder(node, ClassLiteralOrigin(origin)), ids))

	delete(ids, marker)
	return result
}
