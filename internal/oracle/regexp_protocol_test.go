package oracle

// Register this unit's fixtures without editing the shared oracle implementation.
func init() {
	for _, path := range []string{"internal/oracle/testdata/regexp_protocol.a", "internal/oracle/testdata/regexp_indices.a", "internal/oracle/testdata/regexp_replacement_callback.a", "internal/oracle/testdata/regexp_string.a"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/regexp_errors.a", true, false})
}
