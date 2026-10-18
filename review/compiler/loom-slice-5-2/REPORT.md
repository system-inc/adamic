Re-cut parked Loom change chg_a79f7td9k7pv84chh20ehjbqex as compiler/loom-slice-5-2, preserving slice 5's compiler and runtime meaning.
Source bcb03d95cab6a83e6a23cc2e1669d5c145d911ff; one squash commit on main 27d661a838681d3144808e108e3e61e32e940a69; the final delivery message records the pushed SHA.
Focused tests, 63 fixtures, eight witness-ledger tests, TestCallTargetReaders and changed-package vet all pass; counts regenerated once in 89.353 s.
Embedded placeholder, readiness, construction, stack, rewind, payload and retained-stopping-artifact mutants are caught by the focused and witness tests.
Not covered: the full gate, full admission census, complete adapted parser, Program-region lifetime, optional guard-cost measurement, WASI or non-Linux systems.

The only squash conflict was internal/oracle/counts.md. It was resolved by taking current main's table in full, then regenerating it exactly once. All other files merged automatically, retaining main's accessor-spread, static-constructor-spread, template and compiler build infrastructure alongside slice 5. No production file was manually edited for this re-cut.

The runtime delta is exactly the parked slice's own delta: identical ordered added/deleted lines in all 14 runtime paths, compared against its parent c0fd8945b872e57a341ccee5973e3f4d1b75c34b. Main's own runtime edits are retained. The checked proof and path list are in runtime-proof.json. No extra runtime changes were made.

The counts command uses main's current sharded family selector, as established in the first re-cut:

`go test ./internal/oracle -run TestCountsAreRecorded -count=1 -json -timeout 180s -args -update-counts`

It passed in 89.353 s. The final diff against main has 41 added rows, 21 changed existing rows and no removals. Every moved row is one of slice 5's own previously attributed rows. All exact before/after values and specific member/reason attribution are in counts-attribution.json. No unrelated row was changed, and no second regeneration ran. The twelve TestParserAheadCounts* leaves then pass without updating the table.

Toolchain setup used GOPROXY=https://proxy.golang.org|direct, ADAMIC_GOCACHE_OFF=1 and GOMAXPROCS=4. Its timing lines are Go ready 0.078 s, Node ready 0.082 s, submodules ready 0.104 s, markdown ready 0.220 s, clang ready 0.481 s, shared cache off 0.483 s, Go build ready 32.956 s, test binaries deferred 33.124 s, build cache warm 33.125 s and done 33.155 s. nproc=5; cgroup cpu.max=400000 100000, or 4 CPUs. Go 1.27.1, Node 24.19.0 and clang 20.1.8. Tests source /workspace/adamic-tools/env.sh, preserve its -trimpath -buildvcs=false flags, and append -p=4. Evidence lives in this directory; complete setup output is setup.log.

Validation was focused throughout; each test command uses -count=1, -json and a 90 s test timeout unless explicitly stated above. All output was captured directly to logs, without piping a test process. Exact selectors, commands and wall seconds are in focused-results.json and fixture-results.json. Every measured test leaf's seconds are in test-seconds.json. No measured leaf reached 60 s; the largest was a counts fixture, string_views_holders.a, at 16.22 s.

- python3 review/compiler/loom-slice-5-2/run-focused.py passes the slice's lower, native, oracle and CLI selections. This includes all eight TestSlice5Witness* tests, product relocation checks and embedded mutants. The Program-region and optional measurement skips remain explicit.
- python3 review/compiler/loom-slice-5-2/run-fixtures.py runs all 55 slice fixtures plus eight controls for main's merged accessor spread, static constructor spread, template nullish and long Unicode literal behavior. All 63 run exactly once through the selected oracle family cases. They pass against source Node, JavaScript, native ASan/UBSan, release native and successful-program leak checks as required by the oracle.
- TestSlice5Witness* validates all eight exact witness hashes and stopping diagnostics, then validates the type-correct repair hashes and Node/native/JavaScript agreement under sanitizers and leak checks. Every retained stopping-artifact mutant is rejected against its repaired source.
- go test ./internal/ir -run '^(TestCallTargetReaders|TestEveryMapAProgramHoldsIsSorted|TestAProgramIsTheSameBytesEveryTimeGobWritesIt)$' -count=1 -json -timeout 90s passes. TestCallTargetReaders takes 10.67 s; the package takes 10.684 s. No allowlist entry was added.
- go vet ./cmd/adamic ./internal/flow ./internal/fresh ./internal/ir ./internal/javascript ./internal/lower ./internal/native ./internal/oracle exits 0 with no diagnostics.
- gofmt and git diff --check are clean. No new Go package was introduced, so no compiler-dependency declaration is needed.

Historical independent overlay evidence from the first re-cut is preserved in review/compiler/loom-slice-5. Those 27 external overlay runners were not rerun for this counts-conflict repair; this run executes the embedded focused and ledger mutants described above. No cohere source was copied, no PR was opened and no Loom submission was made. The user resubmits the parked change using the new SHA.

Runtime paths preserved exactly:

- internal/native/runtime/adamic.h
- internal/native/runtime/array.c
- internal/native/runtime/class_features.c
- internal/native/runtime/class_static.c
- internal/native/runtime/construction.c
- internal/native/runtime/heap.c
- internal/native/runtime/json_stringify.c
- internal/native/runtime/library_language.c
- internal/native/runtime/library_object.c
- internal/native/runtime/object.c
- internal/native/runtime/region.c
- internal/native/runtime/stack.c
- internal/native/runtime/union.c
- oracle/adamic.mjs

Witness leaf seconds:

| Test | Seconds |
|---|---:|
| TestSlice5WitnessBeforeUse | 1.61 |
| TestSlice5WitnessAssignmentResult | 1.68 |
| TestSlice5WitnessAliasReset | 1.7 |
| TestSlice5WitnessNullBeforeUse | 1.68 |
| TestSlice5WitnessSavedLeak | 1.73 |
| TestSlice5WitnessNullSavedLeak | 1.74 |
| TestSlice5WitnessReturnAssignment | 1.76 |
| TestSlice5WitnessSpeculation | 1.45 |
