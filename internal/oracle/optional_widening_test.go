package oracle

// Register here so the relation unit does not change the oracle's central harness.
func init() {
	for _, name := range []string{"class", "fresh", "declared"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path: "internal/oracle/testdata/optional_widening_" + name + ".a", lowers: true})
	}
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/optional_widening_sound.a", true, false})
}
