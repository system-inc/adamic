package tailwind

import (
	"encoding/json"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"reflect"
)

type AdamicSlot02Reading struct {
	Literals, Templates []string
	Shared              bool
}

// Observe the actual public method and its private prerequisite on a bound reader.
func AdamicSlot02Readings(nodes []*ast.Node) []AdamicSlot02Reading {
	reader := NewClassLiteralReader(DefaultClassLiteralSettings())
	reader.values = map[*ast.Node]classValues{}
	out := []AdamicSlot02Reading{}
	for _, node := range nodes {
		values := reader.classValuesIn(node)
		literals := reader.ClassLiteralsIn(node)
		row := AdamicSlot02Reading{Literals: []string{}, Templates: []string{}, Shared: reflect.ValueOf(literals).Pointer() == reflect.ValueOf(values.literals).Pointer()}
		for _, literal := range literals {
			data, err := json.Marshal(struct {
				Text              string
				Start, End        int
				Origin            ClassLiteralOrigin
				Leading, Trailing bool
			}{literal.Text, literal.Range.Pos(), literal.Range.End(), literal.Origin, literal.Edges.Leading, literal.Edges.Trailing})
			if err != nil {
				panic(err)
			}
			row.Literals = append(row.Literals, string(data))
		}
		for _, template := range values.templates {
			row.Templates = append(row.Templates, template.node.Kind.String())
		}
		out = append(out, row)
	}
	return out
}
