# rules/react shared helpers

34 of the triage's 35 required symbols are certified on cohere `7945d102a6c18dd36adf9114a758ce646e8b2359`.
`isEs5ComponentCallStrict` is stopped on the separately owned `ecmascript/react.IsEs5ComponentCall` package dependency; see [STOPPED.md](STOPPED.md).

The AST helpers use the existing numeric `RuleContext` arena. Capture overlays keep Go bodies unchanged, rename them behind recording wrappers, and save every actual invocation across the complete upstream React test package. Stable arena indexes preserve returned node identity; strings are compared through a lossless UTF-16 observation encoding. Each replay compares source Node, emitted JavaScript, and ASan/UBSan native with the Go result. This certifies the captured contracts, not arbitrary trees or every consuming rule's findings. The proof rule uses the real shared parser and passes its complete upstream corpus.

The isolated AST snapshots come from Go's `ForEachChild`, with source byte positions converted to UTF-16. The consuming rule gate separately proves the parser adapter on its own upstream cases. `parametersOf` returns a fresh read-only view; consuming upstream calls do not mutate it. `attributesOf` retains the parser's child-array identity. Arbitrary invalid UTF-8 strings and decoder nesting beyond `OptionsJson`'s 512-level refusal are outside the captured input domain. No compiler gap or missing fixture is counted as a successful mutant.

Partial-port provenance: `origin/codex/lint-helpers-05` supplies the component-base-name comparison and Unicode uppercase table (adapted for rules/react's `use` and digit acceptance); `origin/codex/lint-helpers-04` supplies the compiler-options acceptance/error structure, replacing its external decoder/sorter callbacks with the landed `OptionsJson` reader and explicit rune-order sorting; `origin/codex/lint-helpers-from-codex/lint-wave1-11` supplies the optional-parentheses contract, adapted to the shared numeric context. The strict ES5 wrapper from slot05 is not delivered because its shelf dependency is not ported on this base.

Actual capture calls: 29,787, plus 234 explicit leaf controls, for 30,021 comparisons per backend. The capture coverage guard checks all 34 helpers and hashes the Go helper/test sources. Every helper has one compiling semantic mutant, checked by result mismatch on Node, emitted JavaScript, and native. The proof rule is `react-hooks/error-boundaries`, with 69 unique upstream source/file/options combinations and one compiling mutant.

Conditional helper readiness alone now includes:

- `react-hooks/error-boundaries`
- `react-hooks/void-use-memo`
- `react/default-props-match-prop-types`
- `react/no-access-state-in-setstate`
- `react/no-set-state`
- `react/no-unused-state`

Rules with other helper-package blockers remain conditional on those landings; see the claim file's complete consumer/blocker list. No cumulative triage forecast is presented as a current integrated finding-parity result.

| Go symbol in rules/react | File | Actual captured calls |
|---|---|---:|
| `attributesOf` | [attributes_of.a](attributes_of.a) | 390 |
| `commentValueOf` | [comment_value_of.a](comment_value_of.a) | 17 |
| `DecodeCompilerRuleOptions` | [decode_compiler_rule_options.a](decode_compiler_rule_options.a) | 120 |
| `DecodeNoMethodSetStateOptions` | [decode_no_method_set_state_options.a](decode_no_method_set_state_options.a) | 140 |
| `enclosingClassOf` | [enclosing_class_of.a](enclosing_class_of.a) | 84 |
| `enclosingComponentOf` | [enclosing_component_of.a](enclosing_component_of.a) | 215 |
| `enclosingFunctionOf` | [enclosing_function_of.a](enclosing_function_of.a) | 233 |
| `functionBodyBlock` | [function_body_block.a](function_body_block.a) | 1019 |
| `hasValidComponentParameters` | [has_valid_component_parameters.a](has_valid_component_parameters.a) | 106 |
| `isComponentClass` | [is_component_class.a](is_component_class.a) | 3634 |
| `isComponentIdentifierName` | [is_component_identifier_name.a](is_component_identifier_name.a) | 388 |
| `isCreateReactClassCall` | [is_create_react_class_call.a](is_create_react_class_call.a) | 2502 |
| `isFunctionLike` | [is_function_like.a](is_function_like.a) | 4176 |
| `isHookIdentifierName` | [is_hook_identifier_name.a](is_hook_identifier_name.a) | 730 |
| `isJavaScriptIdentifier` | [is_java_script_identifier.a](is_java_script_identifier.a) | 10 |
| `isNonNodeExpression` | [is_non_node_expression.a](is_non_node_expression.a) | 126 |
| `isPureComponentBase` | [is_pure_component_base.a](is_pure_component_base.a) | 55 |
| `isReactCompiledFunction` | [is_react_compiled_function.a](is_react_compiled_function.a) | 130 |
| `isReactComponentBase` | [is_react_component_base.a](is_react_component_base.a) | 627 |
| `isReactComponentBaseName` | [is_react_component_base_name.a](is_react_component_base_name.a) | 624 |
| `isRestParameter` | [is_rest_parameter.a](is_rest_parameter.a) | 22 |
| `isThisExpression` | [is_this_expression.a](is_this_expression.a) | 163 |
| `isTopLevelCompilationCandidate` | [is_top_level_compilation_candidate.a](is_top_level_compilation_candidate.a) | 130 |
| `jsxAnnotationIn` | [jsx_annotation_in.a](jsx_annotation_in.a) | 17 |
| `mentionsRef` | [mentions_ref.a](mentions_ref.a) | 8 |
| `parametersOf` | [parameters_of.a](parameters_of.a) | 106 |
| `reactFunctionNameOf` | [react_function_name_of.a](react_function_name_of.a) | 179 |
| `reactPragmaFor` | [react_pragma_for.a](react_pragma_for.a) | 287 |
| `returnsNonNode` | [returns_non_node.a](returns_non_node.a) | 99 |
| `semanticParentOf` | [semantic_parent_of.a](semantic_parent_of.a) | 8311 |
| `skipParenthesesOptional` | [skip_parentheses_optional.a](skip_parentheses_optional.a) | 4836 |
| `sortDefaultPropsInitializerOf` | [sort_default_props_initializer_of.a](sort_default_props_initializer_of.a) | 31 |
| `sourceSliceOf` | [source_slice_of.a](source_slice_of.a) | 194 |
| `stylePropObjectUnwrapParentheses` | [style_prop_object_unwrap_parentheses.a](style_prop_object_unwrap_parentheses.a) | 78 |

From the repository root, after sourcing the setup environment:

```sh
python3 stage1/cohere/lint/helpers/rules-react/testdata/capture.py
python3 stage1/cohere/lint/helpers/rules-react/testdata/capture_ast.py
go test ./stage1/cohere/lint/helpers/rules-react -count=1 -v -timeout=40m
go test ./stage1/cohere/lint -run 'TestRulesAgree|TestOwnedWitnesses|TestMutants/try_block_omitted' -count=1 -v -timeout=30m
ADAMIC_TYPESCRIPT_SOURCE=/path/to/TypeScript-at-050880ce59e30b356b686bd3144efe24f875ebc8 go test ./stage1/cohere/lint -run '^TestCompilerAndStage1Agree$' -count=1 -v -timeout=30m
```

The 910-file compiler/stage1 corpus also matched Go on all three backends; see [VALIDATION.md](VALIDATION.md) for the native command-limit recovery and the complete local check scope.

Capture logs are scratch output, not delivered evidence. `testdata/coverage.json` owns the frozen source hashes and per-helper counts; regenerate it intentionally when changing the pinned corpus.
