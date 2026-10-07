package oracle

func init() {
	for _, name := range []string{"79", "193", "variants", "required", "propagation"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/optional_after_call_" + name + ".a", true, false})
	}
	// Computed optional indexing is an existing lowering gap, separate from the narrowing check.
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/optional_after_call_gaps/computed.a", false, false})
}
