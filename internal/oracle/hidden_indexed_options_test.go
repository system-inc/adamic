package oracle

func init() {
	for _, path := range []string{
		"internal/oracle/testdata/hidden_boundary_indexed_options.a",
		"internal/oracle/testdata/hidden_boundary_indexed_options_order.a",
	} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path, true, false})
	}
}
