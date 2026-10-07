package oracle

func init() {
	for _, path := range []string{"internal/oracle/testdata/user_iterators.a", "internal/oracle/testdata/user_iterators_rest_tdz.a"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
}
