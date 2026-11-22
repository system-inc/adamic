package javascript

import "fmt"

// Hidden readiness state leaves own keys and object spread unchanged.
const fieldReadinessRuntime = `const adamicTupleObjects = new WeakSet();
const adamicTuple = (values) => { adamicTupleObjects.add(values); return values; };
const adamicViewObject = (value) => value !== null && typeof value === "object" && (!Array.isArray(value) || adamicTupleObjects.has(value)) && !(value instanceof Map);
// A class constructor: a function here, an object natively. A view admits one only where its object's contract says StaticsAdmitted.
const adamicViewConstructor = (value) => typeof value === "function" && typeof adamicClassIdentities !== "undefined" && typeof adamicClasses !== "undefined" && adamicClasses[adamicClassIdentities.get(value) - 1]?.static === true;
// adamicViewStaticField is the descriptor of a class constructor's data static named name, walked as Node walks its
// prototype chain, a base's included, as native's static_own_slot walks static_parent; undefined where it's absent or
// refused. An accessor anywhere on the class chain is refused. The first constructor holding the name decides: an
// enumerable data static a class wrote is found, and anything else is refused, the emitted function's own name,
// length and prototype among them, which are never enumerable, so a builtin a base declares is hidden as on Node.
const adamicViewStaticField = (value, name) => {
    if (!adamicViewConstructor(value) || adamicViewStaticAccessor(value, name)) return undefined;
    for (let holder = value; adamicViewConstructor(holder); holder = Object.getPrototypeOf(holder)) {
        const field = Object.getOwnPropertyDescriptor(holder, name);
        if (field === undefined) continue;
        return Object.hasOwn(field, "value") && field.enumerable && !adamicFieldReadiness.get(holder)?.has(name) ? field : undefined;
    }
};
// adamicViewStaticAccessor is whether a class constructor's class, or a base's, declares a static accessor named name. Its
// accessors dispatch through adamicClasses, not a property, so neither a descriptor nor the in operator sees one.
const adamicViewStaticAccessor = (value, name) => { for (let id = adamicClassIdentities.get(value); id; id = adamicClasses[id - 1].base) if (Object.hasOwn(adamicClasses[id - 1].accessors, name)) return true; return false; };
// adamicViewStaticAbsent is whether a class constructor holds nothing at all under name, so an optional field reads undefined
// as Node does: not an accessor, and not a name the in operator finds, which a static a class holds is, and so is a property Function's
// prototype gives every function (call, apply and the rest), which adamicViewStaticField never finds.
const adamicViewStaticAbsent = (value, name) => adamicViewConstructor(value) && !adamicViewStaticAccessor(value, name) && !(name in value);
// adamicViewStatics is whether value is a class constructor holding each of names as a data static, its own or a base's,
// and each of optional as one or nothing: never an accessor, never a builtin only a base declares.
const adamicViewStatics = (value, names, optional = []) => adamicViewConstructor(value) && names.every((name) => adamicViewStaticField(value, name) !== undefined) && optional.every((name) => adamicViewStaticField(value, name) !== undefined || adamicViewStaticAbsent(value, name));
const adamicFieldReadiness = new WeakMap();
const adamicFieldRepresentations = new WeakMap();
const adamicRecordFieldTypes = (object, types) => { adamicFieldRepresentations.set(object, {...adamicFieldRepresentations.get(object), ...types}); return object; };
const adamicViewWrite = (object, name, value, type, declared) => { if (adamicFieldRepresentations.get(object)?.[name] !== type && !(adamicFieldRepresentations.get(object)?.[name] === 13 && ([3,4,5,6,8,10].includes(type))) && !(adamicFieldRepresentations.get(object)?.[name] === 10 && type <= 2) && !(adamicFieldRepresentations.get(object)?.[name] === 7 && type === 1)) adamicViewField(object, name, "<write>." + name, type, declared || adamicViewTypeNames[type] || "unknown"); adamicWriteField(object, name, value); adamicRecordFieldTypes(object, {[name]: adamicFieldRepresentations.get(object)?.[name] === 10 && type <= 2 ? 10 : type}); };
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
    const valid = type === 1 ? typeof value === "number" : type === 2 ? typeof value === "boolean" : type === 3 ? typeof value === "string" : type === 4 || type === 14 ? value !== null && typeof value === "object" && !Array.isArray(value) && !(value instanceof Map) && !ArrayBuffer.isView(value) && adamicRecords.has(value) === (type === 14) : type === 5 ? Array.isArray(value) : type === 6 ? value instanceof Map : type === 8 ? typeof value === "function" : false;
    if (!valid) panic(message);
    return value;
};
const adamicViewTypeNames = Object.freeze({1: "number", 2: "boolean", 3: "string", 4: "object", 5: "array", 6: "Map"});
const adamicReceiverKind = (value) => typeof value !== "object" ? typeof value : value instanceof AdamicClosure ? "function" : Array.isArray(value) ? "array" : value instanceof Map ? "Map" : "object";
// A checked read's receiver as native reads one by name. A closure is an AdamicClosure here, an object with an own
// length, so it's stopped as the function native says it is. Two kinds pass here that native names instead, a
// generator object and a typed array: each stops at the by-name read as found missing, so only the wording differs.
const adamicViewReceiver = (object) => object === undefined || object === null || !(object instanceof AdamicClosure) && adamicViewObject(object) || typeof object === "function" && typeof adamicClassIdentities !== "undefined" && (adamicClassIdentities.has(object) || adamicClassIdentities.has(Object.getPrototypeOf(object)));
// statics names the fields an object read's contract requires where it admits a class constructor (StaticsAdmitted, #s03ndgw), and
// optionalStatics the ones it leaves optional: one holding each required one as a data static, and no optional one the walk refuses,
// is the object read, as native's adamic_view_static_read admits it.
const adamicViewField = (object, name, expression, type, expected = adamicViewTypeNames[type], allowed = [], statics = undefined, optionalStatics = []) => {
    // The receiver's own kind comes first, as native's adamic_object_view reads it: a string has own length and index
    // keys, so a lookup by name would read on where native stops. A class is a function here and an object there.
    if (!adamicViewReceiver(object)) panic("field read failed: " + expression + " is not a " + expected + "; expected " + expected + ", found " + adamicReceiverKind(object));
    const value = adamicReadField(object, name, expression, false, false, expected);
    const valid = type === 1 ? typeof value === "number" : type === 2 ? typeof value === "boolean" : type === 3 ? typeof value === "string" : type === 4 ? adamicViewObject(value) || statics !== undefined && adamicViewStatics(value, statics, optionalStatics) : type === 5 ? Array.isArray(value) : type === 6 ? value instanceof Map : false;
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
