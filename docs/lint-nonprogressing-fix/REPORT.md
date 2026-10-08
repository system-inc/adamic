# Nonprogressing automatic fixes

Built on area/stage1-lint 1f9e223d8. Source commit f141bc838.

`fixed()` resolves every overlap before validating survivors, matching cohere/internal/edit/engine.go:167-168. A no-progress survivor retains its reserved span and is rejected with `the fix replaces text with itself` (cohere/internal/edit/apply.go:194-212). Overlap refusals precede validation refusals; other edits still apply.

`TestNonprogressingFix` uses `type Shape = { value: string }` and option `interface`; Go, Node, emitted JavaScript and sanitized native match findings, individual edits, refusal and final source. The fixture is the actual consistent-type-definitions port copied into a scratch registry, rather than a second shared rule registration.

`TestNonprogressingFixPlanOrder` proves overlap reservation, refusal order, an independent edit, empty insertion rejection and UTF-8 wire positions from a non-BMP prefix. `TestNonprogressingFixPanicMutant` restores the panic in a valid compiled program; all three port runtimes exit 70 at execution. Focused suite: PASS, 120.587 seconds. Registry and vet pass; Go files formatted.

The exact prefer-promise-reject-errors `input.tsx` row `Promise.reject(<string>'x');` is covered by upstream capture fix 9a24f732: the scratch Go overlay of that commit captures recovery metadata and the unchanged typed oracle succeeds with the rejection finding retained. Probe PASS, 30.041 seconds. The capture fix is not changed or bundled here.

Full gate enables ADAMIC_LINT_BENCH=1, ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-07-typescript (v6.0.3), ADAMIC_LINT_PROFILE_DIR and ADAMIC_LINT_PROFILE_SNAPSHOTS. Command: `GOMAXPROCS=4 GOFLAGS=-buildvcs=false go test -json -count=1 -timeout=60m ./stage1/cohere/lint`.

Top-level counts: {'pass': 37, 'fail': 0, 'skip': 1}; all test counts: {'pass': 119, 'fail': 0, 'skip': 1}. Wall: 1905.165s. nproc: 5; mean load: [2.0226134452919946, 1.837659940944882, 1.4429069779363517]; peak load: [4.84619140625, 3.86181640625, 2.45947265625]. Registered mutants caught: 75.

The sole skip is TestCheckerBridgeRefusalPending, awaiting codex/tsgo-errors-as-values and TSGoError in internal/load/prelude.d.ts. All input-dependent tests ran. Existing explicitly unsupported recovery inputs remain recorded by the harness; this unit does not expand parser support. Invalid-range rejection parity is outside this no-progress unit.

Evidence: [focused.log](evidence/focused.log), [capture-probe.log](evidence/capture-probe.log), [full summary](evidence/full/summary.json), [raw events](evidence/full/events.jsonl.gz), [receipt](evidence/full/receipt.json), [registry](evidence/registry.log), [vet](evidence/vet.log).
