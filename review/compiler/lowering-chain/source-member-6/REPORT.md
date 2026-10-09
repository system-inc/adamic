Built a located namespace-value capability boundary and live-binding witnesses; escaped containers remain NotYet.
Commit: see compiler/namespace-value HEAD; base 7a10c877; no source branch imported and no conflicts.
Checks: build, internal vet, changed-file a-check, stage1 probes, stage3 fixtures, focused oracles and counts passed.
Mutants: three object snapshots, the dropped escape guard, eight semantic and three state mutants were caught.
Not covered: runtime namespace containers and full tsc compilation; no language policy changed.

The existing representation flattens namespace exports into singleton slots and a readiness flag. It has no canonical object whose fields share those live slots. Copying current values into an object breaks reads after qualified writes and writes through aliases. The stored and returned witnesses observe Node output `2\n7\n`; snapshots produce `1\n2\n`. The passed witness observes `2\n`; its snapshot produces `1\n`. Both compiled backends match each snapshot's own Node source, finish normally, and differ from the original only in stdout. Sanitizers and leak checks pass.

The new diagnostic names the namespace, its declaration location, the escaping expression location, and a fix: qualified member access or fixed namespace functions plus explicit state. This is a temporary NotYet capability boundary, not a permanent language refusal. The original tracing fixture is unchanged and still stops at 38:11. No ruling is needed for retaining that boundary. Qualified live-state access and function values already work; the two passing controls demonstrate those alternatives, not new container support.

Production change: internal/lower/namespaces.go calls the small namespaceValueNotYet hook in internal/lower/namespace_value.go. No protected lowering, native, JavaScript backend, splitter or oracle runner file changed. No cohere implementation was copied. tracing-qualified.a derives from the existing stage3 reduction and changes only the terminal driver to qualified access.

Commands, all with output redirected to the corresponding evidence log:

- `go build ./cmd/adamic`: pass.
- `go vet ./internal/...`: pass.
- Changed-file a-check: pass on exactly the five new .a files. Current main has no local a-check executable. run-a-check.py invokes the unchanged Gate.aCheck from developer-tools gate pin fbac28c62493f27a788edc02a18bc8edb68de5da, cloud/fast-gate/run.py. Its checked outcome includes named NotYet boundaries; it does not mean all five programs execute.
- `go test ./stage1/... -run 'Gap|Gaps|Probes' -count=1 -timeout 30m`: all 35 package rows pass.
- `go test ./stage3/fixtures -count=1 -timeout 30m`: pass, 27.354s.
- `go test ./internal/oracle -run '^(TestNamespaceValueBoundary|TestTracingNamespaceValueBoundary)$' -count=1 -timeout 30m -v`: pass, 1.996s.
- `go test ./internal/oracle -run 'TestNativeAgreesWithNode/stage3/namespace-value/' -count=1 -timeout 30m -v`: pass, 1.420s. Both valid controls match source Node in JavaScript and release/sanitized native, including stderr and exit status, with leak checks.
- `go test ./internal/lower -run '^(TestNamespace|TestTracingNamespaceEscape|TestDebugNamespace|TestTscNamespace|TestCallableNamespace)' -count=1 -timeout 30m -v`: pass, 1.900s.
- `go test ./internal/oracle -run '^(TestNamespaceSemanticMutants|TestNamespaceStateMutants)$' -count=1 -timeout 30m -v`: pass, 1.616s.
- `go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts`: pass, 55.731s.

Counts: only two rows added. qualified-live.a has A/F/R/L/P/G 4/4/9/15/3/0; tracing-qualified.a has 11/11/12/21/8/0. They count the two new runnable controls. Existing rows are unchanged. Negative witnesses do not enter executable fixture counts.

Mutant evidence: TestNamespaceValueBoundary independently catches stored, passed and returned snapshot replacements against original Node stdout in both backends. A Go overlay removing namespaceValueNotYet fails TestNamespaceValueBoundary and TestTracingNamespaceValueBoundary at the original boundary assertion (exit 1); it is not a later native compilation failure. The compressed patch is evidence/drop-escape.patch.gz. Existing semantic mutants cover scoped function, constant and enum binding, body order, exported state, Debug initialization, returned assignment and factory binding. State mutants cover lost assignment, lost hoisting and skipped readiness check. Every named catcher passed; their logs record each individual subtest.

Stock replay: unchanged TypeScript pin 050880ce59e30b356b686bd3144efe24f875ebc8. `go run ./cmd/adamic c /tmp/checked-any-stock/src/compiler/tracing.ts` stops before lowering at src/compiler/binder.ts:1109:17, TS2412 (undefined into FlowNode under exactOptionalPropertyTypes). The tracing escape's source provenance is src/compiler/tracing.ts:95:19, not a reached full-project stop. observations.json pins the source hash and distinguishes these observations. No revealed-byte claim is made.

Setup: node 0.019s, Go 0.022s, submodule 0.061s, markdown 0.069s, clang 0.181s, build 42.403s, tests deferred 42.609s, cache warm 42.610s, done 42.639s. nproc 5, quota 4. Environment script /workspace/adamic-tools/env.sh; GOPROXY https://proxy.golang.org|direct.

Toward step 15: the remaining escape wall now has a precise location and actionable fix, with evidence preventing a snapshot miscompile. Namespace objects themselves remain unsupported until live slot aliasing, identity, receiver behavior and staged property installation can all be represented.

Evidence reproduction: export GOPROXY and source the environment script above. For a-check, create /tmp/namespace-value-gate, write `git show fbac28c62493f27a788edc02a18bc8edb68de5da:cloud/fast-gate/run.py` to /tmp/namespace-value-gate/gate.py, then run evidence/run-a-check.py from the repository root. For the guard mutant, decompress evidence/drop-escape.patch.gz, apply it to a scratch copy of internal/lower/namespaces.go and use Go's -overlay Replace mapping from the repository file to that copy. Run the focused boundary selector above with `-overlay overlay.json`; expected exit is 1, with both named boundary tests failing. The working-tree production source is never mutated.
