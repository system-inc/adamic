package javascript

import (
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

func (javascriptEmitter *emitter) jsonDescriptor(schema ir.JSONDecodeSchema) string {
	descriptor, encodeError := json.Marshal(schema)
	if encodeError != nil {
		panic(encodeError)
	}
	key := string(descriptor)
	if name, ok := javascriptEmitter.jsonDescriptorNames[key]; ok {
		return name
	}
	if javascriptEmitter.jsonDescriptorNames == nil {
		javascriptEmitter.jsonDescriptorNames = map[string]string{}
	}
	name := fmt.Sprintf("adamicJsonSchema_%d", len(javascriptEmitter.jsonDescriptors))
	javascriptEmitter.jsonDescriptorNames[key] = name
	javascriptEmitter.jsonDescriptors = append(javascriptEmitter.jsonDescriptors, "const "+name+" = "+key+";\n")
	return name
}

func (javascriptEmitter *emitter) jsonDecode(expression ir.JSONDecode) string {
	return "decodeJson(" +
		javascriptEmitter.value(expression.Text) +
		", " +
		javascriptEmitter.jsonDescriptor(expression.Schema) +
		")"
}
