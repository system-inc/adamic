# JSON decoding coverage

Base: origin/cloud/json-decode, 0ca1fbdade455a9e3d9debcd959aa4381320ba6b.
Reviewed CLAUDE.md, README.md, docs/0.1.md, docs/memory.md, the complete
28-file feature diff against current origin/main, both commit messages, and
changed implementations, fixtures, registration, cache and mutant tests.

The source oracle supplies checker-generated schemas but executes original
source and V8 JSON.parse. Native has its own parser and validator. The JavaScript
backend uses the Node validator. Schema agreement alone is not independent
coverage of checker type interpretation.

## Inventory

Names below omit json_decode_ and .a. Existing means present on the feature
branch before this coverage work. New means coverage_ prefixed fixtures.

| Implementation condition or value | Existing oracle program | New program or limit |
|---|---|---|
| number success, wrong JSON kind, null rejection | scalars | unions, nested |
| decimal fraction/exponent, rounding, signed zero, overflow, underflow | scalars, grammar | nested |
| string success and wrong kind | scalars | nested, grammar |
| true and false scalar | scalars | empty mutable boolean array |
| string, numeric and boolean literals, successful and failing matches | scalars, objects | numeric and boolean tags |
| lone high/low surrogates, pair, raw supplementary point, every short escape, NUL | scalars | nested; grammar repeats/reverses surrogate units, mixed-case hexadecimal |
| literal equality with surrogate and supplementary constants | scalars | unicode_tag uses both in discriminants |
| primitive union child selection and boxing number/boolean | only homogeneous literal unions | unions: number/string/boolean, success and each rejecting JSON container/null |
| object plus one primitive, no discriminant | none | object_scalar: success, missing, wrong field and wrong root |
| object union discriminant lookup, missing, unmatched, wrong kind, selected member failure | objects | tags: numeric and boolean; unicode_tag: WTF-8 key and values |
| arrays of references and numbers, wrong container, failing element | objects, supplement | nested multidimensional arrays; empty mutable booleans |
| empty arrays | depth | empty, nested |
| empty object schema, extra fields dropped, root-kind mismatch | none | empty |
| fixed tuple success, wrong kind/length, wrong element | objects | nested object/array tuple and partial-construction failures |
| empty tuple | none | empty_tuple.a differs; excluded from testdata |
| required field absent, present wrong type, null where nonnullable | objects, supplement | nested failures with full paths |
| nested objects and arrays | objects, supplement, depth | nested with multidimensional arrays, tuples and escaped strings |
| optional fields absent, present, wrong type, null | objects | layouts alternates optional positions over repeated allocation lifetimes |
| dropped extra field and fresh type field order | objects | layouts; empty; nested |
| duplicate keys last wins | objects | already covered |
| numeric key enumeration, __proto__, # and surrogate names | objects | unicode_tag uses surrogate name for union dispatch |
| empty input and disallowed whitespace | grammar | already covered |
| all accepted JSON whitespace including CR/LF | grammar | grammar CR-only column check |
| invalid keyword, missing numeric digits, fraction and exponent syntax | grammar | grammar adds false/null prefixes and exponent sign variations |
| invalid string escape, control character, unterminated string, incomplete or invalid hex | grammar | grammar incomplete hex and trailing backslash |
| object key/colon/comma/end and array comma/end errors | grammar | grammar premature endings |
| trailing input, UTF-16 columns and LF line numbers | grammar | grammar raw supplementary key before syntax error |
| depth below/equal/above 128 and recursive schema graph reuse | depth | depth_extra counts actual dropped field containers |
| array/object/tuple partial construction cleanup | objects, supplement | nested and layouts, held by ASan/UBSan/LeakSanitizer |
| result root scalar/reference ownership and error ownership | all | unions, empty, nested |
| canonical shape reuse with optional slot compaction; field offsets disabled | objects | layouts repeated present/absent combinations and required reads after optional fields |
| canonical shapes: same names with differing reference flags; same flags with differing names; same names/flags across string/array payloads; fresh object identity | none | shapes holds numbers, strings, booleans and arrays together and repeats; fresh identical decodes compare false |
| Object.keys on directly bound const decoded result.value, including parentheses | objects | layouts, empty |
| direct typed call, argument variable, loop, type alias/interface/generic type alias | all existing | calls, layouts, tags |
| renamed runtime import, parenthesized callee/argument/result, argument side effects, trailing comma/comment, concrete wrapper, arrow callback | only trailing comma in scalars | calls |
| local same-name function must not be decoded or instrumented | none | calls |
| T containing null anywhere, with visited-set recursion | compile-time tests, no runnable oracle fixture | existing TestJSONDecodeRefusals covers null/nullable root, field, tuple, array, object union and recursive nested type |
| missing explicit type argument, undefined, function, class, Map, Set, ambiguous/overlapping object unions, open generic | compile-time tests, no runnable fixture | TestJSONDecodeRefusals; impossible as accepted programs |
| explicit number or optional number union undefined | no testdata program; optional explicit union in refusal tests | number_undefined.a, required_undefined.a, optional_undefined.a refused |
| optional missing-property type removed rather than declared undefined accepted | objects | layouts |
| mixed primitive field reads and array element representation | none | mixed_read.a, mixed_array.a cannot lower |
| optional boolean read | objects validates presence without reading boolean | optional_boolean_read.a cannot lower |
| index signatures, callable/constructable objects, Date/RegExp, type parameter even constrained, non-JSON unknown/never, NUL field name | no runnable decode fixture | explicitly refused schema branches; no accepted runtime program |
| unsupported union child arrays/tuples, ambiguous/no distinct required literal discriminant, empty-name discriminant | no runnable decode fixture | refused schema/representation paths; no accepted runtime program |
| zero/multiple type arguments or zero/multiple text arguments | checker rejects most before decoder | no well-typed call; hidden oracle argument is generated, not source syntax |
| spread text argument | no runnable fixture | native lowering expressly rejects spread; not a supported call form |
| Object.keys through let, alias, unbound expression, or wrapper result | none | decodedResultValue requires a direct const initializer; unsupported reflection forms |
| reading union discriminant through bracket syntax and optional method calls | none | unrelated lowering limitations; Unicode dispatch checked by kind/error instead, optional arrays by length |

Allocation failure is not deterministic without fault injection; no oracle program
forces calloc to fail. Internal nil symbols/declarations, schema JSON encoding
errors and unknown AST nodes are compiler defenses rather than source features.
These cannot be targeted by a well-typed supported decodeJson program.

## Disagreement

empty_tuple.a is exactly:

```ts
import { decodeJson } from 'adamic';
const result = decodeJson<readonly []>('[]');
console.log(result.kind);
```

Source Node: stdout empty; stderr
`adamic: panic: TypeError: Cannot read properties of undefined (reading 'length')`,
exit 70. Native: stdout `Ok` plus newline; stderr empty; exit 0.
JavaScript backend: identical to source Node, exit 70.

Likely cause: oracle/json_decode.mjs:118 reads node.fields.length, but
internal/ir/library_json_decode.go:19 marks Fields omitempty. An empty tuple's
schema therefore has no fields property. Native uses field_count = 0 and accepts
it, consistent with the documented fixed-tuple contract.

The other notes/*.a files are compile-time limits, not output disagreements.
Their .build captures were inspected; they are deliberately excluded from oracle
registration and counts. Node can execute mixed_read, mixed_array and
optional_boolean_read, but native cannot produce a binary. Undefined schemas are
refused by the shared descriptor generator on the Node side too.

## Commands and completed checks

Shells sourced /workspace/adamic-tools/env.sh. Successful setup:

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (1s)
setup: build cache warm (192s)
setup: done in 192s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
nproc: 5
```

Ran bash cloud/setup.sh twice. The first overlapped checkout and failed warming
the cache with undefined decoder methods; the second completed on 0ca1fbd.
Go 1.27.1, clang 20.1.8, Node v24.19.0.

Read the requested comparison with git diff origin/main...origin/cloud/json-decode
and git log origin/main..origin/cloud/json-decode after fetching both refs.
The initially provided origin/main was stale and cloud/json-decode was absent.
Created coverage/json-decode with git switch -c coverage/json-decode origin/cloud/json-decode.

Executed test commands, with stdout/stderr redirected to the named logs:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestJSONDecode/json_decode_coverage' -count=1 -timeout 15m -v > /tmp/json-decode-new.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestJSONDecode$/json_decode_coverage' -count=1 -timeout 15m -v > /tmp/json-decode-new-exact.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestJSONDecode$/json_decode_coverage' -count=1 -timeout 15m -v > /tmp/json-decode-corrected.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestJSONDecode$/json_decode_coverage' -count=1 -timeout 15m -v > /tmp/json-decode-final-fixtures.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestJSONDecode$/json_decode_coverage' -count=1 -timeout 15m -v > /tmp/json-decode-coverage.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestJSONDecode$|^TestJSONDecodeRefusals$' -count=1 -timeout 15m -v > /tmp/json-decode-all.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts > /tmp/json-decode-counts.log 2>&1
gofmt -l cmd internal > /tmp/json-decode-gofmt.log
go vet ./... > /tmp/json-decode-vet.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./... > /tmp/json-decode-gate.log 2>&1
git diff --check
```

The earlier draft runs found unsupported forms and a fixture console API error;
the final JSON/refusal run passed in 11.313s, and counts passed in 37.664s.
Formatting and vet produced empty logs. The first overly broad regex also ran
TestJSONDecodeErrorLeakMutant, which passed. No test output was piped to head/tail.

Explicit builds/runs, for every new .a fixture and each notes/*.a probe:

```sh
for file in internal/oracle/testdata/json_decode_coverage_*.a notes/json-decode/*.a; do
 name=$(basename "$file" .a)
 go run ./cmd/adamic build "$file" -o "/tmp/json-decode-builds/$name" > "/tmp/json-decode-builds/$name.build" 2>&1
 result=$?
 if [ "$result" -eq 0 ]; then "/tmp/json-decode-builds/$name" > "/tmp/json-decode-builds/$name.native" 2> "/tmp/json-decode-builds/$name.native.err"; fi
 node --disable-warning=ExperimentalWarning oracle/node.mjs "$file" > "/tmp/json-decode-builds/$name.node" 2> "/tmp/json-decode-builds/$name.node.err"
 printf '%s build=%s node=%s\n' "$name" "$result" "$?"
done
```

Ran that loop first on draft fixtures, then on corrected fixtures plus notes;
rebuilt and reran unicode_tag after removing an unsupported bracket read.
Python compared native and Node stdout for the initial ten fixtures: all equal;
all of those builds and runs exit 0 with empty stderr. shapes was then built and run separately through all three backends, with identical stdout, exit 0 and empty stderr. It was registered after the full gate's oracle process had finished, while bridge/stage-1 tests were still running.

Also emitted and executed the JavaScript backend for unicode_tag and empty_tuple:

```sh
go run ./cmd/adamic js "$file" > "/tmp/json-decode-builds/$name.mjs" 2> "/tmp/json-decode-builds/$name.js.build.err"
node --disable-warning=ExperimentalWarning oracle/node.mjs "/tmp/json-decode-builds/$name.mjs" > "/tmp/json-decode-builds/$name.js" 2> "/tmp/json-decode-builds/$name.js.err"
```

unicode_tag passes; empty_tuple matches source Node's panic. The oracle itself
runs the JavaScript backend and native sanitizer/leak checks for all ten.

## Mutation proof

Used an isolated worktree at /tmp/json-decode-mutant, detached at 0ca1fbd,
so the full gate in /workspace/adamic could not observe changed runtime code.
Copied json_decode_coverage_tags.a into it. The worktree's empty submodule
needed TypeScript and TypeScript-shim symlinks to the existing checkout;
initial builds without those links failed before running the mutant.

Changed exactly one native decoder line, in literal():

```c
return value->kind==json_number && value->number==type->number;
```

to:

```c
return value->kind==json_number;
```

Both the mutated native program and source Node compiled/ran successfully,
exit 0, stderr empty. Their stdout comparison failed (diff exit 1):

```diff
--- /tmp/json-decode-mutant-node.log	2026-10-07 02:29:28.461733912 +0000
+++ /tmp/json-decode-mutant-native.log	2026-10-07 02:29:00.741890678 +0000
@@ -1,10 +1,10 @@
 4
-8
-at $.tag: expected N, found number
+at $.value: expected string, found number
+at $: missing field value
 at $.tag: expected N, found string
 at $: missing field tag
 at $: missing field value
-at $.value: expected number, found string
+1
 at $: expected N, found array
 yes
 0
```

Restored the runtime with git restore, rebuilt, and compared again:
cmp exit 0; the original output returned. Runtime source is restored in
both worktrees. This is a stdout witness, not a compiler/Werror failure.

Successful commands after supplying the submodule links:

```sh
cd /tmp/json-decode-mutant
go run ./cmd/adamic build internal/oracle/testdata/json_decode_coverage_tags.a -o /tmp/json-decode-builds/tags-mutant > /tmp/json-decode-mutant-build.log 2>&1
/tmp/json-decode-builds/tags-mutant > /tmp/json-decode-mutant-native.log 2> /tmp/json-decode-mutant-native.err
node --disable-warning=ExperimentalWarning oracle/node.mjs internal/oracle/testdata/json_decode_coverage_tags.a > /tmp/json-decode-mutant-node.log 2> /tmp/json-decode-mutant-node.err
diff -u /tmp/json-decode-mutant-node.log /tmp/json-decode-mutant-native.log > /tmp/json-decode-mutant-diff.log
git restore internal/native/runtime/json_decode.c
go run ./cmd/adamic build internal/oracle/testdata/json_decode_coverage_tags.a -o /tmp/json-decode-builds/tags-restored > /tmp/json-decode-restored-build.log 2>&1
/tmp/json-decode-builds/tags-restored > /tmp/json-decode-restored-native.log 2> /tmp/json-decode-restored-native.err
cmp /tmp/json-decode-mutant-node.log /tmp/json-decode-restored-native.log
```

## Final eleven-program validation

The full gate's oracle package passed in 815.250s on the initial ten programs;
the bridge package passed in 1366.004s. After registering shapes, the final
all-JSON/refusal run passed in 113.902s, including the eleventh program's
ASan/UBSan and leak checks. This separate pass closes the gate snapshot's
registration timing gap.

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestJSONDecode$|^TestJSONDecodeRefusals$' -count=1 -timeout 15m -v > /tmp/json-decode-all-final.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts > /tmp/json-decode-counts-final.log 2>&1
go run ./cmd/adamic build internal/oracle/testdata/json_decode_coverage_shapes.a -o /tmp/json-decode-builds/json_decode_coverage_shapes > /tmp/json-decode-builds/json_decode_coverage_shapes.build 2>&1
/tmp/json-decode-builds/json_decode_coverage_shapes > /tmp/json-decode-builds/json_decode_coverage_shapes.native 2> /tmp/json-decode-builds/json_decode_coverage_shapes.native.err
node --disable-warning=ExperimentalWarning oracle/node.mjs internal/oracle/testdata/json_decode_coverage_shapes.a > /tmp/json-decode-builds/json_decode_coverage_shapes.node 2> /tmp/json-decode-builds/json_decode_coverage_shapes.node.err
cmp /tmp/json-decode-builds/json_decode_coverage_shapes.native /tmp/json-decode-builds/json_decode_coverage_shapes.node
```

Final counts update passed in 83.198s and added exactly eleven rows without
changing previous rows. shapes at its final testdata path was explicitly built
and run successfully, with stdout equal to Node. All eleven rows have allocations
equal to frees, zero in regions, and no oracle leak report.

The broad gate recorded an unrelated failure in
internal/unicodeproperties/TestCanonicalizeUnicodeNode: `node: signal: killed`
after 1120.25s (package 1204.600s). No Unicode value disagreement was reported.
Its runUnicodeBatch uses a 4-minute Node subprocess deadline via CommandContext.
Cgroup memory.events reports oom=0 and oom_kill=0. A subprocess deadline under
concurrent compiler/test load is the likely cause; that is an inference, not a
confirmed context.Err observation. This coverage branch changes no Unicode code.
At this recording point, the broad gate's remaining stage-1 tests were still
running. The final JSON suite and counts gate have completed successfully.

## Twelfth program

arrays covers arrays of discriminated objects, optional member fields, partial
construction cleanup and mixed scalar arrays validated through kind/error. It
also holds an ordinary object sharing the value field name with decoder results.
The targeted three-way oracle passed in 4.188s with sanitizers and leak checks.
Explicit CLI build/run and Node comparison passed. Twelve-program counts passed
in 78.828s. Mixed array reads remain unsupported, but validation is supported.

Exact additional commands (after sourcing /workspace/adamic-tools/env.sh):

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestJSONDecode$/json_decode_coverage_arrays.a$' -count=1 -timeout 15m -v > /tmp/json-decode-arrays.log 2>&1
go run ./cmd/adamic build internal/oracle/testdata/json_decode_coverage_arrays.a -o /tmp/json-decode-builds/json_decode_coverage_arrays > /tmp/json-decode-builds/json_decode_coverage_arrays.build 2>&1
/tmp/json-decode-builds/json_decode_coverage_arrays > /tmp/json-decode-builds/json_decode_coverage_arrays.native 2> /tmp/json-decode-builds/json_decode_coverage_arrays.native.err
node --disable-warning=ExperimentalWarning oracle/node.mjs internal/oracle/testdata/json_decode_coverage_arrays.a > /tmp/json-decode-builds/json_decode_coverage_arrays.node 2> /tmp/json-decode-builds/json_decode_coverage_arrays.node.err
cmp /tmp/json-decode-builds/json_decode_coverage_arrays.native /tmp/json-decode-builds/json_decode_coverage_arrays.node
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts > /tmp/json-decode-counts-twelve.log 2>&1
```

An exploratory second mutation removed `|| hasJsonDecode(program)` from
internal/native/fields.go in the isolated worktree. Built arrays with
`go run ./cmd/adamic build internal/oracle/testdata/json_decode_coverage_arrays.a -o /tmp/json-decode-builds/arrays-offset-mutant --sanitize`
and ran with `ASAN_OPTIONS=detect_leaks=1`. It survived with matching output,
exit 0 and empty stderr. The trap restored fields.go. This probe does not
establish mutation coverage of that guard; the numeric literal mutation above
is the successful branch-dependence witness.

The broad gate also reported 30-minute timeouts in stage1/cohere/css and
stage1/cohere/lint. It is failed, not green; remaining broad-suite completion
is not part of the successful targeted JSON validation. No production changes
were made to address those unrelated failures.
