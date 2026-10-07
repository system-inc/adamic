Built JSON.parse with proven revivers, and completed the admitted stringify families: 41 additional test262 passes.
Commits: claim 778a6791be533f45a019eb62cc3ffa379c7ed0d7; implementation 67482341f9b4008a1d69b232d89376b461561c44; descendant guard a696839cf955eef01dd1ad87bbf551a0489f10e7.
Validation: Linux validity-runner survey and JSON source-Node, backend, release, ASan, UBSan, LeakSanitizer and counts gates pass.
Mutants: eight temporary runtime mutations plus the existing key-order descriptor mutation fail only against Node; nine safety-guard mutations fail their refusal probes.
Uncovered: 24 language refusals and four rawJSON declaration mismatches; dynamic holders and heterogeneous parse results remain refused.

## Before and after

Separate runner checkout: origin/codex/test262-ts-validity at af12899. Stock tsc 6.0.3;
Node 24.19.0 / V8 13.6.233.17-node.51; test262 c8c798898646638cd0c24879f8e0374e847e7d74.
Branch codex/library-json-2 starts at codex/library-json-number tip 735a702.
Only the refused column was reduced. All 41 newly passing tests agree with Node.
Number and Math source files are unchanged.

| built-ins/JSON | pass | disagreement | refused | not-typescript | crashed | skipped | total |
|---|---:|---:|---:|---:|---:|---:|---:|
| Before | 7 | 0 | 69 | 26 | 0 | 63 | 165 |
| After | 48 | 0 | 28 | 26 | 0 | 63 | 165 |

Raw reports: [before.json](before.json), [after.json](after.json).
[remaining.json](remaining.json) records every remaining refused file. The per-file readout used
the unchanged validity runner's classify, compileClass and typescriptVerdict functions through a
temporary Go test, removed afterward. Its remaining reasons were checked against the final survey.

## Library work, largest first

* JSON.parse: all 33 original first refusals now pass. V8's grammar, continuation stack,
  diagnostic templates and UTF-16 positions are ported from json-parser.cc, json-parser.h and
  message-template.h; THIRD_PARTY_NOTICES.md names them. Decimal conversion calls the existing
  adamic_number_parse_float in parse.c. Canonical serialization uses the existing stringify and
  V8 number printer; no new number parser or printer was written.
* The six object-reference first refusals now pass. Scalar fields require a complete literal
  origin without annotations or rebinding. Nested trees require the stronger JSON-local proof:
  the binding never escapes and every use is stringify's first argument. Aliases, shorthand
  escapes, property reads/stores, exports and opaque descendants invalidate that proof.
* The one replacer-function first refusal now passes. Only primitive-returning replacers are
  admitted, proving that SerializeJSONProperty visits the root alone. Argument values and
  callback returns are proven; throws propagate through the ordinary exception cleanup.
* The one literal toJSON first refusal now passes. Non-callable data properties serialize
  normally, including a fresh RegExp's empty enumerable shape. Callable literal toJSON methods
  require zero arguments and proven primitive or literal-array return metadata. Invocation
  precedes the replacer; nested undefined returns omit object keys and become null in arrays.
* Literal holes now serialize correctly, but that test next refuses new Array(3), so it adds
  no passing test. Immediate Number, String and Boolean boxes are consumed only for their proven
  JSON behavior; Number conversion reuses the existing helper without editing Number files.

Observation correcting the claim: its ten constructor diagnoses were initially grouped as
boxed construction from the coarse reason. Per-file source inspection identifies ten new Array()
parse scaffolds. The hole test exposes an eleventh Array-constructor refusal. These are language
work; no Array constructor or general sparse-array representation was built.

## Admission boundaries

Parsing discarded results uses an iterative continuation stack and iterative cleanup. Arrays and
objects nested 20,000 levels pass against Node, including malformed deep input and error position.
Duplicate keys keep the last value and retain the first insertion position. Integer-index keys
enumerate first. UTF-16 decoding preserves lone surrogates; binary64 edge values use the existing
conversion and formatting routines.

A used scalar must have a proven literal root and compatible contextual type, or a proven scalar
reviver return. Broad scalars are not trusted as narrower literal types. Arbitrary objects, any,
unknown, null storage and heterogeneous containers still need a language representation.
Immediate stringify consumption can use a private canonical carrier; it never escapes into the
language's values. A reviver's undefined result is normalized before an object emits its key.

Revivers require literal input of depth at most 64, avoiding an unproven Node recursion-overflow
boundary. Zero/key-only callbacks can visit container trees; two-argument callbacks require a
proven scalar input on every visit. Holder/this/context access, heterogeneous values, nullable
returns without proven result metadata and opaque callable identity refuse at compilation.
A void function view alone does not prove an actual undefined return: its body must establish it.
Private canonical results are bounded to depth 64 and currently exclude stringify replacer/space
arguments, including when the parsed value occurs inside a literal tree. General parse/reviver support is not claimed.

## Language handoff for @system_adamic

Counts are remaining first diagnostics; further blockers can be masked.

| Count | Feature | One-line reproducer |
|---:|---|---|
| 11 | Array construction and general holes | `const keys = new Array<string>(3); keys[1] = 'key'; JSON.stringify({key: 1}, keys);` |
| 5 | var scope | `var value = 1; console.log(value);` |
| 4 | Detached methods and their receiver | `const format = Number.prototype.toString; console.log(format.call(1));` |
| 2 | First-class JSON singleton | `const library = JSON; console.log(typeof library);` |
| 1 | any parse result | `const value = JSON.parse('{"n":1}'); console.log(value.n);` |
| 1 | delete and changing presence | `const value: {n?: number} = {n: 1}; delete value.n;` |
| masked | Structural views hiding descendant shapes | `const full = {n: 1, extra: 2}; const view: {readonly n: number} = full; JSON.stringify(view);` |
| masked | Heterogeneous reviver values and holders | `JSON.parse('{"n":1}', function (key, value) { return key === 'n' ? undefined : value; });` |
| masked | Property descriptors and access behavior | `const value = {n: 1}; Object.defineProperty(value, 'n', {enumerable: false}); JSON.stringify(value);` |
| masked | Prototype changes | `Object.setPrototypeOf({}, {toJSON: () => 7});` |
| masked | Stored boxes and prototype behavior | `const value = new Number(1); JSON.stringify(value);` |
| masked | Default/union callback argument storage | `const fn: (key: string, value: number) => number = (key, value = 99) => value; JSON.stringify(NaN, fn);` |
| masked | Callable array-return view metadata | `const fn: () => readonly (number \| string)[] = () => [1]; JSON.stringify({toJSON: fn});` |

Four further refusals are checker-code mismatches for proposal APIs: two TS2550/stock TS2339 and
two TS2550/stock TS2339+TS2345. Stock ES2024 tsc also rejects JSON.rawJSON/isRawJSON; the runner's
exact-code rule leaves these in refused. They are recorded transparently; no proposal declarations
or implementation were invented to make rejected programs compile.

## Mutants and what caught them

Every temporary mutation was restored. The runtime comparisons first require successful native
compilation, exit zero, empty stderr and a clean LeakSanitizer run. Only then does source Node
comparison fail with stdout differs. The persisted key-order test establishes the same conditions.

| Mutation | Catching fixture/gate |
|---|---|
| Smallest subnormal rounded to zero | library_json_parse_values.a / Node stdout |
| Duplicate key retained the first value | library_json_parse_values.a / Node stdout |
| Reviver children visited in reverse order | library_json_parse_reviver.a / Node stdout |
| Root replacer numeric return incremented | library_json_stringify_origin.a / Node stdout |
| toJSON numeric return incremented | library_json_stringify_origin.a / Node stdout |
| SyntaxError position incremented | library_json_parse_diagnostics.a / Node stdout |
| Fresh RegExp serialized as null | library_json_stringify_origin.a / Node stdout |
| Parsed undefined normalized as null | library_json_parse_reviver.a / Node stdout |
| Integer-index descriptor order reversed | TestJSONStringifyOracleCatchesKeyOrder / Node stdout |
| Nullable reviver return proof removed | nullable_reviver compile-refusal probe |
| Void actual-return proof removed | void_reviver compile-refusal probe |
| Replacer identity/callability proof removed | parsed_replacer compile-refusal probe |
| Binding rebinding proof removed | rebound_object compile-refusal probe |
| Scalar parameter assignability replaced by representation alone | narrow_reviver and narrow_replacer compile-refusal probes |
| JSON-only use proof removed | escaping_nested and shorthand_escape compile-refusal probes |
| Optional toJSON callability proof removed | optional_toJSON compile-refusal probe |
| Array callback return metadata proof removed | array_return_view compile-refusal probe |
| Parsed restriction checked only at root | parsed_descendant_space and parsed_descendant_keys compile-refusal probes |

The nested parsed-undefined composition was additionally reproduced as a failing clean Node-only
comparison before its fix, then passed afterward. Logs are /tmp/library-json-2-mutant-*.log;
regression logs are /tmp/library-json-2-nested-undefined-before.log and -after.log.

## Commands and setup

```sh
source /workspace/adamic-tools/env.sh
export PATH=/tmp/test262-typescript/node_modules/.bin:$PATH
/tmp/library-json-number-runner -adapt -json -root /workspace/adamic \
  -test262 /tmp/test262 -work /tmp/library-json-2-pushed built-ins/JSON \
  > /tmp/library-json-2-pushed.json 2> /tmp/library-json-2-pushed.log
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m \
  ./internal/lower ./internal/native ./internal/fresh ./internal/flow \
  ./internal/javascript ./internal/ir > /tmp/library-json-2-release-packages.log 2>&1
go vet ./... > /tmp/library-json-2-vet.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -v -count=1 -timeout 30m ./internal/oracle \
  -run 'TestJSONStringify|TestJSON(ParseRefusalBoundaries|RuntimeNodeComparison)|TestNativeAgreesWithNode/internal/oracle/testdata/(library_json_|json_stringify_)|TestCountsAreRecorded' \
  -args -update-counts > /tmp/library-json-2-release-oracle.log 2>&1
```

The implementation package gate passes: lower 31.772s, native 114.692s, fresh 44.880s,
flow 70.916s; javascript and ir have no package tests. Repository-wide go vet passes with no output.
The focused JSON oracle and whole-fixture count gate pass (49.489s).
After the descendant-only guard, lower and ir pass again (lower 6.523s), the entire JSON oracle
passes again (4.103s), and repository-wide vet passes again. Logs are
/tmp/library-json-2-descendant-packages.log, -oracle.log and -vet.log. The full repository test gate
was not invoked for this unit; the implementation packages and focused oracle are the stated gate.
All test output was written directly to log files, never piped.

Setup: Go 0s, clang 0s, Node 0s, submodules 0s, cache warm 21s, total 21s.
nproc 5; CPU quota 400000/100000; memory 17.6 GB. Setup succeeded.
