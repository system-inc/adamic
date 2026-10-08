package oracle

func init() {
	for _, path := range []string{
		"internal/oracle/testdata/generic_function_census_contains.a",
		"internal/oracle/testdata/generic_function_census_extra_arguments.a",
		"internal/oracle/testdata/generic_function_census_names.a",
		"internal/oracle/testdata/generic_function_census_positions.a",
	} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path, true, false})
	}
}
