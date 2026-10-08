# Wave 14 shared fixes rerun

Merged `origin/lint-fix/jsx-inventory-discovery` at e7c196a97 and `origin/codex/parser-recovery-land` at dbe2fff02, using two merge commits, no rebase. No shared source was edited by this unit. Archived fixtures remain `.a.txt` and the namespace descriptor remains removed. The only owned control change logs the captured upstream inventory before each rule comparison.

The full lint package runs with TypeScript source pin 050880ce59e30b356b686bd3144efe24f875ebc8, ADAMIC_LINT_BENCH=1 and both profile directory/snapshot inputs set to /tmp/wave14-fixes-profile. GOMAXPROCS=4, nproc=5, GOFLAGS=-buildvcs=false; no diagnostic or mismatch probe enabled. Full package completion counts and timing are recorded beside the raw compressed logs in validation-wave-14-shared-fixes/.

## Original-context comparison completed

The owned overlay test compares each reconstructed Go result with the original upstream count before comparing Go, source Node, emitted JavaScript and sanitized native findings, automatic edits, suggestions and fixed results. It passes three rule subtests and fails three. This additional check is separate from the requested full lint package.

| Rule | Captured unique upstream cases | Original-context result |
| --- | ---: | --- |
| @typescript-eslint/no-array-delete | 44 | Pass, including owned witness |
| @typescript-eslint/no-extraneous-class | 79 | Pass, including five default/option witnesses |
| nexus/correctness-no-leaked-number-render | 52 | Pass, including actual JSX witness |
| @typescript-eslint/no-base-to-string | 323 | Blocked: project reconstruction drops moduleDetection force |
| nexus/correctness-no-global-listener-target-assertion | 6 | Blocked: same dropped compiler setting changes DOM augmentation |
| no-label-var | 22 | Blocked: capture/reconstruction loses checker presence |

The overlay fails at `stage1/cohere/lint/rules/typescript-no-array-delete/evidence/landing-controls.go.txt:103` for each blocked rule. Cohere's `defaultTsConfig`, `cohere/internal/lint/testing/program.go:22`, sets `moduleDetection: "force"`; shared replay's strict-only project is made at `stage1/cohere/lint/lint_test.go:378`. The original shadowed-String case is `cohere/internal/lint/rules/typescript/no_base_to_string_test.go:189`; the DOM augmentation case is `TestCorrectnessNoGlobalListenerTargetAssertionStaysSilentOnAMergedAddEventListener`, `cohere/internal/lint/rules/nexus/correctness_no_global_listener_target_assertion_test.go:188`. Checker-presence coverage is `TestNoLabelVarRequiresTheTypedHarness`, `cohere/internal/lint/rules/core/no_label_var_test.go:171`. Its typed and untyped runs must remain distinct even with identical source and options.

### Smallest verified reduced inputs

* `smallest-base-to-string.ts.txt`: `let String;String({})` (21 bytes). Go count is 1 under strict-only settings and 0 with moduleDetection force.
* `smallest-global-listener.ts.txt`: Window augmentation plus a listener asserting its target to SVGAElement (108 bytes). Go count is 1 under strict-only settings and 0 with moduleDetection force. Exact single-line input is in the fixture, and candidate-reduction observations are in reduced-candidates.json.
* `smallest-checker-presence.ts.txt`: `NaN:;` (5 bytes). Go count is 1 with a program/checker and 0 without a program.

These are raw TypeScript inputs, not Adamic modules; they stay `.ts.txt` and are excluded from the source corpus. Reductions use the independent Go oracle and are not claimed as a new native parity run.

## Full package observations

Both JSX integration tests and TestJsxInventoryDiscovery pass. TestRulesAgree passes, including all captured no-label-var cases and the former `undefined: for(;;) { break undefined; }` failure. That syntax issue no longer needs a #ykgsvgn reproducer.

TestCompilerAndStage1Agree fails at `stage1/cohere/lint/lint_test.go:467`: the broad stage1 walk includes the incoming parser's intentional negative input, `stage1/typescript/parser/testdata/lint_cases/decorated_async_promise_executor.ts:1`, whose source is `new Promise(@dec async () => {})`. The independent Go oracle rejects parse errors as invalid corpus. This is corpus selection, not a rule disagreement. The shared walk and fixture are unchanged. Smallest verified invalid-corpus input is `@` (1 byte), in smallest-invalid-corpus.ts.txt; `f(@x)` preserves the decorated-call-argument shape and is also independently verified rejected. No invalid fixture was silently excluded to make the package green.

### Snapshot disagreement isolated

`TestProfileSnapshotsAgree` fails at `stage1/cohere/lint/profile_test.go:225` on `stage1/typescript/parser/testdata/lint_cases/wave13_top_level_await_new.ts:1`. Native adds a `no-new` diagnostic; Go does not. This is separate from corpus selection. Selecting only `no-new` verifies the smallest reduced input `await new C` (11 bytes): Go count 0, native count 1, both exit 0 with empty stderr. The control `new C` gives 1 on both sides. Results are in isolated-no-new-counts.json. The operand is treated as a bare new-expression statement because the parser's await recognition at `stage1/typescript/parser/parser.ts:1792` accepts Identifier lookahead outside await context but not NewKeyword. The unchanged no-new listener is at `stage1/cohere/lint/rules/no-new/rule.ts:15`, and upstream's selector is `cohere/internal/lint/rules/core/no_new.go:68`. Exact input is in smallest-await-new.ts.txt. No shared parser or rule was edited.

## Final outcome

Full lint package: 33 pass, 2 fail, 1 skip; 82/82 registered mutant subtests pass, including all six owned mutants, and TestOwnedWitnesses passes. The two failures are TestCompilerAndStage1Agree (invalid negative fixture included in corpus) and TestProfileSnapshotsAgree (top-level await/new parser discrepancy). The skip remains TestCheckerBridgeRefusalPending at checker_pending_test.go:49, awaiting TSGoError from codex/tsgo-errors-as-values, not an omitted optional input.

The full package wall time was 1965.070 seconds, with package test execution 1952.997 seconds. nproc=5, CPU quota=4. The original-context overlay separately completes in 306.826 seconds, with three passing rule subtests and three failing as documented above. Gate and overlay ran concurrently initially; performance samples include that contention. The original-context failures must not be hidden by the passing TestRulesAgree replay.

The workspace approached capacity during the run; standard `go clean -cache` cleared rebuildable Go cache and restored space. No test failed due to storage. No full repository gate or standalone legacy suite was rerun. No new rule was claimed.
