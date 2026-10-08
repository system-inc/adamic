Built: native checked array reads and mutation use the shared element_kind byte; scalar join consumes checked elements lazily.
Branch: codex/views-arrays-callables-parser, after 2693041a; the published checkpoint SHA is reported with the push.
Validation: focused lower/native/javascript/ir/oracle view and array regressions pass; 13 new probes run against Node, sanitized native, release native, and emitted JavaScript.
Mutants: array field kind, allocation kind, selected element kind, indexed-write kind, push kind, literal, copied kind, join contract, sparse map, JavaScript write kind, and JavaScript pop-hole controls.
Limits: array own fields, tuples, nullable element contracts, remaining intrinsics, and exact per-read allocation proofs remain; no complete tsc-site unlocking count is claimed.

## Storage and integration

Removed the duplicate adamic_array_view_metadata member. All array helpers now
use Lane 1's adamic_array.element_kind, with 1 number, 2 boolean, 7 packed optional
number, 10 heap pointers, and 0 uncertified scalar storage. Heap headers classify
the selected reference; the physical byte is never a shape/readiness certificate.
Assertion targets never stamp storage. Fresh literals retain their existing
shared emitter stamp, normalized to heap pointers for reference slots. Producer
stamps live in internal/native/view_arrays.go; sparse constructors are included.
Dense/sparse slice copies preserve the source byte; splice/concat copy metadata
without claiming their consumer paths are admitted.

The native selected-read hook uses sparse lookup, decodes packed/boxed values,
checks kind/literals and object contracts, and does not scan a field-read array.
A billion-slot sparse array field reads its length without scanning or allocating
its slots. Map/visits/reduce skip holes before element checks. Sparse pop commits
only after the selected read, releases the removed slot, and maintains absent-slot
accounting. Sparse push and slice preserve indexing, holes and ownership.
Indexed writes and push check the original physical storage before storing; a
number view of a boolean allocation fails loudly rather than reinterpreting bits.
Unknown scalar storage stays uncertified. JavaScript uses the equivalent existing
WeakMap metadata, retaining native JavaScript hole semantics.

Scalar .join now obtains each declared element contract through viewArrayJoin;
its native conversion checks each element at consumption and keeps the original
array/identity/storage. JavaScript maps present slots through the same contract
before join. Missing slots remain empty join positions; stored incompatible
values fail. Receiver and separator run once, in order, and a separator that
pushes into the source affects the join as Node does. Nested joins and optional
receiver chains still have unresolved dependencies and are not claimed complete.

Small named shared hooks in this checkpoint:

- lower/object.go: viewArrayJoin attaches the scalar join read metadata.
- lower/library_array_holes.go: viewArrayHolesConsumer and viewArrayHolesMethod
  lift only the implemented push/pop/slice guards in programs with array views.
- native/emit_expressions.go: emitViewArrayJoin and sparse push/slice helpers.
- native/library_array_holes.go and emit_arrays.go: viewArrayElementSlot after
  skipping absent callback slots.
- javascript/javascript.go: checked join dispatch and the sparse indexed-write flag.

No callable files or certification hooks were changed. The earlier registered
internArrayViewContract remains the shared contract builder. Lane 1's remote was
ab4d6f90 at this checkpoint, without the lazy-admission handoff containing this
branch; the lead assigned the cast conflict to Lane 1, so it is not resolved here.

## Census remainder

These are read-family demand counts, not successful full-compiler lowerings.
Families overlap. Supporting a consumer does not certify every array allocation
or remove other lanes' contracts.

| Family | Census pairs / reads | Remaining at this checkpoint |
| --- | ---: | --- |
| Array contracts | 334 / 3,189 | Kind/presence/readiness path implemented; complex element contracts and per-read proof still partial; no exact residual number inferred |
| Element and consumer reads | 251 / 1,602 | Selected read/conversion and supported callbacks/mutation covered; own fields, unions, generic instances and remaining consumers still partial |
| Own array fields | 30 / 72 | 30 / 72; no own-field admission added |
| Tuple contracts | 6 / 9 | 6 / 9 |
| Previously refused intrinsic names | 35 / 128 | Join primitive conversion implemented for its 5 / 49 demand; 30 / 79 demand remains under other refused intrinsic names |

The 49 join reads are 42 string[], 3 readonly string[], 2 number[], and two
optional string-array receiver chains. The first three shapes (3 pairs / 47 reads)
have the implemented scalar path. The optional chains still need their receiver
admission. This is mechanism coverage, not evidence those 47 whole-source reads
have individually lowered in tsc. Nested join refusal affects none of these five
census receiver shapes. Next by missing demand: own fields (72), then indexOf
(34), includes (16), sort (11), splice (7), unshift (5), concat (3), lastIndexOf (2),
shift (1). Own-field integration intersects the pending Lane 1 lazy contract work.

## Evidence

Node controls: native-array-node-controls.json. Successful probes cover numeric
and boolean join, constant-space array field reads, evaluation order, numeric and
string writes/push/pop, sparse map/reduce/slice/join, and popping missing slots.
Wrong kinds/literals pin exit 70 and the expression, expected type and found kind
in both backends. Existing array fixtures additionally cover presence/readiness,
transitive object elements, alias writes, generic casts and parser rows 2/5.

Command (output captured in native-array-logs/focused.log):

    ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/native ./internal/javascript ./internal/ir ./internal/oracle -run 'Test.*View|TestPhantom|TestArrayHolesAliasRefusals|TestArrayHolesNullableSearchRefusal|Test.*ArrayJoin|Test.*ArraySearch|Test.*ArraySlice|Test.*ArrayPop|Test.*ArrayPush|Test.*Template' -count=1

Observed after mutant restoration: lower 5.914s, native 0.707s, JavaScript
0.890s, IR 0.023s (no matching tests), oracle 31.437s, all pass. go vet on
the four touched compiler packages also passed. The focused new native-array oracle also passed
separately. No full repository gate was run for this checkpoint.

The wider TestArrayHolesMilestone run fails before array lowering on the existing
library_array_holes_callbacks.a predicate: adamic/no-type-predicate rejects
(value) => value !== undefined. The alias and nullable-search refusal tests pass;
this checkpoint does not alter predicate preflight.

Mutant commands and each raw semantic failure are in native-array-logs. Every
mutant is restored in a finally block; a clang/build or JavaScript syntax error
is explicitly rejected as evidence. None edits callable code. The full eleven
mutants are rerun against the integrated array implementation.

Toolchain setup: GOPROXY=https://proxy.golang.org|direct; bash cloud/setup.sh;
source /workspace/adamic-tools/env.sh. Node ready 0.046s, Go 0.068s, markdown
0.119s, submodules 0.124s, clang 0.388s, build 94.503s, warm 95.248s, done 95.336s;
nproc=5, cgroup quota=4 cores. Node 24.19.0, Go 1.27.1, clang 20.1.8.
