package oracle

func init() {
	for _, path := range []string{"internal/oracle/testdata/rest_tuple_create.a", "internal/oracle/testdata/rest_tuple_update.a"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
}
