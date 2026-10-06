package oracle

func init() {
	for _, path := range []string{
		"internal/oracle/testdata/enums_coverage_flag_keys.a",
		"internal/oracle/testdata/enums_coverage_flag_proofs.a",
		"internal/oracle/testdata/enums_coverage_flag_slots.a",
		"internal/oracle/testdata/enums_coverage_generic.a",
		"internal/oracle/testdata/enums_coverage_reflection.a",
		"internal/oracle/testdata/enums_coverage_lookup.a",
		"internal/oracle/testdata/enums_coverage_switches.a",
		"internal/oracle/testdata/enums_coverage_const_flags.a",
		"internal/oracle/testdata/enums_coverage_initialization.a",
		"internal/oracle/testdata/enums_coverage_modules/main.a",
		"internal/oracle/testdata/enums_coverage_numeric_edges.a",
		"internal/oracle/testdata/enums_coverage_string_names.a",
		"internal/oracle/testdata/enums_coverage_constants.a",
		"internal/oracle/testdata/enums_coverage_views.a",
		"internal/oracle/testdata/enums_coverage_shapes.a",
		"internal/oracle/testdata/enums_coverage_mask_edges.a",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
}
