package oracle

// Programs that exercise stdout and stderr edges the ordinary fixtures did not
// combine: a little of each, more than one buffer with both streams, stderr
// alone, panic text around the surrogate range, and a read after both streams.
func init() {
	for _, path := range []string{
		"little_streams.a",
		"lots_streams.a",
		"stderr_only.a",
		"panic_edges.a",
		"read_after_print.a",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/stdout_edges/" + path, true, false})
	}
}
