package flow

// These fixtures have no lowered graph. The oracle's ClassWrongOutput103,
// 106, 107 and 108 tests hold each refusal and source Node observation. Keep
// them out of trace/SSA/range tests without ignoring unexpected lowering errors.
func refusedOracleFixture(name string) bool {
	switch name {
	case "classfeat_init_super.a", "classfeat_init_super_getter.a", "classfeat_init_super_number.a",
		"classfeat_static_private_instance.a", "classfeat_static_private_method.a",
		"iterators_hidden_return.a", "iterators_override_this.a", "iterators_sym_keys_view.a":
		return true
	}
	return false
}
