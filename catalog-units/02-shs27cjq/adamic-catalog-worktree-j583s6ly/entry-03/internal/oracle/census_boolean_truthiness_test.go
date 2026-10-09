package oracle

func init() {
	for _, path := range []string{"internal/oracle/testdata/census_boolean_negation.a", "internal/oracle/testdata/census_boolean_sieve.a", "internal/oracle/testdata/census_boolean_operators.a", "internal/oracle/testdata/census_boolean_dead_branch.a"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
}
