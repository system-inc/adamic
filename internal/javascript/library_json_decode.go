package javascript

import (
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) jsonDescriptor(schema ir.JSONDecodeSchema) string {
	descriptor, err := json.Marshal(schema)
	if err != nil {
		panic(err)
	}
	key := string(descriptor)
	if name, ok := e.jsonDescriptorNames[key]; ok {
		return name
	}
	if e.jsonDescriptorNames == nil {
		e.jsonDescriptorNames = map[string]string{}
	}
	name := fmt.Sprintf("adamicJsonSchema_%d", len(e.jsonDescriptors))
	e.jsonDescriptorNames[key] = name
	e.jsonDescriptors = append(e.jsonDescriptors, "const "+name+" = "+key+";\n")
	return name
}

func (e *emitter) jsonDecode(expression ir.JSONDecode) string {
	return "decodeJson(" + e.value(expression.Text) + ", " + e.jsonDescriptor(expression.Schema) + ")"
}
