package oracle

// The shared oracle runs this fixture against source Node, generated JavaScript, native release,
// ASan/UBSan and the leak check. Register separately to avoid changing oracle_test.go's fixture list.
func init() {
	for _, path := range []string{
		"internal/oracle/testdata/classfeat_private_literal.a",
		"internal/oracle/testdata/classfeat_static_private_instance.a",
		"internal/oracle/testdata/classfeat_static_private_method.a",
		"internal/oracle/testdata/classfeat_maybe_setter.a",
		"internal/oracle/testdata/classfeat_method_view.a",
		"internal/oracle/testdata/class_features_static.a",
		"internal/oracle/testdata/class_features_static_private.a",
		"internal/oracle/testdata/class_features_private.a",
		"internal/oracle/testdata/class_features_accessors.a",
		"internal/oracle/testdata/class_features_twice.a",
		"internal/oracle/testdata/class_features_retained.a",
		"internal/oracle/testdata/class_features_distinct.a",
		"internal/oracle/testdata/class_inheritance.a",
		"internal/oracle/testdata/class_inheritance_exceptions.a",
		"internal/oracle/testdata/class_inheritance_order.a",
		"internal/oracle/testdata/class_identity.a",
		"internal/oracle/testdata/class_inheritance_memory.a",
		"internal/oracle/testdata/class_inheritance_generic.a",
		"internal/oracle/testdata/class_inheritance_interface.a",
		"internal/oracle/testdata/class_private_generic.a",
		"internal/oracle/testdata/class_private_members.a",
		"internal/oracle/testdata/class_inheritance_conditional.a",
		"internal/oracle/testdata/class_super_closure.a",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
}
