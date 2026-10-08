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

func (e *emitter) recordCall(c ir.RecordCall) string {
	if c.DictionaryRead != nil && len(e.program.ViewOrigins) != 0 {
		return e.dictionaryRead(*c.DictionaryRead)
	}
	args := e.values(c.Arguments)
	switch c.Method {
	case "get":
		if c.OwnOnly {
			return "((r,k) => Object.hasOwn(r,k) ? r[k] : undefined)(" + args + ")"
		}
		return "adamicRecordGet(" + args + ")"
	case "has":
		return "adamicRecordHas(" + args + ")"
	case "set":
		return "adamicRecordSet(" + args + ")"
	case "delete":
		return "((r, k) => delete r[k])(" + args + ")"
	case "hasOwn", "keys", "values", "entries":
		return "Object." + c.Method + "(" + args + ")"
	}
	panic("javascript: unknown record operation")
}
func (e *emitter) recordLiteral(r ir.RecordLiteral) string {
	// Computed data properties preserve __proto__ and key/value evaluation order.
	parts := []string{}
	if r.Spread != nil {
		parts = append(parts, "..."+e.value(r.Spread))
	}
	for _, entry := range r.Entries {
		parts = append(parts, "["+e.value(entry.Key)+"]: "+e.value(entry.Value))
	}
	return "({" + strings.Join(parts, ", ") + "})"
}
