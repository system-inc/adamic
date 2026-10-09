# JSON and Number claim

Measured on Linux before implementation, from main `50045bd`, with the runner
built separately from `origin/codex/test262-ts-validity` at `af12899`.
Test262 is `c8c798898646638cd0c24879f8e0374e847e7d74`, adaptation on, no limit,
Node 24.19.0, stock TypeScript 6.0.3, clang 20.1.8, Go 1.27.1.

| Directory | pass | disagreement | refused | not-typescript | crashed | skipped | total |
|---|---:|---:|---:|---:|---:|---:|---:|
| built-ins/JSON | 7 | 0 | 69 | 26 | 0 | 63 | 165 |
| built-ins/Number | 152 | 0 | 70 | 44 | 0 | 74 | 340 |

## All observed refusal reasons

These are first diagnostics, not independent feature counts. Fixing the first
can expose another refusal. The four TS2550 records remain refused because
the validity runner requires a matching stock tsc diagnostic; stock tsc emits
TS2339 instead. They are not evidence that stock tsc accepts rawJSON.

### built-ins/JSON

| Count | First refusal |
|---:|---|
| 33 | refuses JSON.parse: its result's type can't be proven from the text |
| 10 | not yet: new an Identifier |
| 6 | not yet: JSON.stringify object references (structural types can hide fields and toJSON; runtime shapes need complete value metadata) |
| 5 | refuses var |
| 4 | refuses a method read as a value (toString would lose its object, and this with it) |
| 2 | error TS2550: Property '…' does not exist on type '…'. Do you need to change your target library? Try changing the '…' compiler option to '…' or later. (tsc: TS2339) |
| 2 | error TS2550: Property '…' does not exist on type '…'. Do you need to change your target library? Try changing the '…' compiler option to '…' or later. (tsc: TS2339,TS2345) |
| 2 | not yet: JSON as a value outside equality or typeof (overloaded calls and static properties need their own representation) |
| 1 | not yet: JSON.stringify a literal with toJSON semantics |
| 1 | not yet: JSON.stringify a spread or hole in a literal |
| 1 | not yet: JSON.stringify replacer functions (the callback must have a proven type for every visited value and its holder) |
| 1 | not yet: a value of type any |
| 1 | refuses delete |

### built-ins/Number

| Count | First refusal |
|---:|---|
| 37 | not yet: toString through an object view (a value with a different native representation may be hidden by the view) |
| 6 | not yet: new an Identifier |
| 6 | refuses var |
| 3 | refuses isPrototypeOf |
| 2 | not yet: a try around toString, whose failure is a panic natively but a throw a catch can take on Node (docs/memory.md) |
| 2 | not yet: for...in without a proven fixed plain-object origin (arrays, prototypes and absent synthetic fields cannot be enumerated soundly) |
| 2 | not yet: reading isFinite |
| 2 | refuses a method read as a value (toString would lose its object, and this with it) |
| 1 | not yet: Number conversion of an object |
| 1 | not yet: a BinaryExpression with a string and a number |
| 1 | not yet: a try around toExponential, whose failure is a panic natively but a throw a catch can take on Node (docs/memory.md) |
| 1 | not yet: a try around toFixed, whose failure is a panic natively but a throw a catch can take on Node (docs/memory.md) |
| 1 | not yet: a try around toPrecision, whose failure is a panic natively but a throw a catch can take on Node (docs/memory.md) |
| 1 | not yet: a value argument to toExponential |
| 1 | not yet: valueOf through an object view (a value with a different native representation may be hidden by the view) |
| 1 | refuses == |
| 1 | refuses a method read as a value (toFixed would lose its object, and this with it) |
| 1 | refuses the void operator |

## Library priorities, largest first

| Count | Family | Work and boundary |
|---:|---|---|
| 37 | Number formatting through boxed receivers | Immediate `new Number(...)` receivers can be proven to carry a Number internal slot. Reuse existing number printing and parsing. Arbitrary structural views stay refused. |
| 33 | JSON.parse | Implement the V8 grammar and diagnostics, reusing parse.c for binary64 conversion. A discarded result needs no dynamic value representation. Used arbitrary results and general revivers depend on the language blockers below; never treat any as a proven type. |
| 16 | Boxed construction (10 JSON, 6 Number) | Identify the actual constructor before changing anything. General boxed identity, coercion and subclasses depend on runtime value metadata; immediate intrinsic receivers are library work. |
| 6 | JSON.stringify object references | Complete origin evidence can prove literals behind immutable bindings, but structural field types alone cannot prove complete runtime JSON behavior. |
| 5 | Catchable Number range failures | Existing formatters panic; the library needs a catchable error and cleanup/MayThrow dispatch hooks. Do not change the formatting algorithm. |
| 4 | rawJSON/isRawJSON checker discrepancy | Stock ES2024 does not declare these proposal APIs either. Keep refused; do not add an unverified declaration. |
| 2 | Number.isFinite value observation | Intrinsic identity/type observation is library work; general first class overloaded static calls need language representation. |
| 1 each | JSON toJSON, holes, replacer; Number object conversion and format argument conversion | Implement only where the receiver, callback and every argument's behavior is proven. Coercion through arbitrary objects stays refused. |

## Language handoff for @system_adamic

No language feature is implemented by this slice. Counts below are the first
refusal counts across both directories; JSON.parse also masks dynamic-result
and reviver blockers, so their latent count is not yet observed. Reproducers
are one line each and belong in `.a` files.

| First count | Language blocker | One-line reproducer |
|---:|---|---|
| 11 | var adaptation cannot safely rewrite every scope | `var value = 1; console.log(String(value));` |
| 7 | Detached method values | `const format = Number.prototype.toString; format.call(1);` |
| 3 | Prototype reflection | `Object.prototype.isPrototypeOf(Number.prototype);` |
| 2 | Enumeration without fixed complete origin | `for (const key in Number) console.log(key);` |
| 2 | First class JSON singleton | `const json = JSON; console.log(typeof json);` |
| 1 | General any values | `const value: any = 1; console.log(String(value));` |
| 1 | delete and variable property presence | `const value: { key?: number } = { key: 1 }; delete value.key;` |
| 1 | String/number addition coercion | `console.log('value=' + 1);` |
| 1 | Loose equality | `console.log(String(Number.prototype == 0));` |
| 1 | void operator | `void Number(1);` |
| masked | Unknown, null and heterogeneous JSON value representation | `const value: unknown = JSON.parse('{"key":null}'); console.log(typeof value);` |
| masked | Reviver dynamic holder, deletion, heterogeneous callback and this | `JSON.parse('{"key":1}', function (key, value) { return key === 'key' ? undefined : value; });` |
| masked | General boxed primitive internal slots and identity | `const first = new Number(1); const second = new Number(1); console.log(String(first === second));` |

The parser must be exact on syntax diagnostics, UTF-16 positions, binary64
edges, strings including lone surrogates, duplicate keys, and deep input.
Anything not connected to a proven result representation remains refused.
No claim of full JSON conformance follows from a discarded-result parser.

## Commands and setup

`bash cloud/setup.sh > /tmp/library-json-number-setup.log 2>&1`: Go ready 0s,
clang ready 1s, Node ready 1s, submodules ready 1s, cache warm 70s, done 70s.
`nproc`: 5; cgroup quota 400000/100000; memory 17.6 GB. Environment file:
`/workspace/adamic-tools/env.sh`. Setup passed. The first validity-runner build
failed because its separate checkout had no initialized cohere; initializing
its submodules and rebuilding succeeded.

```sh
source /workspace/adamic-tools/env.sh
export PATH=/tmp/test262-typescript/node_modules/.bin:$PATH
/tmp/library-json-number-runner -adapt -json -root /workspace/adamic \
  -test262 /tmp/test262 -work /tmp/library-json-number-before \
  built-ins/JSON built-ins/Number > /tmp/library-json-number-before.json \
  2> /tmp/library-json-number-before.log
```

Raw before report and its complete text log are copied into this directory.
