package javascript

import "fmt"

// Hidden readiness state leaves own keys and object spread unchanged.
const fieldReadinessRuntime = `const adamicFieldReadiness = new WeakMap();
const adamicFieldRepresentations = new WeakMap();
const adamicRecordFieldTypes = (object, types) => { adamicFieldRepresentations.set(object, {...adamicFieldRepresentations.get(object), ...types}); return object; };
const adamicViewWrite = (object, name, value, type) => { if (!Object.hasOwn(object, name) || adamicFieldRepresentations.get(object)?.[name] !== type && !(adamicFieldRepresentations.get(object)?.[name] === 10 && type <= 2) && !(adamicFieldRepresentations.get(object)?.[name] === 7 && type === 1)) adamicViewField(object, name, "<write>." + name, type); adamicWriteField(object, name, value); adamicRecordFieldTypes(object, {[name]: adamicFieldRepresentations.get(object)?.[name] === 10 && type <= 2 ? 10 : type}); };
const adamicUninitializedFields = (object, names) => { adamicFieldReadiness.set(object, new Set(names)); return object; };
const adamicReadField = (object, name, expression, optional = false, allowAbsent = false, fieldView = false) => {
    if (optional && (object === undefined || object === null)) return undefined;
    if (object === undefined || object === null) panic(fieldView ? "field read failed: " + expression + " is not initialized; expected " + fieldView + ", found missing" : "read before assignment: field '" + name + "' in " + expression);
    if (!Object.hasOwn(object, name) && typeof object === 'function' && adamicClassIdentities.has(Object.getPrototypeOf(object))) return adamicReadField(Object.getPrototypeOf(object), name, expression, false, allowAbsent, fieldView);
    if (allowAbsent && !Object.hasOwn(object, name)) return undefined;
    if (!Object.hasOwn(object, name) || adamicFieldReadiness.get(object)?.has(name)) panic(fieldView ? "field read failed: " + expression + " is not initialized; expected " + fieldView + ", found " + (!Object.hasOwn(object, name) ? "missing" : "uninitialized") : "read before assignment: field '" + name + "' in " + expression);
    return object[name];
};
const adamicViewTypeNames = Object.freeze({1: "number", 2: "boolean", 3: "string", 4: "object", 5: "array", 6: "Map"});
const adamicViewField = (object, name, expression, type, expected = adamicViewTypeNames[type], allowed = []) => {
    const value = adamicReadField(object, name, expression, false, false, expected);
    const valid = type === 1 ? typeof value === "number" : type === 2 ? typeof value === "boolean" : type === 3 ? typeof value === "string" : type === 4 ? value !== null && typeof value === "object" && !Array.isArray(value) && !(value instanceof Map) : type === 5 ? Array.isArray(value) : type === 6 ? value instanceof Map : false;
    if (!valid) panic("field read failed: " + expression + " is not a " + expected + "; expected " + expected + ", found " + (value === undefined ? "nullish" : value === null ? "nullish" : Array.isArray(value) ? "array" : value instanceof Map ? "Map" : typeof value));
    if (allowed.length && !allowed.includes(value)) panic("field read failed: " + expression + " expected " + expected + ", found " + typeof value + " " + value);
    return value;
};
const adamicObjectReadCall = (method, expression, target, ...sources) => {
    if (method === "optionalFunctionStorage") { if (!(target instanceof AdamicClosure) || typeof target.code !== "function") panic("optional contract lacks object-return closure storage at " + expression); return true; }
    if (method === "optionalArrayPresence") { if (target == null && sources[0]) return true; if (!Array.isArray(target)) panic("optional contract lacks object array storage at " + expression); for (const object of target) { adamicObjectReadCall("optionalViewStorage", expression, object, sources[2]); if (object != null) adamicObjectReadCall("optionalSpreadPresence", expression, object, sources[1]); } return true; }
    if (method === "optionalViewStorage") { if (target == null ? sources[0] : (typeof target === "object" || typeof target === "function")) return true; panic("optional view lacks own-presence storage at " + expression); }
    if (method === "optionalSpreadKeys") return target == null ? [] : Object.keys(target);
    if (method === "optionalSpreadPresence") { for (const name of sources[0]) if (!Object.hasOwn(target, name)) panic("optional write lost own presence: '" + name + "' at " + expression); return true; }
    if (method === "optionalWritePresence") { if (!Object.hasOwn(target, sources[0])) panic("optional write lost own presence: '" + sources[0] + "' at " + expression); return true; }
    if (method === "assign") { for (const source of sources) for (const name of Object.keys(source)) adamicWriteField(target, name, adamicReadField(source, name, expression)); return target; }
    return Object.keys(target).map(name => { const value = adamicReadField(target, name, expression); return method === "entries" ? [name, value] : value; });
};
const adamicCheckedViewCast = (object, field, type, allowed, message) => allowed.includes(adamicViewField(object, field, field, type)) ? object : panic(message);
const adamicDefineField = (object, name, value, enumerable, ready, type) => { if (type !== undefined) adamicRecordFieldTypes(object, {[name]: type}); Object.defineProperty(object, name, {value, writable: true, enumerable, configurable: true}); if (ready) adamicFieldReadiness.get(object)?.delete(name); else { let fields = adamicFieldReadiness.get(object); if (!fields) adamicFieldReadiness.set(object, fields = new Set()); fields.add(name); } };
const adamicSpreadFields = (object, expression) => { const result = {}; if (object !== undefined && object !== null) for (const name of Object.keys(object)) Object.defineProperty(result, name, {value: adamicReadField(object, name, expression), enumerable: true, writable: true, configurable: true}); return adamicRecordFieldTypes(result, adamicFieldRepresentations.get(object) || {}); };
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
