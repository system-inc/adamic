package javascript

// Normalize only after all three arguments have been evaluated. Copying an array earlier would
// hide mutations made by the space argument. A closure's C-like backend wrapper is represented as
// an actual function for JSON, which omits it from objects and writes null in arrays.
const jsonStringifyRuntime = `const adamicJSONValue = (value) => {
 if (value instanceof AdamicClosure) return () => undefined;
 if (Array.isArray(value)) return value.map(adamicJSONValue);
 if (value !== null && typeof value === 'object' && !(value instanceof Map) && !(value instanceof Set)) {
  const copy = {};
  for (const key of Object.keys(value)) copy[key] = adamicJSONValue(value[key]);
  return copy;
 }
 return value;
};
const adamicJSONStringify = (value, replacer, space) => JSON.stringify(adamicJSONValue(value), replacer, space);
`
