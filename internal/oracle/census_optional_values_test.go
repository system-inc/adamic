package oracle

func init() {
	for _, path := range []string{"internal/oracle/testdata/census_optional_values_system.a", "internal/oracle/testdata/census_optional_values_forms.a", "internal/oracle/testdata/census_optional_values_receivers.a", "internal/oracle/testdata/census_optional_values_padding.a"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
}
