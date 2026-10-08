package oracle

func init() {
	for _, path := range []string{
		"internal/oracle/testdata/rest_union_diagnostic.a",
		"internal/oracle/testdata/rest_union_token.a",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
}
