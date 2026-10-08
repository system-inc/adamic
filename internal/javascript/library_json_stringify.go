package javascript

// Normalize only after all three arguments have been evaluated. Copying an array earlier would
// hide mutations made by the space argument. A closure's C-like backend wrapper is represented as
// an actual function for JSON, which omits it from objects and writes null in arrays.
const jsonStringifyRuntime = `class AdamicJSONParsed { constructor(text) { this.adamicJSONParsed = text; } }
const adamicJSONParse = (text, callback, ignored, mode) => {
 const value = JSON.parse(text, callback === undefined ? undefined : (key, value) => adamicCall(callback, [key, value]));
 if (mode === "discard") return 0;
 if (mode === "canonical") return new AdamicJSONParsed(JSON.stringify(value));
 return value;
};
const adamicJSONValue = (value) => {
 if (value instanceof AdamicJSONParsed) return value.adamicJSONParsed === undefined ? undefined : JSON.parse(value.adamicJSONParsed);
 if (value instanceof AdamicClosure) return () => undefined;
 if (Array.isArray(value)) return value.map(adamicJSONValue);
 if (value !== null && typeof value === 'object' && !(value instanceof Map) && !(value instanceof Set)) {
  const copy = {};
  for (const key of Object.keys(value)) copy[key] = key === "toJSON" && value[key] instanceof AdamicClosure ? (...args) => adamicJSONValue(adamicCall(value[key], args)) : adamicJSONValue(value[key]);
  return copy;
 }
 return value;
};
const adamicJSONStringify = (value, replacer, space) => JSON.stringify(adamicJSONValue(value), replacer instanceof AdamicClosure ? (key, value) => adamicCall(replacer, [key, value]) : adamicJSONValue(replacer), space);
`
