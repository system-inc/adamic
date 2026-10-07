# Environment placement coverage

Base: origin/codex/environment-placement at fc3075a. Coverage branch is cut from that base.
Read CLAUDE.md, README.md, docs/0.1.md, docs/memory.md, the complete three-dot branch diff,
the seven branch commit messages, and the affected lowering, IR, emission, runtime and oracle code.

The oracle requires explicit registration in fixtures, not just a file in testdata.
All fifteen new files are registered with lowers=true and checked=false.
All finish with exit 0. The compiler failure below is deliberately unregistered.

## Case inventory

Names below are files in internal/oracle/testdata. New names start with environment_coverage_.
The existing coverage column records what was already on the feature branch before this work.
Related IR wrapper cases are grouped by the source operation that produces them.

| Condition, value or call | Existing coverage | Added coverage or limitation |
| --- | --- | --- |
| No captured slots, no AllocateEnvironment | nested_minimal, nested_hoisting | Already covered |
| One shared entry record, ordered slots, captured numbers and parameters | nested_captures, environment_siblings, nested_pattern_parameter | primitives adds boolean and optional numbers/strings |
| Captured heap strings, mutation and replacement | environment_direct, nested_captures, environment_exits | values_exits uses runtime strings, replacement arrays and objects |
| Captured arrays and objects, children released on return and propagated throw | No explicit named-helper environment fixture | values_exits, region_values, callback_exits |
| Captured Map and Set references | No explicit named-helper environment fixture | collections, both return and throw |
| Weak slots and strong-cycle refusal | nested_weak; refusals/nested_cycle plus TestNestedCycleRefusal | Already covered |
| Zero-initialized slots, ready bits, uninitialized reads/writes, destructuring | nested_tdz, nested_tdz_write, nested_destructured_tdz, nested_destructured | initializer_throw exits while a later object slot is still uninitialized |
| At most 64 slots on stack | environment_direct, environment_exits | boundary64 checks exactly 64, normal and throw exits |
| More than 64 slots in a call-owned region | environment_large (65 strings) | region_values has 65 strings, arrays and objects, normal and throw exits |
| Heap fallback when any capturing carrier escapes | environment_returned and stores below | alias_join, callback_result, closure_self |
| Sibling direct calls, mutual recursion, shared layout | environment_siblings, nested_mutual | values_exits adds sibling calls over reference slots |
| Synchronous local alias invocation and loop calls | environment_direct, environment_loop | bounded_call passes a helper to a statically bounded const arrow |
| Named direct parameter forwarding, all-target fixed point | environment_local_call, environment_keeping_call | captured_parameter captures an argument in an escaping closure |
| Unknown function-value argument target | environment_unknown_call | Already covered |
| Missing/empty target set or missing parameter summary is conservative; every bounded target must permit borrowing | environment_unknown_call, environment_virtual_call; TestEnvironmentParameterAllTargets | A missing summary cannot be made by well-formed lowered source; the synthetic test covers it |
| Virtual call with a retaining override | environment_virtual_call | Already covered |
| Const declaration and assignment aliases, taint retained after reassignment | environment_direct covers const only | alias_join covers branch assignment, optional alias and reassignment after another alias is kept |
| Conditional, Coalesce, Narrow, Unwrap, MaybeOf wrappers preserve carrier taint | No explicit optional conditional alias fixture | alias_join exercises optional/coalesced aliases and conditional return |
| Captured/global declaration or assignment is an escape | environment_capture, environment_global | captured_parameter |
| MakeClosure retaining ancestor slots; scan nested bodies | nested_three_levels, environment_callback_escape | callback_result returns freshly created capturing closures through map |
| Return of a helper or aggregate holding it | environment_returned, nested_mixed | local_containers, alias_join |
| Return a slot value without returning a carrier: Call results are independent of carrier taint | Strings in environment_exits; no arrays or objects | returned_values returns strings, arrays and objects through helpers; values survive cleanup and calls have independent state |
| Stores into fields, array elements/push, Map and Set, global | environment_field, environment_array, environment_map, environment_set, environment_global | local_containers checks even local object/array/Map containers keep heap storage |
| ClosureSelf current carrier in a named function expression | library_fnexpr_recurse has self use but no named-helper frame | closure_self stores current in a local array with the shared record alive |
| Borrowed forEach, map, filter, sort and find callbacks | environment_callbacks | callback_exits throws through each callback with strings, arrays and objects in the environment |
| Unknown/mutable callback target keeps heap | environment_unknown_callback | Already covered |
| ArrayVisit some and every | Ordinary array fixtures, without explicit named-helper frame | callbacks_extra uses named helpers |
| ArrayReduce, ArrayFrom, MapForEach borrowing | closures_throw has ordinary arrow closures, without explicit named-helper frame | callbacks_extra uses named reduce and Map.forEach helpers plus in-place Array.from arrow |
| Callback creates escaping closure | environment_callback_escape | callback_result escapes through a returned map result |
| Callback operands other than carrier can be retained | Container escape fixtures already keep heap | local_containers, callback_result |
| Primitive consumers do not carry environments | environment_siblings, environment_callbacks | primitives, callbacks_extra |
| Normal, early return, exceptional propagation and finally cleanup | environment_exits, environment_large | values_exits, region_values, initializer_throw, callback_exits |
| Interior cell counts redirect to heap owner, uncounted stack/region counts do nothing | Existing escape and local environment fixtures | New heap, stack and region probes run with ASan, UBSan, release builds and leak checks |
| Unknown operation fallback; thrown carrier | Synthetic TestUnknownEnvironmentTransferStaysOnHeap, TestThrownEnvironmentStaysOnHeap | Cannot express these source programs on this base |

Generated C confirms these actual new placements (one environment allocation per listed program except returned_values, which has ordinary and region-returning C variants):

- Stack: boundary64, bounded_call, callback_exits, callbacks_extra, collections,
  initializer_throw, primitives, values_exits, returned_values.
- Heap: alias_join, callback_result, captured_parameter, closure_self, local_containers.
- Call region: region_values.

## Compiler failure

captured_reference_parameters.a is a legal source probe with a named helper capturing
unmodified string, array and object parameters. Original Node prints:

```text
value0:value0:value0!:value0:true
value1:value1:value1!:value1:true
```

There is no native binary and therefore no native program stdout. Native compilation panics:

```text
panic: native: a store into the borrowed parameter text
```

The oracle panics while generating C; go run ./cmd/adamic build also reports exit status 2
(the go run shell command exits 1). Node exits 0. This is a compiler failure, not a measured
runtime stdout mismatch. Full build trace is native-build.log; original output is node.stdout.

Suspected branch line: internal/native/emit_locals.go:156, makeCell now calls store for an
EnvironmentCell. internal/lower/borrow.go:28 still marks an unassigned reference parameter
Borrowed. Initializing the environment cell reaches store's Borrowed assertion at line 32.
The older independent-cell path initialized the cell without calling store. No compiler fix
is included in this coverage branch.

## Source cases that cannot be exercised

- Throwing a closure: internal/lower/exceptions.go:31 refuses all thrown values except Error.
  A throw new Error(helper()) tests cleanup but does not throw the carrier itself.
- Variable-sized environment layout, parallel-pool and async-frame transfer: no such source/IR
  facility on this base; synthetic fallback tests already exist.
- Named Array.from callback: lowering says Array.from requires an arrow written in place.
  The new program instead uses the supported arrow retaining the named-helper frame.
- A standalone undefined callback parameter and a string|null captured local: NotYet representations.
  The supported optional number/string slots are tested instead.
- Generic/block-scoped nested declarations, optional/default/rest nested parameters, dynamic this,
  rebinding nested declarations, and first-class sibling/ancestor function references: intentionally
  unsupported by nested_functions.go; existing TestNestedFunctionGapsAreLoud covers the refusals.
- A same-signature closure parameter captured by a same-signature returning closure can trigger the
  cycle-capable refusal. captured_parameter uses different input/output signatures to test the
  captured-parameter escape rule without that type-level possible cycle.
- Unmodified captured reference parameters: the compiler failure above prevents native execution.

## Mutation

Changed exactly one line in internal/native/emit_locals.go, allocateEnvironment's stack arm:

```go
// Original:
e.regions.cleanups[environment] = "adamic_environment_end(" + environment + ");"
// Mutant:
e.regions.cleanups[environment] = "(void)" + environment + ";"
```

Ran the uncached oracle on environment_coverage_values_exits.a. It exited 1 solely at the leak
check, reporting 756 bytes in 15 allocations. Node, JavaScript, sanitized native with leak
checking disabled, and release native still agreed. Thus no clang warning or unrelated output
failure killed the mutation. Restored the exact original source in a Python finally block.
The subsequent complete oracle uses the restored compiler. Log: /tmp/environment-coverage-mutant.log.

## Commands and observations

Every build/test shell sources /workspace/adamic-tools/env.sh, the path printed by setup.

```sh
bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
nproc
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/environment_coverage_' -count=1 -timeout 15m
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -count=1 -timeout 15m
go vet ./...
gofmt -l cmd internal
git diff --check
```

Setup was run twice: the first cache warm overlapped the branch checkout and failed against
mixed Go source versions. The second settled-branch setup passed. Timing lines: go ready 0s,
clang ready 1s, node ready 1s, submodules ready 1s, build cache warm 133s, done in 133s.
nproc: 5, cgroup cpu.max: 400000 100000. Go 1.27.1, clang 20.1.8, Node 24.19.0.

Focused oracle runs were logged in /tmp/environment-coverage-focused{,-final,-corrected,-pass}.log.
The first three exposed unsupported probe forms and the recorded compiler panic; after adapting
source forms, the initial fourteen passed in 3.713s. The first counts update passed in 22.489s and added
fourteen rows without changing any existing numeric row. The complete uncached oracle then
passed in 110.002s. A final returned_values probe was added to distinguish returning slot
contents from escaping a carrier; its focused oracle and standalone build/run passed. Counts
and the complete oracle were refreshed once more for the final fifteen-program registry.

The standalone build loop ran, for every new testdata program and the notes probe:

```sh
go run ./cmd/adamic build "$file" -o "/tmp/environment-coverage-builds/$name"
"/tmp/environment-coverage-builds/$name"
cp "$file" "/tmp/environment-coverage-builds/$name.mts"
node --disable-warning=ExperimentalWarning "/tmp/environment-coverage-builds/$name.mts"
```

Native execution was conditional on a successful build. Two corrected fixtures (bounded_call,
callbacks_extra) were rebuilt and rerun with the same commands. All fifteen final testdata
programs build and exit 0; Python compared every Node/native stdout and stderr byte for byte.
Logs: /tmp/environment-coverage-standalone{,-corrected}.log and per-program .build.log,
.native.out/.err and .node.out/.err under /tmp/environment-coverage-builds/.

Generated each program's C with go run ./cmd/adamic c "$file" and inspected allocation calls.
The mutant test command was:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/environment_coverage_values_exits.a' -count=1 -timeout 10m
```

Vet passed without diagnostics; gofmt listed no files. The full repository stage1 gate was not run.

The final additional focused command was:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/environment_coverage_returned_values.a' -count=1 -timeout 10m
```

The returned_values standalone run used the same CLI build, native run, .mts copy and Node
commands above, and cmp compared stdout and stderr. Logs: /tmp/environment-coverage-returned.log,
/tmp/environment-coverage-counts-final.log and /tmp/environment-coverage-oracle-final.log.

Final results: returned_values focused oracle passed in 0.546s; final counts refresh passed
in 16.790s; final complete uncached oracle passed in 104.924s. Final vet and formatting checks
passed with empty logs. git diff --check and git diff --cached --check were clean.
