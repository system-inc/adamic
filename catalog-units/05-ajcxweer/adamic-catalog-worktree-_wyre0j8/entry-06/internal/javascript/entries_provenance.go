package javascript

import (
	"fmt"

	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) checkedObjectEnumeration(call ir.ObjectCall) string {
	primitive := map[ir.Type]string{ir.Number: "number", ir.String: "string", ir.Boolean: "boolean"}[call.Element]
	// Get each enumerated value once. Accessors may mutate later properties, so neither
	// a declared-field snapshot nor a second enumeration may substitute for this one.
	result := "pairs.map(pair => pair[1])"
	if call.Method == "entries" {
		result = "pairs"
	}
	restriction := ""
	if len(call.Allowed) > 0 {
		restriction = " || ![" + e.values(call.Allowed) + "].includes(value)"
	}
	return fmt.Sprintf("((object) => { const pairs = []; for (const key of Object.keys(object)) { if (!Object.getOwnPropertyDescriptor(object, key)?.enumerable) continue; const value = adamicFieldReadiness.get(object)?.has(key) ? undefined : object[key]; if (adamicTypeOf(value) !== %s%s) panic(\"Object enumeration key '\" + key + \"': actual \" + adamicTypeOf(value) + \", declared \" + %s); pairs.push([key, value]); } return %s; })(%s)", quote(primitive), restriction, quote(call.ElementName), result, e.value(call.Arguments[0]))
}
