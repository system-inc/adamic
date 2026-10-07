package oracle

func init() {
	for _, path := range []string{
		"internal/oracle/testdata/method_coverage_alias_order.a",
		"internal/oracle/testdata/method_coverage_array_mutations.a",
		"internal/oracle/testdata/method_coverage_array_values.a",
		"internal/oracle/testdata/method_coverage_bound_strings.a",
		"internal/oracle/testdata/method_coverage_codes.a",
		"internal/oracle/testdata/method_coverage_conversions.a",
		"internal/oracle/testdata/method_coverage_empty_alias_tdz.a",
		"internal/oracle/testdata/method_coverage_format_failure.a",
		"internal/oracle/testdata/method_coverage_map_parsers.a",
		"internal/oracle/testdata/method_coverage_map_strings.a",
		"internal/oracle/testdata/method_coverage_math.a",
		"internal/oracle/testdata/method_coverage_number_formats.a",
		"internal/oracle/testdata/method_coverage_number_statics.a",
		"internal/oracle/testdata/method_coverage_object_descriptors.a",
		"internal/oracle/testdata/method_coverage_object_statics.a",
		"internal/oracle/testdata/method_coverage_object_tags.a",
		"internal/oracle/testdata/method_coverage_receiver_order.a",
		"internal/oracle/testdata/method_coverage_string_at.a",
		"internal/oracle/testdata/method_coverage_string_charat.a",
		"internal/oracle/testdata/method_coverage_string_charcodeat.a",
		"internal/oracle/testdata/method_coverage_string_codepointat.a",
		"internal/oracle/testdata/method_coverage_string_concat.a",
		"internal/oracle/testdata/method_coverage_string_endswith.a",
		"internal/oracle/testdata/method_coverage_string_includes.a",
		"internal/oracle/testdata/method_coverage_string_indexof.a",
		"internal/oracle/testdata/method_coverage_string_lastindexof.a",
		"internal/oracle/testdata/method_coverage_string_normalize.a",
		"internal/oracle/testdata/method_coverage_string_overloads.a",
		"internal/oracle/testdata/method_coverage_string_padend.a",
		"internal/oracle/testdata/method_coverage_string_padstart.a",
		"internal/oracle/testdata/method_coverage_string_repeat.a",
		"internal/oracle/testdata/method_coverage_string_slice.a",
		"internal/oracle/testdata/method_coverage_string_startswith.a",
		"internal/oracle/testdata/method_coverage_string_substring.a",
		"internal/oracle/testdata/method_coverage_string_tolowercase.a",
		"internal/oracle/testdata/method_coverage_string_tostring.a",
		"internal/oracle/testdata/method_coverage_string_touppercase.a",
		"internal/oracle/testdata/method_coverage_string_trim.a",
		"internal/oracle/testdata/method_coverage_string_trimend.a",
		"internal/oracle/testdata/method_coverage_string_trimstart.a",
		"internal/oracle/testdata/method_coverage_string_valueof.a",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
}
