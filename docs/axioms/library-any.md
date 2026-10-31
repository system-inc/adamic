# The library's `any`, audited

Every declaration Adamic believes without proof is an axiom. TypeScript's bundled library types 375 positions as `any`, and an `any` the library hands to a program is a value no one proved. This file gives each of those positions exactly one disposition. It is the only place the relation walk's trust in a library `any` may come from (ruled October 31, 2026 by @system_adamic on #tmdhz1d; the audit is #p76cwjw, owned by @system_adamic_language).

## Scope and method

The files are the ones `lib.es2024.d.ts` reaches through `/// <reference lib>`, as bundled at cohere/TypeScript `d92d9bfe` (`tsc/internal/bundled/libs`, 69 files), and `internal/load/prelude.d.ts`, which declares no `any`. Each `any` keyword is found by the TypeScript parser, never by text search, and given a polarity from its position: a parameter flips it, and a parameter of a callback flips it back.

- **Out:** the library hands the value to the program: a return type, a property, a predicate's target, a type argument in one of those, or a parameter of a callback the library calls. 57 positions, plus 19 that flow both ways, a property the program can also write.
- **In:** the program hands the value to the library: a parameter of a library function. 248 positions. These are the library's own runtime contract and carry nothing into the program, with one exception below.
- **Defaults and constraints:** a type parameter's default or constraint. 51 positions.

Node's declarations in `internal/load/node_types` aren't `IsLibrary` and aren't in this audit. They are the host's axioms, a second audit (#c431sb6), and the relation walk counts their `any` until it lands.

## The three dispositions

1. **Corrected.** The prelude merges in a declaration ahead of the library's, which may only add truth (the adaptation rule, October 6 and 7). It's counted on tsc before it lands. @system_adamic_runtime_library builds it, and Language reviews it.
2. **Modelled.** A lowering rule supplies the real type and checks it at the call, loudly where Node runs on (contract item 7 lets a type hatch stop there). The entry names the rule, its adapter's argument and result checks (a model is only as true as the check its adapter makes, ruled October 31 after M3's push), and a fixture that goes red when the rule is removed. Language attacks every modelled entry: a modelled call whose real type is wider than its target is a must.
3. **Counted.** The default. The relation walk counts the `any` wherever it meets a typed slot, so the program is NotYet or Refused. A position no rule names is counted.

A position that never gives a program a value (a constraint, a pattern inside a conditional type's `extends` clause, a registry read only through `keyof`) has no disposition. It is listed under "Not value-carrying" with the reason.

## Corrected

| Id | Declaration | Where | The correction | State |
| --- | --- | --- | --- | --- |
| C1 | `ArrayConstructor.isArray(arg: any): arg is any[]` | lib.es5.d.ts:1518 | `isArray(arg: unknown): arg is readonly unknown[]`, merged ahead. `string \| string[]` still narrows to `string[]`, since a member that's a subtype keeps its own type. `unknown` narrows to `readonly unknown[]`, whose element reads are already NotYet. Without it, `if (Array.isArray(x)) { const ys: string[] = x; }` on `unknown` reads `[5]` raw natively (live on main 31fdd16d, Language's p2 and p8 on #tmdhz1d) | #97b2hqc, owed with its count on tsc |
| C2 | `MapConstructor` `new (): Map<any, any>` | lib.es2015.collection.d.ts:50, lib.es2015.iterable.d.ts:180 | `new <K = unknown, V = unknown>(): Map<K, V>`, merged ahead. `const m: Map<string, number> = new Map()` infers its arguments from the slot, so the honest idiom keeps its type, and an uncontextual `new Map()` is `Map<unknown, unknown>` instead of a map that relates to anything | #bmbb2vb, only if the census on tsc says it pays. The modelled fallback stays the default: an argumentless `new Map()` is empty, so its `any` arguments carry no value at the relation that receives it. On main 31fdd16d the honest `const t: Map<string, string> = new Map()` is NotYet ("a Map whose keys aren't strings"), so either form is a gain |

## Modelled

| Id | Declaration | Where | The rule that supplies the real type | Its pins | Attacked |
| --- | --- | --- | --- | --- | --- |
| M1 | `String.replace`'s and `String.replaceAll`'s callback, `(substring: string, ...args: any[]) => string`, and the same in `RegExp[Symbol.replace]` and `String.replace`'s symbol overload, **with a RegExp pattern only** | lib.es5.d.ts:468, lib.es2021.string.d.ts:30, lib.es2015.symbol.wellknown.d.ts:204, :253 | regexReplacement (regexp.go:199, library_regexp_replace.go:70) supplies each argument (the captures as `string \| undefined`, then the offset, the input and the groups). Its adapter checks each against the callback's declared parameter, in internal/native/runtime/regexp_replace.c and internal/javascript/library_regexp_replace.go: `panic: RegExp replacement callback argument does not fit its declared type`. A string pattern has no lowering, so it isn't modelled for either method: rule 3100 (refuseStringWidening, invariance.go:561) refuses its callback before any adapter | internal/oracle/testdata/regexp_replace/callback.a, arguments.a, argument_guard.a; notyet_library_regex_offset.a; closure_convention_regexp_count.a | Oct 31 on main 31fdd16d: the offset seen as `string` with no capture group, an unmatched group seen as `string`, and the global form each stop at the call on both backends (audit probes m1 to m3) |
| M2 | `ObjectConstructor.entries(o: {}): [string, any][]` and `values(o: {}): any[]` | lib.es2017.object.d.ts:28, :40 | The object enumeration lowering reads each own field's actual type and checks it against the declared element: `panic: Object enumeration key '<key>': actual <type>, declared <type>`. A shape whose declared fields don't fit the element is refused before that (`Adamic 0.1 refuses Object.entries; every present field must have tsc's result element representation`) | internal/oracle/testdata/entries_record_alias.a, entries_record_misfit.a | Oct 31 on main: a hidden field (`{ a: 5, b: "x" }` seen as `{ a: number }`) through an interface and through an alias, for entries and for values, stops at `'b'` on both backends (m4 to m6). A shape whose field doesn't fit is refused (p3, p4) |
| M3 | `ArrayConstructor.prototype: any[]`, through which a method value (`Array.prototype.slice`) has `any` parameters and results | lib.es5.d.ts:1519 | library_method_values.go's typed adapters. Each proves the receiver's element type and checks what crosses it: indexOf, lastIndexOf, includes, join, slice's result, fill's value (object.go:1365), splice's inserted values, concat's arrays and values. An adapter that doesn't exist is NotYet (unshift) | internal/oracle/testdata/library_method_values.a, method_coverage_array_mutations.a | Oct 31 on main: a result into another element type (k1), concat of another element (k3) or a non-array (k7), splice (k4) and fill (k5) of another type are each NotYet. **push is a must:** `push.call(array, { x: 1 })` on a `string[]` is admitted, and its later string read is a null-pointer load under UBSan natively, while Node throws a TypeError (k10). A number stops only inside clang (k6, k8). **M3 is trusted only once #gw3yrsn lands** (push's argument check). Until then a push method value through call or apply isn't modelled, and the relation counts it |

M3 is the one place an In position carries a value into the program: the library writes its `any` argument into an array the program owns, through `this`. So an In `any` stays harmless only while no adapter stores it in a program value unchecked, and every adapter that stores is held to that.

## Counted

Each is counted at the relation. Most of the calls are refused or NotYet before any relation, and that's named where it holds.

| Declaration | Where | Today |
| --- | --- | --- |
| `eval(x: string): any` | lib.es5.d.ts:31 | `eval` is refused forever (R-T10) |
| `PropertyDescriptor.value`, `get(): any` | lib.es5.d.ts:111, :113 | `Object.getOwnPropertyDescriptor` and `defineProperty` are refused |
| `Object()` and `Object(value)` | lib.es5.d.ts:155, :156 | Counted |
| `Object.getPrototypeOf`, `Object.create` (both), `Object.setPrototypeOf` | lib.es5.d.ts:165, :186, :193; lib.es2015.core.d.ts:337 | Prototype mutation is refused; the rest is counted |
| `Function.apply`, `call`, `bind` on `Function`; `Function.prototype`; `Function.arguments` | lib.es5.d.ts:279, :286, :294, :299, :303 | A value typed `Function` is refused; typed code reaches `CallableFunction`'s overloads, which carry no `any` |
| `IArguments`' index and iterator | lib.es5.d.ts:403; lib.es2015.iterable.d.ts:132 | Reading an element of `arguments` is an expression of type `any`, NotYet |
| `replace`'s and `replaceAll`'s callback with a string pattern | lib.es5.d.ts:468, lib.es2021.string.d.ts:30 | Refused at the relation by rule 3100 ("a function taking string seen as one taking any[]", audit probe m7). It moves to M1 only with a lowering that makes the same per-argument check (#g41ym71) |
| `JSON.parse`'s result and its reviver's `this` and `value`; `JSON.stringify`'s replacer's `this` and `value` | lib.es5.d.ts:1161, :1169 | `JSON.parse`'s result is an expression of type `any`, NotYet |
| `new Array(n)` and `Array(n)`: `any[]` | lib.es5.d.ts:1512, :1515 | NotYet ("an array of any", "new an Identifier") |
| A rejection's reason: `then` and `catch`'s `onrejected`, `PromiseRejectedResult.reason`, `AggregateError.errors` | lib.es5.d.ts:1542, :1555, :1562; lib.es2020.promise.d.ts:24; lib.es2021.promise.d.ts:18 | A callback that types `reason` as anything but `unknown` is counted. `Promise.allSettled` is NotYet today |
| `Object.assign(target, ...sources: any[]): any` | lib.es2015.core.d.ts:311 | The rest overload is refused; the typed overloads carry no `any` |
| `Object.fromEntries(entries: Iterable<readonly any[]>): any`, and `fromEntries<T = any>` | lib.es2019.object.d.ts:24, :30 | Counted |
| `ProxyHandler.apply` and `get` | lib.es2015.proxy.d.ts:22, :52 | Counted |
| `Reflect.apply`, `construct` (both), `get`, `getOwnPropertyDescriptor` | lib.es2015.reflect.d.ts:30, :42, :44, :74, :85 | Counted |
| The constructors' `prototype`: Map, WeakMap, Set, Promise, WeakRef, FinalizationRegistry | lib.es2015.collection.d.ts:52, :86, :120; lib.es2015.promise.d.ts:21; lib.es2021.weakref.d.ts:31, :67 | Counted |
| `ReturnType<T>` and `InstanceType<T>`'s false branch | lib.es5.d.ts:1648, :1653 | Reached only when T is `any` or `never`; counted |
| The defaults `TReturn = any` and `TNext = any` on Iterator, Iterable, IterableIterator, Generator, AsyncIterator, AsyncIterable, AsyncIterableIterator, AsyncGenerator, and `TReturn = any` on IteratorResult | lib.es2015.iterable.d.ts:37, :39, :46, :53; lib.es2015.generator.d.ts:19; lib.es2018.asynciterable.d.ts:28, :35, :42; lib.es2018.asyncgenerator.d.ts:19 | A done result's `value` and a generator's return reach a program only through `next()`; generators are NotYet |
| The defaults `V = any` on WeakMap and `T = any` on Set's constructor | lib.es2015.collection.d.ts:85, :119; lib.es2015.iterable.d.ts:187 | Counted |
| `PromiseWithResolvers.reject` | lib.es2024.promise.d.ts:20 | An In position; its value comes back out as a rejection's reason, counted above |

## Not value-carrying

| Positions | Where | Why no value comes of it |
| --- | --- | --- |
| Constraints: `A extends any[]`, `T extends (...args: any) => any`, `K extends keyof any`, the decorator contexts' `Class` and `Value` | lib.es5.d.ts:341, :348, :363, :377, :384, :399, :1611, :1628, :1638, :1643, :1648, :1653; lib.es2015.reflect.d.ts:25, :39; lib.decorators.d.ts:43, :81; lib.es2015.symbol.wellknown.d.ts:96, :106 | A constraint bounds a type argument and is replaced by it; it never types a value |
| Patterns in a conditional's `extends` clause: `Awaited`'s `then`, `ThisParameterType`, `Parameters` and `ConstructorParameters`' `=> any` | lib.es5.d.ts:322, :1569, :1570, :1638, :1643 | They match a shape and bind an `infer`, and the bound type is what flows |
| `Intl.DateTimeFormatPartTypesRegistry`'s twelve members | lib.es2017.intl.d.ts:19 to :29; lib.es2021.intl.d.ts:19 | A registry read only through `keyof`, never as a value |

## The list the walk reads

The relation walk carries no exemption of its own. It reads the modelled entries (M1 to M3) by their declarations, and a correction needs no entry, since after it lands there is no `any` left to count. A test enumerates this file's modelled ids against the walk's list, and fails on an entry with no pin, or a pin that doesn't go red with its rule removed. No entry lands without Language's read.
