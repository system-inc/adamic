package oracle

func init() {
	for _, path := range []string{
		"internal/oracle/testdata/rest_callable_not.a",
		"internal/oracle/testdata/rest_callable_write.a",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
}
