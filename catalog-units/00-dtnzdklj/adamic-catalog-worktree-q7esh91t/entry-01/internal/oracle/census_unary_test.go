package oracle

func init() {
	for _, path := range []string{"internal/oracle/testdata/census_unary_truthiness.a", "internal/oracle/testdata/census_unary_numeric.a", "internal/oracle/testdata/census_overload_contracts.a"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
}
