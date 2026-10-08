4853a6a9b69a1636c8f7c1c3aa8f41958e5c8bd8 adds proven void closure values on a topic-only branch.
Own commits: 2e33164f, a1772221, 9e441a5f, 0e2db0f1, 4853a6a9; base 6998ebc24ae353193cb1495d3d51308131a4b5c7; no merge commits.
Focused Node oracles and refusal tests pass; counts refresh passes; vet and diff checks pass; setup 31.507s, nproc 5.
Nine mutants fail: five omitted operations, defined result, lost return, erased actual result, accepted unknown target.
Morning coverage: 28 reached, 3 retained, 61 unreached out of 92; unknown callables remain refused; no checked views used.

The fresh branch `codex/notyet-void-value-topic-0730` starts at the newest origin/main resolved for this task. Only my four original non-merge commits were cherry-picked. The counts conflict retained main's read_files row and inserted only my fixture rows. Earlier reports remain historical evidence on their stated compiler bases.

`voidValue` now admits CallClosure only when the existing IR ClosureTargets API proves its actual target and every target has Returns == 0. This covers a closure literal and a const initialized directly with one. Its closure and arguments are evaluated as helper arguments, then the call runs before undefined is returned. Unknown parameters, mutable values and number-returning arrows assigned to () => void remain refused. No other worker's lowering function, backend, IR file or runtime C helper was changed by the new rule. This unit owns the void-value reason; the separate callClosure result-representation reasons remain outside this function.

The new `.a` fixture is registered through internal/oracle/void_value_test.go. It observes argument order, calls through const and literal closures, and throwing calls. TestNativeAgreesWithNode compares it with Node in release native, sanitized native, and JavaScript. Its counts are 8 allocations, 8 frees, 11 retains, 23 releases, peak 4, regions 0.

Commands (each test writes directly to its log):

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/notyet-void-topic-setup.log 2>&1
source /workspace/adamic-tools/env.sh
nproc
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle ./internal/lower -run 'TestNativeAgreesWithNode/internal/oracle/testdata/void_value_|TestVoidValueErasedResultsStayNotYet|TestLibraryMapSetGapsStayRefused' -count=1 -timeout 10m > /tmp/notyet-void-topic-restored.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 10m -args -update-counts > /tmp/notyet-void-topic-counts.log 2>&1
ADAMIC_GATE_UNCACHED=1 python3 internal/lower/testdata/run-void-value-mutants.py > /tmp/notyet-void-topic-mutants-final.log 2>&1
go vet ./internal/lower ./internal/oracle > /tmp/notyet-void-topic-vet.log 2>&1
git diff --check
```

Focused output: oracle 0.555s, lower 0.367s. Counts: oracle 34.746s. Every mutant exits 1 and sources are restored in finally. Drop-call is caught by binder; drop-closure by closure; drop-array-visit, drop-map-clear and drop-map-visit by builtin; defined-result by binder; lose-return by checker. These seven fail Node stdout comparison. Erase-callable-result and accept-unknown-target fail the respective lower refusal assertions (`got <nil>`), proving both actual-result and target-provenance checks can fail. Each log is included here.

The morning table is e8c283b5ed32477805357b652a170b85a04b2469, stage3/notyet-table/rerun-0730/after/roots.csv. Exact NotYet reason filtering and (kind, where, reason, text) deduplication yields 92 roots. All were replayed with the guarded measurement worker from 9a1f14c5, extracted into scratch rather than brought into branch history. Its overlay was generated from the current topic sources, with a scratch-only marker just before successful voidValue return. The measurement command was `python3 /tmp/notyet-void-topic-count.py`; its worker command is `/tmp/notyet-void-topic-replay/worker -project /tmp/notyet-void-adapted/src/tsc/tsc.ts -where /tmp/notyet-void-adapted/<CSV where> -kind NotYet -reason 'a void call used as a value'`. The marker proves exact-site reach for 28 roots across the carried and new rules; it does not claim 28 additional roots from the closure extension or whole-unit compilation. The 61 unreached sites have no coverage claim.

The same binder:3644:15 replay selects bindEnumDeclaration and reports no findings. Checker:8239:52 selects bindElement and advances to NotYet reading add at checker:8242:41. Retained exact roots are sys:569:12 and transformers/module/impliedNodeFormatDependent:71:20 and :73:16. Their mutable or parameter callable targets are unknown. Accepting them as undefined requires actual-callee proof or a result-preserving callable representation; a void annotation alone is insufficient. No checked-view-dependent rule was added.

See sites.csv and coverage.json.gz for all 92 statuses and next findings. Setup timings, focused checks, counts, and every mutant log are included. No PR, area push, main push, or unrelated worker commit is part of this branch.
