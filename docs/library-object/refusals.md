# Fresh Object refusal classification

Classification completed before implementation, from main `5d4c801`, with real `typescript@6.0.3` and test262 `5992dc3b60faf62a48fd6be8a40ae9d9a8c84d81`. Adaptation is enabled. The old 0-pass measurement predates existing Object support.

| Measurement | Pass | Disagreement | Refused | Crashed | Skipped | Total |
|---|---:|---:|---:|---:|---:|---:|
| Fresh before | 17 | 1 | 2317 | 0 | 1076 | 3411 |

## Every refusal checked

| Bucket | Count | Meaning |
|---|---:|---|
| a | 1679 | Real tsc rejects with the same TS diagnostic code |
| b | 388 | Real tsc accepts; first refusal names an Object library feature |
| c | 250 | Real tsc accepts; first refusal names a language or representation feature |

No tsc rejection had a different code. A match means the runner's first normalized TS code appears among real tsc's diagnostics, not that every diagnostic or its order is identical. Every refused adapted program was checked separately against `lib.es2024.d.ts` and the exact `internal/load/prelude.d.ts`, with all `compilerOptions()` from `internal/load/load.go`. The checker implementation is the npm TypeScript package, not typescript-go. Skipped tests are outside the refusal classification.

[classification.json](classification.json) lists every refused test, the full real-tsc diagnostics, feature, bucket and source hash. [adapted-sources.jsonl.gz](adapted-sources.jsonl.gz) preserves the runner's exact sources and per-test outcomes. [classify-tsc.cjs](classify-tsc.cjs) and [summarize-refusals.py](summarize-refusals.py) reproduce the tsc pass and tables. The scratch runner was a copy of `cmd/adamic-test262`, with only a capture after each result. The official runner's aggregate result and the capture run match. No runner source was edited.

Counts describe the first blocker observed. Closing it may expose another language or library blocker, so bucket (b) counts are candidate tests, not promised passes. Bucket (c) includes deliberately unsupported language features such as any, delete and detached methods, not just implementation work owed.

## Bucket (a) reasons

| Reason | Count |
|---|---:|
| `error TS18048` | 343 |
| `error TS2339` | 311 |
| `error TS2345` | 216 |
| `error TS2322` | 181 |
| `error TS7009` | 166 |
| `error TS7006` | 133 |
| `error TS7005` | 114 |
| `error TS2379` | 60 |
| `error TS2554` | 41 |
| `error TS2551` | 23 |
| `error TS2559` | 18 |
| `error TS2532` | 16 |
| `error TS7034` | 14 |
| `error TS2375` | 12 |
| `error TS2769` | 8 |
| `error TS7053` | 6 |
| `error TS2683` | 3 |
| `error TS2704` | 3 |
| `error TS2790` | 2 |
| `error TS2351` | 2 |
| `error TS2584` | 1 |
| `error TS2739` | 1 |
| `error TS2538` | 1 |
| `error TS2542` | 1 |
| `error TS4104` | 1 |
| `error TS2349` | 1 |
| `error TS2695` | 1 |

## Bucket (b) reasons

| Reason | Count |
|---|---:|
| `Object.defineProperty` | 185 |
| `Object.seal` | 38 |
| `Object.isExtensible` | 35 |
| `Object.isSealed` | 28 |
| `Object.defineProperties` | 24 |
| `Object.getOwnPropertyNames` | 16 |
| `Object.keys` | 14 |
| `Object.getOwnPropertyDescriptor` | 11 |
| `Object.preventExtensions` | 9 |
| `Object.freeze` | 6 |
| `Object.create` | 5 |
| `Object.values` | 3 |
| `Object.assign` | 2 |
| `Object.entries` | 2 |
| `Object.fromEntries` | 2 |
| `Object.hasOwn` | 2 |
| `Object.hasOwnProperty` | 1 |
| `Object.getOwnPropertyDescriptors` | 1 |
| `Object.getPrototypeOf` | 1 |
| `Object.groupBy` | 1 |
| `Object.propertyIsEnumerable` | 1 |
| `Object.setPrototypeOf` | 1 |

## Bucket (c) reasons

| Reason | Count |
|---|---:|
| `not yet: a value of type any` | 99 |
| `not yet: new an Identifier` | 15 |
| `refuses a method read as a value (toString would lose its object, and this with it)` | 14 |
| `refuses inherited library member prototype read as an own field` | 10 |
| `refuses a method read as a value (valueOf would lose its object, and this with it)` | 8 |
| `refuses a method read as a value (propertyIsEnumerable would lose its object, and this with it)` | 7 |
| `refuses a method read as a value (hasOwnProperty would lose its object, and this with it)` | 6 |
| `not yet: an OmittedExpression in an array literal` | 6 |
| `refuses arguments` | 6 |
| `refuses a method read as a value (toLocaleString would lose its object, and this with it)` | 6 |
| `refuses delete` | 4 |
| `refuses isPrototypeOf` | 3 |
| `refuses var` | 3 |
| `refuses a method read as a value (create would lose its object, and this with it)` | 2 |
| `refuses a value of type { enumerable: boolean` | 2 |
| `refuses a method read as a value (defineProperties would lose its object, and this with it)` | 2 |
| `refuses a method read as a value (defineProperty would lose its object, and this with it)` | 2 |
| `refuses a method read as a value (freeze would lose its object, and this with it)` | 2 |
| `refuses a method read as a value (getOwnPropertyNames would lose its object, and this with it)` | 2 |
| `refuses a method read as a value (getPrototypeOf would lose its object, and this with it)` | 2 |
| `refuses a method read as a value (hasOwn would lose its object, and this with it)` | 2 |
| `not yet: an array of never` | 2 |
| `not yet: Object as a value outside equality or typeof (overloaded calls and static properties need their own representation)` | 2 |
| `refuses a method read as a value (isExtensible would lose its object, and this with it)` | 2 |
| `refuses a method read as a value (isFrozen would lose its object, and this with it)` | 2 |
| `refuses a method read as a value (isSealed would lose its object, and this with it)` | 2 |
| `refuses a method read as a value (keys would lose its object, and this with it)` | 2 |
| `refuses a method read as a value (preventExtensions would lose its object, and this with it)` | 2 |
| `not yet: toString through an object view (a value with a different native representation may be hidden by the view)` | 2 |
| `not yet: a value of type { a: number; } & { a: number; } & { a: string; }` | 1 |
| `not yet: a value of type never` | 1 |
| `not yet: a value of type Object & "123"` | 1 |
| `not yet: a value of type true & { a: number; }` | 1 |
| `not yet: a value of type 1 & { a: number; }` | 1 |
| `not yet: a value of type "test" & { a: number; }` | 1 |
| `refuses a value of type { get: () => number` | 1 |
| `refuses a value of type { set: () => void` | 1 |
| `refuses a method read as a value (getOwnPropertyDescriptor would lose its object, and this with it)` | 1 |
| `not yet: a value of type boolean & { [x: string]: PropertyDescriptor; }` | 1 |
| `not yet: this outside a method` | 1 |
| `not yet: reading Boolean` | 1 |
| `not yet: Number as a value outside equality or typeof (overloaded calls and static properties need their own representation)` | 1 |
| `not yet: Number.prototype` | 1 |
| `not yet: reading Math` | 1 |
| `not yet: reading Date` | 1 |
| `not yet: reading RegExp` | 1 |
| `not yet: reading Error` | 1 |
| `not yet: reading EvalError` | 1 |
| `not yet: reading RangeError` | 1 |
| `not yet: reading ReferenceError` | 1 |
| `not yet: reading SyntaxError` | 1 |
| `not yet: reading TypeError` | 1 |
| `not yet: reading URIError` | 1 |
| `not yet: JSON as a value outside equality or typeof (overloaded calls and static properties need their own representation)` | 1 |
| `not yet: reading Function` | 1 |
| `not yet: Array as a value outside equality or typeof (overloaded calls and static properties need their own representation)` | 1 |
| `not yet: String as a value outside equality or typeof (overloaded calls and static properties need their own representation)` | 1 |
| `not yet: assigning a field of a value` | 1 |
| `not yet: hasOwnProperty through an object view (a value with a different native representation may be hidden by the view)` | 1 |
| `refuses a method read as a value (seal would lose its object, and this with it)` | 1 |

## Claim and build order

1. The largest library refusal, `Object.defineProperty` (185), and `defineProperties` (24) require observable descriptors, accessors and shape mutation. These remain explicit refusals: CLAUDE.md binds the approved fixed-shape memory and type contract, and docs/0.1.md forbids this mutation. Implementing them needs a representation and soundness design beyond this slice. Prototype mutation has the same constraint. Descriptor reads currently return any, adding a language blocker even if their first refusal were removed.
2. Build the largest remaining family: `Object.seal` (38), `Object.isExtensible` (35), `Object.isSealed` (28), then `Object.preventExtensions` (9). Port V8's integrity tests for represented fixed plain objects and its primitive fast paths. Unknown host constructors, arrays needing integrity mutation, descriptor/prototype mutation, and unrepresented values remain refused. Existing Object.freeze must interoperate with the integrity queries.
3. Investigate own-name reflection (16) and exact-shape keys (14) after integrity, retaining explicit refusals whenever the shape or presence is unproven.
4. Close the observed baseline disagreement `Object.is(undefined, null)`: the existing null/undefined representation cannot be silently compared as equal. Refuse ambiguous comparisons unless this slice can prove the exact result.

## Language handoff for @system_adamic

Each row is a representative adapted test body on one line. To reproduce, prepend the unchanged runner prelude from `cmd/adamic-test262/prelude.go`. The full per-test sources and diagnostics are in the linked artifacts; no language changes are claimed here.

| First language/representation refusal | Tests | One-line reproducer | Representative test |
|---|---:|---|---|
| not yet: a value of type any | 99 | <code>let target = { a: 1 }; let result = Object.assign(target, "1a2c3", { a: "c" }, undefined, { b: 6 }, null, 125, { a: 5 }); assertSameValue(Object.getOwnPropertyNames(result).length, 7, "The length should be 7 in the final object."); assertSameValue(result.a, 5, "The value should be {a:5}."); assertSameValue(result[0], "1", "The value should be {\"0\":\"1\"}."); assertSameValue(result[1], "a", "The value should be {\"1\":\"a\"}."); assertSameValue(result[2], "2", "The value should be {\"2\":\"2\"}."); assertSameValue(result[3], "c", "The value should be {\"3\":\"c\"}."); assertSameValue(result[4], "3", "The value should be {\"4\":\"3\"}."); assertSameValue(result.b, 6, "The value should be {b:6}.");</code> | `built-ins/Object/assign/Override.js` |
| not yet: new an Identifier | 15 | <code>let strObj = new String("a"); Object.freeze(strObj); assert(Object.isFrozen(strObj), 'Object.isFrozen(strObj) !== true');</code> | `built-ins/Object/freeze/15.2.3.9-2-d-3.js` |
| refuses a method read as a value (toString would lose its object, and this with it) | 14 | <code>let o = { x: 1, y: 2 }; let a = Object.keys(o); let s = Object.prototype.toString.call(a); assertSameValue(s, '[object Array]', 's');</code> | `built-ins/Object/keys/15.2.3.14-2-2.js` |
| refuses inherited library member prototype read as an own field | 10 | <code>let arr = [0, 1]; Array.prototype[1] = 2; Object.defineProperties(arr, { length: { value: 1 } }); assertSameValue(arr.length, 1, 'arr.length'); assertSameValue(arr.hasOwnProperty("1"), false, 'arr.hasOwnProperty("1")'); assertSameValue(arr[0], 0, 'arr[0]'); assertSameValue(Array.prototype[1], 2, 'Array.prototype[1]');</code> | `built-ins/Object/defineProperties/15.2.3.7-6-a-167.js` |
| refuses a method read as a value (valueOf would lose its object, and this with it) | 8 | <code>assertSameValue(typeof Object.prototype.valueOf.call(true), "object", 'typeof Object.prototype.valueOf.call(true)');</code> | `built-ins/Object/prototype/valueOf/15.2.4.4-1.js` |
| refuses a method read as a value (propertyIsEnumerable would lose its object, and this with it) | 7 | <code>assert( !!Object.prototype.propertyIsEnumerable.hasOwnProperty("length"), 'The value of !!Object.prototype.propertyIsEnumerable.hasOwnProperty("length") is expected to be true' ); assertSameValue( Object.prototype.propertyIsEnumerable.length, 1, 'The value of Object.prototype.propertyIsEnumerable.length is expected to be 1' );</code> | `built-ins/Object/prototype/propertyIsEnumerable/S15.2.4.7_A11.js` |
| refuses a method read as a value (hasOwnProperty would lose its object, and this with it) | 6 | <code>assert( Object.prototype.hasOwnProperty.call(Object, "length"), "The Object constructor has a 'length' own property" ); assertSameValue(Object.length, 1, 'The value of Object.length is expected to be 1');</code> | `built-ins/Object/S15.2.3_A3.js` |
| not yet: an OmittedExpression in an array literal | 6 | <code>let arr = [0, , 2]; Object.defineProperties(arr, { length: { value: 5 } }); assertSameValue(arr.length, 5, 'arr.length'); assertSameValue(arr[0], 0, 'arr[0]'); assertSameValue(arr.hasOwnProperty("1"), false, 'arr.hasOwnProperty("1")'); assertSameValue(arr[2], 2, 'arr[2]');</code> | `built-ins/Object/defineProperties/15.2.3.7-6-a-155.js` |
| refuses arguments | 6 | <code>let arg = function() { return arguments; }(); Object.defineProperty(arg, "prop", { value: 11, configurable: false }); assertThrows("TypeError", function() { Object.defineProperties(arg, { prop: { value: 12, configurable: true } }); });</code> | `built-ins/Object/defineProperties/15.2.3.7-6-a-22.js` |
| refuses a method read as a value (toLocaleString would lose its object, and this with it) | 6 | <code>assertSameValue( typeof Object.prototype.toLocaleString, "function", 'The value of &#96;typeof Object.prototype.toLocaleString&#96; is expected to be "function"' ); assertSameValue( Object.prototype.toLocaleString(), Object.prototype.toString(), 'Object.prototype.toLocaleString() must return the same value returned by Object.prototype.toString()' ); assertSameValue( {}.toLocaleString(), {}.toString(), '({}).toLocaleString() must return the same value returned by ({}).toString()' );</code> | `built-ins/Object/prototype/toLocaleString/S15.2.4.3_A1.js` |
| refuses delete | 4 | <code>let descObj = {}; Object.defineProperty(descObj, "configurable", { get: function() { return true; } }); let newObj = Object.create({}, { prop: descObj }); let result1 = newObj.hasOwnProperty("prop"); delete newObj.prop; let result2 = newObj.hasOwnProperty("prop"); assertSameValue(result1, true, 'result1'); assertSameValue(result2, false, 'result2');</code> | `built-ins/Object/create/15.2.3.5-4-105.js` |
| refuses isPrototypeOf | 3 | <code>assert( !!Function.prototype.isPrototypeOf(Object), 'The value of !!Function.prototype.isPrototypeOf(Object) is expected to be true' );</code> | `built-ins/Object/S15.2.3_A2.js` |
| refuses var | 3 | <code>var props = {}; var result = false; Object.defineProperty(props, "prop", { get: function() { result = this instanceof Object; return {}; }, enumerable: true }); Object.create({}, props); assert(result, 'result !== true');</code> | `built-ins/Object/create/15.2.3.5-4-4.js` |
| refuses a method read as a value (create would lose its object, and this with it) | 2 | <code>assertSameValue(typeof(Object.create), "function", 'typeof(Object.create)');</code> | `built-ins/Object/create/15.2.3.5-0-1.js` |
| refuses a value of type { enumerable: boolean | 2 | <code>let accessed = false; let descObj = { enumerable: false }; let newObj = Object.create({}, { prop: descObj }); for (let property in newObj) { if (property === "prop") { accessed = true; } } assertSameValue(accessed, false, 'accessed'); assert(newObj.hasOwnProperty("prop"), 'newObj.hasOwnProperty("prop") !== true');</code> | `built-ins/Object/create/15.2.3.5-4-75.js` |
| refuses a method read as a value (defineProperties would lose its object, and this with it) | 2 | <code>let f = Object.defineProperties; assertSameValue(typeof(f), "function", 'typeof(f)');</code> | `built-ins/Object/defineProperties/15.2.3.7-0-1.js` |
| refuses a method read as a value (defineProperty would lose its object, and this with it) | 2 | <code>let f = Object.defineProperty; assertSameValue(typeof(f), "function", 'typeof(f)');</code> | `built-ins/Object/defineProperty/15.2.3.6-0-1.js` |
| refuses a method read as a value (freeze would lose its object, and this with it) | 2 | <code>let f = Object.freeze; assertSameValue(typeof(f), "function", 'typeof(f)');</code> | `built-ins/Object/freeze/15.2.3.9-0-1.js` |
| refuses a method read as a value (getOwnPropertyNames would lose its object, and this with it) | 2 | <code>assertSameValue(typeof(Object.getOwnPropertyNames), "function", 'typeof(Object.getOwnPropertyNames)');</code> | `built-ins/Object/getOwnPropertyNames/15.2.3.4-0-1.js` |
| refuses a method read as a value (getPrototypeOf would lose its object, and this with it) | 2 | <code>assertSameValue(typeof(Object.getPrototypeOf), "function", 'typeof(Object.getPrototypeOf)');</code> | `built-ins/Object/getPrototypeOf/15.2.3.2-0-1.js` |
| refuses a method read as a value (hasOwn would lose its object, and this with it) | 2 | <code>assertSameValue(typeof Object.hasOwn, 'function'); assert(Object.hasOwn(Object, 'hasOwn'));</code> | `built-ins/Object/hasOwn/hasown.js` |
| not yet: an array of never | 2 | <code>assertSameValue(Object.is(true, false), false, "&#96;Object.is(true, false)&#96; returns &#96;false&#96;"); assertSameValue(Object.is(false, true), false, "&#96;Object.is(false, true)&#96; returns &#96;false&#96;"); assertSameValue(Object.is(true, 1), false, "&#96;Object.is(true, 1)&#96; returns &#96;false&#96;"); assertSameValue(Object.is(false, 0), false, "&#96;Object.is(false, 0)&#96; returns &#96;false&#96;"); assertSameValue(Object.is(true, {}), false, "&#96;Object.is(true, {})&#96; returns &#96;false&#96;"); assertSameValue(Object.is(true, undefined), false, "&#96;Object.is(true, undefined)&#96; returns &#96;false&#96;"); assertSameValue(Object.is(false, undefined), false, "&#96;Object.is(false, undefined)&#96; returns &#96;false&#96;"); assertSameValue(Object.is(true, null), false, "&#96;Object.is(true, null)&#96; returns &#96;false&#96;"); assertSameValue(Object.is(false, null), false, "&#96;Object.is(false, null)&#96; returns &#96;false&#96;"); assertSameValue(Object.is(true, NaN), false, "&#96;Object.is(true, NaN)&#96; returns &#96;false&#96;"); assertSameValue(Object.is(false, NaN), false, "&#96;Object.is(false, NaN)&#96; returns &#96;false&#96;"); assertSameValue(Object.is(true, ''), false, "&#96;Object.is(true, '')&#96; returns &#96;false&#96;"); assertSameValue(Object.is(false, ''), false, "&#96;Object.is(false, '')&#96; returns &#96;false&#96;"); assertSameValue(Object.is(true, []), false, "&#96;Object.is(true, [])&#96; returns &#96;false&#96;"); assertSameValue(Object.is(false, []), false, "&#96;Object.is(false, [])&#96; returns &#96;false&#96;");</code> | `built-ins/Object/is/not-same-value-x-y-boolean.js` |
| not yet: Object as a value outside equality or typeof (overloaded calls and static properties need their own representation) | 2 | <code>assertSameValue(Object.is({}, {}), false, "&#96;Object.is({}, {})&#96; returns &#96;false&#96;"); assertSameValue( Object.is(Object(), Object()), false, "&#96;Object.is(Object(), Object())&#96; returns &#96;false&#96;" ); assertSameValue( Object.is(new Object(), new Object()), false, "&#96;Object.is(new Object(), new Object())&#96; returns &#96;false&#96;" ); assertSameValue( Object.is(Object(0), Object(0)), false, "&#96;Object.is(Object(0), Object(0))&#96; returns &#96;false&#96;" ); assertSameValue( Object.is(new Object(''), new Object('')), false, "&#96;Object.is(new Object(''), new Object(''))&#96; returns &#96;false&#96;" );</code> | `built-ins/Object/is/not-same-value-x-y-object.js` |
| refuses a method read as a value (isExtensible would lose its object, and this with it) | 2 | <code>let f = Object.isExtensible; assertSameValue(typeof(f), "function", 'typeof(f)');</code> | `built-ins/Object/isExtensible/15.2.3.13-0-1.js` |
| refuses a method read as a value (isFrozen would lose its object, and this with it) | 2 | <code>let f = Object.isFrozen; assertSameValue(typeof(f), "function", 'typeof(f)');</code> | `built-ins/Object/isFrozen/15.2.3.12-0-1.js` |
| refuses a method read as a value (isSealed would lose its object, and this with it) | 2 | <code>let f = Object.isSealed; assertSameValue(typeof(f), "function", 'typeof(f)');</code> | `built-ins/Object/isSealed/15.2.3.11-0-1.js` |
| refuses a method read as a value (keys would lose its object, and this with it) | 2 | <code>let f = Object.keys; assertSameValue(typeof(f), "function", 'typeof(f)');</code> | `built-ins/Object/keys/15.2.3.14-0-1.js` |
| refuses a method read as a value (preventExtensions would lose its object, and this with it) | 2 | <code>let f = Object.preventExtensions; assertSameValue(typeof(f), "function", 'typeof(f)');</code> | `built-ins/Object/preventExtensions/15.2.3.10-0-1.js` |
| not yet: toString through an object view (a value with a different native representation may be hidden by the view) | 2 | <code>let tostr = Object.prototype.toString(); assertSameValue(tostr, "[object Object]", 'The value of tostr is expected to be "[object Object]"');</code> | `built-ins/Object/prototype/S15.2.4_A2.js` |
| not yet: a value of type { a: number; } & { a: number; } & { a: string; } | 1 | <code>let target = { a: 1 }; let result = Object.assign(target, { a: 2 }, { a: "c" }); assertSameValue(result.a, "c", "The value should be 'c'.");</code> | `built-ins/Object/assign/ObjectOverride-sameproperty.js` |
| not yet: a value of type never | 1 | <code>let target = 12; let result = Object.assign(target, "aaa", "bb2b", "1c"); assertSameValue(Object.getOwnPropertyNames(result).length, 4, "The length should be 4 in the final object."); assertSameValue(result[0], "1", "The value should be {\"0\":\"1\"}."); assertSameValue(result[1], "c", "The value should be {\"1\":\"c\"}."); assertSameValue(result[2], "2", "The value should be {\"2\":\"2\"}."); assertSameValue(result[3], "b", "The value should be {\"3\":\"b\"}.");</code> | `built-ins/Object/assign/Override-notstringtarget.js` |
| not yet: a value of type Object & "123" | 1 | <code>let target = new Object(); let result = Object.assign(target, "123"); assertSameValue(result[0], "1", 'The value of result[0] is expected to be "1"'); assertSameValue(result[1], "2", 'The value of result[1] is expected to be "2"'); assertSameValue(result[2], "3", 'The value of result[2] is expected to be "3"');</code> | `built-ins/Object/assign/Source-String.js` |
| not yet: a value of type true & { a: number; } | 1 | <code>let result = Object.assign(true, { a: 1 }); assertSameValue(typeof result, "object", "Return value should be an object."); assertSameValue(result.valueOf(), true, "Return value should be true.");</code> | `built-ins/Object/assign/Target-Boolean.js` |
| not yet: a value of type 1 & { a: number; } | 1 | <code>let result = Object.assign(1, { a: 1 }); assertSameValue(typeof result, "object", "Return value should be an object."); assertSameValue(result.valueOf(), 1, "Return value should be 1.");</code> | `built-ins/Object/assign/Target-Number.js` |
| not yet: a value of type "test" & { a: number; } | 1 | <code>let result = Object.assign("test", { a: 1 }); assertSameValue(typeof result, "object", "Return value should be an object."); assertSameValue(result.valueOf(), "test", "Return value should be 'test'.");</code> | `built-ins/Object/assign/Target-String.js` |
| refuses a value of type { get: () => number | 1 | <code>let o = {}; let getter = function() { return 1; } let desc = { get: getter, writable: false }; assertThrows("TypeError", function() { Object.defineProperty(o, "foo", desc); }); assertSameValue(o.hasOwnProperty("foo"), false, 'o.hasOwnProperty("foo")');</code> | `built-ins/Object/defineProperty/15.2.3.6-3-2.js` |
| refuses a value of type { set: () => void | 1 | <code>let o = {}; let setter = function() {} let desc = { set: setter, writable: false }; assertThrows("TypeError", function() { Object.defineProperty(o, "foo", desc); }); assertSameValue(o.hasOwnProperty("foo"), false, 'o.hasOwnProperty("foo")');</code> | `built-ins/Object/defineProperty/15.2.3.6-3-4.js` |
| refuses a method read as a value (getOwnPropertyDescriptor would lose its object, and this with it) | 1 | <code>assertSameValue(typeof(Object.getOwnPropertyDescriptor), "function", 'typeof(Object.getOwnPropertyDescriptor)');</code> | `built-ins/Object/getOwnPropertyDescriptor/15.2.3.3-0-1.js` |
| not yet: a value of type boolean & { [x: string]: PropertyDescriptor; } | 1 | <code>let trueResult = Object.getOwnPropertyDescriptors(true); assertSameValue(Object.keys(trueResult).length, 0, 'trueResult has 0 items'); let falseResult = Object.getOwnPropertyDescriptors(false); assertSameValue(Object.keys(falseResult).length, 0, 'falseResult has 0 items');</code> | `built-ins/Object/getOwnPropertyDescriptors/primitive-booleans.js` |
| not yet: this outside a method | 1 | <code>assert(!Object.isFrozen(this));</code> | `built-ins/Object/isFrozen/15.2.3.12-3-1.js` |
| not yet: reading Boolean | 1 | <code>let b = Object.isFrozen(Boolean); assertSameValue(b, false, 'b');</code> | `built-ins/Object/isFrozen/15.2.3.12-3-10.js` |
| not yet: Number as a value outside equality or typeof (overloaded calls and static properties need their own representation) | 1 | <code>let b = Object.isFrozen(Number); assertSameValue(b, false, 'b');</code> | `built-ins/Object/isFrozen/15.2.3.12-3-12.js` |
| not yet: Number.prototype | 1 | <code>let b = Object.isFrozen(Number.prototype); assertSameValue(b, false, 'b');</code> | `built-ins/Object/isFrozen/15.2.3.12-3-13.js` |
| not yet: reading Math | 1 | <code>let b = Object.isFrozen(Math); assertSameValue(b, false, 'b');</code> | `built-ins/Object/isFrozen/15.2.3.12-3-14.js` |
| not yet: reading Date | 1 | <code>let b = Object.isFrozen(Date); assertSameValue(b, false, 'b');</code> | `built-ins/Object/isFrozen/15.2.3.12-3-15.js` |
| not yet: reading RegExp | 1 | <code>let b = Object.isFrozen(RegExp); assertSameValue(b, false, 'b');</code> | `built-ins/Object/isFrozen/15.2.3.12-3-17.js` |
| not yet: reading Error | 1 | <code>let b = Object.isFrozen(Error); assertSameValue(b, false, 'b');</code> | `built-ins/Object/isFrozen/15.2.3.12-3-19.js` |
| not yet: reading EvalError | 1 | <code>let b = Object.isFrozen(EvalError); assertSameValue(b, false, 'b');</code> | `built-ins/Object/isFrozen/15.2.3.12-3-21.js` |
| not yet: reading RangeError | 1 | <code>let b = Object.isFrozen(RangeError); assertSameValue(b, false, 'b');</code> | `built-ins/Object/isFrozen/15.2.3.12-3-22.js` |
| not yet: reading ReferenceError | 1 | <code>let b = Object.isFrozen(ReferenceError); assertSameValue(b, false, 'b');</code> | `built-ins/Object/isFrozen/15.2.3.12-3-23.js` |
| not yet: reading SyntaxError | 1 | <code>let b = Object.isFrozen(SyntaxError); assertSameValue(b, false, 'b');</code> | `built-ins/Object/isFrozen/15.2.3.12-3-24.js` |
| not yet: reading TypeError | 1 | <code>let b = Object.isFrozen(TypeError); assertSameValue(b, false, 'b');</code> | `built-ins/Object/isFrozen/15.2.3.12-3-25.js` |
| not yet: reading URIError | 1 | <code>let b = Object.isFrozen(URIError); assertSameValue(b, false, 'b');</code> | `built-ins/Object/isFrozen/15.2.3.12-3-26.js` |
| not yet: JSON as a value outside equality or typeof (overloaded calls and static properties need their own representation) | 1 | <code>let b = Object.isFrozen(JSON); assertSameValue(b, false, 'b');</code> | `built-ins/Object/isFrozen/15.2.3.12-3-27.js` |
| not yet: reading Function | 1 | <code>let b = Object.isFrozen(Function); assertSameValue(b, false, 'b');</code> | `built-ins/Object/isFrozen/15.2.3.12-3-4.js` |
| not yet: Array as a value outside equality or typeof (overloaded calls and static properties need their own representation) | 1 | <code>let b = Object.isFrozen(Array); assertSameValue(b, false, 'b');</code> | `built-ins/Object/isFrozen/15.2.3.12-3-6.js` |
| not yet: String as a value outside equality or typeof (overloaded calls and static properties need their own representation) | 1 | <code>let b = Object.isFrozen(String); assertSameValue(b, false, 'b');</code> | `built-ins/Object/isFrozen/15.2.3.12-3-8.js` |
| not yet: assigning a field of a value | 1 | <code>function foo() {} foo.x = 1; let a = Object.keys(foo); assertSameValue(a.length, 1, 'a.length'); assertSameValue(a[0], 'x', 'a[0]');</code> | `built-ins/Object/keys/15.2.3.14-3-2.js` |
| not yet: hasOwnProperty through an object view (a value with a different native representation may be hidden by the view) | 1 | <code>let o = {}; assertSameValue(o.hasOwnProperty("foo"), false, 'o.hasOwnProperty("foo")');</code> | `built-ins/Object/prototype/hasOwnProperty/8.12.1-1_1.js` |
| refuses a method read as a value (seal would lose its object, and this with it) | 1 | <code>assertSameValue(typeof Object.seal, "function", 'typeof(f)');</code> | `built-ins/Object/seal/object-seal-is-a-function.js` |
