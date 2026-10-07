package oracle

func init() {
	for _, path := range []string{
		"internal/oracle/testdata/accessors.a",
		"internal/oracle/testdata/accessors_generic.a",
		"internal/oracle/testdata/accessors_super.a",
		"internal/oracle/testdata/accessors_order.a",
		"internal/oracle/testdata/accessors_hierarchy.a",
		"internal/oracle/testdata/accessors_ownership.a",
		"internal/oracle/testdata/accessors_throw.a",
		"internal/oracle/testdata/accessors_analyses.a",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
}
