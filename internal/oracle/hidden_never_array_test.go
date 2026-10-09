package oracle

func init() {
	for _, path := range []string{
		"internal/oracle/testdata/hidden_boundary_never_array.a",
		"internal/oracle/testdata/hidden_boundary_never_array_observations.a",
	} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path, true, false})
	}
}
