// Oracle-only access to Go's private readers. Production cohere is unchanged.
package tailwind

import (
	"encoding/json"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

type AdamicValues struct {
	Literals  []string `json:"literals"`
	Templates []string `json:"templates"`
}

func adamicValues(v classValues, index func(*ast.Node) int) AdamicValues {
	result := AdamicValues{[]string{}, []string{}}
	for _, l := range v.literals {
		b, e := json.Marshal(struct {
			Node              int
			Text              string
			Start, End        int
			Origin            ClassLiteralOrigin
			Leading, Trailing bool
		}{index(l.Node), l.Text, l.Range.Pos(), l.Range.End(), l.Origin, l.Edges.Leading, l.Edges.Trailing})
		if e != nil {
			panic(e)
		}
		result.Literals = append(result.Literals, string(b))
	}
	for _, v := range v.templates {
		b, e := json.Marshal(struct {
			Node              int
			Origin            ClassLiteralOrigin
			Leading, Trailing bool
		}{index(v.node), v.origin, v.edges.Leading, v.edges.Trailing})
		if e != nil {
			panic(e)
		}
		result.Templates = append(result.Templates, string(b))
	}
	return result
}
func AdamicSurfaceValues(reader *ClassLiteralReader, node *ast.Node, mode string, index func(*ast.Node) int) AdamicValues {
	var v classValues
	switch mode {
	case "read":
		v = reader.readClassValues(node)
	case "attribute":
		v = reader.attributeValues(node)
	case "callee":
		v = reader.calleeValues(node)
	case "variable":
		v = reader.variableValues(node)
	case "collect":
		collectClassValues(node, ClassLiteralOriginCallee, classValueEdges{}, &v)
	case "under":
		v = classValuesUnder(node, ClassLiteralOriginVariable)
	default:
		panic("unknown mode")
	}
	return adamicValues(v, index)
}
func AdamicPatternMatches(reader *ClassLiteralReader, text string) []bool {
	result := []bool{}
	result = append(result, reader.variables.matches(text))
	return result
}

func AdamicSurfaceRefuses(reader *ClassLiteralReader, node *ast.Node, mode string) (refused bool) {
	defer func() {
		if recover() != nil {
			refused = true
		}
	}()
	switch mode {
	case "callee":
		reader.calleeValues(node)
	case "variable":
		reader.variableValues(node)
	default:
		panic("bad refusal mode")
	}
	return false
}

func AdamicReadsCallee(reader *ClassLiteralReader, node *ast.Node) bool {
	return reader.readsCallee(node)
}
