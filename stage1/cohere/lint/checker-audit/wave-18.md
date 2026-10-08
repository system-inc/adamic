# Wave 18 checker and analysis audit

Audit only; no port changes. Worker: codex/typeaware-wave-18 at 8e77e2e2e0dcf5939e4df2ca6c5b38a0649ab4dc. Area inspected: c8f6d74f5387df926ae12d8babd2780564b60097. Paths and line numbers below refer to the worker's pinned cohere checkout. This audit branch starts at the area and adds only this file.

## Missing native analysis implementations

These are analysis dependencies, not requests to return Go's lint verdict through the checker. Existing prepared-input reporters do not implement native source analysis.

| Exact symbol as cohere calls it | Rules blocked | Definition / call site |
| --- | --- | --- |
| high_level_intermediate_representation.ForFunction | react-hooks/static-components; react-hooks/set-state-in-render | cohere/internal/lint/ecmascript/high_level_intermediate_representation/cache.go:63; cohere/internal/lint/rules/react/static_components.go:120; set_state_in_render.go:183 |
| high_level_intermediate_representation.ForFunctionWithoutManualMemoization | react-hooks/set-state-in-effect | cohere/internal/lint/ecmascript/high_level_intermediate_representation/cache.go:178; cohere/internal/lint/rules/react/set_state_in_effect.go:267 |
| high_level_intermediate_representation.MayHoldComponentOrHook | react-hooks/static-components; react-hooks/set-state-in-render; react-hooks/set-state-in-effect | cohere/internal/lint/ecmascript/high_level_intermediate_representation/spelling.go:88; calls static_components.go:113, set_state_in_render.go:176, set_state_in_effect.go:259 |
| walk.functionEvaluation (receiver *jsxNoConstructedContextValuesStabilityWalk) | react/jsx-no-constructed-context-values | cohere/internal/lint/rules/react/jsx_no_constructed_context_values_stability.go:532; calls :351, :513 |
| walk.anyEscapes (receiver *jsxNoConstructedContextValuesStabilityWalk) | react/jsx-no-constructed-context-values | cohere/internal/lint/rules/react/jsx_no_constructed_context_values_stability.go:603; calls :369, :575 |
| jsxNoConstructedContextValuesStaysHome | react/jsx-no-constructed-context-values | cohere/internal/lint/rules/react/jsx_no_constructed_context_values_stability.go:649; call :629 |
| jsxNoConstructedContextValuesCheckMemo | react/jsx-no-constructed-context-values | cohere/internal/lint/rules/react/jsx_no_constructed_context_values_stability.go:786; call cohere/internal/lint/rules/react/jsx_no_constructed_context_values.go:236 |

ForFunction needs native lowering, SSA and capture analysis. ForFunctionWithoutManualMemoization additionally needs the memo-erasure path. Context-value stability needs callback-return and holder escape evaluation. MayHoldComponentOrHook is the shared source eligibility helper, not a type-checker method.

## Checker facts absent from the area

The worker already implements the following questions in separate bridge files. They are pending landing/migration, not wholly unimplemented work. The area's bridge/tsgo/checker/facts.go does not register these questions. Similar granular questions already present do not expose all the required payloads; especially symbol identity, shorthand binding, declaration containers and inherited modifier flags must not be replaced by spelling comparisons.

| Exact upstream checker call or analysis helper | Rules needing the missing payload | Existing worker question / implementation |
| --- | --- | --- |
| ctx.TypeChecker.IsArrayLikeType; ctx.TypeChecker.GetNumberIndexType; ctx.TypeChecker.GetTypeArguments; type_checking.GetWellKnownSymbolPropertyOfType; type_checking.NeedsToBeAwaited | @typescript-eslint/await-thenable | iteration-type-facts; bridge/tsgo/checker/iteration_type_facts.go. Calls in cohere/internal/lint/rules/typescript/await_thenable.go:194-195, :230-231, :306, :317, :331, :349-358; helper implementations cohere/internal/lint/checking/types.go. Existing type-shape/raw-shape/property-shape/call-count questions remain usable. |
| ctx.TypeChecker.GetDeclaredTypeOfSymbol; ctx.TypeChecker.GetBaseTypes; ctx.TypeChecker.GetPropertyOfType | @typescript-eslint/class-literal-property-style | base-member-facts; bridge/tsgo/checker/base_member_facts.go. Calls cohere/internal/lint/rules/typescript/class_literal_property_style.go:262, :266-267. GetBaseTypes itself is already represented by base-shapes; missing member payload includes inherited property/declaration/modifier facts. |
| ctx.TypeChecker.GetSymbolAtLocation; ctx.TypeChecker.GetShorthandAssignmentValueSymbol; rule.DeclarationsIn | no-class-assign; no-const-assign; no-constant-binary-expression; no-new-native-nonconstructor; prefer-promise-reject-errors; prefer-rest-params | node-symbol-details and binding-declarations; bridge/tsgo/checker/declaration_facts.go. Representative calls cohere/internal/lint/rules/core/no_class_assign.go:171, :189-191; no_constant_binary_expression.go:87; no_new_native_nonconstructor.go:103; prefer_promise_reject_errors.go:182, :215, :278, :307; prefer_rest_params.go:114. no-const-assign reads declarations as described at no_const_assign.go:134. GetSymbolAtLocation already underlies symbol-origin; the richer identity/binding payload is absent. |
| ctx.TypeChecker.GetSymbolAtLocation | react/jsx-no-undef | jsx-syntax-facts; bridge/tsgo/checker/jsx_syntax_facts.go. Call cohere/internal/lint/rules/react/jsx_no_undef.go:101. Same raw AST payload is used by react/jsx-fragments and react/no-adjacent-inline-elements, whose missing payload is syntax, not a checker method. |
| ctx.TypeChecker.GetSymbolAtLocation; ctx.TypeChecker.GetShorthandAssignmentValueSymbol | react/static-property-placement; react/style-prop-object | component-property-syntax-facts; bridge/tsgo/checker/component_property_syntax_facts.go. Calls cohere/internal/lint/rules/react/static_property_placement.go:406; style_prop_object.go:205-207. |
| regexpattern.Walk; regexsyntax.ParseRegexFlags; regexsyntax.SkipPatternEscape; regexsyntax.ClassEnd | prefer-regex-literals | regex-pattern-facts; bridge/tsgo/checker/regex_pattern_facts.go. Calls cohere/internal/lint/rules/core/prefer_regex_literals.go:566, :835, :871, :884, :895. These parse ECMAScript regex grammar; they are not Go checker calls or replacements for Go regex predicates. |

Do not infer that each listed checker method is entirely absent from the area. This table identifies the additional fact contracts consumed by these ports. It separates existing worker implementations from the missing native analysis above. The area already exposes GetTypeAtLocation-derived type shapes and ordinary signature/property/base queries.

## Legacy bridge APIs and landing moves

No evidence that the area's low-level bridge removed tsgoProgram, tsgoInspect or tsgoRelease: stage1/cohere/lint/checker_bridge.a still imports them. What changed is the rule-facing API and ownership.

| Old API / caller | Area-facing move |
| --- | --- |
| tsgoProgram(config, paths), in wave_18_suite.a, wave_18_core_suite.a, wave_18_constructor_suite.a, wave_18_preference_suite.a, wave_18_title_suite.a, wave_18_jsx/main.a and wave_18_component_props/main.a | Harness owns one program per run. Rules use context.checker; no rule-owned handle. |
| tsgoInspect(program, path, byteStart, byteEnd, kind, question), via stage1/cohere/typeaware/rules.ts:104 | context.checker.ask(index, question), with the provided parser node and area's selector/offset mapping. Result/refusal handling follows Answer. |
| SourceFile tsgoInspect in wave_18_jsx/jsx_syntax_facts.a:55 and wave_18_component_props/component_property_syntax_facts.a:62 | context.checker.askFile(new FileQuestion('ReadsOtherFiles', question)); declare programReads in rule.json. Five React rule directories already migrated on worker; runtime parity not yet revalidated. |
| tsgoRelease(program), same standalone suite entrypoints | Harness releases the run handle; rules do not release it. |
| tsgoTypeParts, in shared stage1/cohere/typeaware/unary_minus.ts:44 and released/query-cost fixtures | No Checker.typeParts facade. Low-level Program.TypeParts remains. These are inherited helper/fixture usages, not a direct Wave 18 rule requirement; do not add a query solely because this import is transitive. |

No direct tsgoQuery call was found in Wave 18's rule/suite sources. Query/typeParts are absent from the facade, not proven removed from the low-level library. Existing custom question strings requiring landing are listed above; private transcript recording must become the area's recording/replay rather than a second checker.

## Other blockers and validation limits

Previous all-input lint run on worker failed before tests: SourceFile.FileName() became tspath.RootedFilePath (cohere/TypeScript/tsc/internal/ast/ast.go:2703); SourceFile.Path() became PathKey() (:2707); Program.IsSourceFileDefaultLibrary takes tspath.PathKey (cohere/TypeScript/tsc/internal/compiler/program.go:1804). Remaining worker errors were bridge/tsgo/checker/declaration_facts.go:17, :22 and facts.go:292. This is a pin/API adaptation problem, not a missing checker question. Current area uses FileName().AsString(); do not treat the earlier worker failure as evidence that the newly fetched area itself fails.

Worker evidence: stage1/cohere/typeaware/wave_18_component_props/UNPARK_CHECKER_REPORT.md and validation-unpark-checker/summary.json on codex/typeaware-wave-18. Seventeen standalone ports are historical work, five React registry migrations are prepared, four analysis-dependent rules are unfinished; none is claimed current-pin green by this audit. class-literal-property-style's getters option also remains incomplete native verdict work, not a missing checker API. No runtime tests, mutants, rule migrations or new claims were performed in this audit-only turn.
