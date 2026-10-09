package javascript

// Normalize only after all three arguments have been evaluated. Copying an array earlier would
// hide mutations made by the space argument. A closure's C-like backend wrapper is represented as
// an actual function for JSON, which omits it from objects and writes null in arrays.
const jsonStringifyRuntime = `const adamicJSONStringify = (value, replacer, space) => {
 const views = new WeakMap();
 const view = (value) => {
  if (value instanceof AdamicClosure) return () => undefined;
  if (value === null || typeof value !== 'object' || value instanceof Map || value instanceof Set) return value;
  if (views.has(value)) return views.get(value);
  const proxy = new Proxy(value, {get(target, key) {
   if (key === 'toJSON' && typeof target[key] === 'function') {
    return (key) => view(target.toJSON(target, key));
   }
   return view(Reflect.get(target, key));
  }});
  views.set(value, proxy);
  return proxy;
 };
 return JSON.stringify(view(value), view(replacer), space);
};
`
