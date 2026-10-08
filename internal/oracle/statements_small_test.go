package oracle

func init() {
	for _, path := range []string{"internal/oracle/testdata/statements_small_prefix.a", "internal/oracle/testdata/statements_small_parameters.a", "internal/oracle/testdata/statements_small_throw.a"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
}
