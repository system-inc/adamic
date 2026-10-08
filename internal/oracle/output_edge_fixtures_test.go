package oracle

// Finite edge fixtures also run through the ordinary three-backend oracle, leak checks and counts.
// The spin fixture is held only by timed signal tests: it never finishes or has stable counts.
func init() {
	for _, path := range []string{"fsize.a", "fsize_out.a", "closed.a", "panic_surrogate.a", "usr1.a"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/output_edges/" + path, true, false})
	}
}
