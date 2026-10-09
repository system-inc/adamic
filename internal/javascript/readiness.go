package javascript

import "fmt"

// Hidden readiness state leaves own keys and object spread unchanged.
const fieldReadinessRuntime = `const adamicNamespaces = new WeakMap();
const adamicNamespace = () => { const object = {}; adamicNamespaces.set(object, new Set()); return object; };
const adamicNamespaceInstall = (object, name, value, readonly) => { object[name] = value; adamicRecordFieldTypes(object, {[name]: 10}); if (readonly) adamicNamespaces.get(object).add(name); };
const adamicNamespaceCheckWrite = (object, name) => { if (adamicNamespaces.get(object)?.has(name)) panic("namespace write failed: " + name + " is readonly"); };
const adamicNamespaceRead = (object, name, type) => { const value = object[name]; if (type === 10 || (type === 7 || type === 9 || type >= 3 && type <= 6 || type === 8) && value === undefined || (type === 1 || type === 7) && typeof value === "number" || (type === 2 || type === 9) && typeof value === "boolean" || type === 3 && typeof value === "string" || type === 4 && value !== null && typeof value === "object" && !Array.isArray(value) && !(value instanceof Map) || type === 5 && Array.isArray(value) || type === 6 && value instanceof Map || type === 8 && value instanceof AdamicClosure) return value; panic("namespace read failed: " + name + " has an unexpected runtime type"); };
const adamicFieldReadiness = new WeakMap();
const adamicFieldRepresentations = new WeakMap();
const adamicRecordFieldTypes = (object, types) => { adamicFieldRepresentations.set(object, {...adamicFieldRepresentations.get(object), ...types}); return object; };
const adamicViewWrite = (object, name, value, type) => { if (!Object.hasOwn(object, name) || adamicFieldRepresentations.get(object)?.[name] !== type && !(adamicFieldRepresentations.get(object)?.[name] === 13 && ([3,4,5,6,8,10].includes(type))) && !(adamicFieldRepresentations.get(object)?.[name] === 10 && type <= 2) && !(adamicFieldRepresentations.get(object)?.[name] === 7 && type === 1)) adamicViewField(object, name, "<write>." + name, type); adamicWriteField(object, name, value); adamicRecordFieldTypes(object, {[name]: adamicFieldRepresentations.get(object)?.[name] === 10 && type <= 2 ? 10 : type}); };
const adamicUninitializedFields = (object, names) => { adamicFieldReadiness.set(object, new Set(names)); return object; };
const adamicReadField = (object, name, expression, optional = false, allowAbsent = false, fieldView = false) => {
    if (optional && (object === undefined || object === null)) return undefined;
    if (object === undefined || object === null) panic(fieldView ? "field read failed: " + expression + " is not initialized; expected " + fieldView + ", found missing" : "read before assignment: field '" + name + "' in " + expression);
    if (!Object.hasOwn(object, name) && typeof object === 'function' && adamicClassIdentities.has(Object.getPrototypeOf(object))) return adamicReadField(Object.getPrototypeOf(object), name, expression, false, allowAbsent, fieldView);
    if (allowAbsent && !Object.hasOwn(object, name)) return undefined;
    if (!Object.hasOwn(object, name) || adamicFieldReadiness.get(object)?.has(name)) panic(fieldView ? "field read failed: " + expression + " is not initialized; expected " + fieldView + ", found " + (!Object.hasOwn(object, name) ? "missing" : "uninitialized") : "read before assignment: field '" + name + "' in " + expression);
    return object[name];
};
const adamicPlaceholderNarrow = (value, type, message) => {
    const valid = type === 1 ? typeof value === "number" : type === 2 ? typeof value === "boolean" : type === 3 ? typeof value === "string" : type === 4 ? value !== null && typeof value === "object" && !Array.isArray(value) && !(value instanceof Map) && !ArrayBuffer.isView(value) : type === 5 ? Array.isArray(value) : type === 6 ? value instanceof Map : type === 8 ? typeof value === "function" : false;
    if (!valid) panic(message);
    return value;
};
const adamicViewTypeNames = Object.freeze({1: "number", 2: "boolean", 3: "string", 4: "object", 5: "array", 6: "Map"});
const adamicViewField = (object, name, expression, type, expected = adamicViewTypeNames[type], allowed = []) => {
    const value = adamicReadField(object, name, expression, false, false, expected);
    const valid = type === 1 ? typeof value === "number" : type === 2 ? typeof value === "boolean" : type === 3 ? typeof value === "string" : type === 4 ? value !== null && typeof value === "object" && !Array.isArray(value) && !(value instanceof Map) : type === 5 ? Array.isArray(value) : type === 6 ? value instanceof Map : type === 8 ? value instanceof AdamicClosure : false;
    if (!valid) panic("field read failed: " + expression + " is not a " + expected + "; expected " + expected + ", found " + (value === undefined ? "nullish" : value === null ? "nullish" : Array.isArray(value) ? "array" : value instanceof Map ? "Map" : typeof value));
    if (allowed.length && !allowed.includes(value)) panic("field read failed: " + expression + " expected " + expected + ", found " + typeof value + " " + value);
    return value;
};
const adamicPlaceholderViewField = (object, name, expression, type, expected, allowed, optional, absent) => {
    if (optional && (object === undefined || object === null)) return undefined;
    if (object !== undefined && object !== null) {
        if (!Object.hasOwn(object, name) && typeof object === 'function' && adamicClassIdentities.has(Object.getPrototypeOf(object))) return adamicPlaceholderViewField(Object.getPrototypeOf(object), name, expression, type, expected, allowed, false, absent);
        if (absent && !Object.hasOwn(object, name)) return undefined;
        if (Object.hasOwn(object, name) && (object[name] === undefined || object[name] === null)) return object[name];
    }
    return adamicViewField(object, name, expression, type, expected, allowed);
};
const adamicCheckedViewCast = (object, field, type, allowed, message) => allowed.includes(adamicViewField(object, field, field, type)) ? object : panic(message);
const adamicDefineField = (object, name, value, enumerable, ready, type) => { if (type !== undefined) adamicRecordFieldTypes(object, {[name]: type}); Object.defineProperty(object, name, {value, writable: true, enumerable, configurable: true}); if (ready) adamicFieldReadiness.get(object)?.delete(name); else { let fields = adamicFieldReadiness.get(object); if (!fields) adamicFieldReadiness.set(object, fields = new Set()); fields.add(name); } };
const adamicSpreadFields = (object, expression) => { const result = {}; if (object !== undefined && object !== null) for (const name of Object.keys(object)) Object.defineProperty(result, name, {value: adamicReadField(object, name, expression), enumerable: true, writable: true, configurable: true}); return adamicRecordFieldTypes(result, adamicFieldRepresentations.get(object) || {}); };
const adamicWriteField = (object, name, value) => { adamicNamespaceCheckWrite(object, name); object[name] = value; adamicFieldReadiness.get(object)?.delete(name); };
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
