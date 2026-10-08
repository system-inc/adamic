package oracle

func init() {
	for _, path := range []string{
		"internal/oracle/testdata/function_values_return.a",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
}
