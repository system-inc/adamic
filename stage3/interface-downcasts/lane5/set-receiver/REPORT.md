Built: Set receivers keep their actual map allocation through structural calls; callable views certify intrinsic add/has and add returns the same receiver.
Commits: shared changes 09d5525f1aaf35523d39b5ff9c82e88462c63a84; intrinsic adapters, tests and evidence 058635b99b58bea2b242e63afdcf80ebf96c2f6e.
Commands and outputs: focused Node/native/JavaScript/sanitizer/release/leak/Werror tests PASS; nearby callable and direct Set regressions PASS; vet PASS; seven counts rows updated and verified.
Mutants: uncast map caught by clang -Werror; native and JavaScript missing self returns caught by Node comparison; missing self result certificate and both skipped signature validators caught by runtime oracle checks. All six fail independently.
Not covered: other intrinsic Set methods, aggregate element representations, consumed structural argument paths, full package/gate runs; global counts refresh still fails on 39 fixtures outside this unit.

## Representation and contract

The producer-to-receiver boundary is `emitter.arguments` in `internal/native/emit_functions.go`: a Set lowers as ir.Map, while the structural parameter lowers as ir.Object. The explicit pointer conversion retains the allocation and its heap tag. Checked callable and field reads inspect that tag before accessing object layout and dispatch to the Set adapter. This is an ABI conversion, not a forged object shape.

Conservative assumption: structural Set views preserve the original map allocation. Set constructors alone establish intrinsic identity and element representation. Native method code identities select static const signatures; JavaScript tracks producer representation in an internal WeakMap and returns frozen signatures only for the actual Set.prototype method. The view cannot manufacture its producer's certificate. Number, boolean and string elements are certified; unsupported elements/methods refuse loudly.

`add` retains the returned receiver for its caller and holds reference keys before insertion. `has` returns boolean. Checked `size` reads dispatch to the actual Set representation. The identity fixture uses a dynamic string, proves result === receiver, calls has through the returned view, and checks duplicate insertion size. Wrong parameter and result views stop at read time with exit 70 and the field, expected signature and incompatible representation in the diagnostic.

## Shared ownership commit

09d5525f isolates changes outside the callable adapter files:

- internal/native/emit_functions.go
- internal/native/emit_expressions.go
- internal/native/runtime/adamic.h
- internal/native/runtime/map.c
- internal/native/runtime/object.c
- internal/javascript/javascript.go
- internal/javascript/readiness.go

No protected emit.go, lower.go, native.go or oracle_test.go edits. Branch begins at the requested origin/codex/views-callables 926a1d39d1a0d6b4cf49bbccf521a4cc02d10f56. No other lane or main merged. No cohere source copied.

## Commands and evidence

Commands ran from the repository, with output redirected to logs. Each Go command sourced `/workspace/adamic-tools/env.sh`.

| Command | Result | Log |
|---|---|---|
| `export GOPROXY='https://proxy.golang.org\|direct'; bash cloud/setup.sh` | PASS | logs/views-set-setup.log |
| `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewSetIntrinsic' -count=1 -v -timeout=5m` | PASS, 4.909s; five Node controls, seven C builds, two incompatible-signature probes, counts verification | logs/views-set-final.log |
| `python3 stage3/interface-downcasts/lane5/set-receiver/run-mutants.py` | PASS: six independent mutants each caught with exit 1; sources restored | logs/views-set-mutants.log and mutant-*.log |
| `go test ./internal/native ./internal/javascript -run 'TestViewCallable' -count=1 -timeout=5m` | PASS, native 15.066s, JavaScript 3.230s | logs/views-set-regressions.log |
| `go test ./internal/oracle -run '^TestCheckedViewCallableMethods$\|^TestCheckedViewCallables$' -count=1 -timeout=5m` | PASS, 2.201s | logs/views-set-oracle-regressions.log |
| `go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/^(sets\|set_foreach_numbers\|set_foreach_objects\|set_undefined\|library_map_set_keys)\.a$' -count=1 -timeout=5m` | PASS, 2.181s | logs/views-set-ordinary-sets.log |
| `go vet ./internal/native ./internal/javascript ./internal/oracle` | PASS, empty output | logs/views-set-vet.log |
| `go test ./internal/oracle -run '^TestCheckedViewSetIntrinsicCounts$' -count=1 -timeout=5m -args -update-counts` | PASS, 0.922s; added seven rows, changed no existing rows | logs/views-set-counts-focused-update.log |
| `npm ci --prefix stage3/api` | PASS; installed pinned Node types | logs/views-set-node-types.log |
| `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout=30m -args -update-counts` | FAIL, 52.636s, 39 fixture failures after dependency installation | logs/views-set-counts-global-node-types.log |

Setup: Node ready 0.077s; Go 0.088s; clang 0.513s; markdown dependencies 0.985s; submodules 397.510s; go build 751.155s; cache warm 751.296s; done 751.342s. `nproc` is 5; CPU quota is 4 cores. Go 1.27.1, clang 20.1.8, Node 24.19.0. The setup environment file was `/workspace/adamic-tools/env.sh`.

The first counts attempt lacked pinned Node types; npm ci removed those load failures. The retry still observes graph-region invalid frees, unsupported lowering and other failures outside the Set probes. No base-branch comparison was run to establish their provenance independently. The focused updater preserves all existing rows, and the final suite verifies all seven new rows. Panic fixtures record allocations at their deliberate exit 70; successful fixtures have matching allocations and frees.

An initial count/regression invocation used invalid `-timeout30m`/`-timeout5m` syntax; corrected commands above supersede those invocation failures. Earlier exploratory tests exposed the missing JavaScript size read and were superseded by the final green suite. No full package or gate run was performed.

## Mutant observations

| Mutation | Catcher |
|---|---|
| Remove `(adamic_object *)` receiver conversion | Every-probe C build reports incompatible pointer types under production -Werror |
| Native add returns NULL after insertion | Node control detects exit/output mismatch |
| JavaScript add discards Reflect.apply result | Node control detects exit/output mismatch |
| Native Set<string>.add certificate claims void result | Positive add probe refuses incompatible result representation |
| Skip native callable signature validator | Wrong-parameter/result fixtures fail required exit-70 refusal |
| Skip JavaScript callable signature validator | Wrong-parameter/result fixtures fail required exit-70 refusal |

Runtime mutants compiled successfully and failed semantically; runner rejects sanitizer/compiler-error-only failures for those mutants. The compiler mutant is intentionally caught before execution.
