package javascript

import "fmt"

// Hidden readiness state leaves own keys and object spread unchanged.
const fieldReadinessRuntime = `const adamicFieldReadiness = new WeakMap();
const adamicUninitializedFields = (object, names) => { adamicFieldReadiness.set(object, new Set(names)); return object; };
const adamicReadField = (object, name, expression, optional = false, allowAbsent = false) => {
    if (optional && (object === undefined || object === null)) return undefined;
    if (!Object.hasOwn(object, name) && typeof object === 'function' && adamicClassIdentities.has(Object.getPrototypeOf(object))) return adamicReadField(Object.getPrototypeOf(object), name, expression, false, allowAbsent);
    if (allowAbsent && !Object.hasOwn(object, name)) return undefined;
    if (!Object.hasOwn(object, name) || adamicFieldReadiness.get(object)?.has(name)) panic("read before assignment: field '" + name + "' in " + expression);
    return object[name];
};
const adamicDefineField = (object, name, value, enumerable, ready) => { Object.defineProperty(object, name, {value, writable: true, enumerable, configurable: true}); if (ready) adamicFieldReadiness.get(object)?.delete(name); else { let fields = adamicFieldReadiness.get(object); if (!fields) adamicFieldReadiness.set(object, fields = new Set()); fields.add(name); } };
const adamicSpreadFields = (object, expression) => { const result = {}; if (object !== undefined && object !== null) for (const name of Object.keys(object)) Object.defineProperty(result, name, {value: adamicReadField(object, name, expression), enumerable: true, writable: true, configurable: true}); return result; };
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
