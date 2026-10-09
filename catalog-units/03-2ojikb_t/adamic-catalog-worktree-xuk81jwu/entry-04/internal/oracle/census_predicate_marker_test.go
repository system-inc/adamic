package oracle

func init() {
	for _, path := range []string{
		"internal/oracle/testdata/census_predicate_marker.a",
		"internal/oracle/testdata/census_predicate_marker_live.a",
	} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path, true, false})
	}
}
