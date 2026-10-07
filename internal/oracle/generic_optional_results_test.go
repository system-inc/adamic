package oracle

func init() {
	for _, path := range []string{"internal/oracle/testdata/generic-optional-callback-result.a", "internal/oracle/testdata/generic_optional_callback_results.a"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path, true, false})
	}
}
