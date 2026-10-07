package oracle

// Register the consumed argument oracle without changing the central harness.
func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{path: "internal/oracle/testdata/node_options_widening.a", lowers: true})
}
