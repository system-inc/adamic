package javascript

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

// Hidden readiness state leaves own keys and object spread unchanged.
var fieldReadinessRuntime = representationTagsRuntime() + `const adamicFieldReadiness = new WeakMap();
const adamicFieldRepresentations = new WeakMap();
const adamicSlotContracts = new WeakMap();
const adamicRealTypes = new WeakMap();
const adamicRecordSlotContracts = (object, slots, real) => { adamicSlotContracts.set(object, slots); adamicRealTypes.set(object, real); return object; };
const adamicCheckedWrite = (object, name, value, allowed, where) => {
 if (object === null || object === undefined || !Object.hasOwn(object, name) || !allowed.includes(adamicSlotContracts.get(object)?.[name])) panic("field write failed: property '" + name + "' on " + (object === null || object === undefined ? "undefined" : adamicRealTypes.get(object) || "record") + " at " + where + " has no compatible declared slot");
 adamicWriteField(object, name, value);
};
const adamicRecordFieldTypes = (object, types) => { adamicFieldRepresentations.set(object, {...adamicFieldRepresentations.get(object), ...types}); return object; };
const adamicViewWrite = (object, name, value, type) => { if (!Object.hasOwn(object, name) || adamicFieldRepresentations.get(object)?.[name] !== type && !(adamicFieldRepresentations.get(object)?.[name] === adamicTags.Union && type <= adamicTags.Boolean) && !(adamicFieldRepresentations.get(object)?.[name] === adamicTags.MaybeNumber && type === adamicTags.Number)) adamicViewField(object, name, "<write>." + name, type); adamicWriteField(object, name, value); adamicRecordFieldTypes(object, {[name]: adamicFieldRepresentations.get(object)?.[name] === adamicTags.Union && type <= adamicTags.Boolean ? adamicTags.Union : type}); };
const adamicUninitializedFields = (object, names) => { adamicFieldReadiness.set(object, new Set(names)); return object; };
const adamicReadField = (object, name, expression, optional = false, allowAbsent = false, fieldView = false) => {
    if (optional && (object === undefined || object === null)) return undefined;
    if (object === undefined || object === null) panic(fieldView ? "field read failed: " + expression + " is not initialized; expected " + fieldView + ", found missing" : "read before assignment: field '" + name + "' in " + expression);
    if (!Object.hasOwn(object, name) && typeof object === 'function' && adamicClassIdentities.has(Object.getPrototypeOf(object))) return adamicReadField(Object.getPrototypeOf(object), name, expression, false, allowAbsent, fieldView);
    if (allowAbsent && !Object.hasOwn(object, name)) return undefined;
    if (!Object.hasOwn(object, name) || adamicFieldReadiness.get(object)?.has(name)) panic(fieldView ? "field read failed: " + expression + " is not initialized; expected " + fieldView + ", found " + (!Object.hasOwn(object, name) ? "missing" : "uninitialized") : "read before assignment: field '" + name + "' in " + expression);
    return object[name];
};
const adamicViewTypeNames = Object.freeze({[adamicTags.Number]: "number", [adamicTags.Boolean]: "boolean", [adamicTags.String]: "string", [adamicTags.Object]: "object", [adamicTags.Record]: "object", [adamicTags.Array]: "array", [adamicTags.Map]: "Map"});
const adamicViewField = (object, name, expression, type, expected = adamicViewTypeNames[type], allowed = [], absent = false, optional = false, undefinedMember = false) => {
    const value = adamicReadField(object, name, expression, optional, absent, expected);
    if (value === undefined && (absent || undefinedMember || type === adamicTags.MaybeNumber || type === adamicTags.MaybeBoolean || optional && (object === undefined || object === null))) return undefined;
    const valid = (type === adamicTags.Number || type === adamicTags.MaybeNumber) ? typeof value === "number" : (type === adamicTags.Boolean || type === adamicTags.MaybeBoolean) ? typeof value === "boolean" : type === adamicTags.String ? typeof value === "string" : (type === adamicTags.Object || type === adamicTags.Record) ? value !== null && typeof value === "object" && !Array.isArray(value) && !(value instanceof Map) : type === adamicTags.Array ? Array.isArray(value) : type === adamicTags.Map ? value instanceof Map : type === adamicTags.Closure ? value instanceof AdamicClosure || typeof value === "function" : false;
    if (!valid) panic("field read failed: " + expression + " is not a " + expected + "; expected " + expected + ", found " + (value === undefined ? "nullish" : value === null ? (absent || undefinedMember ? "null" : "nullish") : Array.isArray(value) ? "array" : value instanceof Map ? "Map" : typeof value));
    if (allowed.length && !allowed.includes(value)) panic("field read failed: " + expression + " expected " + expected + ", found " + typeof value + " " + value);
    return value;
};
const adamicLogicalKind = value => value === undefined ? adamicTags.UndefinedRepresentation : value === null ? adamicTags.NullRepresentation : typeof value === "number" ? adamicTags.Number : typeof value === "boolean" ? adamicTags.Boolean : typeof value === "string" ? adamicTags.String : value instanceof AdamicClosure ? adamicTags.Closure : Array.isArray(value) ? adamicTags.Array : value instanceof Map ? adamicTags.Map : typeof value === "object" ? adamicTags.Object : 0;
const adamicNarrow = (value, wanted) => {
 const kind=adamicLogicalKind(value);
 if (wanted === adamicTags.Union || kind === wanted || (wanted === adamicTags.Record && kind === adamicTags.Object) || (wanted === adamicTags.MaybeNumber && (kind === adamicTags.Number || kind === adamicTags.UndefinedRepresentation)) || (wanted === adamicTags.MaybeBoolean && (kind === adamicTags.Boolean || kind === adamicTags.UndefinedRepresentation))) return value;
 panic("a union value does not match its narrowed type");
};
const adamicViewNullish = (object, name, expression, expected, kinds, nullAllowed, undefinedAllowed, allowed, absent, optional) => {
 const value=adamicReadField(object,name,expression,optional,absent,expected);
 if (value === undefined && (optional && (object === undefined || object === null) || absent && !Object.hasOwn(object,name))) return undefined;
 const kind=adamicLogicalKind(value);
 const valid=kind === adamicTags.NullRepresentation ? nullAllowed : kind === adamicTags.UndefinedRepresentation ? undefinedAllowed : (kinds & (1 << kind)) !== 0 && (!allowed.length || allowed.includes(value));
 if (!valid) panic("field read failed: " + expression + " matches no member of " + expected + "; expected " + expected + ", found " + (kind === adamicTags.NullRepresentation ? "null" : kind === adamicTags.UndefinedRepresentation ? "undefined" : kind === adamicTags.Array ? "array" : kind === adamicTags.Map ? "Map" : kind === adamicTags.Closure ? "function" : kind === 0 ? "unsupported representation" : typeof value));
 return value;
};
const adamicCheckedViewCast = (object, field, type, allowed, message) => allowed.includes(adamicViewField(object, field, field, type)) ? object : panic(message);
const adamicDefineField = (object, name, value, enumerable, ready, type) => { if (type !== undefined) adamicRecordFieldTypes(object, {[name]: type}); Object.defineProperty(object, name, {value, writable: true, enumerable, configurable: true}); if (ready) adamicFieldReadiness.get(object)?.delete(name); else { let fields = adamicFieldReadiness.get(object); if (!fields) adamicFieldReadiness.set(object, fields = new Set()); fields.add(name); } };
const adamicSpreadFields = (object, expression) => { const result = {}; if (object !== undefined && object !== null) for (const name of Object.keys(object)) Object.defineProperty(result, name, {value: adamicReadField(object, name, expression), enumerable: true, writable: true, configurable: true}); return adamicRecordSlotContracts(adamicRecordFieldTypes(result, adamicFieldRepresentations.get(object) || {}), {...adamicSlotContracts.get(object)}, "record"); };
const adamicObjectReadCall = (method, expression, target, ...sources) => {
    if (method === "assign") { for (const source of sources) for (const name of Object.keys(source)) adamicWriteField(target, name, adamicReadField(source, name, expression)); return target; }
    return Object.keys(target).map(name => { const value = adamicReadField(target, name, expression); return method === "entries" ? [name, value] : value; });
};
const adamicWriteField = (object, name, value) => { object[name] = value; adamicFieldReadiness.get(object)?.delete(name); };
`

func (e *emitter) localReady(local int) string {
	if e.program.Locals[local].Captured && !e.program.Locals[local].Global {
		if e.function != nil {
			for index, captured := range e.function.Environment {
				if captured == local {
					return fmt.Sprintf("self.cells[%d].ready", index)
				}
			}
		}
		return e.cellName(local) + ".ready"
	}
	return readyName(local)
}

const mapCertificateRuntime = `
const adamicMapCertificates = new WeakMap();
function adamicMapProducer(map,key,value,name){adamicMapCertificates.set(map,[key,value,name || "uncertified Map"]);return map;}
function adamicMapView(map,pairs,where,expected){if(map==null)return map;const c=adamicMapCertificates.get(map);if(c&&c[0]&&c[1]&&pairs.some(p=>p[0]===c[0]&&p[1]===c[1]))return map;panic('Map contract failed: '+where+'; expected '+expected+', found '+(c?c[2]:'uncertified Map'));}
`

// Emit representation numbers from the IR table, including semantic slot tags.
func representationTagsRuntime() string {
	return fmt.Sprintf("const adamicTags = Object.freeze({Number: %d, Boolean: %d, String: %d, Object: %d, Array: %d, Map: %d, MaybeNumber: %d, Closure: %d, MaybeBoolean: %d, Union: %d, Weak: %d, NullRepresentation: %d, UndefinedRepresentation: %d, Record: %d, Uint8Array: %d, Int32Array: %d, Float64Array: %d});\n",
		ir.Number,
		ir.Boolean,
		ir.String,
		ir.Object,
		ir.Array,
		ir.Map,
		ir.MaybeNumber,
		ir.Closure,
		ir.MaybeBoolean,
		ir.Union,
		ir.Weak,
		ir.NullRepresentation,
		ir.UndefinedRepresentation,
		ir.Record,
		ir.Uint8Array,
		ir.Int32Array,
		ir.Float64Array,
	)
}
