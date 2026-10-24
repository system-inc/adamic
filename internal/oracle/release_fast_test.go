package oracle

// Each witness belongs to the ordinary Node/native/JavaScript, WASI and count
// families. Solo leaves keep graph destruction independent of shared aliases.
func init() {
	for _, name := range []string{"graph", "shared"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{
			"internal/oracle/testdata/release_fast_" + name + ".a", true, false,
		})
	}
}
