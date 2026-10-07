package oracle

// Register in the shared Node, native, sanitizer and leak oracle without editing its main list.
func init() {
	for _, path := range []string{
		"internal/oracle/testdata/class_static.a",
		"internal/oracle/testdata/class_static_calls.a",
		"internal/oracle/testdata/class_static_alias.a",
		"internal/oracle/testdata/class_static_alias_tdz.a",
		"internal/oracle/testdata/class_static_tdz.a",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
}
