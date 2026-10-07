package oracle

func init() {
	for _, path := range []string{"internal/oracle/testdata/library_method_values.a", "internal/oracle/testdata/library_method_values_dead_zone.a"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
}
