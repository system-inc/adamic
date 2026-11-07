package lint

import "testing"

// One independently discoverable mutant and product declaration per registered rule.

func TestMutants_adamic_no_type_predicate(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "adamic-no-type-predicate")
}

func TestProduct_LintMutant_adamic_no_type_predicate(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "adamic-no-type-predicate")
}

func TestMutants_base_boundary_no_global_container(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "base-boundary-no-global-container")
}

func TestProduct_LintMutant_base_boundary_no_global_container(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "base-boundary-no-global-container")
}

func TestMutants_base_consistency_no_console(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "base-consistency-no-console")
}

func TestProduct_LintMutant_base_consistency_no_console(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "base-consistency-no-console")
}

func TestMutants_base_consistency_no_hand_built_declared_error(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "base-consistency-no-hand-built-declared-error")
}

func TestProduct_LintMutant_base_consistency_no_hand_built_declared_error(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "base-consistency-no-hand-built-declared-error")
}

func TestMutants_base_consistency_require_pagination_argument_name(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "base-consistency-require-pagination-argument-name")
}

func TestProduct_LintMutant_base_consistency_require_pagination_argument_name(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "base-consistency-require-pagination-argument-name")
}

func TestMutants_eqeqeq(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "eqeqeq")
}

func TestProduct_LintMutant_eqeqeq(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "eqeqeq")
}

func TestMutants_max_classes_per_file(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "max-classes-per-file")
}

func TestProduct_LintMutant_max_classes_per_file(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "max-classes-per-file")
}

func TestMutants_max_depth(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "max-depth")
}

func TestProduct_LintMutant_max_depth(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "max-depth")
}

func TestMutants_max_nested_callbacks(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "max-nested-callbacks")
}

func TestProduct_LintMutant_max_nested_callbacks(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "max-nested-callbacks")
}

func TestMutants_nexus_consistency_no_abbreviated_identifier(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "nexus-consistency-no-abbreviated-identifier")
}

func TestProduct_LintMutant_nexus_consistency_no_abbreviated_identifier(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "nexus-consistency-no-abbreviated-identifier")
}

func TestMutants_nexus_consistency_no_ambiguous_identifier(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "nexus-consistency-no-ambiguous-identifier")
}

func TestProduct_LintMutant_nexus_consistency_no_ambiguous_identifier(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "nexus-consistency-no-ambiguous-identifier")
}

func TestMutants_nexus_consistency_no_enum(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "nexus-consistency-no-enum")
}

func TestProduct_LintMutant_nexus_consistency_no_enum(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "nexus-consistency-no-enum")
}

func TestMutants_nexus_consistency_no_for_in(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "nexus-consistency-no-for-in")
}

func TestProduct_LintMutant_nexus_consistency_no_for_in(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "nexus-consistency-no-for-in")
}

func TestMutants_nexus_consistency_no_long_line_comment(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "nexus-consistency-no-long-line-comment")
}

func TestProduct_LintMutant_nexus_consistency_no_long_line_comment(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "nexus-consistency-no-long-line-comment")
}

func TestMutants_nexus_consistency_no_multiline_arrow_function(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "nexus-consistency-no-multiline-arrow-function")
}

func TestProduct_LintMutant_nexus_consistency_no_multiline_arrow_function(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "nexus-consistency-no-multiline-arrow-function")
}

func TestMutants_nexus_consistency_no_shouting(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "nexus-consistency-no-shouting")
}

func TestProduct_LintMutant_nexus_consistency_no_shouting(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "nexus-consistency-no-shouting")
}

func TestMutants_nexus_consistency_no_single_line_jsdoc(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "nexus-consistency-no-single-line-jsdoc")
}

func TestProduct_LintMutant_nexus_consistency_no_single_line_jsdoc(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "nexus-consistency-no-single-line-jsdoc")
}

func TestMutants_nexus_consistency_no_stuttering_name(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "nexus-consistency-no-stuttering-name")
}

func TestProduct_LintMutant_nexus_consistency_no_stuttering_name(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "nexus-consistency-no-stuttering-name")
}

func TestMutants_nexus_consistency_require_type_suffix(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "nexus-consistency-require-type-suffix")
}

func TestProduct_LintMutant_nexus_consistency_require_type_suffix(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "nexus-consistency-require-type-suffix")
}

func TestMutants_no_await_in_loop(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "no-await-in-loop")
}

func TestProduct_LintMutant_no_await_in_loop(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "no-await-in-loop")
}

func TestMutants_no_bitwise(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "no-bitwise")
}

func TestProduct_LintMutant_no_bitwise(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "no-bitwise")
}

func TestMutants_no_caller(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "no-caller")
}

func TestProduct_LintMutant_no_caller(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "no-caller")
}

func TestMutants_no_continue(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "no-continue")
}

func TestProduct_LintMutant_no_continue(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "no-continue")
}

func TestMutants_no_debugger(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "no-debugger")
}

func TestProduct_LintMutant_no_debugger(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "no-debugger")
}

func TestMutants_no_div_regex(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "no-div-regex")
}

func TestProduct_LintMutant_no_div_regex(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "no-div-regex")
}

func TestMutants_no_duplicate_case(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "no-duplicate-case")
}

func TestProduct_LintMutant_no_duplicate_case(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "no-duplicate-case")
}

func TestMutants_no_empty(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "no-empty")
}

func TestProduct_LintMutant_no_empty(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "no-empty")
}

func TestMutants_no_empty_character_class(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "no-empty-character-class")
}

func TestProduct_LintMutant_no_empty_character_class(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "no-empty-character-class")
}

func TestMutants_no_empty_pattern(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "no-empty-pattern")
}

func TestProduct_LintMutant_no_empty_pattern(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "no-empty-pattern")
}

func TestMutants_no_empty_static_block(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "no-empty-static-block")
}

func TestProduct_LintMutant_no_empty_static_block(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "no-empty-static-block")
}

func TestMutants_no_ex_assign(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "no-ex-assign")
}

func TestProduct_LintMutant_no_ex_assign(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "no-ex-assign")
}

func TestMutants_no_iterator(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "no-iterator")
}

func TestProduct_LintMutant_no_iterator(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "no-iterator")
}

func TestMutants_no_labels(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "no-labels")
}

func TestProduct_LintMutant_no_labels(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "no-labels")
}

func TestMutants_no_multi_assign(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "no-multi-assign")
}

func TestProduct_LintMutant_no_multi_assign(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "no-multi-assign")
}

func TestMutants_no_negated_condition(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "no-negated-condition")
}

func TestProduct_LintMutant_no_negated_condition(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "no-negated-condition")
}

func TestMutants_no_new(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "no-new")
}

func TestProduct_LintMutant_no_new(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "no-new")
}

func TestMutants_no_octal_escape(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "no-octal-escape")
}

func TestProduct_LintMutant_no_octal_escape(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "no-octal-escape")
}

func TestMutants_no_plusplus(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "no-plusplus")
}

func TestProduct_LintMutant_no_plusplus(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "no-plusplus")
}

func TestMutants_no_proto(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "no-proto")
}

func TestProduct_LintMutant_no_proto(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "no-proto")
}

func TestMutants_no_return_assign(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "no-return-assign")
}

func TestProduct_LintMutant_no_return_assign(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "no-return-assign")
}

func TestMutants_no_script_url(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "no-script-url")
}

func TestProduct_LintMutant_no_script_url(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "no-script-url")
}

func TestMutants_no_self_compare(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "no-self-compare")
}

func TestProduct_LintMutant_no_self_compare(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "no-self-compare")
}

func TestMutants_no_sequences(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "no-sequences")
}

func TestProduct_LintMutant_no_sequences(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "no-sequences")
}

func TestMutants_no_sparse_arrays(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "no-sparse-arrays")
}

func TestProduct_LintMutant_no_sparse_arrays(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "no-sparse-arrays")
}

func TestMutants_no_template_curly_in_string(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "no-template-curly-in-string")
}

func TestProduct_LintMutant_no_template_curly_in_string(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "no-template-curly-in-string")
}

func TestMutants_no_underscore_dangle(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "no-underscore-dangle")
}

func TestProduct_LintMutant_no_underscore_dangle(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "no-underscore-dangle")
}

func TestMutants_no_unexpected_multiline(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "no-unexpected-multiline")
}

func TestProduct_LintMutant_no_unexpected_multiline(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "no-unexpected-multiline")
}

func TestMutants_no_unneeded_ternary(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "no-unneeded-ternary")
}

func TestProduct_LintMutant_no_unneeded_ternary(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "no-unneeded-ternary")
}

func TestMutants_no_unsafe_finally(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "no-unsafe-finally")
}

func TestProduct_LintMutant_no_unsafe_finally(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "no-unsafe-finally")
}

func TestMutants_no_unsafe_negation(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "no-unsafe-negation")
}

func TestProduct_LintMutant_no_unsafe_negation(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "no-unsafe-negation")
}

func TestMutants_no_unsafe_optional_chaining(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "no-unsafe-optional-chaining")
}

func TestProduct_LintMutant_no_unsafe_optional_chaining(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "no-unsafe-optional-chaining")
}

func TestMutants_no_unused_private_class_members(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "no-unused-private-class-members")
}

func TestProduct_LintMutant_no_unused_private_class_members(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "no-unused-private-class-members")
}

func TestMutants_no_useless_catch(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "no-useless-catch")
}

func TestProduct_LintMutant_no_useless_catch(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "no-useless-catch")
}

func TestMutants_no_useless_concat(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "no-useless-concat")
}

func TestProduct_LintMutant_no_useless_concat(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "no-useless-concat")
}

func TestMutants_no_useless_constructor(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "no-useless-constructor")
}

func TestProduct_LintMutant_no_useless_constructor(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "no-useless-constructor")
}

func TestMutants_no_var(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "no-var")
}

func TestProduct_LintMutant_no_var(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "no-var")
}

func TestMutants_no_warning_comments(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "no-warning-comments")
}

func TestProduct_LintMutant_no_warning_comments(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "no-warning-comments")
}

func TestMutants_no_with(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "no-with")
}

func TestProduct_LintMutant_no_with(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "no-with")
}

func TestMutants_prefer_destructuring(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "prefer-destructuring")
}

func TestProduct_LintMutant_prefer_destructuring(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "prefer-destructuring")
}

func TestMutants_prefer_template(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "prefer-template")
}

func TestProduct_LintMutant_prefer_template(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "prefer-template")
}

func TestMutants_react_forward_ref_uses_ref(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "react-forward-ref-uses-ref")
}

func TestProduct_LintMutant_react_forward_ref_uses_ref(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "react-forward-ref-uses-ref")
}

func TestMutants_react_jsx_no_comment_textnodes(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "react-jsx-no-comment-textnodes")
}

func TestProduct_LintMutant_react_jsx_no_comment_textnodes(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "react-jsx-no-comment-textnodes")
}

func TestMutants_react_no_find_dom_node(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "react-no-find-dom-node")
}

func TestProduct_LintMutant_react_no_find_dom_node(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "react-no-find-dom-node")
}

func TestMutants_react_no_is_mounted(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "react-no-is-mounted")
}

func TestProduct_LintMutant_react_no_is_mounted(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "react-no-is-mounted")
}

func TestMutants_react_no_redundant_should_component_update(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "react-no-redundant-should-component-update")
}

func TestProduct_LintMutant_react_no_redundant_should_component_update(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "react-no-redundant-should-component-update")
}

func TestMutants_require_yield(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "require-yield")
}

func TestProduct_LintMutant_require_yield(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "require-yield")
}

func TestMutants_typescript_adjacent_overload_signatures(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "typescript-adjacent-overload-signatures")
}

func TestProduct_LintMutant_typescript_adjacent_overload_signatures(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "typescript-adjacent-overload-signatures")
}

func TestMutants_typescript_init_declarations(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "typescript-init-declarations")
}

func TestProduct_LintMutant_typescript_init_declarations(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "typescript-init-declarations")
}

func TestMutants_typescript_method_signature_style(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "typescript-method-signature-style")
}

func TestProduct_LintMutant_typescript_method_signature_style(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "typescript-method-signature-style")
}

func TestMutants_typescript_no_this_alias(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "typescript-no-this-alias")
}

func TestProduct_LintMutant_typescript_no_this_alias(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "typescript-no-this-alias")
}

func TestMutants_typescript_no_wrapper_object_types(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "typescript-no-wrapper-object-types")
}

func TestProduct_LintMutant_typescript_no_wrapper_object_types(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "typescript-no-wrapper-object-types")
}

func TestMutants_typescript_prefer_literal_enum_member(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "typescript-prefer-literal-enum-member")
}

func TestProduct_LintMutant_typescript_prefer_literal_enum_member(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "typescript-prefer-literal-enum-member")
}

func TestMutants_unicode_bom(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "unicode-bom")
}

func TestProduct_LintMutant_unicode_bom(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "unicode-bom")
}

func TestMutants_vars_on_top(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "vars-on-top")
}

func TestProduct_LintMutant_vars_on_top(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "vars-on-top")
}

func TestMutants_no_cond_assign(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "no-cond-assign")
}

func TestProduct_LintMutant_no_cond_assign(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "no-cond-assign")
}

func TestMutants_no_extra_bind(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "no-extra-bind")
}

func TestProduct_LintMutant_no_extra_bind(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "no-extra-bind")
}

func TestMutants_no_unused_labels(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "no-unused-labels")
}

func TestProduct_LintMutant_no_unused_labels(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "no-unused-labels")
}

func TestMutants_no_useless_rename(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "no-useless-rename")
}

func TestProduct_LintMutant_no_useless_rename(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "no-useless-rename")
}

func TestMutants_default_param_last(test *testing.T) {
	test.Parallel()
	mutantRuleShard(test, "default-param-last")
}

func TestProduct_LintMutant_default_param_last(test *testing.T) {
	test.Parallel()
	requireMutantRuleProduct(test, "default-param-last")
}
