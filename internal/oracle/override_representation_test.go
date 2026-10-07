package oracle

// Supported neighbors run against source Node, both backends, sanitizers and the leak check.
func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/override_same_representation.a", true, false})
}
