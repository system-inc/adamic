package javascript

// DictionaryRuntime is included by the shared indexed-read dispatcher once
// producer certificates and transitive result contracts are wired. No cast
// admission or source producer is enabled merely by including this helper.
func DictionaryRuntime() string { return viewDictionariesRuntime }

const viewDictionariesRuntime = `
const adamicViewDictionaryRead = (record, key, kinds, childContract, expression, declared) => {
 const fail = found => panic('cast failed: field read failed: ' + expression + '; expected ' + declared + ', found ' + found);
 if (record === null || typeof record !== 'object' || Array.isArray(record) || record instanceof Map) fail(record === null ? 'null' : Array.isArray(record) ? 'array' : record instanceof Map ? 'Map' : typeof record);
 // Match the existing own-key record contract. Never call a getter or inherited member.
 if (!Object.hasOwn(record, key) && key in Object.prototype) panic("record member '" + key + "' is missing; records hold own keys only");
 const slot = Object.getOwnPropertyDescriptor(record, key);
 if (slot !== undefined && !Object.hasOwn(slot, 'value')) fail('accessor');
 if (typeof adamicFieldReadiness !== 'undefined' && adamicFieldReadiness.get(record)?.has(key)) fail('uninitialized');
 const value = slot === undefined ? undefined : slot.value;
 const kind = value === undefined ? 'undefined' : value === null ? 'null' : Array.isArray(value) ? 'array' : value instanceof Map ? 'Map' : typeof adamicTypeOf === 'function' ? adamicTypeOf(value) : typeof value;
 if (!kinds.includes(kind)) fail(kind);
 const reference = ['object','array','Map','function'].includes(kind);
 if (reference && !childContract) fail('unsupported representation');
 return {value, contract: reference ? childContract : 0};
};
`
