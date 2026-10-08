package oracle

func init() {
	for _, path := range []string{
		"internal/oracle/testdata/notyet_void_undefined_push.a",
		"internal/oracle/testdata/notyet_void_undefined_pop.a",
	} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path, false, false})
	}
}
