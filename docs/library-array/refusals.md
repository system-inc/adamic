# Array refusals before implementation

Linux baseline: Adamic `5d4c801`, test262 `c8c798898646638cd0c24879f8e0374e847e7d74`, Node 24.19.0, TypeScript package 6.0.3. Adaptation enabled. Observations below were recorded before implementation.

| pass | disagreement | refused | crashed | skipped | total |
|---:|---:|---:|---:|---:|---:|
| 125 | 0 | 2339 | 8 | 610 | 3082 |

The fresh baseline has eight crashes, all clang rejecting the same heterogeneous number/object array representation in indexOf or lastIndexOf. The earlier report had one crash. The first is `prototype/indexOf/15.4.4.14-5-10.js`; reproduce it alone before changing code.

## Method

`go run ./cmd/adamic-test262 -adapt -json -test262 /workspace/test262 -work /tmp/library-array-before built-ins/Array` wrote `/tmp/library-array-before.json` and `/tmp/library-array-before.log`. No runner files were edited. A scratch copy of its adapt.go, classify.go, frontmatter.go, prelude.go and rewrite.go exported its exact Program text. The real npm typescript package checked all 2,472 attempted inputs, including the entire adapted harness. Adamic separately checked each identical file.

Compiler options exactly follow internal/load/load.go: strict, noUncheckedIndexedAccess, exactOptionalPropertyTypes, noImplicitReturns, noFallthroughCasesInSwitch, erasableSyntaxOnly, verbatimModuleSyntax, allowImportingTsExtensions, noEmit; ESNext modules, forced module detection, bundler resolution, ES2024 target, lib.es2024.d.ts, empty types, plus internal/load/prelude.d.ts. Each input is an independent forced module; batches share only library declarations. No skipLibCheck or source suppressions were added.

Bucket (a) requires an actual matching diagnostic code in both compilers, not just rejection by both. No code mismatches were observed. Counts use the first Adamic blocker; a library blocker may mask additional language blockers. Removing it does not promise the test will pass. The per-test ledger in refusals.json records every refusal, its bucket, named feature, diagnostic codes and SHA256 of its adapted input. before.json preserves the runner table and every reason.

| bucket | count | meaning |
|---|---:|---|
| a | 1845 | real tsc rejects with matching code |
| b | 282 | real tsc accepts; missing library interface or overload |
| c | 212 | real tsc accepts; missing or refused language representation |

## Bucket (a) top features

| count | feature |
|---:|---|
| 547 | TS7006 |
| 412 | TS2345 |
| 269 | TS2339 |
| 162 | TS2554 |
| 111 | TS2683 |
| 89 | TS7009 |
| 85 | TS7005 |
| 53 | TS2769 |
| 23 | TS2322 |
| 18 | TS2790 |
| 18 | TS7053 |
| 12 | TS2538 |
| 9 | TS2403 |
| 9 | TS2367 |
| 6 | TS2304 |
| 5 | TS2542 |
| 4 | TS2555 |
| 4 | TS2551 |
| 2 | TS2454 |
| 2 | TS7030 |
| 2 | TS18048 |
| 2 | TS2378 |
| 1 | TS18046 |

Top observed Adamic diagnostic messages, with source locations removed and quoted names normalized:

| count | reason |
|---:|---|
| 828 | TS7006: Parameter '…' implicitly has an '…' type. |
| 238 | TS2345: Argument of type '…' is not assignable to parameter of type '…'. |
| 177 | TS2339: Property '…' does not exist on type '…'. |
| 135 | TS7009: '…' expression, whose target lacks a construct signature, implicitly has an '…' type. |
| 70 | TS7005: Variable '…' implicitly has an '…' type. |
| 53 | TS2769: No overload matches this call. |
| 32 | TS7053: Element implicitly has an '…' type because expression of type '…' can'…'{}'. |
| 27 | TS2554: Expected 2-3 arguments, but got 1. |
| 27 | TS2554: Expected 3 arguments, but got 2. |
| 24 | TS7034: Variable '…' implicitly has type '…' in some locations where its type cannot be determined. |
| 22 | TS7053: Element implicitly has an '…' type because expression of type '…' can'…'Object'. |
| 21 | TS2554: Expected 1-2 arguments, but got 0. |

## Bucket (b) top features

| count | feature |
|---:|---|
| 51 | generic Array.prototype.indexOf.call |
| 50 | Array or boxed primitive constructors |
| 49 | generic Array.prototype.lastIndexOf.call |
| 18 | Array.isArray observations |
| 17 | built-in prototype observations |
| 9 | Array.prototype.reduceRight |
| 5 | Array constructor calls and aliases |
| 5 | Array.toString or prototype method observations |
| 4 | generic Array.prototype.copyWithin.call |
| 4 | default sort and comparator coercion |
| 4 | generic Array.prototype.toReversed.call |
| 4 | generic Array.prototype.toSorted.call |
| 3 | generic Array.prototype.pop.call |
| 3 | reduce without initial value |
| 3 | generic Array.prototype.slice.call |
| 3 | generic Array.prototype.toSpliced.call |
| 2 | Array.isArray or prototype method observations |
| 2 | Array.of or prototype method observations |
| 2 | generic Array.prototype.find.call |
| 2 | generic Array.prototype.findIndex.call |
| 2 | Array.findIndex or prototype method observations |
| 2 | generic Array.prototype.findLast.call |
| 2 | generic Array.prototype.findLastIndex.call |
| 2 | Array.findLastIndex or prototype method observations |
| 2 | Array.forEach or prototype method observations |
| 2 | generic Array.prototype.push.call |
| 2 | generic Array.prototype.shift.call |
| 2 | generic Array.prototype.unshift.call |
| 1 | Array.from iterable overload |
| 1 | Array.concat or prototype method observations |
| 1 | generic Array.prototype.concat.call |
| 1 | copyWithin explicit undefined end |
| 1 | generic Array.prototype.entries.call |
| 1 | Array.every or prototype method observations |
| 1 | generic Array.prototype.every.call |
| 1 | generic Array.prototype.fill.call |
| 1 | generic Array.prototype.filter.call |
| 1 | Array.find or prototype method observations |
| 1 | Array.findLast or prototype method observations |
| 1 | Array.flatMap or prototype method observations |
| 1 | Array.includes or prototype method observations |
| 1 | Array.join or prototype method observations |
| 1 | generic Array.prototype.join.call |
| 1 | generic Array.prototype.keys.call |
| 1 | generic Array.prototype.map.call |
| 1 | generic Array.prototype.reduce.call |
| 1 | generic Array.prototype.reduceRight.call |
| 1 | Object.freeze on arrays (outside Array territory) |
| 1 | generic Array.prototype.reverse.call |
| 1 | generic Array.prototype.some.call |
| 1 | generic Array.prototype.splice.call |
| 1 | generic Array.prototype.toString.call |
| 1 | generic Array.prototype.values.call |
| 1 | generic Array.prototype.with.call |

Top observed Adamic reasons:

| count | reason |
|---:|---|
| 51 | Adamic 0.1 refuses a method read as a value (indexOf would lose its object, and this with it); call it in an arrow that keeps the object: (searchElement, fromIndex) => its object.indexOf(searchElement, fromIndex) (unbound-method) |
| 50 | stage 0 can't lower new an Identifier yet |
| 49 | Adamic 0.1 refuses a method read as a value (lastIndexOf would lose its object, and this with it); call it in an arrow that keeps the object: (searchElement, fromIndex) => its object.lastIndexOf(searchElement, fromIndex) (unbound-method) |
| 18 | Adamic 0.1 refuses inherited library member isArray read as an own field; prototype members are not stored in an object's shape; call the method on its receiver, or wrap that call in an arrow (unbound-method) |
| 17 | Adamic 0.1 refuses inherited library member prototype read as an own field; prototype members are not stored in an object's shape; call the method on its receiver, or wrap that call in an arrow (unbound-method) |
| 9 | Adamic 0.1 refuses inherited library member reduceRight read as an own field; prototype members are not stored in an object's shape; call the method on its receiver, or wrap that call in an arrow (unbound-method) |
| 6 | Adamic 0.1 refuses a method read as a value (toString would lose its object, and this with it); call it in an arrow that keeps the object: () => its object.toString() (unbound-method) |
| 5 | stage 0 can't lower Array as a value outside equality or typeof (overloaded calls and static properties need their own representation) yet |
| 4 | Adamic 0.1 refuses a method read as a value (copyWithin would lose its object, and this with it); call it in an arrow that keeps the object: (target, start, end) => its object.copyWithin(target, start, end) (unbound-method) |
| 4 | Adamic 0.1 refuses a method read as a value (findIndex would lose its object, and this with it); call it in an arrow that keeps the object: (predicate, thisArg) => its object.findIndex(predicate, thisArg) (unbound-method) |
| 4 | Adamic 0.1 refuses a method read as a value (findLastIndex would lose its object, and this with it); call it in an arrow that keeps the object: (predicate, thisArg) => its object.findLastIndex(predicate, thisArg) (unbound-method) |
| 4 | Adamic 0.1 refuses a method read as a value (toReversed would lose its object, and this with it); call it in an arrow that keeps the object: () => its object.toReversed() (unbound-method) |

## Bucket (c) top features

| count | feature |
|---:|---|
| 37 | truthiness |
| 36 | empty or heterogeneous array representation |
| 26 | property deletion |
| 24 | property descriptors and prototype mutation |
| 18 | sparse array literals |
| 18 | arguments object |
| 14 | unproven any, void, undefined or never values |
| 13 | closure reads before local registration |
| 6 | var declarations |
| 6 | unrepresented global values |
| 4 | getters |
| 3 | statement forms |
| 2 | assignment through a temporary |
| 2 | mixed string and number operators |
| 1 | in operator |
| 1 | void operator |
| 1 | generic array element representation |

Top observed Adamic reasons:

| count | reason |
|---:|---|
| 26 | Adamic 0.1 refuses delete; an object's shape is fixed; use a Map for keys that come and go |
| 23 | stage 0 can't lower an array of never yet |
| 20 | Adamic 0.1 refuses Object.defineProperty; property descriptors can change the presence, type or access behavior of fields; Adamic fields have a fixed shape and are plain loads and stores |
| 18 | stage 0 can't lower an OmittedExpression in an array literal yet |
| 18 | Adamic 0.1 refuses arguments; name the parameters, or take a rest parameter |
| 14 | Adamic 0.1 refuses a filter callback that doesn't return a boolean; return a comparison, like word.length > 0: 0.1 has no truthiness |
| 11 | Adamic 0.1 refuses a every callback that doesn't return a boolean; return a comparison, like word.length > 0: 0.1 has no truthiness |
| 11 | Adamic 0.1 refuses a some callback that doesn't return a boolean; return a comparison, like word.length > 0: 0.1 has no truthiness |
| 9 | stage 0 can't lower reading arr yet |
| 7 | stage 0 can't lower a value of type any yet |
| 6 | Adamic 0.1 refuses var; use const or let |
| 4 | Adamic 0.1 refuses isPrototypeOf; Adamic has no observable prototype chain; use instanceof for class identity or an explicit discriminant |

## Claim and build order

1. Resolve the eight heterogeneous-array search crashes first: refuse unproven slot representations before emitting C, with a small oracle reproducer.
2. Build the largest library family: generic indexOf.call (51 refused tests) and lastIndexOf.call (49). Cover dense arrays, strings, fixed plain array-like shapes and primitive bounds. Preserve HasProperty, strict equality, ToLength and relative indexes. Refuse descriptors, structural views with hidden fields, dynamic coercion and boxed receivers until their representations exist.
3. Address Array.isArray metadata/observations next where the existing identity and fixed representations permit it. Then close explicit undefined copyWithin bounds if time permits. Every newly passing test must match Node.

This is a claim of the named features, not a claim that all 100 generic-search tests can pass: some also require property descriptors or boxed objects. Bucket (c) remains owned by @system_adamic.

## Language handoff to @system_adamic

| feature | one-line reproducer |
|---|---|
| empty or heterogeneous array representation | `const empty = []; empty.length;` |
| property descriptors and prototype mutation | `Object.defineProperty({x: 1}, "x", {value: 2});` |
| var declarations | `var n = 1;` |
| statement forms | `new Array(1);` |
| sparse array literals | `const sparse = [1, , 3]; sparse.indexOf(3);` |
| assignment through a temporary | `({x: 1}).x = 2;` |
| closure reads before local registration | `function f() { return a.length; } let a = [1]; f();` |
| property deletion | `const a = [1]; delete a[0];` |
| arguments object | `function f() { return arguments.length; }` |
| truthiness | `[1].some(() => 1);` |
| unrepresented global values | `const f = JSON; f.stringify(1);` |
| unproven any, void, undefined or never values | `const a = [1].map(() => {}); a.length;` |
| mixed string and number operators | `const text = "x" + 1;` |
| in operator | `"x" in {x: 1};` |
| void operator | `void 1;` |
| getters | `const x = {get length() { return 1; }};` |
| generic array element representation | `function f<T>(a: T[]) { return a.toSorted(); }` |

These reproducers name the language blocker; full adapted test paths and reasons are in the ledger. Empty arrays, void/never callback results and hoisted closures are distinct from missing Array methods. Property descriptors and deletion require a presence/prototype model, not an approximation in this slice.

## Setup

Required bash cloud/setup.sh succeeded: go ready (0s); clang ready (1s); node ready (1s); submodules ready (1s); build cache warm (100s); done in 100s on 5 processors, cgroup cpu.max 400000 100000, 17.6 GB. Sourced /workspace/adamic-tools/env.sh. nproc: 5.

## Language blockers exposed while building

The crash reproducer is `const target = {}; [0, target, 2].indexOf(target, 2);`. TypeScript infers an object element type that also admits primitives; native slots require a tagged representation. The reduced negative oracle is in `internal/oracle/testdata/library_array_refused/library_array_heterogeneous.a` and is exercised explicitly by the Array lowering test. The subdirectory follows `fresh_refused/`: top-level oracle files are globbed as positive graph fixtures.

A fixed object view may hide fields: `const full = {0: 1, 2: 9, length: 3}; const view: {0: number; length: number} = full; Array.prototype.indexOf.call(view, 9);`. Search therefore requires a plain literal or its unannotated, unreassigned binding. Structural views need a presence/shape proof before searches can use them.

`const value: {} = [1]; Array.isArray(value);` needs a dynamic object tag; the object type alone cannot prove false. Optional array representations are also refused rather than guessed.

The 50 constructor refusals include length constructors and boxed primitives. Dense native arrays cannot represent the absent elements of `Array(3)`; boxed primitives require runtime object identities and prototype behavior. Extending those representations requires shared IR/emission work outside this unit's allowed dispatch hooks. These remain explicit refusals.

Fixed-shape generic searches also refuse optional/nullable or mixed search representations and field values before strict comparison. For example, `function f(x: number | undefined) { const a = {0: x, length: 1}; return Array.prototype.indexOf.call(a, undefined); }` needs a comparison that preserves the undefined tag. An object field initialized from a structural array view is refused: `const a = [1]; const v: {length: number} = a; Array.prototype.indexOf.call({0: v, length: 1}, a);`. Pure literal object identities remain supported.
