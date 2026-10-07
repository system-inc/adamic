package oracle

func init() {
	for _, path := range []string{
		"internal/oracle/testdata/proven_guards.a",
		"internal/oracle/testdata/proven_class_guards.a",
		"internal/oracle/testdata/proven_assertions.a",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
}
