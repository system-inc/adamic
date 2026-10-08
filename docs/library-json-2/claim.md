# JSON-only claim

Branch `codex/library-json-2` starts at `735a702` on
`codex/library-json-number`. Number and Math files are outside this unit.

Linux baseline, separate validity runner from `af12899`, stock tsc 6.0.3,
Node 24.19.0, test262 `c8c798898646638cd0c24879f8e0374e847e7d74`:

| Directory | pass | disagreement | refused | not-typescript | crashed | skipped |
|---|---:|---:|---:|---:|---:|---:|
| built-ins/JSON | 7 | 0 | 69 | 26 | 0 | 63 |

## Exact first refusal counts

| Count | Reason |
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

## Library work, largest first

| First count | Library family and boundary |
|---:|---|
| 33 | JSON.parse: port V8 syntax scanning and diagnostics, reuse existing binary64 conversion, preserve UTF-16 strings, duplicate keys and iterative deep parsing. Admit only proven result uses. Support revivers only when every argument, return and holder behavior is proven; general dynamic holders belong to the handoff. |
| 10 | Boxed primitive construction seen first: recognize immediate intrinsic receivers only inside JSON calls where their complete behavior is proven. Stored boxed identity is language/runtime representation work. Number conversion algorithms remain owned by the new worker and are reused. |
| 6 | JSON.stringify structural object references: complete runtime shape and toJSON evidence required; otherwise refuse. |
| 1 each | Literal toJSON, spread/hole and replacer callbacks: build only proven JSON-local behavior, without adding language reflection or dynamic shapes. |
| 4 | rawJSON declaration mismatch: stock ES2024 tsc rejects the proposal APIs too (TS2339 rather than our TS2550); keep refused, do not invent declarations. |

Counts are first diagnostics: one test can expose another blocker after the first closes.

## Language handoff for @system_adamic

| First count | Language blocker | One-line reproducer |
|---:|---|
| 5 | var scope and adaptation | `var value = 1; console.log(`${value}`);` |
| 4 | Detached methods and this | `const format = Object.prototype.toString; format.call(JSON);` |
| 2 | First-class JSON singleton | `const json = JSON; console.log(typeof json);` |
| 1 | any values with no proven representation | `const value: any = JSON.parse("1"); console.log(value);` |
| 1 | delete and changing property presence | `const value: { key?: number } = { key: 1 }; delete value.key;` |
| masked | Arbitrary parse result, null and heterogeneous container representation | `const value: unknown = JSON.parse("{\"key\":null}"); console.log(typeof value);` |
| masked | Reviver holder, this, heterogeneous value and deletion | `JSON.parse("{\"key\":1}", function (key, value) { return key === "key" ? undefined : value; });` |
| masked | Stored boxed primitive identity and prototype behavior | `const value = new Number(1); console.log(JSON.stringify(value));` |
| masked | Complete structural object origin | `const full = { key: 1, extra: 2 }; const view: { key: number } = full; JSON.stringify(view);` |

No language feature is implemented by this unit. Refusal must happen before native execution whenever the required evidence is absent.

## Measurement and setup

```sh
source /workspace/adamic-tools/env.sh
export PATH=/tmp/test262-typescript/node_modules/.bin:$PATH
/tmp/library-json-number-runner -adapt -json -root /workspace/adamic \
  -test262 /tmp/test262 -work /tmp/library-json-2-before built-ins/JSON \
  > /tmp/library-json-2-before.json 2> /tmp/library-json-2-before.log
```

`bash cloud/setup.sh`: Go 0s, clang 0s, Node 0s, submodules 0s, cache warm
21s, total 21s. `nproc`: 5; CPU quota 400000/100000; memory 17.6 GB.
Environment `/workspace/adamic-tools/env.sh`. Setup and measurement passed.
