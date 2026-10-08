# Shared checker landing blockers

Merged area c8f6d74f5 without rebasing. No rule verdict or private recording was moved into the shared checker. The new registered local rules obtain answers through RuleContext.checker; their replay is the area's native recording, bound to program and source hashes.

Three existing ports remain outside the unified registry because the shared checker has no selector for a node belonging to another file. Checker.ask(index, question) always selects this.path and this.parser; Checker.askFile(FileQuestion) enforces declared ProgramReads but still selects the current root. The raw bridge questions already exist. The missing helper is a shared, descriptor-checked foreign-file node selector and its transcript/replay path, rather than another checker program. The old standalone drivers remain historical controls, not a claim of unified certification.

| Rule | Exact upstream dependency | Existing owned reproducer |
| --- | --- | --- |
| nexus/correctness-no-process-exit-after-output | correctnessNoProcessExitAfterOutputWriters.followedCallee, cohere/internal/lint/rules/nexus/correctness_no_process_exit_after_output.go:313; GetResolvedSignature at 314 and foreign declaration file at 330 | wave28_next/no_process_exit_after_output.a:86 follows an imported writer body through World.get(signature.path) |
| nexus/correctness-require-blocking-standard-streams | correctnessRequireBlockingStandardStreamsBuildIndex, cohere/internal/lint/rules/nexus/correctness_require_blocking_standard_streams.go:251; program.SourceFiles at 259 and program.GetSourceFileForResolvedModule at 283 | wave28_next/require_blocking_standard_streams.a:76 follows an imported initialization body through World.get(next) |
| react/jsx-no-constructed-context-values | jsxNoConstructedContextValuesStabilityWalk.callEvaluation, cohere/internal/lint/rules/react/jsx_no_constructed_context_values_stability.go:488; signature.Declaration at 492 and functionEvaluation(callee, depth+1) at 513 | wave28_fifth/jsx-no-constructed-context-values/ast.a:39 locates a foreign declaration with World.get(path); the fifth-batch cross-file callee control is retained |

The three previously parked React Hooks rules still need native HIR/SSA analysis:

| Rule | Exact upstream dependency |
| --- | --- |
| react-hooks/set-state-in-effect | high_level_intermediate_representation.ForFunctionWithoutManualMemoization, cohere/internal/lint/rules/react/set_state_in_effect.go:267; ControlDominators at 477 |
| react-hooks/set-state-in-render | high_level_intermediate_representation.ForFunction, cohere/internal/lint/rules/react/set_state_in_render.go:183; UnconditionalBlocks at 260 |
| react-hooks/static-components | high_level_intermediate_representation.ForFunction, cohere/internal/lint/rules/react/static_components.go:120 |

These six are not registered as clean skeletons, and their earlier parked/probe checks were kept. No new rules were claimed. Shared harness, checker and generator sources were not changed by this migration.

The foreign-selector probe is `gaps/foreign-selector-probe.a`. `adamic types` refuses it with TS2339 on `Checker.askOtherFile`; the probe intentionally lives under gaps and is not registered as a successful rule.

The area also retains `TestCheckerBridgeRefusalPending`, which skips until the prelude exposes `TSGoError` and `tsgoInspect` returns the C error buffer as that value. This is an existing shared bridge dependency, not a missing test input. The check is kept unchanged.
