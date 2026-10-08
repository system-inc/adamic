# Wave 10 unified checker landing

Area merged, not rebased: `c8f6d74f5387df926ae12d8babd2780564b60097`, merge `24b69ed57b0b5bdaaa4b5d207c1eb317dda6d697`. The sole conflict was the checker question imports; both registrations and all existing checks remain. The retained wave question sources were updated for the area's rooted-path TypeScript API. Source migration `e203ee3b`; loop pending move `6868fb83`. No main or area branch is pushed.

`codex/lint-port-typescript-eslint-init-declarations` at `12a7c6a1cc53f2e55eabcc1df449f3f7d31c7b64` is already an ancestor of the area. Its current-area implementation is rechecked with the rest. No new rule is claimed.

## Certified active adapters

The active adapters use only the area's shared `RuleContext.checker` and shared recording/replay. Typed factories initialize their decision objects on first listener use because syntax-only fix replay constructs factories without a checker. Each descriptor has `node: true`, named kinds, a full upstream test prefix, an unchanged Go oracle and a firing witness. Helpers used by one rule stay in its directory.

| Rule | Unique captured upstream cases | Unified directory |
| --- | ---: | --- |
| @typescript-eslint/no-redundant-type-constituents | 108 | ../../lint/rules/no-redundant-type-constituents/ |
| @typescript-eslint/prefer-includes | 67 | ../../lint/rules/prefer-includes/ |
| nexus/correctness-no-uncleared-race-timeout | 21 | ../../lint/rules/no-uncleared-race-timeout/ |
| require-await | 88 | ../../lint/rules/require-await/ |
| symbol-description | 30 | ../../lint/rules/symbol-description/ |
| @typescript-eslint/init-declarations, already landed | 92 | ../../lint/rules/typescript-init-declarations/ |

`TestRulesAgree` passed in 201.900 seconds, comparing 4,193 captured upstream cases and generated cases. Its 416 live typed comparisons (the five new rules plus the existing pilot) match findings, fixes and suggestions byte for byte between Go, sanitized native, source Node and emitted JavaScript. Go's aggregate program-and-lint time was 21.029 seconds; native program, lint and transcript recording was 29.469 seconds (1.401 times Go). This aggregate includes the existing pilot and is not a steady-state per-rule benchmark.

`TestCompilerAndStage1Agree` passed in 409.010 seconds on 964 files, with 31,863,700 bytes identical on those four runtimes. The TypeScript checkout is pinned and clean, with its exact commit and path in the input record. These corpus rows do not open a checker program, so typed rules produce explicit no-program coverage records there; this is not fresh typed compiler-corpus certification. The prior standalone corpus evidence remains historical.

## Pending shared dependencies

The pending adapters retain their decisions, witnesses, oracle adapters and mutant descriptors outside active registry discovery. They are not certified as unified rules. The legacy standalone implementations and their previous evidence remain in place.

- `no-unmodified-loop-condition`: valid unnamed default function declaration. `TestNoUnmodifiedLoopConditionUnnamedFunctionDeclarationHasNoNameToReach`, `cohere/internal/lint/rules/core/no_unmodified_loop_condition_test.go:241`, fixture at line 251. Native parser `primary`, `stage1/typescript/parser/parser.ts:176`, rejects `CloseParenToken` at 52. Its unified witness matched all four runtimes, but full upstream replay failed and the unified mutant is not certified. See [pending report](unified_pending/no-unmodified-loop-condition/REPORT.md), [reproducer](evidence/unified-landing/loop-repro.ts.txt), [failed run](evidence/unified-landing/loop-parity-failure.jsonl.gz), and [passing Go case](evidence/unified-landing/loop-upstream.log).
- `nexus/correctness-no-process-exit-after-output`: `correctnessNoProcessExitAfterOutputIsProcessMember`, `cohere/internal/lint/rules/nexus/correctness_no_process_exit_after_output.go:410`. Shared upstream capture discards companion declaration files and modules; foreign-node raw questions are also not implemented. See [pending report](../wave_10_next/unified_pending/no-process-exit-after-output/REPORT.md).
- `nexus/correctness-require-blocking-standard-streams`: `correctnessRequireBlockingStandardStreamsBuildIndex`, `cohere/internal/lint/rules/nexus/correctness_require_blocking_standard_streams.go:251`, needs `SourceFiles` at 259 and `ResolveModule` at 279. Same multi-file capture and foreign-node dependencies. See [pending report](../wave_10_next/unified_pending/require-blocking-standard-streams/REPORT.md).
- `structure/react-hook-no-any-type`: `reactHookNoAnyTypeBlindedRules`, `cohere/internal/lint/rules/structure/react_hook_no_any_type.go:290`, needs the linked rule catalog (`rule.Registered`, filtering `ResolvesReactValueTypes`). There is no shared catalog helper. A frozen catalog is not used as a substitute. See [pending report](react_hook_no_any_type/unified_pending/REPORT.md).
- Previously parked and unimplemented React claims (`react/forbid-elements`, `react/forbid-prop-types`, `react/iframe-missing-sandbox`) are not activated or claimed green in this turn. No new work was taken.

## Inputs, checks and evidence

The full lint command, all environment inputs, source commits and clean TypeScript status are in [all-inputs.json](evidence/unified-landing/all-inputs.json). [run-all.py](evidence/unified-landing/run-all.py) supplies the corpus, benchmark and fresh profile-directory inputs and records counts, wall time and sampled load in [all-inputs-results.json](evidence/unified-landing/all-inputs-results.json). Output goes directly to a log; the final log is compressed only after completion.

The existing `TestCheckerBridgeRefusalPending` skips at `stage1/cohere/lint/checker_pending_test.go:49` because `internal/load/prelude.d.ts` lacks the pending `TSGoError` API. It explicitly awaits `codex/tsgo-errors-as-values`. No input enables this check on this area revision; no guard or skip is changed. This prevents a zero-skip result.

Registry rejection and mutation controls pass: [registry-tests.log](evidence/unified-landing/registry-tests.log). Gofmt reports no files: [gofmt.log](evidence/unified-landing/gofmt.log). Vet passes: [vet.log](evidence/unified-landing/vet.log). Raw checker question tests pass: [bridge-questions.log](evidence/unified-landing/bridge-questions.log). All pending adapters type-check: the three pending-types logs in the evidence directory.

`TestBridge` passed in 103.998 seconds: released-handle checks, independent Go/native comparisons, sanitizer and leak checks, and source-position, length, lifetime and ownership mutants. See [released-handles.log](evidence/unified-landing/released-handles.log). Setup succeeded after retained legacy rooted-path API fixes; its full timing lines are in [setup.log](evidence/unified-landing/setup.log). Total setup was 25.583 seconds, nproc 5, cgroup quota 4 CPUs.

The six owned mutant statuses and three-runtime output differences are in [owned-mutants.json](evidence/unified-landing/owned-mutants.json). No compile failure is counted as a caught rule mutant. The loop adapter and other pending rules are explicitly excluded from that certification.

Final full package result: **118 pass, 0 fail, 1 skip** (34 pass, 0 fail, 1 skip at top level), package 2115.202 seconds, outer wall 2117.159 seconds. The sole skip is `TestCheckerBridgeRefusalPending`, named above. nproc 5; sampled one-minute load mean 2.161, max 5.065. All 80 registered mutants were caught, including all six owned mutants on native, Node and emitted JavaScript. All owned witnesses passed. Required checks and their timings are in [required-checks.json](evidence/unified-landing/required-checks.json); the complete log is [lint-all.jsonl.gz](evidence/unified-landing/lint-all.jsonl.gz). The TypeScript checkout remained clean. The branch is pushed with these explicit shared blockers; a zero-skip certification is not asserted.
