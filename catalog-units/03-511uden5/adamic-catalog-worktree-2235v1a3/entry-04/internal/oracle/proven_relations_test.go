package oracle

// Register separately so the shared Node, native, sanitizer and leak oracle owns these fixtures.
func init() {
	for _, path := range []string{
		"internal/oracle/testdata/proven_satisfies.a",
		"internal/oracle/testdata/proven_upcasts.a",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
}
