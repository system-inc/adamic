# Wave 24 checker-area merge

Merged `origin/area/stage1-lint` at `c4bdc23fa86d55cf7e579989201c11258f4d3a62` with a merge commit, `d392c82e24e20f71e64d99cd58ea0a703e1c8e39`. No rebase, new claim, protected compiler edit or test removal. `codex/lint-port-no-caller` is already contained in this area tip.

The wave Rules fact gateway and fifth-batch node gateway now call the area's `RuleContext.checker.ask`; they retain their existing single program handle. The bridge question implementations and every inherited assertion are adapted to the new cohere pin's `RootedFilePath`, `PathKey`, and resolved-module API. No recorded-answer stand-in is added.

`symbol-description` and `valid-typeof` have unified typed descriptor directories, handed-node listeners, verbatim messages, upstream prefix capture, JSON options adapters, witnesses and mutants. `valid-typeof` now covers requireStringLiterals and quoting-undefined suggestions. Current runtime results and the complete package counts are recorded in evidence/unpark below.

## Standalone driver blocker

A fresh CLI compiler refuses `Checker.askFile` at `stage1/cohere/lint/checker.a:66:20` through `lowering.useOfThis`, `internal/lower/class.go:570`. Reproducer, from the repository root after building the archive:

```sh
go build -o /tmp/unpark-adamic ./cmd/adamic
go build -buildmode=c-archive -o /tmp/unpark-checker.a ./bridge/tsgo/archive
/tmp/unpark-adamic build stage1/cohere/typeaware/wave-24-fifth/suite.a -o /tmp/unpark-fifth --tsgo /tmp/unpark-checker.a
```

The original, next, third and fifth standalone suites stop before parity, mutants or released-handle checks. The next suite first reports the same refusal at `Frames.field`, `stage1/cohere/typeaware/frames.ts:42:23`. The fresh fifth recheck confirms the shared Checker refusal after the owned constructor fix. These suites are not freshly green; earlier reports remain historical evidence only. The unified harness compiles and runs via its existing checked compilation path.

## Existing React parking

The area contains a standalone static_single_assignment port, but no native high-level function-lowering/capture service corresponding to the calls below. The dependencies remain `high_level_intermediate_representation.ForFunctionWithoutManualMemoization` in cohere/internal/lint/rules/react/set_state_in_effect.go:267, `high_level_intermediate_representation.ForFunction` in set_state_in_render.go:183 and static_components.go:120. A checker fact does not implement this analysis. Parked claims and all existing probes are retained.

## Verification limitations

The shared `TestCheckerBridgeRefusalPending` explicitly skips until `codex/tsgo-errors-as-values` supplies the C error-buffer result for tsgoInspect (stage1/cohere/lint/checker_pending_test.go:49). Every input-controlled lint check is enabled; this dependency skip is preserved and named rather than bypassed. Setup initially failed against stale submodules/bridge APIs. After their adaptation it passes in 43.396 seconds, nproc 5, quota four CPUs. Bridge tests and vet pass. Logs preserve the initial build failure, the aborted pre-array-fix attempt, and the clean full retry separately.

## Corpus finding outside these rules

TestCompilerAndStage1Agree fails on tracked validation-wave-24/control-002.a: Go reports @typescript-eslint/no-this-alias / thisAssignment at the self binding, while Node omits it. The area's Rule.visit in stage1/cohere/lint/rules/typescript-no-this-alias/rule.a:14 filters out .a. Cohere's NoThisAlias invokes isTypeScriptSourceFile at cohere/internal/lint/rules/typescript/no_this_alias.go:139; isTypeScriptSourceFile at no_explicit_any.go:160 uses sourcename.TreatedAs, accepting .a, and reports at no_this_alias.go:154. No other owner's rule or comparator is changed.

The isolated evidence/unpark/no-this-alias-reproducer.a reproduces this with const self = this; export {};. Same selected manifest, live Go oracle 619 bytes with thisAssignment, Node 54 bytes without it; neither command fails. Both outputs are retained. Existing tracked corpus inputs are not removed. Later corpus/profile checks may report the same extension mismatch; every failure remains in the full event log.

## Owned timing and an additional input limitation

All 183 typed upstream cases in TestRulesAgree total live Go program/lint 17.874644364s and sanitized native program/lint/recording 21.797291528s (including the area's pilot; not per-rule isolated totals). Three separate successful witness samples per migrated rule compare complete bytes: symbol-description median Go 61.138ms, native 85.724ms, 926 bytes; valid-typeof median Go 60.933ms, native 83.206ms, 1536 bytes. These are process totals with sanitizers and recording, sampled while the package was running, not a release throughput claim. Detailed samples and outputs are in evidence/unpark/witness-timing.json and witness-*.

An earlier probe named the exact same Symbol witness witness.a in the strict tsconfig. The live Go oracle rejected it before lint: program.Build (cohere/internal/types/program/program.go:242) reaches the root match-count guard at program.go:684, naming one config file and zero matched program paths. The successful measurements use the harness's TypeScript witness input spelling. This does not establish typed .a input support; the native rule implementations themselves remain .a. The failed output is retained as a-root-program-failure.log. A root-file repro config is beside root-program-reproducer.a in evidence/unpark; run it through a manifest beginning with program <absolute config path> and a symbol-description row. No guard is weakened.

## Final complete lint package result

Command: go test -json -count=1 -timeout 60m ./stage1/cohere/lint, with cloud env sourced, GOMAXPROCS=4, ADAMIC_LINT_BENCH=1, ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-24-corpus (clean 050880ce59e30b356b686bd3144efe24f875ebc8), both profile inputs set to /workspace/unpark-wave-profile. Exit 1. Package 1848.217s; outer wall 1850.289973446s. Top level: 33 pass, 1 fail, 1 skip; including subtests: 114 pass, 1 fail, 1 skip. All 77 registry mutants caught, including no-caller-property, symbol-description-global and valid-typeof-options on Node, emitted JavaScript and sanitized native.

TestRulesAgree passes: 13,778,745 complete bytes and all 183 typed cases. TestOwnedWitnesses, all profile tests, shards, throughput, registration, factory hooks, source-content cache controls and option controls pass. Setup, lint-registry, gofmt, vet and bridge/tsgo tests pass after the pin API adaptation. The filtered comparison retry passed TestRulesAgree; its redundant mutant process was stopped once the clean full run took over. The full run above is completed, not stopped. Every original wave assertion remains, but its standalone drivers fail before comparison as documented.

Failure: TestCompilerAndStage1Agree on the no-this-alias .a extension guard. Skip: TestCheckerBridgeRefusalPending, awaiting tsgoInspect's TSGoError C-buffer result. No input-controlled test skips, relaxed guard, deleted input or removed test. These results do not make the whole package green. nproc=5, CPU quota=4; sampled one-minute load minimum/median/maximum=1.000/1.984/7.451. Complete events, stderr, runner, environment, counts and load samples are in evidence/unpark/lint*.

This push lands the merge and migration work for integration, with the ten legacy rule ports blocked on the fresh CLI refusal and the three React claims still parked. codex/lint-port-no-caller needs no new merge: its pushed tip is already an ancestor of area c4bdc23f; its rule, upstream cases, witness and mutant pass again in this completed run. No main or area branch is pushed and no new rule is claimed.
