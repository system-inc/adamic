package javascript

// Errors are returned to the generated helper, whose existing exception IR
// raises TypeError on both backends and participates in cleanup and catches.
const objectDescriptorRuntime = `
const adamicObjectEnumerable = (object, key) => !key.startsWith('#') && Object.prototype.propertyIsEnumerable.call(object, key);
const adamicDefinePropertyError = (object, key, value, writable, enumerable, configurable, mask, representation) => {
 const descriptor = {};
 if (mask & 1) descriptor.value = value;
 if (mask & 2) descriptor.writable = !!writable;
 if (mask & 4) descriptor.enumerable = !!enumerable;
 if (mask & 8) descriptor.configurable = !!configurable;
 try { Object.defineProperty(object, key, descriptor); return undefined; }
 catch (error) { return error.message; }
};
`
