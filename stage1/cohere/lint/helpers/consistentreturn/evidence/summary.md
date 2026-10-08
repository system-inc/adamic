# Local completed gates

- Owned helper capture/Node/emitted-JS/ASan+UBSan/mutants: PASS 55.412s.
- Complete helpers package, every default input: PASS 728.781s.
- Separately packaged comments, consumers/mutants/explicit JSX gaps: PASS 306.216s.
- Complete lint package with ADAMIC_TYPESCRIPT_SOURCE pinned v6.0.3: PASS 1616.427s.
- All 83 registered rule mutants caught on Node, emitted JS and native.
- Go vet, gofmt and committed fixture metadata: clean.
- Nine new helper mutants caught on all three backends, successful compilation and normal exit.

No full consuming-rule port is claimed: canRunOffEnd, judgeReturns, judgeScope
and Judge are stopped on the unported/reserved control_flow_graph.Build.
Optional timing/artifact exporters remain opt-in. The known checker error-value
contract remains a pre-existing explicit pending test. All required input
comparisons ran; no JSX count map, registration or compiler/runtime file changed.

Logs include an intentionally failing child comparison used by
TestEmittedJavaScriptMismatch; that parent test passes. The package exit was 0.

## Selected full lint evidence

```
    lint_test.go:283: cohere cases: 4576 unique source/rule/options combinations
    lint_test.go:283: cohere cases: 4576 unique source/rule/options combinations
    lint_test.go:459: compiler and stage1: 926 files
    lint_test.go:283: cohere cases: 4576 unique source/rule/options combinations
    lint_test.go:283: cohere cases: 4576 unique source/rule/options combinations
--- PASS: TestJsxLintTrees (87.55s)
--- PASS: TestRulesAgree (403.24s)
    lint_test.go:283: cohere cases: 4576 unique source/rule/options combinations
--- PASS: TestShardsAgree (230.35s)
--- PASS: TestCompilerAndStage1Agree (663.41s)
--- PASS: TestMutants (0.04s)
    --- PASS: TestMutants/debugger_fix_suppressed (28.65s)
    --- PASS: TestMutants/nexus-stuttering-name-equality (143.56s)
    --- PASS: TestMutants/directive_prefix_ignored (62.21s)
    --- PASS: TestMutants/prefer-destructuring-object-fix (21.70s)
    --- PASS: TestMutants/bom_removes_two_marks (155.02s)
    --- PASS: TestMutants/comment_self_directive_exemption_removed (50.89s)
    --- PASS: TestMutants/no-useless-rename-import-side (44.27s)
    --- PASS: TestMutants/nested_generator_owns_outer_yield (155.41s)
    --- PASS: TestMutants/empty-string-is-whitespace (159.13s)
    --- PASS: TestMutants/no-useless-catch_finally_repair_range_reversed (23.49s)
    --- PASS: TestMutants/continue_target_ignored (22.38s)
    --- PASS: TestMutants/and-left-undefined-ignored (31.68s)
    --- PASS: TestMutants/finally_continue_incorrectly_absorbed_by_switch (29.32s)
    --- PASS: TestMutants/with_statement_underlined_whole (137.43s)
    --- PASS: TestMutants/boolean_inverse_changed (36.25s)
    --- PASS: TestMutants/label_option_widened (26.32s)
    --- PASS: TestMutants/no-iterator-static-key (32.75s)
    --- PASS: TestMutants/miss-postfix-writes (22.97s)
    --- PASS: TestMutants/no-empty-pattern_reports_real_bindings (21.00s)
    --- PASS: TestMutants/concat_substituted_template_ignored (146.88s)
    --- PASS: TestMutants/empty_character_class_silenced (46.49s)
    --- PASS: TestMutants/regex_fix_eats_extra_byte (37.00s)
    --- PASS: TestMutants/ordering-relations-ignored (136.89s)
    --- PASS: TestMutants/no-multi-assign-option-reversed (26.31s)
    --- PASS: TestMutants/empty_placeholder_reported (31.59s)
    --- PASS: TestMutants/comment_exemption_removed (146.08s)
    --- PASS: TestMutants/comma_chain_reports_inner (52.64s)
    --- PASS: TestMutants/self-comparison-equality-reversed (30.65s)
    --- PASS: TestMutants/continue_finding_stops_before_its_semicolon (139.26s)
    --- PASS: TestMutants/constructor-property-name-ignored (119.58s)
    --- PASS: TestMutants/no-script-url-case-fold (35.34s)
    --- PASS: TestMutants/return_parentheses_option_reversed (31.70s)
    --- PASS: TestMutants/for_update_option_ignored (26.47s)
    --- PASS: TestMutants/destructuring_hole_reported (142.75s)
    --- PASS: TestMutants/condition_polarity_reversed (24.81s)
    --- PASS: TestMutants/proto_accessor_suppressed (126.80s)
    --- PASS: TestMutants/omit-is-prototype-of (130.95s)
    --- PASS: TestMutants/no-caller-property (97.67s)
    --- PASS: TestMutants/interface_accepts_type_suffix (28.64s)
    --- PASS: TestMutants/int32_hint_option_ignored (32.36s)
    --- PASS: TestMutants/boolean_type_verdict_inverted (32.29s)
    --- PASS: TestMutants/parentheses_around_a_discarded_construction_not_seen_through (155.74s)
    --- PASS: TestMutants/one_excess_class_is_silently_allowed (29.00s)
    --- PASS: TestMutants/JSDoc_replacement_text_corrupted (21.50s)
    --- PASS: TestMutants/pagination_bad_name_accepted (35.58s)
    --- PASS: TestMutants/shouting_fourth_token_omitted (19.69s)
    --- PASS: TestMutants/declared_BaseError_constructor_ignored (35.08s)
    --- PASS: TestMutants/console_receiver_widened (34.11s)
    --- PASS: TestMutants/nexus_long_line_comment_configured_threshold_ignored (28.68s)
    --- PASS: TestMutants/await_crosses_function_boundary (205.60s)
    --- PASS: TestMutants/global_container_name_judgement_reversed (33.97s)
    --- PASS: TestMutants/max-depth-keyword-span (122.58s)
    --- PASS: TestMutants/nexus-for-in-message (28.71s)
    --- PASS: TestMutants/enum_reported_at_its_declaration_rather_than_its_name (21.92s)
    --- PASS: TestMutants/enum_bitwise_option_ignored (28.08s)
    --- PASS: TestMutants/ambiguous-identifier-span (31.87s)
    --- PASS: TestMutants/prefer-find_array_verdict_suppressed (24.61s)
    --- PASS: TestMutants/nexus-abbreviated-identifier_whole-word_finding_suppressed (29.52s)
    --- PASS: TestMutants/multiline-arrow-single-line-exemption-removed (127.29s)
    --- PASS: TestMutants/wrapper_Number_ignored (25.58s)
    --- PASS: TestMutants/callback_at_the_limit_is_reported (29.87s)
    --- PASS: TestMutants/module_empty_export_fix_suppressed (21.34s)
    --- PASS: TestMutants/no-unexpected-multiline_mutant_from_batch_8 (42.88s)
    --- PASS: TestMutants/react-jsx-no-comment-textnodes_mutant_from_batch_8 (56.32s)
    --- PASS: TestMutants/react-forward-ref-uses-ref_mutant_from_batch_8 (37.41s)
    --- PASS: TestMutants/number-inference (25.34s)
    --- PASS: TestMutants/prefer-template_mutant_from_batch_8 (21.97s)
    --- PASS: TestMutants/predicate_finding_widened_to_its_owner (154.24s)
    --- PASS: TestMutants/inline_type_gate (35.84s)
    --- PASS: TestMutants/no-useless-constructor_mutant_from_batch_8 (21.43s)
    --- PASS: TestMutants/method_listener_lost (21.19s)
    --- PASS: TestMutants/no-unused-private-class-members_mutant_from_batch_8 (31.11s)
    --- PASS: TestMutants/init_declarations_always_span_includes_annotation (20.46s)
    --- PASS: TestMutants/react-no-redundant-should-component-update_mutant_from_batch_8 (19.06s)
    --- PASS: TestMutants/react-no-is-mounted_mutant_from_batch_8 (28.80s)
    --- PASS: TestMutants/adjacent_overload_separation_ignored (21.57s)
    --- PASS: TestMutants/react-no-find-dom-node_mutant_from_batch_8 (19.05s)
    --- PASS: TestMutants/no-octal-escape_mutant_from_batch_8 (30.48s)
    --- PASS: TestMutants/var_declaration_suppressed (21.17s)
    --- PASS: TestMutants/duplicate_case_suppressed (19.01s)
    --- PASS: TestMutants/parenthesized_alias_lost (140.28s)
    --- PASS: TestMutants/empty_function_body_reported (18.94s)
    --- PASS: TestMutants/suggestion_applied_as_fix (92.78s)
ok  	github.com/system-inc/adamic/stage1/cohere/lint	1616.427s
```
