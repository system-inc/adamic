# Wasm split decisions

Run the command on a checked Adamic source:

    go run ./cmd/adamic-split internal/split/testdata/decisions.a

Each line contains the IR function name, wasm or javascript, an optional entry
marker, and the first failed eligibility rule. Analyze also returns the original
function index and declaration name-token position. Decisions are sorted by file, line and
column, with IR order breaking ties. Synthetic functions without declarations
sort last and retain nil boundary metadata.

Purity is propagated through direct and virtual calls and closure bindings whose
complete target set can be resolved. Unknown callees, console, Adamic I/O,
external global observations, Weak reads and captured state
prevent crossing. Panic and compiler panic checks are Wasm-safe call aborts.
JSON decode and encode are pure operations; crossability remains a separate
requirement. Closed-program global writes include assignments and mutations
through aliases, calls and contained references. A global initializer alone does
not count as a write.

Boundary types come from Function.Boundary, populated in lowering by the same
checker-type schema builder used by decodeJson. Unsupported types have nil
schemas. One BoundaryStringTable on ir.Program interns metadata literals separately
from codegen, with no lowering copies or per-parameter table copies. Boundary mode preserves undefined field alternatives; decodeJson's
existing admission rules and descriptors remain unchanged. Numbers, booleans,
strings and their literal subtypes are scalars. Arrays contain scalars; objects
contain those types and objects with at most two object levels. Only fields may
have a scalar or data type combined with undefined.

Mutation summaries distinguish container identities from references stored
inside them. Parameter and global origins follow aliases, fields, elements,
returns, closures and callees until a fixed point. Scalar array copies do not
share mutable elements. Ambiguous aliases are treated as possible writes, so
the analysis may leave safe functions in JavaScript when it cannot prove a
parameter remains unchanged. A writes-parameter reason reports this static
may-write result, not an observed execution.

Loops and recursion, including mutual recursion, propagate through callees.
A pure function with a crossable signature and no such work stays JavaScript
with reason too small to cross. An eligible function is an entry only when an
ineligible IR function calls it. Top-level statements are not IR functions and
do not independently create entries.

Every .a fixture has a checked-in .txt decision table and runs independently
through source Node, emitted JavaScript and sanitized native execution. The
cherry-picked erasure witness now checks that boundary metadata distinguishes
element types while the approved relaxed rules do not require const or readonly
declaration markers.

Verification commands:

    source /workspace/adamic-tools/env.sh
    go test -count=1 -v ./internal/ir ./internal/lower ./internal/split ./cmd/adamic-split > /tmp/workers-boundary-packages.log 2>&1
    ADAMIC_GATE_UNCACHED=1 go test -count=1 -v ./internal/oracle -run 'TestJSONDecode' > /tmp/workers-boundary-json-oracle.log 2>&1
    go vet ./... > /tmp/workers-boundary-vet.log 2>&1
    ADAMIC_SPLIT_SWEEP=1 go test -count=1 -run TestAnalyzeOracleFixtures -v ./internal/split > /tmp/workers-boundary-sweep.log 2>&1

The opt-in TestCodegenSnapshot records every oracle source's C, JavaScript,
diagnostic status and decodeJson source-oracle descriptors. Run the same harness
against the base and hook, setting ADAMIC_SPLIT_SNAPSHOT to separate directories,
then compare the directories with diff -qr. The final base is origin/area/platforms
at be2e9c1 (including f573c146 and decoder coverage 01af78b8). All 399 sources
match byte for byte: 344 lower successfully and have identical C and JavaScript;
refused sources have matching diagnostic artifacts rather than invented codegen.
All decoder source-oracle descriptors match as well.

No Wasm emission, bridge generation, deployment or changes to native or
JavaScript codegen are part of this unit. The lowering hook requires compiler
owner review before merge.


The only schema-builder factor is lowering.jsonDecodeSchema(node, rootType,
mode, intern) in internal/lower/library_json_decode.go. decodeJson and encodeJson
pass jsonSchemaDecode and jsonSchemaEncode respectively, with the codegen interner
l.constant. Function metadata passes jsonSchemaBoundary and
ir.Program.BoundaryStrings.Intern. The shared walker uses this explicit mode,
never the callee, and boundary relaxations do not apply to decode or encode.

## Mutation evidence

Every mutation below was applied alone and restored. The twelve analysis/hook
mutants exited 1 because assertions failed, with no build failures. Logs are
/tmp/workers-boundary-platform-mutant-NAME.log; the summary is
/tmp/workers-boundary-platform-mutants.log.

| Mutant | Check that caught it |
| --- | --- |
| transitive-purity: ignore impure callees | TestFixtures: callsLogger and mutual recursion tables |
| mutable-global: ignore closed-program writes on reads | TestFixtures: readsMutable and globalRead |
| map-signature: accept nil parameter schemas | TestFixtures: takesMap and takesSet |
| internal-entry: count eligible callers as crossing callers | TestFixtures: helper sum; command golden test |
| wrong-parameter: reuse first parameter checker type | TestBoundaryParameterOrderAndPositions: second schema must be string |
| parameter-write: omit write exclusion | TestFixtures: parameter aliases and nested writes |
| recursion: omit recursive work | TestFixtures: factorial becomes too small |
| small-leaf: omit crossing cost rule | TestFixtures: tiny becomes eligible |
| unknown-callee: omit unknown-call impurity | TestFixtures: takesClosure reason |
| object-depth: accept depth 200 | TestFixtures: tooDeep becomes eligible |
| contained-reference: omit stored reference propagation | TestFixtures: containedWrite loses parameter origin |
| future-operation: accept unknown IR expressions | TestUnknownOperationIsImpure |
| string-table: share emission table with metadata builder | Base/hook snapshot diff detects changed C and JavaScript |

The string-table mutation passed the ordinary execution tests but failed the
byte-identity comparison. Its separate logs are
/tmp/workers-boundary-platform-mutant-string-table-{tests,snapshot,diff}.log.

Toolchain setup reported go/clang/node/submodules ready in 0s, build cache warm
in 98s, total 98s; nproc was 5. The final platform validation logs use the prefix
/tmp/workers-boundary-platform-. Package tests, the uncached TestJSONDecode
oracle, go vet ./..., the 344-program analysis sweep, and the complete byte
comparison passed. The complete repository test gate was not run.


## Compiler review follow-up

The changes above 8f4da8e preserve history. Declaration positions come directly
from scanner.GetLineAndCharacterOfPosition at the name token (unnamed closures
and constructors use the declaration token). Constructor metadata receives a
nil receiver exactly where runtime Parameters receives its receiver, including
the metadata-only addition beside insertion in class_inheritance.go. Public
allocators copy the complete declared parameter boundary and are refused with
signature: constructor not crossable.

The unions.a fixture pins both an undiscriminated object union and a broad
number | string | boolean union as signature: parameter 1 not crossable. It
also runs ordinary source Node, emitted JavaScript and sanitized native, and
pins live parent/child initializer alignment and a boundary-only string literal
absent from codegen. The two consecutive boundary visitor blocks are merged.

Review validation logs use /tmp/workers-boundary-review-. The 399-source base
and hook snapshots compare byte for byte, including all 344 C/JavaScript pairs,
refusal diagnostics and decodeJson descriptors. The uncached TestJSONDecode
oracle passed in 29.1s (38 native and 63 Node misses); touched package tests,
repository vet and the 344-program sweep passed. All eighteen distinct mutants
were caught by assertion failures, with each mutation restored: the twelve
original classifier and hook mutants, plus these review checks:

| Mutant | Check that caught it |
| --- | --- |
| string-table: use codegen interner for boundary | TestBoundaryStringTable and byte snapshot diff |
| name-position: use declaration instead of name | TestBoundaryParameterOrderAndPositions |
| receiver-alignment: omit initializer receiver schema | TestBoundaryConstructorAlignment |
| allocator-alignment: discard first declared schema | TestConstructorBoundary |
| constructor-refusal: omit named refusal | TestFixtures constructor tables |
| union-fence: accept union nodes | TestBoundaryUnionFences and TestFixtures |

No encodeJson code is present in this base, so its integration is not tested.
The shared builder's new interning argument is the encoder coordination point.


The interner mutant also passes ordinary fixture execution but fails the
399-source byte comparison. Its logs are
/tmp/workers-boundary-review-mutant-string-table-{tests,snapshot,diff}.log.
The assertion mutant summary is /tmp/workers-boundary-review-mutants.log.
No full repository test gate, Wasm emission or bridge generation was run.


The panic/JSON refinement and the complete real compute Worker before/after
decision tables are recorded in [refine-report.md](refine-report.md). Encoder
fixtures run automatically when its independently landing compiler dependency
is present; until then the base reports a skip and the compute integration
witness runs the full source/JavaScript/native comparison.
