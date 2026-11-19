package javascript

import (
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

// The checked runtime messages are synchronized with the record C interface.
//
// A record and a fixed object are both plain objects here, so every record the backend makes is marked in
// adamicRecords, as native's shape marks it: a record literal (recordLiteral) and an entries record (an
// ObjectLiteral with Record), the two IR nodes that make one natively (emit_records.go, entries_records.go). A
// checked narrow to a fixed object refuses a marked object, and one to a record requires the mark
// (adamicPlaceholderNarrow), so a dictionary under a declared fixed-object key stops here as natively (#6c4swvt).
const recordRuntime = `
const adamicRecords = new WeakSet();
const adamicRecord = (record) => (adamicRecords.add(record), record);
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
	case "defineHidden":
		// Node throws the redefinition's TypeError itself, so the call is always true here, and the
		// caller's throw on false never runs.
		return "((r, k, v) => (Object.defineProperty(r, k, {enumerable: false, writable: false, value: v}), true))(" + args + ")"
	case "entries":
		return "Object.entries(" + args + ").map(adamicTuple)"
	case "hasOwn", "keys", "values":
		return "Object." + c.Method + "(" + args + ")"
	case "forInKeys":
		// A for...in's own-key list, which native may share between loops; here each loop has its own.
		return "Object.keys(" + args + ")"
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
	return "adamicRecord({" + strings.Join(parts, ", ") + "})"
}
