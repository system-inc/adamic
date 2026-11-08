package javascript

import (
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

// The checked runtime messages are synchronized with the record C interface.
const recordRuntime = `
const adamicRecordGet = (record, key) => {
 if (Object.hasOwn(record, key)) return record[key];
 if (Object.hasOwn(Object.prototype, key)) panic("record member '" + key + "' is missing; records hold own keys only");
 return undefined;
};
const adamicRecordHas = (key, record) => {
 if (Object.hasOwn(record, key)) return true;
 if (Object.hasOwn(Object.prototype, key)) panic("record member '" + key + "' is missing; records hold own keys only");
 return false;
};
const adamicRecordSet = (record, key, value) => {
 if (key === '__proto__') panic('NotYet: record assignment to __proto__ requires the Object.prototype setter; use an own data property');
 record[key] = value;
 return value;
};
`

func (javascriptEmitter *emitter) recordCall(call ir.RecordCall) string {
	arguments := javascriptEmitter.values(call.Arguments)
	switch call.Method {
	case "get":
		if call.OwnOnly {
			return "((r,k) => Object.hasOwn(r,k) ? r[k] : undefined)(" + arguments + ")"
		}
		return "adamicRecordGet(" + arguments + ")"
	case "has":
		return "adamicRecordHas(" + arguments + ")"
	case "set":
		return "adamicRecordSet(" + arguments + ")"
	case "delete":
		return "((r, k) => delete r[k])(" + arguments + ")"
	case "defineHidden":
		// Node throws the redefinition's TypeError itself, so the call is always true here, and the
		// caller's throw on false never runs.
		return "((r, k, v) => (Object.defineProperty(r, k, {enumerable: false, writable: false, value: v}), true))(" +
			arguments +
			")"
	case "entries":
		return "Object.entries(" + arguments + ").map(adamicTuple)"
	case "hasOwn", "keys", "values":
		return "Object." + call.Method + "(" + arguments + ")"
	}
	panic("javascript: unknown record operation")
}
func (javascriptEmitter *emitter) recordLiteral(literal ir.RecordLiteral) string {
	// Computed data properties preserve __proto__ and key/value evaluation order.
	parts := []string{}
	if literal.Spread != nil {
		parts = append(parts, "..."+javascriptEmitter.value(literal.Spread))
	}
	for _, entry := range literal.Entries {
		parts = append(parts, "["+javascriptEmitter.value(entry.Key)+"]: "+javascriptEmitter.value(entry.Value))
	}
	return "({" + strings.Join(parts, ", ") + "})"
}
