package oracle

import "testing"

// Each checked-write witness owns a bounded top-level test for Loom.
func TestCheckedWiderWrites_overload_callback_fit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[0])
}
func TestCheckedWiderWrites_overload_callback_misfit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[1])
}
func TestCheckedWiderWrites_optional_boolean_fit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[2])
}
func TestCheckedWiderWrites_optional_boolean_misfit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[3])
}
func TestCheckedWiderWrites_required_boolean_fit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[4])
}
func TestCheckedWiderWrites_required_boolean_misfit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[5])
}
func TestCheckedWiderWrites_optional_false_fit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[6])
}
func TestCheckedWiderWrites_optional_false_misfit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[7])
}
func TestCheckedWiderWrites_branch_file_fit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[8])
}
func TestCheckedWiderWrites_branch_file_misfit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[9])
}
func TestCheckedWiderWrites_branch_start_fit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[10])
}
func TestCheckedWiderWrites_branch_start_misfit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[11])
}
func TestCheckedWiderWrites_branch_length_fit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[12])
}
func TestCheckedWiderWrites_branch_length_misfit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[13])
}
func TestCheckedWiderWrites_branch_bottom_number_fit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[14])
}
func TestCheckedWiderWrites_branch_bottom_number_misfit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[15])
}
func TestCheckedWiderWrites_branch_bottom_string_fit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[16])
}
func TestCheckedWiderWrites_branch_bottom_string_misfit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[17])
}
func TestCheckedWiderWrites_generic_site_fit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[18])
}
func TestCheckedWiderWrites_generic_site_misfit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[19])
}
func TestCheckedWiderWrites_overload_result_fit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[20])
}
func TestCheckedWiderWrites_overload_result_misfit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[21])
}
func TestCheckedWiderWrites_emit_comment_fit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[22])
}
func TestCheckedWiderWrites_emit_comment_misfit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[23])
}
func TestCheckedWiderWrites_emit_resolution_fit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[24])
}
func TestCheckedWiderWrites_emit_resolution_misfit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[25])
}
func TestCheckedWiderWrites_emit_callback_fit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[26])
}
func TestCheckedWiderWrites_emit_callback_misfit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[27])
}
func TestCheckedWiderWrites_emit_node_fit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[28])
}
func TestCheckedWiderWrites_emit_node_misfit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[29])
}
func TestCheckedWiderWrites_emit_auto_fit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[30])
}
func TestCheckedWiderWrites_emit_auto_misfit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[31])
}
func TestCheckedWiderWrites_container_slice_fit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[32])
}
func TestCheckedWiderWrites_container_slice_misfit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[33])
}
func TestCheckedWiderWrites_container_map_values_fit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[34])
}
func TestCheckedWiderWrites_container_map_values_misfit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[35])
}
func TestCheckedWiderWrites_container_boolean_fit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[36])
}
func TestCheckedWiderWrites_container_boolean_misfit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[37])
}
func TestCheckedWiderWrites_container_fill_fit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[38])
}
func TestCheckedWiderWrites_container_fill_misfit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[39])
}
func TestCheckedWiderWrites_container_map_object_fit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[40])
}
func TestCheckedWiderWrites_container_map_object_misfit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[41])
}
func TestCheckedWiderWrites_container_alias_fit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[42])
}
func TestCheckedWiderWrites_container_alias_misfit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[43])
}
func TestCheckedWiderWrites_container_number_fit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[44])
}
func TestCheckedWiderWrites_container_number_misfit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[45])
}
func TestCheckedWiderWrites_container_string_fit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[46])
}
func TestCheckedWiderWrites_container_string_misfit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[47])
}
func TestCheckedWiderWrites_container_map_fit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[48])
}
func TestCheckedWiderWrites_container_map_misfit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[49])
}
func TestCheckedWiderWrites_container_splice_fit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[50])
}
func TestCheckedWiderWrites_container_splice_misfit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[51])
}
func TestCheckedWiderWrites_container_object_fit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[52])
}
func TestCheckedWiderWrites_container_object_misfit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[53])
}
func TestCheckedWiderWrites_container_nested_fit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[54])
}
func TestCheckedWiderWrites_container_nested_misfit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[55])
}
func TestCheckedWiderWrites_flow_node_fit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[56])
}
func TestCheckedWiderWrites_flow_node_misfit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[57])
}
func TestCheckedWiderWrites_flow_undefined_fit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[58])
}
func TestCheckedWiderWrites_flow_undefined_misfit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[59])
}
func TestCheckedWiderWrites_flow_array_fit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[60])
}
func TestCheckedWiderWrites_flow_array_misfit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[61])
}
func TestCheckedWiderWrites_never_nullable_fit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[62])
}
func TestCheckedWiderWrites_never_nullable_misfit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[63])
}
func TestCheckedWiderWrites_never_index_misfit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[64])
}
func TestCheckedWiderWrites_never_splice_misfit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[65])
}
func TestCheckedWiderWrites_never_wide_fit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[66])
}
func TestCheckedWiderWrites_never_number_fit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[67])
}
func TestCheckedWiderWrites_never_number_misfit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[68])
}
func TestCheckedWiderWrites_never_string_fit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[69])
}
func TestCheckedWiderWrites_never_string_misfit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[70])
}
func TestCheckedWiderWrites_never_object_fit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[71])
}
func TestCheckedWiderWrites_never_object_misfit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[72])
}
func TestCheckedWiderWrites_diagnostic_alias_fit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[73])
}
func TestCheckedWiderWrites_diagnostic_alias_misfit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[74])
}
func TestCheckedWiderWrites_diagnostic_proof_fit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[75])
}
func TestCheckedWiderWrites_diagnostic_proof_misfit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[76])
}
func TestCheckedWiderWrites_diagnostic_rich_fit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[77])
}
func TestCheckedWiderWrites_diagnostic_rich_misfit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[78])
}
func TestCheckedWiderWrites_diagnostic_related_fit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[79])
}
func TestCheckedWiderWrites_diagnostic_related_misfit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[80])
}
func TestCheckedWiderWrites_diagnostic_nested_fit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[81])
}
func TestCheckedWiderWrites_diagnostic_nested_misfit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[82])
}
func TestCheckedWiderWrites_diagnostic_inner_fit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[83])
}
func TestCheckedWiderWrites_diagnostic_inner_misfit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[84])
}
func TestCheckedWiderWrites_reference_structure(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[85])
}
func TestCheckedWiderWrites_flags_fit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[86])
}
func TestCheckedWiderWrites_flags_misfit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[87])
}
func TestCheckedWiderWrites_string_fit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[88])
}
func TestCheckedWiderWrites_string_misfit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[89])
}
func TestCheckedWiderWrites_boolean_fit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[90])
}
func TestCheckedWiderWrites_boolean_misfit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[91])
}
func TestCheckedWiderWrites_diagnostic_fit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[92])
}
func TestCheckedWiderWrites_diagnostic_misfit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[93])
}
func TestCheckedWiderWrites_number_fit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[94])
}
func TestCheckedWiderWrites_number_misfit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[95])
}
func TestCheckedWiderWrites_compound_fit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[96])
}
func TestCheckedWiderWrites_compound_misfit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[97])
}
func TestCheckedWiderWrites_nullable_number_fit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[98])
}
func TestCheckedWiderWrites_nullable_number_misfit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[99])
}
func TestCheckedWiderWrites_range_fit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[100])
}
func TestCheckedWiderWrites_range_misfit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[101])
}
func TestCheckedWiderWrites_receiver_fit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[102])
}
func TestCheckedWiderWrites_parent_fit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[103])
}
func TestCheckedWiderWrites_parent_misfit(t *testing.T) {
	t.Parallel()
	runCheckedWiderWrites(t, checkedWriteFixtures[104])
}
func TestCheckedWiderWriteMutants_drop_check(t *testing.T) {
	t.Parallel()
	runCheckedWiderWriteMutants(t, "drop check")
}
func TestCheckedWiderWriteMutants_drop_literal_set(t *testing.T) {
	t.Parallel()
	runCheckedWiderWriteMutants(t, "drop literal set")
}
func TestCheckedDiagnosticReferenceMutants_drop_alias_check(t *testing.T) {
	t.Parallel()
	runCheckedDiagnosticReferenceMutants(t, struct{ name, fixture string }{"drop alias check", "diagnostic-alias-misfit"})
}
func TestCheckedDiagnosticReferenceMutants_drop_allocation_proof(t *testing.T) {
	t.Parallel()
	runCheckedDiagnosticReferenceMutants(t, struct{ name, fixture string }{"drop allocation proof", "diagnostic-proof-misfit"})
}
func TestCheckedDiagnosticReferenceMutants_drop_nested_contract(t *testing.T) {
	t.Parallel()
	runCheckedDiagnosticReferenceMutants(t, struct{ name, fixture string }{"drop nested contract", "diagnostic-nested-misfit"})
}
func TestCheckedFlowContainerContractMutants_flow_node_misfit(t *testing.T) {
	t.Parallel()
	runCheckedFlowContainerContractMutants(t, "flow-node-misfit")
}
func TestCheckedFlowContainerContractMutants_flow_undefined_misfit(t *testing.T) {
	t.Parallel()
	runCheckedFlowContainerContractMutants(t, "flow-undefined-misfit")
}
func TestCheckedFlowContainerContractMutants_flow_array_misfit(t *testing.T) {
	t.Parallel()
	runCheckedFlowContainerContractMutants(t, "flow-array-misfit")
}
func TestCheckedFlowContainerContractMutants_container_number_misfit(t *testing.T) {
	t.Parallel()
	runCheckedFlowContainerContractMutants(t, "container-number-misfit")
}
func TestCheckedFlowContainerContractMutants_container_string_misfit(t *testing.T) {
	t.Parallel()
	runCheckedFlowContainerContractMutants(t, "container-string-misfit")
}
func TestCheckedFlowContainerContractMutants_container_map_misfit(t *testing.T) {
	t.Parallel()
	runCheckedFlowContainerContractMutants(t, "container-map-misfit")
}
func TestCheckedFlowContainerContractMutants_container_splice_misfit(t *testing.T) {
	t.Parallel()
	runCheckedFlowContainerContractMutants(t, "container-splice-misfit")
}
func TestCheckedFlowContainerContractMutants_container_object_misfit(t *testing.T) {
	t.Parallel()
	runCheckedFlowContainerContractMutants(t, "container-object-misfit")
}
func TestCheckedFlowContainerContractMutants_container_nested_misfit(t *testing.T) {
	t.Parallel()
	runCheckedFlowContainerContractMutants(t, "container-nested-misfit")
}
func TestCheckedFlowContainerContractMutants_container_boolean_misfit(t *testing.T) {
	t.Parallel()
	runCheckedFlowContainerContractMutants(t, "container-boolean-misfit")
}
func TestCheckedFlowContainerContractMutants_container_fill_misfit(t *testing.T) {
	t.Parallel()
	runCheckedFlowContainerContractMutants(t, "container-fill-misfit")
}
func TestCheckedFlowContainerContractMutants_container_map_object_misfit(t *testing.T) {
	t.Parallel()
	runCheckedFlowContainerContractMutants(t, "container-map-object-misfit")
}
func TestCheckedFlowContainerContractMutants_container_alias_misfit(t *testing.T) {
	t.Parallel()
	runCheckedFlowContainerContractMutants(t, "container-alias-misfit")
}
func TestCheckedEmitContractsMutants_emit_drop_check(t *testing.T) {
	t.Parallel()
	runCheckedEmitContractsMutants(t, "emit-drop-check")
}
func TestCheckedEmitContractsMutants_emit_node_misfit(t *testing.T) {
	t.Parallel()
	runCheckedEmitContractsMutants(t, "emit-node-misfit")
}
func TestCheckedEmitContractsMutants_emit_auto_misfit(t *testing.T) {
	t.Parallel()
	runCheckedEmitContractsMutants(t, "emit-auto-misfit")
}
func TestCheckedEmitContractsMutants_emit_callback_misfit(t *testing.T) {
	t.Parallel()
	runCheckedEmitContractsMutants(t, "emit-callback-misfit")
}
func TestCheckedEmitContractsMutants_emit_comment_misfit(t *testing.T) {
	t.Parallel()
	runCheckedEmitContractsMutants(t, "emit-comment-misfit")
}
func TestCheckedEmitContractsMutants_emit_resolution_misfit(t *testing.T) {
	t.Parallel()
	runCheckedEmitContractsMutants(t, "emit-resolution-misfit")
}
func TestCheckedViewsNextMutants_optional_boolean_misfit(t *testing.T) {
	t.Parallel()
	runCheckedViewsNextMutants(t, "optional-boolean-misfit")
}
func TestCheckedViewsNextMutants_optional_false_misfit(t *testing.T) {
	t.Parallel()
	runCheckedViewsNextMutants(t, "optional-false-misfit")
}
func TestCheckedViewsNextMutants_required_boolean_misfit(t *testing.T) {
	t.Parallel()
	runCheckedViewsNextMutants(t, "required-boolean-misfit")
}
func TestCheckedViewsNextMutants_branch_file_misfit(t *testing.T) {
	t.Parallel()
	runCheckedViewsNextMutants(t, "branch-file-misfit")
}
func TestCheckedViewsNextMutants_branch_start_misfit(t *testing.T) {
	t.Parallel()
	runCheckedViewsNextMutants(t, "branch-start-misfit")
}
func TestCheckedViewsNextMutants_branch_length_misfit(t *testing.T) {
	t.Parallel()
	runCheckedViewsNextMutants(t, "branch-length-misfit")
}
func TestCheckedViewsNextMutants_branch_bottom_number_misfit(t *testing.T) {
	t.Parallel()
	runCheckedViewsNextMutants(t, "branch-bottom-number-misfit")
}
func TestCheckedViewsNextMutants_branch_bottom_string_misfit(t *testing.T) {
	t.Parallel()
	runCheckedViewsNextMutants(t, "branch-bottom-string-misfit")
}
func TestCheckedViewsNextMutants_generic_site_misfit(t *testing.T) {
	t.Parallel()
	runCheckedViewsNextMutants(t, "generic-site-misfit")
}
func TestCheckedViewsNextMutants_overload_result_misfit(t *testing.T) {
	t.Parallel()
	runCheckedViewsNextMutants(t, "overload-result-misfit")
}
func TestCheckedViewsNextMutants_overload_callback_misfit(t *testing.T) {
	t.Parallel()
	runCheckedViewsNextMutants(t, "overload-callback-misfit")
}
