# Index signatures: one object, two stores

Proposal for @system_adamic, from @system_adamic_typescript, October 6. It follows the escape-hatch ruling on #bw3xg7c: index signatures are NotYet, not refused, and they need a sound map-backed representation because stage 3 compiles tsc's source unchanged. This is design only. Nothing here is built.

## What tsc's src/compiler actually needs (TypeScript 6.0.3, grep counts, observed)

These counts come from grep over the pinned source, not from the checker, so they're a floor and a rough shape. The census will give the type-aware numbers.

- **Pure records.** `MapLike<T>` (`corePublic.ts:13`, `[index: string]: T`) appears 60 times, `Record<string, T>` 5 times, inline `{ [k: string]: T }` 13 times. Uses: `paths`, `wildcardDirectories`, package.json `dependencies`, `exports`, `imports` and `typesVersions`, localized diagnostic messages, `textToKeywordObj` (`scanner.ts:135`).
- **Mixed shapes**, the important case. `CompilerOptions` (`types.ts:7571`), `WatchOptions`, `TypeAcquisition`, `BuildOptions` (`tsbuildPublic.ts:166`) and tracing's `Args` each declare dozens of named optional fields **and** `[option: string]: ...`. `commandLineParser.ts:2050-2111` writes them by dynamic key (`options[opt.name] = value`, including `= undefined`), and the rest of the compiler reads them by name (`options.strict`). The same object is reached both ways.
- **JSON-born records.** package.json and tsconfig content arrive through `JSON.parse` (`utilities.ts:7811`, `sourcemap.ts:424`, `utilitiesPublic.ts:749`) and get cast to `MapLike<unknown>` in `moduleNameResolver.ts` and `moduleSpecifiers.ts`.
- **Operations.** 38 `for...in`, 32 `hasProperty` (which is `hasOwnProperty.call`, `core.ts:1266`), 13 `getOwnKeys`, 75 `Object.entries` (57 in `utilities.ts`, mostly `new Map(Object.entries({...}))` over literals), 2 real `delete record[key]` (`commandLineParser.ts:4159`, `transformers/module/system.ts:1779`), 1 `Object.create(null)` (`factory/emitHelpers.ts:1480`).

## JavaScript behavior the representation must keep (Node 24, checked)

1. Key order is array-index keys ascending (canonical decimal, below 2^32 - 1), then every other string key in creation order. `{b, 10, a, 2, 01}` enumerates as `2, 10, b, a, 01`. Map's insertion order alone gets this wrong. `library_object.c`'s `ordered` already has the rule for shapes.
2. A plain record inherits `Object.prototype`. `({})["constructor"]` is a function, `"toString" in {}` is true, and `Object.hasOwn({}, "toString")` is false. TypeScript types that first read as `T | undefined`. That's a type lie in TypeScript itself, and tsc guards against it with `hasProperty`.
3. Writing `record["__proto__"] = x` sets the prototype and adds no own key. `JSON.parse('{"__proto__": 7}')` makes an own key named `__proto__`.
4. Deleting a key and adding it again appends it. Absent and present-undefined are different (`in`, `hasOwn`, keys, spread).

## tsc's assignability that forces the shape of the design (tsc 6.0.3, strict, exactOptionalPropertyTypes, noUncheckedIndexedAccess, checked)

- A **type alias's** object is assignable to `Record<string, number>` (an implicit index signature), and an **interface's** isn't. So a plain object built from a literal can later be read and *written* by dynamic key through a record view.
- `Record<string, number>` is assignable to `{ a?: number }`. So a static read of an optional field can land on an object whose value lives only in the dictionary. On today's optional-field path (`adamic_object_optional_field` misses, returns absent) that's a **silent miscompile**: the read gives undefined where Node gives the value. Any design must close this.
- A literal-key write of a wider value into a named field is rejected (`mixed["strict"] = "yes"`), but a *dynamic* key isn't: `options[opt.name] = value` with `CompilerOptionsValue` can put a string where `strict?: boolean` is declared. That's TypeScript's own unsoundness, and tsc's source does exactly that write.

Conversion at the boundary (copy the object into a Map when it's seen as a record) is ruled out: writes through the record view have to be visible through the original, and the original through the record.

## The representation

**Every plain object can carry a dictionary.** `adamic_object` gains `adamic_dictionary *dictionary`, which stays NULL until a write lands on a key the shape doesn't hold. Class instances need none, since an interface or class type has no implicit index signature (to be confirmed for class instances seen through a type alias). The shape keeps its fixed slots and its fast cached reads. The dictionary is map.c's ordered, string-keyed table, with tombstones and live iteration as Map has them.

- **Dictionary values are Unions** (`ir.Union`, self-describing, boxed numbers) in the first version, so a value written through one static type can be read through another and checked. A homogeneous specialization (as Map's `reference_values`) comes after measurement, not before.
- **Named reads stay fast.** A required field is unchanged. A present optional field is unchanged (shape cache hit). An *absent* optional field, the miss path, tests `object->dictionary` before answering absent: one load and a NULL test, only on a miss. This is the fix for the record-to-optional-field miscompile, and it's what lets a dynamic write to an absent optional field (`options["strict"] = true` on `{} as CompilerOptions`) be seen by `options.strict`.
- **Dynamic reads `o[k]`** (static type has a string index signature; the result is `T | undefined`, which Adamic's noUncheckedIndexedAccess already forces): the shape by name (a per-shape name hash, built once and cached by shape pointer, replacing today's linear strcmp), then the dictionary. On a miss, if `k` names an `Object.prototype` member (`constructor`, `toString`, `valueOf`, `hasOwnProperty`, `isPrototypeOf`, `propertyIsEnumerable`, `toLocaleString`, `__proto__`, `__defineGetter__`, `__defineSetter__`, `__lookupGetter__`, `__lookupSetter__`), it **panics loudly**: `a string-keyed record has no own 'toString'; JavaScript would give Object.prototype's member, which isn't its value type`. Otherwise the answer is undefined. The check costs nothing on a hit.
- **`k in o`** matches JavaScript exactly, prototype names included (a boolean is representable). **`hasOwnProperty.call(o, k)` and `Object.hasOwn`** mean shape or dictionary, own only.
- **Dynamic writes `o[k] = v`.** A named slot takes the value through a checked conversion into the slot's representation, and a value the slot can't hold panics loudly (`'strict' is declared boolean | undefined; this write holds a string`). That's TypeScript's own lie, caught at the write. Any other key goes to the dictionary. `__proto__` panics loudly (it's refused at compile time when written as a literal).
- **`delete o[k]`** is allowed only through an index-signature type. In the dictionary, it leaves a tombstone. On a named slot it panics loudly in the first version (`deleting a declared field`), because shapes are fixed. Both deletes in tsc hit dictionary keys.
- **Enumeration** (`for...in`, `Object.keys`, `values`, `entries`, `getOwnKeys`, spread, `JSON.stringify`) merges the two stores: array-index keys ascending across both, then the shape's string keys in shape order, then the dictionary's in insertion order. That's exactly JavaScript's order, because shape fields exist from construction, before any dictionary key, and an optional field first written later lands in the dictionary, appended, as JavaScript appends it.
- **Literals contextually typed by a record** (`textToKeywordObj`) are built with their literal keys in the shape, as today, and the dictionary grows only if a later write needs it.
- **`Object.fromEntries`** becomes supported (it's refused today, citing index signatures), and so does `JSON.parse` into a record, once `unknown` has a runtime representation (a separate question the census will raise).
- **Number index signatures** (`[i: number]: T`) would canonicalize keys through ToString(number) into the same dictionary. tsc's src/compiler declares none, so they stay NotYet.

## Memory, no collector

The dictionary is owned by its object and freed with it. Keys are retained strings, and values are retained Unions. A record whose value type can reach the record's own type is a cycle-capable mutable slot, under the same `weak`-or-refuse rule as every other slot (#w318mqs). tsc's records hold strings, arrays and option values, so they're expected to be acyclic, and the census will confirm it. `Object.freeze` already marks the object, and `adamic_object_check_write` covers dictionary writes too. A record is mutable, so it isn't shareable across cores unless it's frozen.

## Cost

- +8 bytes on every plain object for the dictionary pointer, measured by the honest benchmarks (#0aja3b1) before and after, with RSS and retain/release counts.
- One NULL test on absent optional-field reads.
- A hash lookup (not a strcmp scan) for dynamic reads of shaped objects.
- Boxing for number values in dictionaries until specialization.

## Decisions I'm bringing you

1. **The pointer on every plain object** (recommended) vs. refusing the type-alias-object-to-record edge in cohere:adamic, which saves the 8 bytes but needs the relation found at every implicit conversion site (calls, returns, spreads, generics), and leaves the `Record` to optional-field edge to close separately anyway. I'd rather pay the 8 bytes and measure them.
2. **Inherited-key reads panic loudly** (recommended) vs. returning `Object.prototype`'s member, which isn't representable as `T`.
3. **The dynamic write of a wrong-typed value into a named slot panics loudly at the write** (recommended), not at a later read.
4. **Union-boxed dictionary values first**, then specialization after measurement.

## How it would be built, once you rule

Four units, one worker each, all held by the Node oracle:

1. The dictionary runtime (`runtime/dictionary.c`) with JavaScript key order and the `__proto__` rule.
2. Lowering element access, `in`, `delete` and `for...in` on index-signature types.
3. The optional-field miss path and the shape name hash.
4. Merged enumeration, with `Object.keys`, `values`, `entries` and `fromEntries` on records.

The mutants that must be caught:

- Enumerate the dictionary in pure insertion order: the key-order fixture.
- Skip the dictionary on an optional-field miss: the `Record` to `{ a?: number }` fixture.
- Drop the inherited-key check: the loud-check fixture, under the convention the `!` check uses.
- Drop the named-slot conversion check: the CompilerOptions-shaped dynamic-write fixture.
- Free a dictionary value early: ASan on a fixture whose values are strings the program builds.

Rulings (@system_adamic, October 6): yes to all four. The pointer goes on every plain object for now, then a later unit drops it from shapes the whole-program type graph proves never reach a record slot. Inherited-key reads panic naming the key. A wrong-typed dynamic write panics at the write, naming the slot and both types. Values start as union boxes.
