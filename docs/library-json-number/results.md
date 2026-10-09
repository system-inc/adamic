# JSON and Number final measurement

The four requested stringify corrections are delivered. This is a partial
delivery of the broader JSON/Number unit: JSON.parse is still refused.

| Directory | When | pass | disagreement | refused | not-typescript | crashed | skipped |
|---|---|---:|---:|---:|---:|---:|---:|
| built-ins/JSON | before | 7 | 0 | 69 | 26 | 0 | 63 |
| built-ins/JSON | after | 7 | 0 | 69 | 26 | 0 | 63 |
| built-ins/Number | before | 152 | 0 | 70 | 44 | 0 | 74 |
| built-ins/Number | after | 190 | 0 | 32 | 44 | 0 | 74 |

The before/after sets retain every original pass. The 38 newly passing cases
are 37 `Number.prototype.toString` cases and one `valueOf` case. Every pass is
a native/Node comparison by the separately built validity runner. No skip or
not-typescript classification changed. Stock tsc 6.0.3 performed 74 checks
afterward, taking 11438ms. Raw reports are `before.json` and `after.json`.

## Remaining exact refusal reasons

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
| 1 | refuses == |
| 1 | refuses a method read as a value (toFixed would lose its object, and this with it) |
| 1 | refuses the void operator |

## Uncovered work

JSON.parse was not implemented. Its 33 first refusals mask further result
representation and reviver problems listed in `claim.md`. A discarded-result
parser remains feasible library work; it is not delivered here. There is no
claim of coverage for parser diagnostics, UTF-16 positions, duplicate-key
resolution, lone surrogates, deep parsing or reviver order. Consequently no
reviver-order or duplicate-key parser mutant was run.

Other remaining library work includes catchable Number range failures, boxed
identity, object coercion, object-reference stringify metadata, toJSON and
replacer callbacks. The language handoff with one-line reproducers remains in
the claim commit. None of those language features was implemented.

## After command

```sh
source /workspace/adamic-tools/env.sh
export PATH=/tmp/test262-typescript/node_modules/.bin:$PATH
/tmp/library-json-number-runner -adapt -json -root /workspace/adamic \
  -test262 /tmp/test262 -work /tmp/library-json-number-after \
  built-ins/JSON built-ins/Number > /tmp/library-json-number-after.json \
  2> /tmp/library-json-number-after.log
```

## Linux validation

`gofmt -l cmd internal` produced no paths. `go vet ./...` produced no diagnostics.
The complete command was:

```sh
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./... \
  > /tmp/library-json-number-full-gate.log 2>&1
```

Its flow tests exposed the misplaced deliberately refused fixtures. They are
now in `docs/library-json-number/probes/`, with no change to shared flow tests.
The full lower package passed in 26.304s, native in 213.872s, and the uncached
full oracle in 180.000s. After the fixture move, the affected flow package was
rerun and passed in 137.609s. The final uncached focused oracle passed in
5.781s, including all eleven stringify fixtures, the Number fixture, both
Node-held compile-time generic refusals and all six stringify refusal cases.

```sh
go test -count=1 -timeout 30m ./internal/flow \
  > /tmp/library-json-number-flow-fixed.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -v -count=1 -timeout 30m ./internal/oracle \
  -run 'TestJSONStringifyRefusals|TestLibraryJSONGenericInstantiations|TestNativeAgreesWithNode/internal/oracle/testdata/(json_stringify|library_json|library_number_immediate)' \
  > /tmp/library-json-number-final-focused.log 2>&1
```

Mutant commands, outcomes and restoration are recorded in
`integration-fixes.md`, `number-immediate.md` and `shapes-mutants.json`.

The complete uncached gate finished with exit 1 solely from the original flow
fixture placement failure. Every other package passed, including Unicode
properties (696.261s) and all stage-1 packages. Unicode finished while the
filtered-gate fallback was being prepared; no test process was stopped. The
full gate was not repeated after relocation: the affected flow package and
focused uncached oracle were rerun and both exited zero, as recorded above.
No claim of a single green full-gate invocation is made.
