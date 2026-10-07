Built integrated array and callable checked views in lowering, native C and JavaScript, including the unchanged sameMap generic array probe.
Commits: dependency merge ea68e9b5 from lane 1 609ed395; implementation 9fc5c01ab850d92b24b66d2a6782f2a820bdb929.
Commands: complete lower/JavaScript/native package tests, focused checked-view oracles on both backends, Node controls and vet pass; timing details below.
Mutants: integrated array kind, element kind and opaque signature acceptance are caught by semantic assertions; helper callable-kind mutant also caught.
Not covered: full TypeScript compiler lowering census, lane 3 erasure, tuples, mixed scalar unions, array callable elements, opaque signatures and unsupported array consumers.

## Implemented behavior

Both array families use lane 1's contract registry and `view()` entry point.
Reading an array field checks presence, initialization and array kind without
walking its elements. Selected indexed, `.at`, loop and callback reads validate
the declared element representation and literal contract. Object elements use
lane 1's transitive field contracts and tagged object-union checks. Aliases retain
identity: the alias-mutation fixture succeeds once and then fails after a write.
Out-of-range indexes return undefined; a present undefined element fails a
required number contract. Bad reads stop with exit 70 and expression, expected
contract and found category. No check erasure is claimed.

`map`, `forEach`, `filter`, `some`, `every`, `find`, `findIndex`, `reduce`, `slice`,
`at`, `pop`, `push`, indexed writes and `for..of` have named read/storage hooks.
Consumers requiring another conversion path, including join, sort, concat and
spreads, fail closed in a program containing array views. The policy is
conservative and program-wide, including reads outside the immediate cast path.
Optional/nullish contracts remain subject to lane 1's unsupported-contract rules.

Callable field reads and calls check presence, readiness and function kind.
A closed-program census of same-named implementations supplies conservative
signature evidence: strict parameter variance, result variance, physical
representations, nominal ancestry and mutable-slot relations must agree.
Bodies must be visible; generic/rest/optional/overloaded signatures and opaque
implementations are not certified. Escaping member reads require that same proof
so a captured alias cannot bypass call admission. A discarded read can use only
the presence/kind checks. This is an overapproximation, not lane 3's allocation
flow. Compatible tagged class constructions preserve existing free method
dispatch after the tag check; unknown class identity keeps the refusal.

Opaque calls pin this message:

    Adamic 0.1 refuses a checked view call to member Runner.run; its signature cannot be checked at runtime and no compatible implementation is proven;

## Authorized shared hooks and metadata

The October 7 approval covers the small dispatch and IR integration edits:

- Lower dispatch: `viewArrayCast`, `viewProvenClassCast`, `viewArrayFields`,
  `viewCallableFieldUses`; registration through `viewArrayContractHook` and
  `viewCallableContractHook` from the two lane-owned files.
- Read lowering/finalization: `markViewArrayRead`, `viewArrayUse`,
  `markProgramViewArrayRead`, `markProgramViewArrayUse`; readiness retains
  callable reads and resolves contracts lowered after earlier reads.
- IR: `HasArrayViews` and operand-free `ArrayViewRead`; array index carries
  contract/type IDs, literals, source text, required-read and undefined facts.
  Loop/map/visit/reduce/pop operations carry that metadata separately from their
  executable expressions, preserving existing CFG and ownership traversal.
- Native: `evaluate` wraps existing `evaluateWithoutViewArrays` to stamp
  allocations, `emitViewArrayRead`, `viewArrayElementSlot`, `viewArrayMutation`,
  `emitViewArrayPop`, and `emitViewCallableRead` connect existing dispatches.
- JavaScript: `value` wraps `valueWithoutViewArrays`; `emitViewArrayRead`,
  `viewArrayElementCheck`, `viewArrayChecker`, `emitViewCallableProperty` connect
  existing read/call/loop/array-operation dispatches.
- Shared runtime/header: include the new array/callable headers, recognize closure
  kind in checked object reads, and preserve storage metadata in array copies.
  Array-field writes without a source-slot certificate remain refused.

`internal/native/runtime/view_arrays.h` adds one physical-storage byte embedded
in every native array after its heap header. It describes actual allocation
storage, never an asserted target type; zero is uncertified scalar storage.
It survives slice/copy paths and aliases. Selected reads inspect reference heap
kinds and safely unbox numeric/boolean values, or unpack maybe-numbers. They do
not reinterpret unproven scalar bits. Runtime helpers and metadata live in new
`view_arrays.c/.h`; callable method/closure lookup lives in `view_callables.c/.h`.
No new allocation or garbage collector is introduced by this metadata.

Protected emit.go, lower.go, native.go and oracle_test.go were not edited.
No code was copied from cohere. Shared counts.md and the original audit are not
rewritten by this lane.

## sameMap and the 2,936-site audit

The exact probe from origin/codex/stage3-parser-proof at 2179dd8 is retained as
`native-generic-array-cast.a`. Its parser BLOCKERS row 2 is core.ts:195; its
original stock-source audit row is core.ts:356:37:

    array.slice(0, i) as unknown[] as U[]

Before integration the parser report records adamic/no-unchecked-cast and Node
stdout `1`. After integration the unchanged `prefix<number, number>` probe
lowers and executes on native debug, native release and JavaScript with stdout
`1`, exactly matching Node. A `prefix<number, string>` witness stops when its
number element is read as string, proving that a resolved type parameter uses
the same lazy element rule. Its Node control prints `1` because Node erases types.

One corrected implementation finding matters: an early array hook compared the
mutable relation backwards and reported this probe blocked. Comparing source to
target is the correct write-safety relation and admits the unchanged concrete
probe. The final tests replace that earlier observation; failure logs are retained
as development evidence, not final acceptance results.

The original audit still has 2,936 undecidable shape-conformance rows. Lane 2's
primary partition remains 1,071: 25 tagged and 1,046 untagged. Its untagged
families include 1,010 arrays and 71 callables with 35 overlapping. Verified
reduced witnesses tied to original sites increase from zero to one, the sameMap
row above. The other 1,070 primary-partition sites were not remeasured. Whole-file
corpus lowering successes are not measured; no eligibility tally is presented as
whole-program lowering, and no additional conformance certificates are claimed.

## Validation and limits

All Go commands source /workspace/adamic-tools/env.sh and set
GOPROXY='https://proxy.golang.org|direct'. Setup succeeded: Node 0.041s, Go 0.056s,
submodule 0.180s, markdown 0.184s, clang 0.508s, build 57.019s, warm 57.205s,
done 57.259s. nproc=5, cpu.max=400000/100000. Go 1.27.1, clang 20.1.8,
Node 24.19.0.

- Complete touched packages: `go test ./internal/lower ./internal/javascript
  ./internal/native -count=1 -timeout 30m`: lower 27.903s, JavaScript 0.897s,
  native 229.906s, pass. Final complete lower/JavaScript tests pass in
  19.131s and 0.806s respectively. Subsequent final changes are covered by the focused
  oracle and final complete lower/JavaScript run; no repeated full native gate.
- Final focused command: `go test ./internal/lower ./internal/javascript
  ./internal/oracle -run 'TestCheckedView|TestView|TestDefaultTaggedSourceViews'
  -count=1 -timeout 30m`: lower 0.656s, JavaScript 0.699s, oracle 24.883s, pass.
  Positive outputs match Node; deliberate malformed views pin exit 70 and stderr
  on native debug, native release and JavaScript. Node controls for all 26 source
  fixtures are saved separately in node-controls.json.
- `go vet ./internal/lower ./internal/javascript ./internal/native`: pass.
- Integrated mutants, restored in finally blocks: dropping array kind fails
  TestCheckedViewArrays/array-length-kind; dropping selected element kind fails
  TestCheckedViewArrays/array-second; accepting an opaque signature fails
  TestCheckedViewOpaqueSignature. The read mutants produce executable debug and
  release C and fail semantic output/exit assertions, not clang warnings.
- The helper mutation runner additionally drops callable kind and is caught by
  TestViewCallablesNode/kind, alongside its array/element/signature mutants.

Logs in logs/integration-* record final commands, Node observations and mutant
failures. The full gate, original-source TypeScript corpus build and benchmarks
were not run. Unsupported families stop explicitly. Current origin/main
71d7e491 is already an ancestor through the lane 1 dependency merge.
