package oracle

func init() {
	for _, path := range []string{"internal/oracle/testdata/class_set_property_union_runtime.a", "internal/oracle/testdata/class_set_property_union_static.a"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
}
