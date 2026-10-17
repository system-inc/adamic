package lint

import "testing"

// One independently discoverable mutant and product declaration per registered rule.

func TestMutants_adamic_no_type_predicate(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "adamic-no-type-predicate")
}

func TestProduct_LintMutant_adamic_no_type_predicate(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "adamic-no-type-predicate")
}

func TestMutants_base_boundary_no_global_container(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "base-boundary-no-global-container")
}

func TestProduct_LintMutant_base_boundary_no_global_container(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "base-boundary-no-global-container")
}

func TestMutants_base_consistency_no_console(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "base-consistency-no-console")
}

func TestProduct_LintMutant_base_consistency_no_console(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "base-consistency-no-console")
}

func TestMutants_base_consistency_no_hand_built_declared_error(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "base-consistency-no-hand-built-declared-error")
}

func TestProduct_LintMutant_base_consistency_no_hand_built_declared_error(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "base-consistency-no-hand-built-declared-error")
}

func TestMutants_base_consistency_require_pagination_argument_name(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "base-consistency-require-pagination-argument-name")
}

func TestProduct_LintMutant_base_consistency_require_pagination_argument_name(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "base-consistency-require-pagination-argument-name")
}

func TestMutants_eqeqeq(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "eqeqeq")
}

func TestProduct_LintMutant_eqeqeq(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "eqeqeq")
}

func TestMutants_max_classes_per_file(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "max-classes-per-file")
}

func TestProduct_LintMutant_max_classes_per_file(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "max-classes-per-file")
}

func TestMutants_max_depth(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "max-depth")
}

func TestProduct_LintMutant_max_depth(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "max-depth")
}

func TestMutants_max_nested_callbacks(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "max-nested-callbacks")
}

func TestProduct_LintMutant_max_nested_callbacks(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "max-nested-callbacks")
}

func TestMutants_nexus_consistency_no_abbreviated_identifier(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "nexus-consistency-no-abbreviated-identifier")
}

func TestProduct_LintMutant_nexus_consistency_no_abbreviated_identifier(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "nexus-consistency-no-abbreviated-identifier")
}

func TestMutants_nexus_consistency_no_ambiguous_identifier(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "nexus-consistency-no-ambiguous-identifier")
}

func TestProduct_LintMutant_nexus_consistency_no_ambiguous_identifier(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "nexus-consistency-no-ambiguous-identifier")
}

func TestMutants_nexus_consistency_no_enum(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "nexus-consistency-no-enum")
}

func TestProduct_LintMutant_nexus_consistency_no_enum(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "nexus-consistency-no-enum")
}

func TestMutants_nexus_consistency_no_for_in(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "nexus-consistency-no-for-in")
}

func TestProduct_LintMutant_nexus_consistency_no_for_in(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "nexus-consistency-no-for-in")
}

func TestMutants_nexus_consistency_no_long_line_comment(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "nexus-consistency-no-long-line-comment")
}

func TestProduct_LintMutant_nexus_consistency_no_long_line_comment(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "nexus-consistency-no-long-line-comment")
}

func TestMutants_nexus_consistency_no_multiline_arrow_function(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "nexus-consistency-no-multiline-arrow-function")
}

func TestProduct_LintMutant_nexus_consistency_no_multiline_arrow_function(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "nexus-consistency-no-multiline-arrow-function")
}

func TestMutants_nexus_consistency_no_shouting(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "nexus-consistency-no-shouting")
}

func TestProduct_LintMutant_nexus_consistency_no_shouting(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "nexus-consistency-no-shouting")
}

func TestMutants_nexus_consistency_no_single_line_jsdoc(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "nexus-consistency-no-single-line-jsdoc")
}

func TestProduct_LintMutant_nexus_consistency_no_single_line_jsdoc(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "nexus-consistency-no-single-line-jsdoc")
}

func TestMutants_nexus_consistency_no_stuttering_name(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "nexus-consistency-no-stuttering-name")
}

func TestProduct_LintMutant_nexus_consistency_no_stuttering_name(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "nexus-consistency-no-stuttering-name")
}

func TestMutants_nexus_consistency_require_type_suffix(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "nexus-consistency-require-type-suffix")
}

func TestProduct_LintMutant_nexus_consistency_require_type_suffix(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "nexus-consistency-require-type-suffix")
}

func TestMutants_no_await_in_loop(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "no-await-in-loop")
}

func TestProduct_LintMutant_no_await_in_loop(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "no-await-in-loop")
}

func TestMutants_no_bitwise(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "no-bitwise")
}

func TestProduct_LintMutant_no_bitwise(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "no-bitwise")
}

func TestMutants_no_caller(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "no-caller")
}

func TestProduct_LintMutant_no_caller(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "no-caller")
}

func TestMutants_no_continue(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "no-continue")
}

func TestProduct_LintMutant_no_continue(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "no-continue")
}

func TestMutants_no_debugger(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "no-debugger")
}

func TestProduct_LintMutant_no_debugger(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "no-debugger")
}

func TestMutants_no_div_regex(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "no-div-regex")
}

func TestProduct_LintMutant_no_div_regex(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "no-div-regex")
}

func TestMutants_no_duplicate_case(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "no-duplicate-case")
}

func TestProduct_LintMutant_no_duplicate_case(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "no-duplicate-case")
}

func TestMutants_no_empty(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "no-empty")
}

func TestProduct_LintMutant_no_empty(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "no-empty")
}

func TestMutants_no_empty_character_class(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "no-empty-character-class")
}

func TestProduct_LintMutant_no_empty_character_class(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "no-empty-character-class")
}

func TestMutants_no_empty_pattern(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "no-empty-pattern")
}

func TestProduct_LintMutant_no_empty_pattern(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "no-empty-pattern")
}

func TestMutants_no_empty_static_block(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "no-empty-static-block")
}

func TestProduct_LintMutant_no_empty_static_block(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "no-empty-static-block")
}

func TestMutants_no_ex_assign(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "no-ex-assign")
}

func TestProduct_LintMutant_no_ex_assign(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "no-ex-assign")
}

func TestMutants_no_iterator(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "no-iterator")
}

func TestProduct_LintMutant_no_iterator(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "no-iterator")
}

func TestMutants_no_labels(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "no-labels")
}

func TestProduct_LintMutant_no_labels(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "no-labels")
}

func TestMutants_no_multi_assign(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "no-multi-assign")
}

func TestProduct_LintMutant_no_multi_assign(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "no-multi-assign")
}

func TestMutants_no_negated_condition(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "no-negated-condition")
}

func TestProduct_LintMutant_no_negated_condition(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "no-negated-condition")
}

func TestMutants_no_new(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "no-new")
}

func TestProduct_LintMutant_no_new(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "no-new")
}

func TestMutants_no_octal_escape(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "no-octal-escape")
}

func TestProduct_LintMutant_no_octal_escape(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "no-octal-escape")
}

func TestMutants_no_plusplus(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "no-plusplus")
}

func TestProduct_LintMutant_no_plusplus(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "no-plusplus")
}

func TestMutants_no_proto(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "no-proto")
}

func TestProduct_LintMutant_no_proto(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "no-proto")
}

func TestMutants_no_return_assign(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "no-return-assign")
}

func TestProduct_LintMutant_no_return_assign(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "no-return-assign")
}

func TestMutants_no_script_url(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "no-script-url")
}

func TestProduct_LintMutant_no_script_url(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "no-script-url")
}

func TestMutants_no_self_compare(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "no-self-compare")
}

func TestProduct_LintMutant_no_self_compare(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "no-self-compare")
}

func TestMutants_no_sequences(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "no-sequences")
}

func TestProduct_LintMutant_no_sequences(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "no-sequences")
}

func TestMutants_no_sparse_arrays(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "no-sparse-arrays")
}

func TestProduct_LintMutant_no_sparse_arrays(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "no-sparse-arrays")
}

func TestMutants_no_template_curly_in_string(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "no-template-curly-in-string")
}

func TestProduct_LintMutant_no_template_curly_in_string(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "no-template-curly-in-string")
}

func TestMutants_no_underscore_dangle(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "no-underscore-dangle")
}

func TestProduct_LintMutant_no_underscore_dangle(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "no-underscore-dangle")
}

func TestMutants_no_unexpected_multiline(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "no-unexpected-multiline")
}

func TestProduct_LintMutant_no_unexpected_multiline(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "no-unexpected-multiline")
}

func TestMutants_no_unneeded_ternary(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "no-unneeded-ternary")
}

func TestProduct_LintMutant_no_unneeded_ternary(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "no-unneeded-ternary")
}

func TestMutants_no_unsafe_finally(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "no-unsafe-finally")
}

func TestProduct_LintMutant_no_unsafe_finally(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "no-unsafe-finally")
}

func TestMutants_no_unsafe_negation(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "no-unsafe-negation")
}

func TestProduct_LintMutant_no_unsafe_negation(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "no-unsafe-negation")
}

func TestMutants_no_unsafe_optional_chaining(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "no-unsafe-optional-chaining")
}

func TestProduct_LintMutant_no_unsafe_optional_chaining(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "no-unsafe-optional-chaining")
}

func TestMutants_no_unused_private_class_members(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "no-unused-private-class-members")
}

func TestProduct_LintMutant_no_unused_private_class_members(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "no-unused-private-class-members")
}

func TestMutants_no_useless_catch(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "no-useless-catch")
}

func TestProduct_LintMutant_no_useless_catch(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "no-useless-catch")
}

func TestMutants_no_useless_concat(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "no-useless-concat")
}

func TestProduct_LintMutant_no_useless_concat(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "no-useless-concat")
}

func TestMutants_no_useless_constructor(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "no-useless-constructor")
}

func TestProduct_LintMutant_no_useless_constructor(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "no-useless-constructor")
}

func TestMutants_no_var(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "no-var")
}

func TestProduct_LintMutant_no_var(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "no-var")
}

func TestMutants_no_warning_comments(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "no-warning-comments")
}

func TestProduct_LintMutant_no_warning_comments(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "no-warning-comments")
}

func TestMutants_no_with(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "no-with")
}

func TestProduct_LintMutant_no_with(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "no-with")
}

func TestMutants_prefer_destructuring(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "prefer-destructuring")
}

func TestProduct_LintMutant_prefer_destructuring(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "prefer-destructuring")
}

func TestMutants_prefer_template(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "prefer-template")
}

func TestProduct_LintMutant_prefer_template(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "prefer-template")
}

func TestMutants_react_forward_ref_uses_ref(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "react-forward-ref-uses-ref")
}

func TestProduct_LintMutant_react_forward_ref_uses_ref(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "react-forward-ref-uses-ref")
}

func TestMutants_react_jsx_no_comment_textnodes(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "react-jsx-no-comment-textnodes")
}

func TestProduct_LintMutant_react_jsx_no_comment_textnodes(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "react-jsx-no-comment-textnodes")
}

func TestMutants_react_no_find_dom_node(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "react-no-find-dom-node")
}

func TestProduct_LintMutant_react_no_find_dom_node(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "react-no-find-dom-node")
}

func TestMutants_react_no_is_mounted(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "react-no-is-mounted")
}

func TestProduct_LintMutant_react_no_is_mounted(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "react-no-is-mounted")
}

func TestMutants_react_no_redundant_should_component_update(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "react-no-redundant-should-component-update")
}

func TestProduct_LintMutant_react_no_redundant_should_component_update(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "react-no-redundant-should-component-update")
}

func TestMutants_require_yield(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "require-yield")
}

func TestProduct_LintMutant_require_yield(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "require-yield")
}

func TestMutants_typescript_adjacent_overload_signatures(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "typescript-adjacent-overload-signatures")
}

func TestProduct_LintMutant_typescript_adjacent_overload_signatures(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "typescript-adjacent-overload-signatures")
}

func TestMutants_typescript_init_declarations(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "typescript-init-declarations")
}

func TestProduct_LintMutant_typescript_init_declarations(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "typescript-init-declarations")
}

func TestMutants_typescript_method_signature_style(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "typescript-method-signature-style")
}

func TestProduct_LintMutant_typescript_method_signature_style(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "typescript-method-signature-style")
}

func TestMutants_typescript_no_this_alias(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "typescript-no-this-alias")
}

func TestProduct_LintMutant_typescript_no_this_alias(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "typescript-no-this-alias")
}

func TestMutants_typescript_no_wrapper_object_types(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "typescript-no-wrapper-object-types")
}

func TestProduct_LintMutant_typescript_no_wrapper_object_types(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "typescript-no-wrapper-object-types")
}

func TestMutants_typescript_prefer_literal_enum_member(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "typescript-prefer-literal-enum-member")
}

func TestProduct_LintMutant_typescript_prefer_literal_enum_member(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "typescript-prefer-literal-enum-member")
}

func TestMutants_unicode_bom(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "unicode-bom")
}

func TestProduct_LintMutant_unicode_bom(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "unicode-bom")
}

func TestMutants_vars_on_top(t *testing.T) {
	t.Parallel()
	mutantRuleShard(t, "vars-on-top")
}

func TestProduct_LintMutant_vars_on_top(t *testing.T) {
	t.Parallel()
	mutantRuleProduct(t, "vars-on-top")
}
