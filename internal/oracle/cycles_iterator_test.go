package oracle

// The controls use the shared Node, JavaScript, native and sanitizer oracle, including leaks.
func init() {
	for _, path := range []string{
		"internal/oracle/testdata/cyc_iface_control.a",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
}
