# Remaining wave 16 claims at d845dccde

These are stopped, not certified ports. No shared files were changed.

| Rule | Exact upstream call | Location under cohere/internal/lint/rules/ | Required question or helper |
|---|---|---|---|
| prefer-arrow-callback | Valid upstream callback with logical fallback and bind | `core/prefer_arrow_callback_test.go:73`, captured `PreferArrowCallback.ts` | The stage 1 parser refuses `evidence/callback-logical-bind.ts.txt` at byte 28, expecting CloseParenToken and finding SemicolonToken. Candidate source remains in `/workspace/wave16-artifacts/parked-prefer-arrow-callback/`, outside discovery and the landing commit. |
| no-throw-literal | `rule_testing.RunTyped` recovery case | `core/no_throw_literal_test.go:164` | The typed upstream path in `stage1/cohere/lint/lint_test.go:382` does not classify recovery. `testdata/oracle.go:186` panics before comparison. Reproducer: `evidence/no-throw-argument.ts.txt`. The candidate source is retained in `/workspace/wave16-artifacts/parked-no-throw-literal/`, outside discovery and the landing commit. |
| no-import-assign | `imports.BindingsOf(node)` | `core/no_import_assign.go:133` | Shared imports helper is absent; reference.IdentifiersNamed at line 168 is also absent. |
| prefer-exponentiation-operator | `text.TrimWhitespace(ctx.SourceFile.Text()[:start])` | `core/prefer_exponentiation_operator.go:551` | Shared ecmascript text helper is absent. |
| no-alert | `property.AccessedName(callee, property.Static)` | `core/no_alert.go:129` | Shared property helper is absent. |
| no-new-func | `property.AccessedName(callee, property.Static)` | `core/no_new_func.go:127` | Shared property helper is absent. |
| no-useless-backreference | `reference.NewTracker(ctx.SourceFile, ctx.TypeChecker, nil)` | `core/no_useless_backreference.go:100` | Shared native reference tracker and regex analysis are absent. |
| nexus/correctness-no-global-listener-target-assertion | `ctx.TypeChecker.GetNonNullableType(known)` | `nexus/correctness_no_global_listener_target_assertion.go:257` | Non-nullable type projection plus exact assignability between those projections is unanswered. |
| nexus/correctness-no-leaked-number-render | `part.AsLiteralType().Value()` | `nexus/correctness_no_leaked_number_render.go:220` | Exact numeric literal payload is unanswered. |
| nexus/correctness-no-mock-on-module-namespace | `ctx.TypeChecker.GetResolvedSignature(call)` | `nexus/correctness_no_mock_on_module_namespace.go:194` | Resolved signature declaration and complete enclosing module chain are unanswered. |
| nexus/security-no-interpolated-shell-command | `ctx.TypeChecker.GetResolvedSignature(call)` | `nexus/security_no_interpolated_shell_command.go:185` | Resolved signature declaration and enclosing module chain are unanswered; literal payloads are also missing. |
| nexus/security-no-interpolated-sql-string | `part.AsLiteralType().Value().(string)` | `nexus/security_no_interpolated_sql_string.go:591` | Exact string literal payload is unanswered. |
| react/jsx-fragments | `jsx.ElementParts(opening)` | `react/jsx_fragments.go:207` | Shared JSX helper is absent. |
| react/jsx-no-undef | `jsx.ElementParts(node)` | `react/jsx_no_undef.go:93` | Shared JSX helper is absent. |
| react/jsx-no-constructed-context-values | `utilsreact.IsLikelyComponentName(name.Text())` | `react/jsx_no_constructed_context_values.go:645` | Shared React helper is absent; resolved signatures at jsx_no_constructed_context_values_stability.go:488 remain unanswered. |
| react-hooks/set-state-in-effect | `high_level_intermediate_representation.ForFunctionWithoutManualMemoization(ctx, functionNode)` | `react/set_state_in_effect.go:267` | React Compiler IR is absent. |
| react-hooks/set-state-in-render | `high_level_intermediate_representation.ForFunction(ctx, functionNode)` | `react/set_state_in_render.go:183` | React Compiler IR is absent. |
| react-hooks/static-components | `high_level_intermediate_representation.ForFunction(ctx, functionNode)` | `react/static_components.go:120` | React Compiler IR is absent. |

`no-new-wrappers` and `no-new-native-nonconstructor` already exist in the facts base. Their descriptors are owned by typeaware-08 and were not edited here.

## Program reads

The upstream declarations of `use-isnan`, `no-throw-literal` and `prefer-arrow-callback` omit `ProgramReads`. The landing descriptor for `use-isnan` and the two parked candidate descriptors omit `programReads` as well. They ask node-level symbol questions and do not call `askFile`, reopen foreign source files or inspect default-library flags. Symbol presence, identity and declaration-file provenance answer their upstream checker calls.
