package oracle

// Keep fixture registration separate from the shared oracle harness.
func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/devirtualize.a", true, false})
}
