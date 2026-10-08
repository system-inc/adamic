# Remaining refusal handoff

All 517 remaining programs were stock-tsc accepted in the baseline, with identical adapted sources after implementation. Counts identify first blockers; later dependencies can remain hidden. Full diagnostics and source hashes are in [after-classification.json](after-classification.json); exact programs are in [refused-sources.jsonl.gz](refused-sources.jsonl.gz).

## Library features

| First blocker | Tests |
|---|---:|
| Object.defineProperty | 46 |
| Object.seal | 35 |
| boxed/builtin constructors | 15 |
| Object.getOwnPropertyDescriptor | 11 |
| Object.defineProperties | 10 |
| Number builtin value | 6 |
| Object.create | 5 |
| Object.freeze | 4 |
| Object.prototype.isPrototypeOf | 3 |
| Object builtin value | 3 |
| RegExp builtin value | 3 |
| Error builtin value | 3 |
| JSON builtin value | 3 |
| Array builtin value | 3 |
| String builtin value | 3 |
| Boolean builtin value | 3 |
| Math builtin value | 3 |
| Date builtin value | 3 |
| Object.values | 3 |
| Object.assign | 2 |
| Object.entries | 2 |
| Object.fromEntries | 2 |
| Object.getOwnPropertyNames | 2 |
| EvalError builtin value | 2 |
| RangeError builtin value | 2 |
| ReferenceError builtin value | 2 |
| SyntaxError builtin value | 2 |
| TypeError builtin value | 2 |
| URIError builtin value | 2 |
| Function builtin value | 2 |
| Object.setPrototypeOf | 2 |
| Object.hasOwnProperty | 1 |
| Object.getOwnPropertyDescriptors | 1 |
| Object.keys | 1 |
| Object.getPrototypeOf | 1 |
| Object.groupBy | 1 |
| Object.preventExtensions | 1 |
| Object.propertyIsEnumerable | 1 |

## Language features for @system_adamic

These features were not implemented. Each row includes a one-line reproducer.

| First blocker | Tests | Reproducer |
|---|---:|---|
| not yet: a value of type any | 99 | <code>const descriptor = Object.getOwnPropertyDescriptor({}, "x"); if (descriptor !== undefined) console.log(descriptor.value);</code> |
| accessor descriptors | 47 | <code>const object = {}; Object.defineProperty(object, "x", { get: () =&gt; 1 });</code> |
| for...in after descriptor mutation | 32 | <code>const object = { x: 1 }; Object.defineProperty(object, "x", { enumerable: false }); for (const key in object) console.log(key);</code> |
| refuses inherited library member prototype read as an own field | 25 | <code>const prototype = Number.prototype;</code> |
| refuses a method read as a value (toString would lose its object, and this with it) | 14 | <code>const value = {}; const detached = value.toString;</code> |
| refuses a method read as a value (valueOf would lose its object, and this with it) | 8 | <code>const value = {}; const detached = value.valueOf;</code> |
| refuses a method read as a value (propertyIsEnumerable would lose its object, and this with it) | 7 | <code>const value = {}; const detached = value.propertyIsEnumerable;</code> |
| refuses a method read as a value (hasOwnProperty would lose its object, and this with it) | 6 | <code>const value = {}; const detached = value.hasOwnProperty;</code> |
| not yet: an OmittedExpression in an array literal | 6 | <code>const values = [1, , 3];</code> |
| refuses arguments | 6 | <code>function count(): number { return arguments.length; }</code> |
| refuses a method read as a value (toLocaleString would lose its object, and this with it) | 6 | <code>const value = {}; const detached = value.toLocaleString;</code> |
| refuses delete | 4 | <code>const object: { value?: number } = { value: 1 }; delete object.value;</code> |
| not yet: for...in over an array (holes and own enumerable properties are not represented; use for...of for elements) | 4 | <code>for (const key in [1, 2]) console.log(key);</code> |
| refuses var | 3 | <code>var value = 1;</code> |
| not yet: this outside a method | 3 | <code>console.log(this);</code> |
| refuses a method read as a value (create would lose its object, and this with it) | 2 | <code>const detached = Object.create;</code> |
| refuses a value of type { enumerable: boolean | 2 | <code>const descriptors: { [key: string]: PropertyDescriptor } = { x: { enumerable: true } };</code> |
| refuses a method read as a value (defineProperties would lose its object, and this with it) | 2 | <code>const detached = Object.defineProperties;</code> |
| refuses a method read as a value (defineProperty would lose its object, and this with it) | 2 | <code>const detached = Object.defineProperty;</code> |
| refuses a method read as a value (freeze would lose its object, and this with it) | 2 | <code>const detached = Object.freeze;</code> |
| refuses a method read as a value (getOwnPropertyNames would lose its object, and this with it) | 2 | <code>const detached = Object.getOwnPropertyNames;</code> |
| refuses inherited library member isArray read as an own field | 2 | <code>const { isArray } = Array;</code> |
| not yet: instanceof against a value that isn't a declared class | 2 | <code>const value = {}; console.log(`${value instanceof Object}`);</code> |
| not yet: hasOwnProperty with a key that isn't a string | 2 | <code>console.log(`${({ value: 1 }).hasOwnProperty(1)}`);</code> |
| refuses a method read as a value (getPrototypeOf would lose its object, and this with it) | 2 | <code>const detached = Object.getPrototypeOf;</code> |
| refuses a method read as a value (hasOwn would lose its object, and this with it) | 2 | <code>const detached = Object.hasOwn;</code> |
| not yet: an array of never | 2 | <code>const values = [];</code> |
| refuses a method read as a value (isExtensible would lose its object, and this with it) | 2 | <code>const detached = Object.isExtensible;</code> |
| refuses a method read as a value (isFrozen would lose its object, and this with it) | 2 | <code>const detached = Object.isFrozen;</code> |
| refuses a method read as a value (isSealed would lose its object, and this with it) | 2 | <code>const detached = Object.isSealed;</code> |
| refuses a method read as a value (keys would lose its object, and this with it) | 2 | <code>const detached = Object.keys;</code> |
| refuses a method read as a value (preventExtensions would lose its object, and this with it) | 2 | <code>const detached = Object.preventExtensions;</code> |
| not yet: toString through an object view (a value with a different native representation may be hidden by the view) | 2 | <code>function show(value: {}): string { return value.toString(); } console.log(show(42));</code> |
| not yet: a value of type { a: number; } & { a: number; } & { a: string; } | 1 | <code>const value = Object.assign({ a: 1 }, { a: 2 }, { a: "x" });</code> |
| not yet: a value of type never | 1 | <code>function fail(): never { throw new Error("no"); } const value = fail();</code> |
| not yet: a value of type Object & "123" | 1 | <code>const value = Object.assign(new Object(), "123");</code> |
| not yet: a value of type true & { a: number; } | 1 | <code>const value = Object.assign(true, { a: 1 });</code> |
| not yet: a value of type 1 & { a: number; } | 1 | <code>const value = Object.assign(1, { a: 1 });</code> |
| not yet: a value of type "test" & { a: number; } | 1 | <code>const value = Object.assign("test", { a: 1 });</code> |
| refuses a value of type { get: () => number | 1 | <code>const descriptors: { [key: string]: { get: () =&gt; number } } = { x: { get: () =&gt; 1 } };</code> |
| refuses a value of type { set: () => void | 1 | <code>const descriptors: { [key: string]: { set: () =&gt; void } } = { x: { set: () =&gt; {} } };</code> |
| refuses a method read as a value (getOwnPropertyDescriptor would lose its object, and this with it) | 1 | <code>const detached = Object.getOwnPropertyDescriptor;</code> |
| not yet: a value of type boolean & { [x: string]: PropertyDescriptor; } | 1 | <code>const descriptors: { [key: string]: PropertyDescriptor } = { x: { value: 1 } }; const value = Object.assign(true as boolean, descriptors);</code> |
| refuses inherited library member constructor read as an own field | 1 | <code>const value = {}; const constructor = value.constructor;</code> |
| not yet: assigning a field of a value | 1 | <code>function make(): { x: number } { return { x: 1 }; } make().x = 2;</code> |
| refuses a method read as a value (seal would lose its object, and this with it) | 1 | <code>const detached = Object.seal;</code> |
| not yet: a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions | 1 | <code>const map = new Map&lt;unknown, number&gt;();</code> |
| not yet: a Set of unknown (a Set holds strings, numbers, booleans, objects, arrays, maps or functions so far) | 1 | <code>const set = new Set&lt;unknown&gt;();</code> |

Additional representation dependency exposed by library guards: `const values = [1, 2]; Object.defineProperty(values, "length", { value: 1, writable: false });` needs descriptor-aware array presence and length. Plain-object shape proof refusals also include host objects, functions, boxed receivers, and dynamic keys; this unit does not approximate them.
