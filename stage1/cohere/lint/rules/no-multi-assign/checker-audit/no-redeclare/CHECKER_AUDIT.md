# Wave 12 shared-checker audit

All eighteen claims were checked. The current checker integration is present; the narrower symbol queries and shared analysis helpers below are absent. No private implementation was added.

| Rule | Exact dependency | Note |
|---|---|---|
| @typescript-eslint/no-redeclare | `checker.Checker.GetGlobalSymbol` at `cohere/internal/lint/rules/typescript/builtin_globals.go:59` | [BLOCKED.md](../no-redeclare/BLOCKED.md) |
| nexus/correctness-no-test-on-global-regex | `checker.Checker.GetSymbolAtLocation` at `cohere/internal/lint/rules/nexus/correctness_no_test_on_global_regex.go:118` | [BLOCKED.md](../correctness-no-test-on-global-regex/BLOCKED.md) |
| nexus/correctness-no-write-only-collection | `checker.Checker.GetShorthandAssignmentValueSymbol` at `cohere/internal/lint/rules/nexus/correctness_no_write_only_collection.go:124` | [BLOCKED.md](../correctness-no-write-only-collection/BLOCKED.md) |
| nexus/correctness-no-process-exit-after-output | `checker.Checker.GetAliasedSymbol` at `cohere/internal/lint/rules/nexus/correctness_no_process_exit_after_output.go:413` | [BLOCKED.md](../correctness-no-process-exit-after-output/BLOCKED.md) |
| nexus/correctness-no-uncleared-race-timeout | `checker.Checker.GetShorthandAssignmentValueSymbol` at `cohere/internal/lint/rules/nexus/correctness_no_uncleared_race_timeout.go:337` | [BLOCKED.md](../correctness-no-uncleared-race-timeout/BLOCKED.md) |
| nexus/correctness-require-blocking-standard-streams | `checker.Checker.GetAliasedSymbol` at `cohere/internal/lint/rules/nexus/correctness_require_blocking_standard_streams.go:586` | [BLOCKED.md](../correctness-require-blocking-standard-streams/BLOCKED.md) |
| no-new-func | `checker.Checker.GetSymbolAtLocation via resolvesToAGlobal (core/no_new_func.go:118)` at `cohere/internal/lint/rules/core/no_new_native_nonconstructor.go:103` | [BLOCKED.md](../no-new-func/BLOCKED.md) |
| no-new-native-nonconstructor | `checker.Checker.GetSymbolAtLocation via resolvesToAGlobal` at `cohere/internal/lint/rules/core/no_new_native_nonconstructor.go:103` | [BLOCKED.md](../no-new-native-nonconstructor/BLOCKED.md) |
| no-new-wrappers | `checker.Checker.GetSymbolAtLocation via resolvesToAGlobal (core/no_new_wrappers.go:118)` at `cohere/internal/lint/rules/core/no_new_native_nonconstructor.go:103` | [BLOCKED.md](../no-new-wrappers/BLOCKED.md) |
| prefer-regex-literals | `reference.NewTracker` at `cohere/internal/lint/rules/core/prefer_regex_literals.go:196` | [BLOCKED.md](../prefer-regex-literals/BLOCKED.md) |
| prefer-rest-params | `checker.Checker.GetSymbolAtLocation` at `cohere/internal/lint/rules/core/prefer_rest_params.go:114` | [BLOCKED.md](../prefer-rest-params/BLOCKED.md) |
| react-hooks/exhaustive-deps | `checker.Checker.GetShorthandAssignmentValueSymbol` at `cohere/internal/lint/rules/react/exhaustive_deps.go:1888` | [BLOCKED.md](../exhaustive-deps/BLOCKED.md) |
| react-hooks/set-state-in-effect | `high_level_intermediate_representation.ForFunctionWithoutManualMemoization` at `cohere/internal/lint/rules/react/set_state_in_effect.go:267` | [BLOCKED.md](../set-state-in-effect/BLOCKED.md) |
| react-hooks/set-state-in-render | `high_level_intermediate_representation.ForFunction` at `cohere/internal/lint/rules/react/set_state_in_render.go:183` | [BLOCKED.md](../set-state-in-render/BLOCKED.md) |
| react-hooks/static-components | `high_level_intermediate_representation.ForFunction` at `cohere/internal/lint/rules/react/static_components.go:120` | [BLOCKED.md](../static-components/BLOCKED.md) |
| react/jsx-fragments | `checker.Checker.GetSymbolAtLocation` at `cohere/internal/lint/rules/react/jsx_fragments.go:290` | [BLOCKED.md](../jsx-fragments/BLOCKED.md) |
| react/jsx-no-constructed-context-values | `checker.Checker.GetSymbolAtLocation` at `cohere/internal/lint/rules/react/jsx_no_constructed_context_values.go:320` | [BLOCKED.md](../jsx-no-constructed-context-values/BLOCKED.md) |
| react/jsx-no-undef | `checker.Checker.GetSymbolAtLocation` at `cohere/internal/lint/rules/react/jsx_no_undef.go:101` | [BLOCKED.md](../jsx-no-undef/BLOCKED.md) |

Each directory contains its reproducer. No new rule is registered and no upstream case or mutant is claimed as passing.
