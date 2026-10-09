# Fresh Object refusal classification

The initial classification was committed and pushed before implementation as `c48709e`, from main `5d4c801`, with real `typescript@6.0.3` and test262 `5992dc3b60faf62a48fd6be8a40ae9d9a8c84d81`. Adaptation is enabled. The old 0-pass measurement predates existing Object support.

| Measurement | Pass | Disagreement | Refused | Crashed | Skipped | Total |
|---|---:|---:|---:|---:|---:|---:|
| Fresh before | 17 | 1 | 2317 | 0 | 1076 | 3411 |

## Every refusal checked

| Bucket | Count | Meaning |
|---|---:|---|
| a | 1679 | Real tsc rejects with the same TS diagnostic code |
| b | 425 | Real tsc accepts; first refusal names a library feature or dependency |
| c | 213 | Real tsc accepts; first refusal names a language or representation feature |

No tsc rejection had a different code. A match means the runner's first normalized TS code appears among real tsc's diagnostics, not that every diagnostic or its order is identical. Every refused adapted program was checked separately against `lib.es2024.d.ts` and the exact `internal/load/prelude.d.ts`, and Adamic's pinned declaration files from `cohere/TypeScript/tsc/internal/bundled/libs` with every `regexp_library.go` correction, with all `compilerOptions()` from `internal/load/load.go`. The checker implementation is the npm TypeScript package, not typescript-go. Skipped tests are outside the refusal classification.

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
| `String constructor` | 7 |
| `Object.freeze` | 6 |
| `Object.create` | 5 |
| `Object.prototype.isPrototypeOf` | 3 |
| `Object.values` | 3 |
| `Object.assign` | 2 |
| `Object.entries` | 2 |
| `Boolean constructor` | 2 |
| `Number constructor` | 2 |
| `Date constructor` | 2 |
| `Object.fromEntries` | 2 |
| `Object.hasOwn` | 2 |
| `Object builtin value` | 2 |
| `Object.hasOwnProperty` | 1 |
| `SyntaxError constructor` | 1 |
| `Object.getOwnPropertyDescriptors` | 1 |
| `Object.getPrototypeOf` | 1 |
| `Object.groupBy` | 1 |
| `Object constructor` | 1 |
| `Boolean builtin value` | 1 |
| `Number builtin value` | 1 |
| `Number.prototype builtin value` | 1 |
| `Math builtin value` | 1 |
| `Date builtin value` | 1 |
| `RegExp builtin value` | 1 |
| `Error builtin value` | 1 |
| `EvalError builtin value` | 1 |
| `RangeError builtin value` | 1 |
| `ReferenceError builtin value` | 1 |
| `SyntaxError builtin value` | 1 |
| `TypeError builtin value` | 1 |
| `URIError builtin value` | 1 |
| `JSON builtin value` | 1 |
| `Function builtin value` | 1 |
| `Array builtin value` | 1 |
| `String builtin value` | 1 |
| `Object.propertyIsEnumerable` | 1 |
| `Object.setPrototypeOf` | 1 |

## Bucket (c) reasons

| Reason | Count |
|---|---:|
| `not yet: a value of type any` | 99 |
| `refuses a method read as a value (toString would lose its object, and this with it)` | 14 |
| `refuses inherited library member prototype read as an own field` | 10 |
| `refuses a method read as a value (valueOf would lose its object, and this with it)` | 8 |
| `refuses a method read as a value (propertyIsEnumerable would lose its object, and this with it)` | 7 |
| `refuses a method read as a value (hasOwnProperty would lose its object, and this with it)` | 6 |
| `not yet: an OmittedExpression in an array literal` | 6 |
| `refuses arguments` | 6 |
| `refuses a method read as a value (toLocaleString would lose its object, and this with it)` | 6 |
| `refuses delete` | 4 |
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
| `not yet: assigning a field of a value` | 1 |
| `not yet: hasOwnProperty through an object view (a value with a different native representation may be hidden by the view)` | 1 |
| `refuses a method read as a value (seal would lose its object, and this with it)` | 1 |

## Classification refinement

The first claim used npm's es2024 declarations and classified generic builtin constructor/value diagnostics as representation gaps: (a) 1679, (b) 388, (c) 250. The final real-tsc pass uses the checker's exact pinned declaration text and RegExp corrections. Constructor/value expressions were inspected and named as library dependencies rather than generic language blockers; the final tables reflect that refinement. The accepted/rejected partition and matching diagnostic codes are checked again, not inferred. These dependencies may belong to other library slices and do not authorize edits there.

## Claim and build order

1. The largest library refusal, `Object.defineProperty` (185), and `defineProperties` (24) require observable descriptors, accessors and shape mutation. These remain explicit refusals: CLAUDE.md binds the approved fixed-shape memory and type contract, and docs/0.1.md forbids this mutation. Implementing them needs a representation and soundness design beyond this slice. Prototype mutation has the same constraint. Descriptor reads currently return any, adding a language blocker even if their first refusal were removed.
2. Build the largest remaining family: `Object.seal` (38), `Object.isExtensible` (35), `Object.isSealed` (28), then `Object.preventExtensions` (9). Port V8's integrity tests for represented fixed plain objects and its primitive fast paths. Unknown host constructors, arrays needing integrity mutation, descriptor/prototype mutation, and unrepresented values remain refused. Existing Object.freeze must interoperate with the integrity queries.
3. Investigate own-name reflection (16) and exact-shape keys (14) after integrity, retaining explicit refusals whenever the shape or presence is unproven.
4. Close the observed baseline disagreement `Object.is(undefined, null)`: the existing null/undefined representation cannot be silently compared as equal. Refuse ambiguous comparisons unless this slice can prove the exact result.

## Language handoff for @system_adamic

Each row is a representative adapted test body on one line. To reproduce, prepend the unchanged runner prelude from `cmd/adamic-test262/prelude.go`. The full per-test sources and diagnostics are in the linked artifacts; no language changes are claimed here.

| First language/representation refusal | Tests | One-line reproducer | Representative test |
|---|---:|---|---|
| not yet: a value of type any | 99 | <code>let target = { a: 1 }; let result = Object.assign(target, &quot;1a2c3&quot;, { a: &quot;c&quot; }, undefined, { b: 6 }, null, 125, { a: 5 }); assertSameValue(Object.getOwnPropertyNames(result).length, 7, &quot;The length should be 7 in the final object.&quot;); assertSameValue(result.a, 5, &quot;The value should be {a:5}.&quot;); assertSameValue(result[0], &quot;1&quot;, &quot;The value should be {\&quot;0\&quot;:\&quot;1\&quot;}.&quot;); assertSameValue(result[1], &quot;a&quot;, &quot;The value should be {\&quot;1\&quot;:\&quot;a\&quot;}.&quot;); assertSameValue(result[2], &quot;2&quot;, &quot;The value should be {\&quot;2\&quot;:\&quot;2\&quot;}.&quot;); assertSameValue(result[3], &quot;c&quot;, &quot;The value should be {\&quot;3\&quot;:\&quot;c\&quot;}.&quot;); assertSameValue(result[4], &quot;3&quot;, &quot;The value should be {\&quot;4\&quot;:\&quot;3\&quot;}.&quot;); assertSameValue(result.b, 6, &quot;The value should be {b:6}.&quot;);</code> | `built-ins/Object/assign/Override.js` |
| refuses a method read as a value (toString would lose its object, and this with it) | 14 | <code>let o = { x: 1, y: 2 }; let a = Object.keys(o); let s = Object.prototype.toString.call(a); assertSameValue(s, &#x27;[object Array]&#x27;, &#x27;s&#x27;);</code> | `built-ins/Object/keys/15.2.3.14-2-2.js` |
| refuses inherited library member prototype read as an own field | 10 | <code>let arr = [0, 1]; Array.prototype[1] = 2; Object.defineProperties(arr, { length: { value: 1 } }); assertSameValue(arr.length, 1, &#x27;arr.length&#x27;); assertSameValue(arr.hasOwnProperty(&quot;1&quot;), false, &#x27;arr.hasOwnProperty(&quot;1&quot;)&#x27;); assertSameValue(arr[0], 0, &#x27;arr[0]&#x27;); assertSameValue(Array.prototype[1], 2, &#x27;Array.prototype[1]&#x27;);</code> | `built-ins/Object/defineProperties/15.2.3.7-6-a-167.js` |
| refuses a method read as a value (valueOf would lose its object, and this with it) | 8 | <code>assertSameValue(typeof Object.prototype.valueOf.call(true), &quot;object&quot;, &#x27;typeof Object.prototype.valueOf.call(true)&#x27;);</code> | `built-ins/Object/prototype/valueOf/15.2.4.4-1.js` |
| refuses a method read as a value (propertyIsEnumerable would lose its object, and this with it) | 7 | <code>assert( !!Object.prototype.propertyIsEnumerable.hasOwnProperty(&quot;length&quot;), &#x27;The value of !!Object.prototype.propertyIsEnumerable.hasOwnProperty(&quot;length&quot;) is expected to be true&#x27; ); assertSameValue( Object.prototype.propertyIsEnumerable.length, 1, &#x27;The value of Object.prototype.propertyIsEnumerable.length is expected to be 1&#x27; );</code> | `built-ins/Object/prototype/propertyIsEnumerable/S15.2.4.7_A11.js` |
| refuses a method read as a value (hasOwnProperty would lose its object, and this with it) | 6 | <code>assert( Object.prototype.hasOwnProperty.call(Object, &quot;length&quot;), &quot;The Object constructor has a &#x27;length&#x27; own property&quot; ); assertSameValue(Object.length, 1, &#x27;The value of Object.length is expected to be 1&#x27;);</code> | `built-ins/Object/S15.2.3_A3.js` |
| not yet: an OmittedExpression in an array literal | 6 | <code>let arr = [0, , 2]; Object.defineProperties(arr, { length: { value: 5 } }); assertSameValue(arr.length, 5, &#x27;arr.length&#x27;); assertSameValue(arr[0], 0, &#x27;arr[0]&#x27;); assertSameValue(arr.hasOwnProperty(&quot;1&quot;), false, &#x27;arr.hasOwnProperty(&quot;1&quot;)&#x27;); assertSameValue(arr[2], 2, &#x27;arr[2]&#x27;);</code> | `built-ins/Object/defineProperties/15.2.3.7-6-a-155.js` |
| refuses arguments | 6 | <code>let arg = function() { return arguments; }(); Object.defineProperty(arg, &quot;prop&quot;, { value: 11, configurable: false }); assertThrows(&quot;TypeError&quot;, function() { Object.defineProperties(arg, { prop: { value: 12, configurable: true } }); });</code> | `built-ins/Object/defineProperties/15.2.3.7-6-a-22.js` |
| refuses a method read as a value (toLocaleString would lose its object, and this with it) | 6 | <code>assertSameValue( typeof Object.prototype.toLocaleString, &quot;function&quot;, &#x27;The value of &#96;typeof Object.prototype.toLocaleString&#96; is expected to be &quot;function&quot;&#x27; ); assertSameValue( Object.prototype.toLocaleString(), Object.prototype.toString(), &#x27;Object.prototype.toLocaleString() must return the same value returned by Object.prototype.toString()&#x27; ); assertSameValue( {}.toLocaleString(), {}.toString(), &#x27;({}).toLocaleString() must return the same value returned by ({}).toString()&#x27; );</code> | `built-ins/Object/prototype/toLocaleString/S15.2.4.3_A1.js` |
| refuses delete | 4 | <code>let descObj = {}; Object.defineProperty(descObj, &quot;configurable&quot;, { get: function() { return true; } }); let newObj = Object.create({}, { prop: descObj }); let result1 = newObj.hasOwnProperty(&quot;prop&quot;); delete newObj.prop; let result2 = newObj.hasOwnProperty(&quot;prop&quot;); assertSameValue(result1, true, &#x27;result1&#x27;); assertSameValue(result2, false, &#x27;result2&#x27;);</code> | `built-ins/Object/create/15.2.3.5-4-105.js` |
| refuses var | 3 | <code>var props = {}; var result = false; Object.defineProperty(props, &quot;prop&quot;, { get: function() { result = this instanceof Object; return {}; }, enumerable: true }); Object.create({}, props); assert(result, &#x27;result !== true&#x27;);</code> | `built-ins/Object/create/15.2.3.5-4-4.js` |
| refuses a method read as a value (create would lose its object, and this with it) | 2 | <code>assertSameValue(typeof(Object.create), &quot;function&quot;, &#x27;typeof(Object.create)&#x27;);</code> | `built-ins/Object/create/15.2.3.5-0-1.js` |
| refuses a value of type { enumerable: boolean | 2 | <code>let accessed = false; let descObj = { enumerable: false }; let newObj = Object.create({}, { prop: descObj }); for (let property in newObj) { if (property === &quot;prop&quot;) { accessed = true; } } assertSameValue(accessed, false, &#x27;accessed&#x27;); assert(newObj.hasOwnProperty(&quot;prop&quot;), &#x27;newObj.hasOwnProperty(&quot;prop&quot;) !== true&#x27;);</code> | `built-ins/Object/create/15.2.3.5-4-75.js` |
| refuses a method read as a value (defineProperties would lose its object, and this with it) | 2 | <code>let f = Object.defineProperties; assertSameValue(typeof(f), &quot;function&quot;, &#x27;typeof(f)&#x27;);</code> | `built-ins/Object/defineProperties/15.2.3.7-0-1.js` |
| refuses a method read as a value (defineProperty would lose its object, and this with it) | 2 | <code>let f = Object.defineProperty; assertSameValue(typeof(f), &quot;function&quot;, &#x27;typeof(f)&#x27;);</code> | `built-ins/Object/defineProperty/15.2.3.6-0-1.js` |
| refuses a method read as a value (freeze would lose its object, and this with it) | 2 | <code>let f = Object.freeze; assertSameValue(typeof(f), &quot;function&quot;, &#x27;typeof(f)&#x27;);</code> | `built-ins/Object/freeze/15.2.3.9-0-1.js` |
| refuses a method read as a value (getOwnPropertyNames would lose its object, and this with it) | 2 | <code>assertSameValue(typeof(Object.getOwnPropertyNames), &quot;function&quot;, &#x27;typeof(Object.getOwnPropertyNames)&#x27;);</code> | `built-ins/Object/getOwnPropertyNames/15.2.3.4-0-1.js` |
| refuses a method read as a value (getPrototypeOf would lose its object, and this with it) | 2 | <code>assertSameValue(typeof(Object.getPrototypeOf), &quot;function&quot;, &#x27;typeof(Object.getPrototypeOf)&#x27;);</code> | `built-ins/Object/getPrototypeOf/15.2.3.2-0-1.js` |
| refuses a method read as a value (hasOwn would lose its object, and this with it) | 2 | <code>assertSameValue(typeof Object.hasOwn, &#x27;function&#x27;); assert(Object.hasOwn(Object, &#x27;hasOwn&#x27;));</code> | `built-ins/Object/hasOwn/hasown.js` |
| not yet: an array of never | 2 | <code>assertSameValue(Object.is(true, false), false, &quot;&#96;Object.is(true, false)&#96; returns &#96;false&#96;&quot;); assertSameValue(Object.is(false, true), false, &quot;&#96;Object.is(false, true)&#96; returns &#96;false&#96;&quot;); assertSameValue(Object.is(true, 1), false, &quot;&#96;Object.is(true, 1)&#96; returns &#96;false&#96;&quot;); assertSameValue(Object.is(false, 0), false, &quot;&#96;Object.is(false, 0)&#96; returns &#96;false&#96;&quot;); assertSameValue(Object.is(true, {}), false, &quot;&#96;Object.is(true, {})&#96; returns &#96;false&#96;&quot;); assertSameValue(Object.is(true, undefined), false, &quot;&#96;Object.is(true, undefined)&#96; returns &#96;false&#96;&quot;); assertSameValue(Object.is(false, undefined), false, &quot;&#96;Object.is(false, undefined)&#96; returns &#96;false&#96;&quot;); assertSameValue(Object.is(true, null), false, &quot;&#96;Object.is(true, null)&#96; returns &#96;false&#96;&quot;); assertSameValue(Object.is(false, null), false, &quot;&#96;Object.is(false, null)&#96; returns &#96;false&#96;&quot;); assertSameValue(Object.is(true, NaN), false, &quot;&#96;Object.is(true, NaN)&#96; returns &#96;false&#96;&quot;); assertSameValue(Object.is(false, NaN), false, &quot;&#96;Object.is(false, NaN)&#96; returns &#96;false&#96;&quot;); assertSameValue(Object.is(true, &#x27;&#x27;), false, &quot;&#96;Object.is(true, &#x27;&#x27;)&#96; returns &#96;false&#96;&quot;); assertSameValue(Object.is(false, &#x27;&#x27;), false, &quot;&#96;Object.is(false, &#x27;&#x27;)&#96; returns &#96;false&#96;&quot;); assertSameValue(Object.is(true, []), false, &quot;&#96;Object.is(true, [])&#96; returns &#96;false&#96;&quot;); assertSameValue(Object.is(false, []), false, &quot;&#96;Object.is(false, [])&#96; returns &#96;false&#96;&quot;);</code> | `built-ins/Object/is/not-same-value-x-y-boolean.js` |
| refuses a method read as a value (isExtensible would lose its object, and this with it) | 2 | <code>let f = Object.isExtensible; assertSameValue(typeof(f), &quot;function&quot;, &#x27;typeof(f)&#x27;);</code> | `built-ins/Object/isExtensible/15.2.3.13-0-1.js` |
| refuses a method read as a value (isFrozen would lose its object, and this with it) | 2 | <code>let f = Object.isFrozen; assertSameValue(typeof(f), &quot;function&quot;, &#x27;typeof(f)&#x27;);</code> | `built-ins/Object/isFrozen/15.2.3.12-0-1.js` |
| refuses a method read as a value (isSealed would lose its object, and this with it) | 2 | <code>let f = Object.isSealed; assertSameValue(typeof(f), &quot;function&quot;, &#x27;typeof(f)&#x27;);</code> | `built-ins/Object/isSealed/15.2.3.11-0-1.js` |
| refuses a method read as a value (keys would lose its object, and this with it) | 2 | <code>let f = Object.keys; assertSameValue(typeof(f), &quot;function&quot;, &#x27;typeof(f)&#x27;);</code> | `built-ins/Object/keys/15.2.3.14-0-1.js` |
| refuses a method read as a value (preventExtensions would lose its object, and this with it) | 2 | <code>let f = Object.preventExtensions; assertSameValue(typeof(f), &quot;function&quot;, &#x27;typeof(f)&#x27;);</code> | `built-ins/Object/preventExtensions/15.2.3.10-0-1.js` |
| not yet: toString through an object view (a value with a different native representation may be hidden by the view) | 2 | <code>let tostr = Object.prototype.toString(); assertSameValue(tostr, &quot;[object Object]&quot;, &#x27;The value of tostr is expected to be &quot;[object Object]&quot;&#x27;);</code> | `built-ins/Object/prototype/S15.2.4_A2.js` |
| not yet: a value of type { a: number; } & { a: number; } & { a: string; } | 1 | <code>let target = { a: 1 }; let result = Object.assign(target, { a: 2 }, { a: &quot;c&quot; }); assertSameValue(result.a, &quot;c&quot;, &quot;The value should be &#x27;c&#x27;.&quot;);</code> | `built-ins/Object/assign/ObjectOverride-sameproperty.js` |
| not yet: a value of type never | 1 | <code>let target = 12; let result = Object.assign(target, &quot;aaa&quot;, &quot;bb2b&quot;, &quot;1c&quot;); assertSameValue(Object.getOwnPropertyNames(result).length, 4, &quot;The length should be 4 in the final object.&quot;); assertSameValue(result[0], &quot;1&quot;, &quot;The value should be {\&quot;0\&quot;:\&quot;1\&quot;}.&quot;); assertSameValue(result[1], &quot;c&quot;, &quot;The value should be {\&quot;1\&quot;:\&quot;c\&quot;}.&quot;); assertSameValue(result[2], &quot;2&quot;, &quot;The value should be {\&quot;2\&quot;:\&quot;2\&quot;}.&quot;); assertSameValue(result[3], &quot;b&quot;, &quot;The value should be {\&quot;3\&quot;:\&quot;b\&quot;}.&quot;);</code> | `built-ins/Object/assign/Override-notstringtarget.js` |
| not yet: a value of type Object & "123" | 1 | <code>let target = new Object(); let result = Object.assign(target, &quot;123&quot;); assertSameValue(result[0], &quot;1&quot;, &#x27;The value of result[0] is expected to be &quot;1&quot;&#x27;); assertSameValue(result[1], &quot;2&quot;, &#x27;The value of result[1] is expected to be &quot;2&quot;&#x27;); assertSameValue(result[2], &quot;3&quot;, &#x27;The value of result[2] is expected to be &quot;3&quot;&#x27;);</code> | `built-ins/Object/assign/Source-String.js` |
| not yet: a value of type true & { a: number; } | 1 | <code>let result = Object.assign(true, { a: 1 }); assertSameValue(typeof result, &quot;object&quot;, &quot;Return value should be an object.&quot;); assertSameValue(result.valueOf(), true, &quot;Return value should be true.&quot;);</code> | `built-ins/Object/assign/Target-Boolean.js` |
| not yet: a value of type 1 & { a: number; } | 1 | <code>let result = Object.assign(1, { a: 1 }); assertSameValue(typeof result, &quot;object&quot;, &quot;Return value should be an object.&quot;); assertSameValue(result.valueOf(), 1, &quot;Return value should be 1.&quot;);</code> | `built-ins/Object/assign/Target-Number.js` |
| not yet: a value of type "test" & { a: number; } | 1 | <code>let result = Object.assign(&quot;test&quot;, { a: 1 }); assertSameValue(typeof result, &quot;object&quot;, &quot;Return value should be an object.&quot;); assertSameValue(result.valueOf(), &quot;test&quot;, &quot;Return value should be &#x27;test&#x27;.&quot;);</code> | `built-ins/Object/assign/Target-String.js` |
| refuses a value of type { get: () => number | 1 | <code>let o = {}; let getter = function() { return 1; } let desc = { get: getter, writable: false }; assertThrows(&quot;TypeError&quot;, function() { Object.defineProperty(o, &quot;foo&quot;, desc); }); assertSameValue(o.hasOwnProperty(&quot;foo&quot;), false, &#x27;o.hasOwnProperty(&quot;foo&quot;)&#x27;);</code> | `built-ins/Object/defineProperty/15.2.3.6-3-2.js` |
| refuses a value of type { set: () => void | 1 | <code>let o = {}; let setter = function() {} let desc = { set: setter, writable: false }; assertThrows(&quot;TypeError&quot;, function() { Object.defineProperty(o, &quot;foo&quot;, desc); }); assertSameValue(o.hasOwnProperty(&quot;foo&quot;), false, &#x27;o.hasOwnProperty(&quot;foo&quot;)&#x27;);</code> | `built-ins/Object/defineProperty/15.2.3.6-3-4.js` |
| refuses a method read as a value (getOwnPropertyDescriptor would lose its object, and this with it) | 1 | <code>assertSameValue(typeof(Object.getOwnPropertyDescriptor), &quot;function&quot;, &#x27;typeof(Object.getOwnPropertyDescriptor)&#x27;);</code> | `built-ins/Object/getOwnPropertyDescriptor/15.2.3.3-0-1.js` |
| not yet: a value of type boolean & { [x: string]: PropertyDescriptor; } | 1 | <code>let trueResult = Object.getOwnPropertyDescriptors(true); assertSameValue(Object.keys(trueResult).length, 0, &#x27;trueResult has 0 items&#x27;); let falseResult = Object.getOwnPropertyDescriptors(false); assertSameValue(Object.keys(falseResult).length, 0, &#x27;falseResult has 0 items&#x27;);</code> | `built-ins/Object/getOwnPropertyDescriptors/primitive-booleans.js` |
| not yet: this outside a method | 1 | <code>assert(!Object.isFrozen(this));</code> | `built-ins/Object/isFrozen/15.2.3.12-3-1.js` |
| not yet: assigning a field of a value | 1 | <code>function foo() {} foo.x = 1; let a = Object.keys(foo); assertSameValue(a.length, 1, &#x27;a.length&#x27;); assertSameValue(a[0], &#x27;x&#x27;, &#x27;a[0]&#x27;);</code> | `built-ins/Object/keys/15.2.3.14-3-2.js` |
| not yet: hasOwnProperty through an object view (a value with a different native representation may be hidden by the view) | 1 | <code>let o = {}; assertSameValue(o.hasOwnProperty(&quot;foo&quot;), false, &#x27;o.hasOwnProperty(&quot;foo&quot;)&#x27;);</code> | `built-ins/Object/prototype/hasOwnProperty/8.12.1-1_1.js` |
| refuses a method read as a value (seal would lose its object, and this with it) | 1 | <code>assertSameValue(typeof Object.seal, &quot;function&quot;, &#x27;typeof(f)&#x27;);</code> | `built-ins/Object/seal/object-seal-is-a-function.js` |
