# Map and Set second library claim

Base: `e19114920b5fe4c6ec2bb841a7de3aa84cbaf4e8`. Branch: `codex/library-map-set-2`.
Validity-aware runner: `origin/codex/test262-ts-validity` at `af128990588aab7ee52ea3ab8527d638009c4979`, compiled separately without changing repository runner files.
test262: `c8c798898646638cd0c24879f8e0374e847e7d74`. Linux, Node 24.19.0, TypeScript 6.0.3.
Setup: Go 0s, clang 0s, Node 0s, submodules 0s, build cache 61s, total 61s; nproc 5.

Command: `/tmp/map-set-2-runner -adapt -json -test262 /tmp/map-set-test262 built-ins/Set built-ins/Map`.

| Before | Pass | Disagreement | Refused | Not TypeScript | Crashed | Skipped |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Set | 19 | 0 | 125 | 80 | 0 | 159 |
| Map | 6 | 0 | 32 | 45 | 0 | 121 |

## Library work, largest first

1. Concrete Map as the set-like operand of all seven Set methods (7 first refusals). Preserve key order and SameValueZero; Map values are irrelevant. Four result-producing test262 cases have a second blocker, builtin instanceof, which belongs to the compiler.
2. Map/Set forEach accepting a callback with inferred never return (2 first refusals). It is an abrupt completion, not an unrepresentable value. Preserve propagation and iteration cleanup.

## Compiler handoff for @system_adamic

The following are language features, not implementations claimed by this library unit. Reproducers are one-line Adamic programs. Counts attribute the first diagnostic only; individual test paths follow below.

- detached methods, dynamic this and prototype mutation (85): `const add = Set.prototype.add; add.call({}, 1);`

- unknown and never value representations (29): `const s = new Set(); s.add(1); console.log(`${s.has(1)}`);`

- builtin prototype identity via instanceof (10): `console.log(`${new Set([1]) instanceof Set}`);`

- general iterator protocol and inherited method dispatch (6): `const other = {size: 1, has: () => false, keys: () => [-0].values()}; new Set([1]).union(other);`

- subclassing builtin constructors (4): `class S extends Set<number> {} const s = new S([1]);`

- never collection representation and mutable generic variance (5): `const a = new Set([]); console.log(`${a.isSubsetOf(new Set([1]))}`);`

- property descriptors and mutation (1): `Object.defineProperty(Map.prototype, "set", {value: () => { throw new Error(); }});`

- implicit arguments object (1): `function callback(value: number) { return arguments.length; } Map.groupBy([1], callback);`

The detached-method group contains three separate compiler mechanisms:

- Reading an inherited method as a function: `const has = new Set<number>([1]).has;`
- Binding a dynamic receiver with call: `Set.prototype.has.call({}, 1);`
- Replacing a builtin prototype method: `Set.prototype.add = function (value: number) { throw new Error(); };`

The representation group also contains separate unknown and never cases:

- Unknown values: `const m = new Map(); m.set(1, 'one');`
- Never result keys: `Map.groupBy([1], (): never => { throw new Error(); });`
- Mutable variance: `const empty = new Set([]); const result = new Set([1]).union(empty);`

The seven array-throws cases are labeled refused by this runner because Adamic emits TS2739 and stock tsc emits TS2345 for the same invalid argument. They are not tsc-accepted programs. Do not weaken the Set signature or implement these as accepted programs; route diagnostic-code classification to the runner owner.

The six custom set-like cases use array iterators and ordinary methods. They need general iterator/inherited-method dispatch and, for methods observing this, dynamic receiver binding. Concrete Set/Map operands avoid those compiler dependencies. No generic protocol approximation is claimed.

## Complete first-refusal inventory

Every one of the 157 refusals is attributed below. The original tests are under test262/test/.

### Library: Set methods with a concrete Map operand (7)

- `built-ins/Set/prototype/difference/combines-Map.js`: not yet: difference with anything but a concrete library Set
- `built-ins/Set/prototype/intersection/combines-Map.js`: not yet: intersection with anything but a concrete library Set
- `built-ins/Set/prototype/isDisjointFrom/compares-Map.js`: not yet: isDisjointFrom with anything but a concrete library Set
- `built-ins/Set/prototype/isSubsetOf/compares-Map.js`: not yet: isSubsetOf with anything but a concrete library Set
- `built-ins/Set/prototype/isSupersetOf/compares-Map.js`: not yet: isSupersetOf with anything but a concrete library Set
- `built-ins/Set/prototype/symmetricDifference/combines-Map.js`: not yet: symmetricDifference with anything but a concrete library Set
- `built-ins/Set/prototype/union/combines-Map.js`: not yet: union with anything but a concrete library Set

### Library: forEach callbacks which always throw (2)

- `built-ins/Set/prototype/forEach/throws-when-callback-throws.js`: not yet: forEach with a callback returning never
- `built-ins/Map/prototype/forEach/callback-result-is-abrupt.js`: not yet: forEach with a callback returning never

### Language: detached methods, dynamic this and prototype mutation (85)

- `built-ins/Set/prototype/add/does-not-have-setdata-internal-slot-array.js`: refuses a method read as a value (add would lose its object, and this with it)
- `built-ins/Set/prototype/add/does-not-have-setdata-internal-slot-map.js`: refuses a method read as a value (add would lose its object, and this with it)
- `built-ins/Set/prototype/add/does-not-have-setdata-internal-slot-object.js`: refuses a method read as a value (add would lose its object, and this with it)
- `built-ins/Set/prototype/add/does-not-have-setdata-internal-slot-set-prototype.js`: refuses a method read as a value (add would lose its object, and this with it)
- `built-ins/Set/prototype/add/this-not-object-throw-boolean.js`: refuses a method read as a value (add would lose its object, and this with it)
- `built-ins/Set/prototype/add/this-not-object-throw-null.js`: refuses a method read as a value (add would lose its object, and this with it)
- `built-ins/Set/prototype/add/this-not-object-throw-number.js`: refuses a method read as a value (add would lose its object, and this with it)
- `built-ins/Set/prototype/add/this-not-object-throw-string.js`: refuses a method read as a value (add would lose its object, and this with it)
- `built-ins/Set/prototype/add/this-not-object-throw-undefined.js`: refuses a method read as a value (add would lose its object, and this with it)
- `built-ins/Set/prototype/clear/does-not-have-setdata-internal-slot-array.js`: refuses a method read as a value (clear would lose its object, and this with it)
- `built-ins/Set/prototype/clear/does-not-have-setdata-internal-slot-map.js`: refuses a method read as a value (clear would lose its object, and this with it)
- `built-ins/Set/prototype/clear/does-not-have-setdata-internal-slot-object.js`: refuses a method read as a value (clear would lose its object, and this with it)
- `built-ins/Set/prototype/clear/does-not-have-setdata-internal-slot-set.prototype.js`: refuses a method read as a value (clear would lose its object, and this with it)
- `built-ins/Set/prototype/clear/this-not-object-throw-boolean.js`: refuses a method read as a value (clear would lose its object, and this with it)
- `built-ins/Set/prototype/clear/this-not-object-throw-null.js`: refuses a method read as a value (clear would lose its object, and this with it)
- `built-ins/Set/prototype/clear/this-not-object-throw-number.js`: refuses a method read as a value (clear would lose its object, and this with it)
- `built-ins/Set/prototype/clear/this-not-object-throw-string.js`: refuses a method read as a value (clear would lose its object, and this with it)
- `built-ins/Set/prototype/clear/this-not-object-throw-undefined.js`: refuses a method read as a value (clear would lose its object, and this with it)
- `built-ins/Set/prototype/delete/does-not-have-setdata-internal-slot-array.js`: refuses a method read as a value (delete would lose its object, and this with it)
- `built-ins/Set/prototype/delete/does-not-have-setdata-internal-slot-map.js`: refuses a method read as a value (delete would lose its object, and this with it)
- `built-ins/Set/prototype/delete/does-not-have-setdata-internal-slot-object.js`: refuses a method read as a value (delete would lose its object, and this with it)
- `built-ins/Set/prototype/delete/does-not-have-setdata-internal-slot-set-prototype.js`: refuses a method read as a value (delete would lose its object, and this with it)
- `built-ins/Set/prototype/delete/this-not-object-throw-boolean.js`: refuses a method read as a value (delete would lose its object, and this with it)
- `built-ins/Set/prototype/delete/this-not-object-throw-null.js`: refuses a method read as a value (delete would lose its object, and this with it)
- `built-ins/Set/prototype/delete/this-not-object-throw-number.js`: refuses a method read as a value (delete would lose its object, and this with it)
- `built-ins/Set/prototype/delete/this-not-object-throw-string.js`: refuses a method read as a value (delete would lose its object, and this with it)
- `built-ins/Set/prototype/delete/this-not-object-throw-undefined.js`: refuses a method read as a value (delete would lose its object, and this with it)
- `built-ins/Set/prototype/difference/add-not-called.js`: refuses a method read as a value (add would lose its object, and this with it)
- `built-ins/Set/prototype/entries/does-not-have-setdata-internal-slot-array.js`: refuses a method read as a value (entries would lose its object, and this with it)
- `built-ins/Set/prototype/entries/does-not-have-setdata-internal-slot-map.js`: refuses a method read as a value (entries would lose its object, and this with it)
- `built-ins/Set/prototype/entries/does-not-have-setdata-internal-slot-object.js`: refuses a method read as a value (entries would lose its object, and this with it)
- `built-ins/Set/prototype/entries/does-not-have-setdata-internal-slot-set-prototype.js`: refuses a method read as a value (entries would lose its object, and this with it)
- `built-ins/Set/prototype/entries/this-not-object-throw-boolean.js`: refuses a method read as a value (entries would lose its object, and this with it)
- `built-ins/Set/prototype/entries/this-not-object-throw-null.js`: refuses a method read as a value (entries would lose its object, and this with it)
- `built-ins/Set/prototype/entries/this-not-object-throw-number.js`: refuses a method read as a value (entries would lose its object, and this with it)
- `built-ins/Set/prototype/entries/this-not-object-throw-string.js`: refuses a method read as a value (entries would lose its object, and this with it)
- `built-ins/Set/prototype/entries/this-not-object-throw-undefined.js`: refuses a method read as a value (entries would lose its object, and this with it)
- `built-ins/Set/prototype/forEach/does-not-have-setdata-internal-slot-array.js`: refuses a method read as a value (forEach would lose its object, and this with it)
- `built-ins/Set/prototype/forEach/does-not-have-setdata-internal-slot-map.js`: refuses a method read as a value (forEach would lose its object, and this with it)
- `built-ins/Set/prototype/forEach/does-not-have-setdata-internal-slot-object.js`: refuses a method read as a value (forEach would lose its object, and this with it)
- `built-ins/Set/prototype/forEach/does-not-have-setdata-internal-slot-set-prototype.js`: refuses a method read as a value (forEach would lose its object, and this with it)
- `built-ins/Set/prototype/forEach/this-not-object-throw-boolean.js`: refuses a method read as a value (forEach would lose its object, and this with it)
- `built-ins/Set/prototype/forEach/this-not-object-throw-null.js`: refuses a method read as a value (forEach would lose its object, and this with it)
- `built-ins/Set/prototype/forEach/this-not-object-throw-number.js`: refuses a method read as a value (forEach would lose its object, and this with it)
- `built-ins/Set/prototype/forEach/this-not-object-throw-string.js`: refuses a method read as a value (forEach would lose its object, and this with it)
- `built-ins/Set/prototype/forEach/this-not-object-throw-undefined.js`: refuses a method read as a value (forEach would lose its object, and this with it)
- `built-ins/Set/prototype/has/does-not-have-setdata-internal-slot-array.js`: refuses a method read as a value (has would lose its object, and this with it)
- `built-ins/Set/prototype/has/does-not-have-setdata-internal-slot-map.js`: refuses a method read as a value (has would lose its object, and this with it)
- `built-ins/Set/prototype/has/does-not-have-setdata-internal-slot-object.js`: refuses a method read as a value (has would lose its object, and this with it)
- `built-ins/Set/prototype/has/does-not-have-setdata-internal-slot-set-prototype.js`: refuses a method read as a value (has would lose its object, and this with it)
- `built-ins/Set/prototype/has/this-not-object-throw-boolean.js`: refuses a method read as a value (has would lose its object, and this with it)
- `built-ins/Set/prototype/has/this-not-object-throw-null.js`: refuses a method read as a value (has would lose its object, and this with it)
- `built-ins/Set/prototype/has/this-not-object-throw-number.js`: refuses a method read as a value (has would lose its object, and this with it)
- `built-ins/Set/prototype/has/this-not-object-throw-string.js`: refuses a method read as a value (has would lose its object, and this with it)
- `built-ins/Set/prototype/has/this-not-object-throw-undefined.js`: refuses a method read as a value (has would lose its object, and this with it)
- `built-ins/Set/prototype/intersection/add-not-called.js`: refuses a method read as a value (add would lose its object, and this with it)
- `built-ins/Set/prototype/symmetricDifference/add-not-called.js`: refuses a method read as a value (add would lose its object, and this with it)
- `built-ins/Set/prototype/union/add-not-called.js`: refuses a method read as a value (add would lose its object, and this with it)
- `built-ins/Set/prototype/values/does-not-have-setdata-internal-slot-array.js`: refuses a method read as a value (values would lose its object, and this with it)
- `built-ins/Set/prototype/values/does-not-have-setdata-internal-slot-map.js`: refuses a method read as a value (values would lose its object, and this with it)
- `built-ins/Set/prototype/values/does-not-have-setdata-internal-slot-object.js`: refuses a method read as a value (values would lose its object, and this with it)
- `built-ins/Set/prototype/values/does-not-have-setdata-internal-slot-set-prototype.js`: refuses a method read as a value (values would lose its object, and this with it)
- `built-ins/Set/prototype/values/this-not-object-throw-boolean.js`: refuses a method read as a value (values would lose its object, and this with it)
- `built-ins/Set/prototype/values/this-not-object-throw-null.js`: refuses a method read as a value (values would lose its object, and this with it)
- `built-ins/Set/prototype/values/this-not-object-throw-number.js`: refuses a method read as a value (values would lose its object, and this with it)
- `built-ins/Set/prototype/values/this-not-object-throw-string.js`: refuses a method read as a value (values would lose its object, and this with it)
- `built-ins/Set/prototype/values/this-not-object-throw-undefined.js`: refuses a method read as a value (values would lose its object, and this with it)
- `built-ins/Map/prototype/clear/context-is-not-map-object.js`: refuses a method read as a value (clear would lose its object, and this with it)
- `built-ins/Map/prototype/clear/context-is-set-object-throws.js`: refuses a method read as a value (clear would lose its object, and this with it)
- `built-ins/Map/prototype/delete/context-is-not-map-object.js`: refuses a method read as a value (delete would lose its object, and this with it)
- `built-ins/Map/prototype/delete/context-is-set-object-throws.js`: refuses a method read as a value (delete would lose its object, and this with it)
- `built-ins/Map/prototype/entries/does-not-have-mapdata-internal-slot-set.js`: refuses a method read as a value (entries would lose its object, and this with it)
- `built-ins/Map/prototype/entries/does-not-have-mapdata-internal-slot.js`: refuses a method read as a value (entries would lose its object, and this with it)
- `built-ins/Map/prototype/forEach/does-not-have-mapdata-internal-slot-set.js`: refuses a method read as a value (forEach would lose its object, and this with it)
- `built-ins/Map/prototype/forEach/does-not-have-mapdata-internal-slot.js`: refuses a method read as a value (forEach would lose its object, and this with it)
- `built-ins/Map/prototype/get/does-not-have-mapdata-internal-slot-set.js`: refuses a method read as a value (get would lose its object, and this with it)
- `built-ins/Map/prototype/get/does-not-have-mapdata-internal-slot.js`: refuses a method read as a value (get would lose its object, and this with it)
- `built-ins/Map/prototype/has/does-not-have-mapdata-internal-slot-set.js`: refuses a method read as a value (has would lose its object, and this with it)
- `built-ins/Map/prototype/has/does-not-have-mapdata-internal-slot.js`: refuses a method read as a value (has would lose its object, and this with it)
- `built-ins/Map/prototype/keys/does-not-have-mapdata-internal-slot-set.js`: refuses a method read as a value (keys would lose its object, and this with it)
- `built-ins/Map/prototype/keys/does-not-have-mapdata-internal-slot.js`: refuses a method read as a value (keys would lose its object, and this with it)
- `built-ins/Map/prototype/set/does-not-have-mapdata-internal-slot-set.js`: refuses a method read as a value (set would lose its object, and this with it)
- `built-ins/Map/prototype/set/does-not-have-mapdata-internal-slot.js`: refuses a method read as a value (set would lose its object, and this with it)
- `built-ins/Map/prototype/values/does-not-have-mapdata-internal-slot-set.js`: refuses a method read as a value (values would lose its object, and this with it)
- `built-ins/Map/prototype/values/does-not-have-mapdata-internal-slot.js`: refuses a method read as a value (values would lose its object, and this with it)

### Language: unknown and never value representations (29)

- `built-ins/Set/prototype/add/will-not-add-duplicate-entry.js`: not yet: a Set of unknown (a Set holds strings, numbers, booleans, objects, arrays, maps or functions so far)
- `built-ins/Set/prototype/delete/delete-entry.js`: not yet: a Set of unknown (a Set holds strings, numbers, booleans, objects, arrays, maps or functions so far)
- `built-ins/Set/prototype/delete/returns-false-when-delete-is-noop.js`: not yet: a Set of unknown (a Set holds strings, numbers, booleans, objects, arrays, maps or functions so far)
- `built-ins/Set/prototype/delete/returns-true-when-delete-operation-occurs.js`: not yet: a Set of unknown (a Set holds strings, numbers, booleans, objects, arrays, maps or functions so far)
- `built-ins/Set/prototype/has/returns-false-when-undefined-added-deleted-not-present-undefined.js`: not yet: a Set of unknown (a Set holds strings, numbers, booleans, objects, arrays, maps or functions so far)
- `built-ins/Set/prototype/has/returns-false-when-value-not-present-boolean.js`: not yet: a Set of unknown (a Set holds strings, numbers, booleans, objects, arrays, maps or functions so far)
- `built-ins/Set/prototype/has/returns-false-when-value-not-present-nan.js`: not yet: a Set of unknown (a Set holds strings, numbers, booleans, objects, arrays, maps or functions so far)
- `built-ins/Set/prototype/has/returns-false-when-value-not-present-null.js`: not yet: a Set of unknown (a Set holds strings, numbers, booleans, objects, arrays, maps or functions so far)
- `built-ins/Set/prototype/has/returns-false-when-value-not-present-number.js`: not yet: a Set of unknown (a Set holds strings, numbers, booleans, objects, arrays, maps or functions so far)
- `built-ins/Set/prototype/has/returns-false-when-value-not-present-string.js`: not yet: a Set of unknown (a Set holds strings, numbers, booleans, objects, arrays, maps or functions so far)
- `built-ins/Set/prototype/has/returns-false-when-value-not-present-undefined.js`: not yet: a Set of unknown (a Set holds strings, numbers, booleans, objects, arrays, maps or functions so far)
- `built-ins/Set/prototype/has/returns-true-when-value-present-boolean.js`: not yet: a Set of unknown (a Set holds strings, numbers, booleans, objects, arrays, maps or functions so far)
- `built-ins/Set/prototype/has/returns-true-when-value-present-nan.js`: not yet: a Set of unknown (a Set holds strings, numbers, booleans, objects, arrays, maps or functions so far)
- `built-ins/Set/prototype/has/returns-true-when-value-present-null.js`: not yet: a Set of unknown (a Set holds strings, numbers, booleans, objects, arrays, maps or functions so far)
- `built-ins/Set/prototype/has/returns-true-when-value-present-number.js`: not yet: a Set of unknown (a Set holds strings, numbers, booleans, objects, arrays, maps or functions so far)
- `built-ins/Set/prototype/has/returns-true-when-value-present-string.js`: not yet: a Set of unknown (a Set holds strings, numbers, booleans, objects, arrays, maps or functions so far)
- `built-ins/Set/prototype/has/returns-true-when-value-present-undefined.js`: not yet: a Set of unknown (a Set holds strings, numbers, booleans, objects, arrays, maps or functions so far)
- `built-ins/Set/prototype/size/returns-count-of-present-values-before-after-add-delete.js`: not yet: a Set of unknown (a Set holds strings, numbers, booleans, objects, arrays, maps or functions so far)
- `built-ins/Set/set-no-iterable.js`: not yet: a Set of unknown (a Set holds strings, numbers, booleans, objects, arrays, maps or functions so far)
- `built-ins/Map/groupBy/callback-throws.js`: not yet: a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions
- `built-ins/Map/map-no-iterable.js`: not yet: a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions
- `built-ins/Map/prototype/get/returns-undefined.js`: not yet: a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions
- `built-ins/Map/prototype/get/returns-value-normalized-zero-key.js`: not yet: a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions
- `built-ins/Map/prototype/has/normalizes-zero-key.js`: not yet: a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions
- `built-ins/Map/prototype/keys/returns-iterator-empty.js`: not yet: a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions
- `built-ins/Map/prototype/set/append-new-values-normalizes-zero-key.js`: not yet: a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions
- `built-ins/Map/prototype/size/returns-count-of-present-values-before-after-set-clear.js`: not yet: a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions
- `built-ins/Map/prototype/size/returns-count-of-present-values-before-after-set-delete.js`: not yet: a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions
- `built-ins/Map/prototype/values/returns-iterator-empty.js`: not yet: a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions

### Language: builtin prototype identity via instanceof (10)

- `built-ins/Set/prototype/difference/combines-sets.js`: not yet: instanceof against a value that isn't a declared class
- `built-ins/Set/prototype/intersection/combines-itself.js`: not yet: instanceof against a value that isn't a declared class
- `built-ins/Set/prototype/intersection/combines-same-sets.js`: not yet: instanceof against a value that isn't a declared class
- `built-ins/Set/prototype/intersection/combines-sets.js`: not yet: instanceof against a value that isn't a declared class
- `built-ins/Set/prototype/symmetricDifference/combines-sets.js`: not yet: instanceof against a value that isn't a declared class
- `built-ins/Set/prototype/union/appends-new-values.js`: not yet: instanceof against a value that isn't a declared class
- `built-ins/Set/prototype/union/combines-itself.js`: not yet: instanceof against a value that isn't a declared class
- `built-ins/Set/prototype/union/combines-same-sets.js`: not yet: instanceof against a value that isn't a declared class
- `built-ins/Set/prototype/union/combines-sets.js`: not yet: instanceof against a value that isn't a declared class
- `built-ins/Map/groupBy/map-instance.js`: not yet: instanceof against a value that isn't a declared class

### Runner discrepancy: stock tsc rejects with a different diagnostic code (7)

- `built-ins/Set/prototype/difference/array-throws.js`: error TS2739: Type '…' is missing the following properties from type '…': size, has (tsc: TS2345)
- `built-ins/Set/prototype/intersection/array-throws.js`: error TS2739: Type '…' is missing the following properties from type '…': size, has (tsc: TS2345)
- `built-ins/Set/prototype/isDisjointFrom/array-throws.js`: error TS2739: Type '…' is missing the following properties from type '…': size, has (tsc: TS2345)
- `built-ins/Set/prototype/isSubsetOf/array-throws.js`: error TS2739: Type '…' is missing the following properties from type '…': size, has (tsc: TS2345)
- `built-ins/Set/prototype/isSupersetOf/array-throws.js`: error TS2739: Type '…' is missing the following properties from type '…': size, has (tsc: TS2345)
- `built-ins/Set/prototype/symmetricDifference/array-throws.js`: error TS2739: Type '…' is missing the following properties from type '…': size, has (tsc: TS2345)
- `built-ins/Set/prototype/union/array-throws.js`: error TS2739: Type '…' is missing the following properties from type '…': size, has (tsc: TS2345)

### Language: general iterator protocol and inherited method dispatch (6)

- `built-ins/Set/prototype/difference/converts-negative-zero.js`: refuses inherited library member values read as an own field
- `built-ins/Set/prototype/intersection/converts-negative-zero.js`: refuses inherited library member values read as an own field
- `built-ins/Set/prototype/isDisjointFrom/converts-negative-zero.js`: refuses inherited library member values read as an own field
- `built-ins/Set/prototype/isSupersetOf/converts-negative-zero.js`: refuses inherited library member values read as an own field
- `built-ins/Set/prototype/symmetricDifference/converts-negative-zero.js`: refuses inherited library member values read as an own field
- `built-ins/Set/prototype/union/converts-negative-zero.js`: refuses inherited library member values read as an own field

### Language: never collection representation and mutable generic variance (5)

- `built-ins/Set/prototype/isDisjointFrom/compares-empty-sets.js`: not yet: a Set of never (a Set holds strings, numbers, booleans, objects, arrays, maps or functions so far)
- `built-ins/Set/prototype/isSubsetOf/compares-empty-sets.js`: not yet: a Set of never (a Set holds strings, numbers, booleans, objects, arrays, maps or functions so far)
- `built-ins/Set/prototype/isSupersetOf/compares-empty-sets.js`: not yet: a Set of never (a Set holds strings, numbers, booleans, objects, arrays, maps or functions so far)
- `built-ins/Set/prototype/symmetricDifference/combines-empty-sets.js`: refuses a value of type Set<never> seen as Set<number>, which can write number where never is read
- `built-ins/Set/prototype/union/combines-empty-sets.js`: refuses a value of type Set<never> seen as Set<number>, which can write number where never is read

### Language: subclassing builtin constructors (4)

- `built-ins/Set/prototype/difference/subclass.js`: not yet: a base that isn't a declared class
- `built-ins/Set/prototype/intersection/subclass.js`: not yet: a base that isn't a declared class
- `built-ins/Set/prototype/symmetricDifference/subclass.js`: not yet: a base that isn't a declared class
- `built-ins/Set/prototype/union/subclass.js`: not yet: a base that isn't a declared class

### Language: implicit arguments object (1)

- `built-ins/Map/groupBy/callback-arg.js`: refuses arguments

### Language: property descriptors and mutation (1)

- `built-ins/Map/get-set-method-failure.js`: refuses Object.defineProperty
