# Wasm split decisions

Run the command on a checked Adamic source:

    go run ./cmd/adamic-split internal/split/testdata/decisions.a

Each line contains the IR function name, wasm or javascript, an optional entry
marker, and the first failed eligibility rule. Analyze also returns the original
function index and declaration position. Decisions are sorted by file, line and
column, with IR order breaking ties. Synthetic functions without declarations
sort last and retain nil boundary metadata.

Purity is propagated through direct and virtual calls and closure bindings whose
complete target set can be resolved. Unknown callees, console, Adamic I/O,
external global observations, Weak reads, explicit panic and captured state
prevent crossing. Closed-program global writes include assignments and mutations
through aliases, calls and contained references. A global initializer alone does
not count as a write.

Boundary types come from Function.Boundary, populated in lowering by the same
checker-type schema builder used by decodeJson. Unsupported types have nil
schemas. A private string table prevents metadata-only literals from changing
emission. Boundary mode preserves undefined field alternatives; decodeJson's
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
    go test -count=1 -v ./internal/oracle -run 'TestJSONDecode' > /tmp/workers-boundary-json-oracle.log 2>&1
    go vet ./... > /tmp/workers-boundary-vet.log 2>&1
    ADAMIC_SPLIT_SWEEP=1 go test -count=1 -run TestAnalyzeOracleFixtures -v ./internal/split > /tmp/workers-boundary-sweep.log 2>&1

The opt-in TestCodegenSnapshot records every oracle source's C, JavaScript,
diagnostic status and decodeJson source-oracle descriptors. Run the same harness
against the base and hook, setting ADAMIC_SPLIT_SNAPSHOT to separate directories,
then compare the directories with diff -qr. The base is 7ad3826. Its 353 sources
include 311 that lower successfully; refused sources have matching diagnostic
artifacts rather than invented codegen.

No Wasm emission, bridge generation, deployment or changes to native or
JavaScript codegen are part of this unit. The lowering hook requires compiler
owner review before merge.
