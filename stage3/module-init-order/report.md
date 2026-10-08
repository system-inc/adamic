Built: binding-specific enum readiness, source-order proofs, and zero-argument Map<never, never> construction.
Commits: codex/module-init-order starts at compiler/stage3-front-2 860a0d5; implementation tip is reported with the push.
Commands: loader/lower tests, independent uncached Node/native oracle, counts update, go build ./..., go vet ./..., gofmt.
Mutants: restoring whole-module enum guarding rejects the parser probe; dropping binding checks fails both premature-read pins; deleting source-order proof fails the no-check assertion.
Not covered: full parser slice, catchable enum-before-initialization exceptions, inhabited never values, and general Map operations with never arguments.

The exact parser probe from codex/stage3-parser-proof 6a420d2 originally failed at native-enum-map.a:3:52 with:

```
stage 0 can't lower an indirect call or class construction before enum initialization; declare enums before executable module code yet
```

The old enumInitialization pass counted every remaining enum in the graph, and rejected unknown calls and constructions while that count was nonzero. It did not establish a dependency of the Map on Pending. Removing that blanket condition exposed a separate rejection of the Map's never key. The fix permits only zero-argument construction of Map<never, never>, using existing empty-map storage; operations still require representable key/value types. No runtime never value is manufactured.

Enum initialization now proves direct enum binding reads only after their declaration has been reached in source order within ESM module order. Deferred function bodies and instance initializers retain checks on the actual enum read. A later enum does not obstruct an unrelated construction. Neither native emitter nor ready-check runtime changed.

Observations from TestModuleInitOrder:

| Fixture | Node | Native, JavaScript IR, release |
|---|---|---|
| Exact native-enum-map.a | stdout 0, exit 0 | stdout 0, exit 0; no Pending check in emitted C |
| Two-module cycle, Map module first | stdout map, enum, 0; exit 0 | stdout map, enum, 0; exit 0; no Pending check in emitted C |
| Actual early enum read | TypeError, exit 70 | pinned ReferenceError readiness panic, exit 70 |
| Construction reading later enum | TypeError, exit 70 | pinned ReferenceError readiness panic, exit 70 |

Node's native ESM execution uses the repository's TypeScript transform. Direct enum member constants are inlined by that transform, even before the declaration. The negative fixtures therefore use a computed key, preserving an actual runtime read. Node reports `TypeError: Cannot read properties of undefined (reading 'Zero')`: regular enums have hoisted var storage. Per the ruling, Adamic instead reports `adamic: panic: ReferenceError: Cannot access 'Pending' before initialization`. Both stop at the read, before further output. This is an intentional pinned error-class difference, not a claim of identical enum TDZ semantics.

Verification logs are under /tmp/module-init-*. Loader tests passed in 0.742s and lower tests in 13.416s. The full uncached TestNativeAgreesWithNode passed in 71.570s. The final four-fixture test passed in 1.976s, including sanitizers and release binaries. Go build and go vet passed. The original whole-module guard mutant compiles but fails map-before-enum at lowering. The unchecked-read mutant compiles but hits a sanitizer null-object access in both negative fixtures instead of the pinned readiness panic.

Setup: go ready 0s, clang ready 0s, node ready 0s, submodules ready 0s, build cache warm 165s; nproc 5 (cgroup quota four CPUs). Source /workspace/adamic-tools/env.sh for every command. No pushes or merges into main or area branches.

Counts update passed in 15.625s. Only two new rows appear: native-enum-map.a and module_init_order/enum.a, each allocations 2, frees 2, retains 1, releases 4, peak live 2, regions 0. Each constructs one Map and one enum object, and both are freed. Existing rows are unchanged. Negative probes are custom pinned-error tests, not generic Node-equality fixtures or new count-table rows.

Commands, with output redirected to logs:

```
ADAMIC_TOOLS=/workspace/adamic-tools bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
go test -count=1 ./internal/load ./internal/lower
go build ./...
go vet ./...
ADAMIC_GATE_UNCACHED=1 go test -count=1 ./internal/oracle -run '^TestNativeAgreesWithNode$'
ADAMIC_GATE_UNCACHED=1 go test -count=1 -v ./internal/oracle -run '^TestModuleInitOrder$'
go test -count=1 ./internal/oracle -run '^TestCountsAreRecorded$' -args -update-counts
```

Mutants use Go overlays on the corresponding lower file, leaving the working tree intact. Restore enum_initialization.go from 860a0d5 for the whole-module guard; replace checkedModuleRead with return false for unchecked reads; replace enumInitialization with return nil for missing proof. Run TestModuleInitOrder/map-before-enum for the first and third, and TestModuleInitOrder/(early-read|construction-reads-enum) for the second, all with ADAMIC_GATE_UNCACHED=1 and -count=1. All three compile and fail the intended assertions; logs are in evidence/.

Front-3 integration at the exact pinned 28e366fa9ffa8ad5a44e2ee8a0d97a38faf2c916 preserves the newer namespace/enum memoized reachability proof. Known premature function and construction reads remain NotYet; the two original negative source fixtures and their Node diagnostics are unchanged. Unknown mutable aliases and property callbacks still compile with readiness checks. The incoming direct source-order enum proof is composed with validation of namespace enum declarations. Existing never-map operations retain their prior closed empty-storage proof; the incoming zero-argument construction does not create an inhabited never value.

The integration controls restore the old blanket guard (safe map rejected), remove the direct enum proof (unexpected readiness check in C), and erase checked module reads (three unknown-callback probes hit UBSan null-object access instead of the pinned readiness stop). Raw evidence is in stage3/front-3/item19.
