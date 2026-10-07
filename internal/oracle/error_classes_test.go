package oracle

func init() {
	for _, path := range []string{
		"internal/oracle/testdata/error_record_narrowing_identity.a",
		"internal/oracle/testdata/error_host_uncaught.a",
		"internal/oracle/testdata/coverage_error_normalize_forms.a",
		"internal/oracle/testdata/coverage_error_argument_order.a",
		"internal/oracle/testdata/coverage_error_catch_finally.a",
		"internal/oracle/testdata/coverage_error_uncaught_surrogate.a",
		"internal/oracle/testdata/coverage_error_codepoint_uncaught.a",
		"internal/oracle/testdata/coverage_error_array_with.a",
		"internal/oracle/testdata/coverage_error_bounds_mutation.a",
		"internal/oracle/testdata/coverage_error_callback_library.a",
		"internal/oracle/testdata/coverage_error_causes.a",
		"internal/oracle/testdata/coverage_error_interface_throw.a",
		"internal/oracle/testdata/coverage_error_kinds.a",
		"internal/oracle/testdata/coverage_error_nested_rethrow.a",
		"internal/oracle/testdata/coverage_error_number_bounds.a",
		"internal/oracle/testdata/coverage_error_number_prototypes.a",
		"internal/oracle/testdata/coverage_error_optional_messages.a",
		"internal/oracle/testdata/coverage_error_padding_dispatch.a",
		"internal/oracle/testdata/coverage_error_prototype.a",
		"internal/oracle/testdata/coverage_error_repeat_unicode.a",
		"internal/oracle/testdata/coverage_error_tdz_interface.a",
		"internal/oracle/testdata/error_checks.a", "internal/oracle/testdata/error_classes.a", "internal/oracle/testdata/error_classes_uncaught.a", "internal/oracle/testdata/error_classes_uncaught_name.a", "internal/oracle/testdata/error_classes_uncaught_empty.a"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
}
