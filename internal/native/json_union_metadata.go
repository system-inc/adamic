package native

import "github.com/system-inc/adamic/internal/ir"

type jsonShapeOptions struct {
	Null  []bool
	Tuple bool
}

// Descriptors belong to allocations, not structural views of those allocations.
func jsonStorageSchema(of ir.Type, null bool) *ir.JSONSchema {
	kind := map[ir.Type]string{ir.Number: "number", ir.Boolean: "boolean", ir.String: "string", ir.MaybeNumber: "maybe_number", ir.MaybeBoolean: "maybe_boolean", ir.Closure: "function", ir.Map: "map"}[of]
	if kind == "" {
		kind = "dynamic"
	}
	return &ir.JSONSchema{Kind: kind, Null: null && of != ir.Union}
}

func (e *emitter) jsonCreatedArray(expression ir.Expression, result string) string {
	var element ir.Type
	switch v := expression.(type) {
	case ir.ArrayLiteral:
		element = v.Element
	case ir.ArrayMap:
		element = v.Result
	case ir.ArrayFrom:
		element = v.Element
	case ir.ArrayFill:
		if v.Array == nil {
			element = v.Element
		}
	case ir.ArrayVisit:
		if v.Method == "filter" {
			element = v.Element
		}
	case ir.MapKeys:
		element = v.Key
	case ir.MapValues:
		element = v.Value
	case ir.SetValues:
		element = v.Element
	case ir.CodePoints, ir.ProgramArguments:
		element = ir.String
	case ir.StringCall:
		if v.Method == "split" {
			element = ir.String
		}
	}
	if element != 0 {
		schema := e.jsonSchema(jsonStorageSchema(element, false))
		e.line("%s->json_element = %s;", result, schema)
	}
	return result
}
