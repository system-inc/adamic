# Wave 14 shared-checker landing

Merged `origin/area/stage1-lint` at 9b7547976bea1b00f8b04508b9102d03511f65b3 into the existing branch; no rebase and no new claims. The incoming checker, diagnostics, leak checks and oracle guards remain intact.

## Changes

Seven existing owned ports now have node listeners in the unified registry, named kinds, verbatim messages, independent upstream Go adapters and owned witnesses. Their queries go through `RuleContext.checker.ask`: no private program, fake answers or rule-owned recording. Source Node and emitted JavaScript replay the shared native checker transcript. The facade maps byte offsets to UTF-16 at diagnostic construction and preserves every automatic and suggestion edit.

The old standalone proofs are retained, with mechanical updates for current rooted path APIs and named kinds. Existing checker question registrations are retained; shared `facts.go` has only the required rooted-path cast. No shared registration generator or test harness was changed. The additional fidelity check is an owned Go-overlay test in `../lint/rules/typescript-no-array-delete/evidence/landing-controls.go.txt`.

## Original upstream fidelity

The overlay compares each fresh Go finding count with the original upstream capture before comparing findings, fixes and suggestions across Go, Node, emitted JavaScript and sanitized native. This prevents both implementations agreeing vacuously after losing project context.

* No-array-delete: all 44 upstream cases and witnesses pass.
* No-extraneous-class: all 79 upstream cases and default plus four option witnesses pass.
* No-leaked-number-render: all 52 upstream cases and actual JSX witness pass.
* No-base-to-string: shared capture loses `moduleDetection: "force"` from `defaultTsConfig`, `cohere/internal/lint/testing/program.go:22`. Reproducer: `const String = (value: unknown) => 'safe'; String({});` (`cohere/internal/lint/rules/typescript/no_base_to_string_test.go:189`). Original Go count 0; reconstructed shared Go count 1. Shared reconstruction is `stage1/cohere/lint/lint_test.go:378`.
* No-global-listener-target-assertion: the same dropped project setting changes `declare global` DOM augmentation in `cohere/internal/lint/rules/nexus/correctness_no_global_listener_target_assertion_test.go:193`. Original Go count 0; reconstructed shared Go count 1. Full source is in the retained overlay log.
* No-mock-on-module-namespace: capture stores other-file count rather than contents or compiler settings (`capturedRun`, `cohere/internal/lint/testing/docs_capture.go:18`). `correctnessNoMockOnModuleNamespaceRunsAsModule`, `cohere/internal/lint/rules/nexus/correctness_no_mock_on_module_namespace.go:148`, requires ESM settings. The default shared witness config (`registration_test.go:47`) supplies only strict mode. Original upstream count 2; shared count 0. A `.mts` witness cannot work around this because `registry/registry.go:85` accepts only ts/tsx/js/jsx. The shared mutant therefore survives; the retained module-aware legacy mutant is caught.
* No-label-var: capture loses the no-checker condition of `TestNoLabelVarRequiresTheTypedHarness`, `cohere/internal/lint/rules/core/no_label_var_test.go:171`. Reproducer: `function bar(x) { x: for(;;) { break x; } }`; original count 0, shared count 1. Separately, the native parser rejects `undefined: for(;;) { break undefined; }`, upstream line 88, with `expected semicolon at 9`. This blocks TestRulesAgree, TestNodeTableIsLinkOnly and TestShardsAgree.

## Retained partial regex ports

These two old partial ports are not claimed complete and do not have new unified descriptors. Current native lowering still refuses `new RegExp(pattern, 'u')` when pattern is nonconstant. The owned reproducer runs successfully in Node and fails native lowering at 2:23. Exact upstream dependencies are `invalidPatternMessage` / `esregexp.Compile`, `cohere/internal/lint/rules/core/no_invalid_regexp.go:251,263`, and `regexsyntax.ParseRegexCharacterClassWithEnd`, `cohere/internal/lint/rules/core/no_misleading_character_class.go:464`. No new hand-rolled regex matcher was added.

## Shared package blockers

The JSX inventory check has a fixed expected set at `stage1/cohere/lint/jsx_integration_test.go:61`; this branch adds 51 genuine upstream JSX sources, so both JSX integration checks fail despite the rule's original-case comparison passing. The source corpus also includes previously committed raw negative regex evidence, `validation-wave-14-constructors/upstream/upstream-067.a`: `new RegExp("[ \\ufe\60f]")`. The Go corpus guard correctly rejects its octal escape. That input is preserved and the guard is not relaxed.

`TestCheckerBridgeRefusalPending` still skips at `stage1/cohere/lint/checker_pending_test.go:49`: `TSGoError` is absent pending `codex/tsgo-errors-as-values`. No optional input was omitted to induce a skip. Full package result, timings, machine load and retained logs follow in `validation-wave-14-shared-checker/RESULTS.md`.

## Observed runtime timing

The per-rule original-case run logged the following paired times (milliseconds). Go includes program creation and lint; native includes program creation, lint and transcript recording. Concurrent verification causes contention; these are observations rather than isolated performance claims.

{
  "@typescript-eslint/no-array-delete": {
    "pairs": 45,
    "Go_median_ms": 61.722075,
    "native_median_ms": 78.044222
  },
  "@typescript-eslint/no-base-to-string": {
    "pairs": 177,
    "Go_median_ms": 69.425792,
    "native_median_ms": 86.560853
  },
  "@typescript-eslint/no-extraneous-class": {
    "pairs": 84,
    "Go_median_ms": 82.41096300000001,
    "native_median_ms": 101.569264
  },
  "nexus/correctness-no-leaked-number-render": {
    "pairs": 53,
    "Go_median_ms": 86.46751,
    "native_median_ms": 107.672905
  },
  "no-label-var": {
    "pairs": 7,
    "Go_median_ms": 87.034021,
    "native_median_ms": 106.889886
  }
}

## Retained checks completed

`go test ./stage1/cohere/typeaware -run '^TestWave14' -count=1 -v -timeout=30m` passes all five top-level tests in 1187.684 seconds. Both compiler and repository manifest inputs were set for the original and next suites. The log proves 21 semantic mutants, nine named-kind declaration mutants and three released-handle registry controls caught. Normal and sanitized corpus comparisons pass, and the old partial regex/label parser refusals remain explicit checked outcomes. These retained proofs do not certify the two partial regex ports as complete.

`go test ./bridge/tsgo/checker -count=1 -v` passes eight tests; lint-registry, gofmt, go vet on the touched checker/typeaware/lint packages and git diff --check pass. The corrected class diagnostic mutant is caught on native, Node and emitted JavaScript (`wave14-shared-class-mutant.log`); its first full-gate attempt failed to compile and remains retained. The namespace shared mutant is blocked by the missing ESM witness configuration described above.
