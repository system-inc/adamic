# Object unit 3 claim

Base: codex/library-object-prototype at 55302e5a87cf189f0f77df5031684155cc355572.
Runner: origin/codex/test262-ts-validity at af128990588aab7ee52ea3ab8527d638009c4979,
run from an isolated archive with `-root /workspace/adamic -adapt -json built-ins/Object`.
Stock TypeScript 6.0.3, es2024 and the runner's strict options and prelude.
Pinned test262: 5992dc3b60faf62a48fd6be8a40ae9d9a8c84d81.

| Before | pass | disagreement | refused | not-typescript | crashed | skipped |
|---|---:|---:|---:|---:|---:|---:|
| built-ins/Object | 58 | 0 | 598 | 1679 | 0 | 1076 |

The older 638/425 classification predates 40 refusal closures. The current
first blockers split into **359 library** and **239 language/representation**.
These are first blockers, not promised new passes. A library call can conceal
an accessor, array-hole or unproven-any dependency later in the same program.
The runner checks stock tsc for checker refusals; not-typescript is excluded
from work. A scratch capture copy is collecting per-test sources and verdicts;
the unmodified runner's completed baseline is the table above.

## Build order and claim

1. Object.defineProperty (186 first blockers): scalar data descriptors on
   proven plain objects, descriptor flags and SameValue redefinition checks,
   explicit refusal for accessors or unproven writes, catchable library TypeError
   through the existing exception IR, and observable enumeration/integrity.
2. Remaining Object.seal receivers (37), with isExtensible/isSealed/isFrozen
   agreement: expand only representations whose own descriptors can be proved.
   Primitive/plain-object queries are already built; there are no remaining
   first blockers named Object.isExtensible or Object.isSealed in this baseline.
3. Object.defineProperties (25): data-descriptor maps under the same proofs,
   preserve conversion-before-application and property order.
4. Evaluate the smaller Object families below after those changes expose their
   next blockers. Do not implement language features or approximate dynamic
   prototype, accessor, array-hole, unproven result, or host-object behavior.

New lowering belongs in library_object*.go, runtime in object*.c/.h. Shared
hooks will be limited to descriptor dispatch, ownership teardown and any
necessary reflection checks, named in the final report. New Adamic fixtures
are .a. No shared runner edits. Claim is committed and pushed before code.

## Library first blockers

| Feature | Tests |
|---|---:|
| Object.defineProperty | 186 |
| Object.seal | 37 |
| Object.defineProperties | 25 |
| boxed/builtin constructors | 15 |
| Object.getOwnPropertyDescriptor | 11 |
| Number builtin value | 6 |
| Object.getOwnPropertyNames | 5 |
| Object.create | 5 |
| Object.freeze | 4 |
| Array builtin value | 3 |
| JSON builtin value | 3 |
| Object builtin value | 3 |
| Object.keys | 3 |
| Object.values | 3 |
| String builtin value | 3 |
| Boolean builtin value | 3 |
| Date builtin value | 3 |
| Error builtin value | 3 |
| Math builtin value | 3 |
| RegExp builtin value | 3 |
| Object.prototype.isPrototypeOf | 3 |
| Object.entries | 2 |
| EvalError builtin value | 2 |
| Function builtin value | 2 |
| RangeError builtin value | 2 |
| ReferenceError builtin value | 2 |
| SyntaxError builtin value | 2 |
| TypeError builtin value | 2 |
| URIError builtin value | 2 |
| Object.fromEntries | 2 |
| Object.setPrototypeOf | 2 |
| Object.assign | 2 |
| Object.groupBy | 1 |
| Object.hasOwnProperty | 1 |
| Object.preventExtensions | 1 |
| Object.propertyIsEnumerable | 1 |
| Object.getOwnPropertyDescriptors | 1 |
| Object.getPrototypeOf | 1 |
| Object.hasOwn | 1 |

## Language handoff for @system_adamic

Each row names a first refusal and a one-line reproducer; these are not built in this unit.

| First blocker | Tests | One-line reproducer |
|---|---:|---|
| `not yet: a value of type any` | 99 | `const descriptor = Object.getOwnPropertyDescriptor({}, "x"); if (descriptor !== undefined) console.log(descriptor.value);` |
| `refuses inherited library member prototype read as an own field` | 25 | `const prototype = Number.prototype;` |
| `refuses a method read as a value (toString would lose its object, and this with it)` | 14 | `const value = {}; const detached = value.toString;` |
| `refuses a method read as a value (valueOf would lose its object, and this with it)` | 8 | `const value = {}; const detached = value.valueOf;` |
| `refuses a method read as a value (propertyIsEnumerable would lose its object, and this with it)` | 7 | `const value = {}; const detached = value.propertyIsEnumerable;` |
| `not yet: an OmittedExpression in an array literal` | 6 | `const values = [1, , 3];` |
| `refuses a method read as a value (hasOwnProperty would lose its object, and this with it)` | 6 | `const value = {}; const detached = value.hasOwnProperty;` |
| `refuses a method read as a value (toLocaleString would lose its object, and this with it)` | 6 | `const value = {}; const detached = value.toLocaleString;` |
| `refuses arguments` | 6 | `function count(): number { return arguments.length; }` |
| `refuses delete` | 4 | `const object: { value?: number } = { value: 1 }; delete object.value;` |
| `not yet: for...in over an array (holes and own enumerable properties are not represented; use for...of for elements)` | 3 | `for (const key in [1, 2]) console.log(key);` |
| `not yet: this outside a method` | 3 | `console.log(this);` |
| `refuses var` | 3 | `var value = 1;` |
| `not yet: an array of never` | 2 | `const values = [];` |
| `not yet: instanceof against a value that isn't a declared class` | 2 | `const value = {}; console.log(`${value instanceof Object}`);` |
| `not yet: toString through an object view (a value with a different native representation may be hidden by the view)` | 2 | `function show(value: {}): string { return value.toString(); } console.log(show(42));` |
| `refuses a method read as a value (create would lose its object, and this with it)` | 2 | `const detached = Object.create;` |
| `refuses a method read as a value (defineProperties would lose its object, and this with it)` | 2 | `const detached = Object.defineProperties;` |
| `refuses a method read as a value (defineProperty would lose its object, and this with it)` | 2 | `const detached = Object.defineProperty;` |
| `refuses a method read as a value (freeze would lose its object, and this with it)` | 2 | `const detached = Object.freeze;` |
| `refuses a method read as a value (getOwnPropertyNames would lose its object, and this with it)` | 2 | `const detached = Object.getOwnPropertyNames;` |
| `refuses a method read as a value (getPrototypeOf would lose its object, and this with it)` | 2 | `const detached = Object.getPrototypeOf;` |
| `refuses a method read as a value (hasOwn would lose its object, and this with it)` | 2 | `const detached = Object.hasOwn;` |
| `refuses a method read as a value (isExtensible would lose its object, and this with it)` | 2 | `const detached = Object.isExtensible;` |
| `refuses a method read as a value (isFrozen would lose its object, and this with it)` | 2 | `const detached = Object.isFrozen;` |
| `refuses a method read as a value (isSealed would lose its object, and this with it)` | 2 | `const detached = Object.isSealed;` |
| `refuses a method read as a value (keys would lose its object, and this with it)` | 2 | `const detached = Object.keys;` |
| `refuses a method read as a value (preventExtensions would lose its object, and this with it)` | 2 | `const detached = Object.preventExtensions;` |
| `refuses a value of type { enumerable: boolean` | 2 | `const descriptors: { [key: string]: PropertyDescriptor } = { x: { enumerable: true } };` |
| `refuses inherited library member isArray read as an own field` | 2 | `const { isArray } = Array;` |
| `not yet: a value of type "test" & { a: number; }` | 1 | `const value = Object.assign(1, { a: 1 });` |
| `not yet: a value of type 1 & { a: number; }` | 1 | `const value = Object.assign(1, { a: 1 });` |
| `not yet: a value of type Object & "123"` | 1 | `const value = Object.assign(1, { a: 1 });` |
| `not yet: a value of type boolean & { [x: string]: PropertyDescriptor; }` | 1 | `const value = Object.assign(1, { a: 1 });` |
| `not yet: a value of type never` | 1 | `function fail(): never { throw new Error("no"); } const value = fail();` |
| `not yet: a value of type true & { a: number; }` | 1 | `const value = Object.assign(1, { a: 1 });` |
| `not yet: a value of type { a: number; } & { a: number; } & { a: string; }` | 1 | `const value = Object.assign(1, { a: 1 });` |
| `not yet: assigning a field of a value` | 1 | `let value: { n: number } / undefined = { n: 1 }; value.n = 2;` |
| `not yet: hasOwnProperty through an object view (a value with a different native representation may be hidden by the view)` | 1 | `function show(value: {}): string { return value.toString(); } console.log(show(42));` |
| `not yet: hasOwnProperty with a key that isn't a string` | 1 | `console.log(`${({ value: 1 }).hasOwnProperty(1)}`);` |
| `refuses a method read as a value (getOwnPropertyDescriptor would lose its object, and this with it)` | 1 | `const detached = Object.getOwnPropertyDescriptor;` |
| `refuses a method read as a value (seal would lose its object, and this with it)` | 1 | `const detached = Object.seal;` |
| `refuses a value of type { get: () => number` | 1 | `const descriptors: { [key: string]: PropertyDescriptor } = { x: { enumerable: true } };` |
| `refuses a value of type { set: () => void` | 1 | `const descriptors: { [key: string]: PropertyDescriptor } = { x: { enumerable: true } };` |
| `refuses inherited library member constructor read as an own field` | 1 | `const value = {}; const constructor = value.constructor;` |
