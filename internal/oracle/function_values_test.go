package oracle

func init() {
	for _, path := range []string{
		"internal/oracle/testdata/function_values_return.a",
		"internal/oracle/testdata/function_values_options.a",
		"internal/oracle/testdata/function_values_signature.a",
		"internal/oracle/testdata/function_values_chain.a",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
}
