package oracle

// Release fast path fixtures. Registered here so oracle_test.go's list stays as it was.
func init() {
	for _, path := range []string{
		"internal/oracle/testdata/release_fast_shared.a",
		"internal/oracle/testdata/release_fast_graph.a",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
}
