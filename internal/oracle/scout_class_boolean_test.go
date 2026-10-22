package oracle

func init() {
	for _, path := range []string{"internal/oracle/testdata/scout_class_boolean_fields.a", "internal/oracle/testdata/scout_selector_optional_boolean.a", "internal/oracle/testdata/scout_values_optional_boolean.a"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path, true, false})
	}
}
