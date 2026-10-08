package oracle

func init() {
	for _, path := range []string{
		"internal/oracle/testdata/notyet_union_callable_name.a",
		"internal/oracle/testdata/notyet_union_callable_location.a",
	} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path, false, false})
	}
}
