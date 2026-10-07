package javascript

// Hidden readiness state leaves own keys and object spread unchanged.
const fieldReadinessRuntime = `const adamicFieldReadiness = new WeakMap();
const adamicUninitializedFields = (object, names) => { adamicFieldReadiness.set(object, new Set(names)); return object; };
const adamicReadField = (object, name, expression, optional = false) => {
    if (optional && (object === undefined || object === null)) return undefined;
    if (!Object.hasOwn(object, name) || adamicFieldReadiness.get(object)?.has(name)) panic("read before assignment: field '" + name + "' in " + expression);
    return object[name];
};
const adamicViewTypeNames = Object.freeze({1: "number", 2: "boolean", 3: "string", 4: "object", 5: "array", 6: "Map"});
const adamicViewField = (object, name, expression, type) => {
    if (object === null || object === undefined || !Object.hasOwn(object, name) || adamicFieldReadiness.get(object)?.has(name)) panic("field read failed: " + expression + " is not initialized");
    const value = object[name];
    const valid = type === 1 ? typeof value === "number" : type === 2 ? typeof value === "boolean" : type === 3 ? typeof value === "string" : type === 4 ? value !== null && typeof value === "object" && !Array.isArray(value) && !(value instanceof Map) : type === 5 ? Array.isArray(value) : type === 6 ? value instanceof Map : false;
    if (!valid) panic("field read failed: " + expression + " is not a " + (adamicViewTypeNames[type] || "checkable value"));
    return value;
};
const adamicWriteField = (object, name, value) => { object[name] = value; adamicFieldReadiness.get(object)?.delete(name); };
`
