package oracle

func init() {
	for _, path := range []string{"internal/oracle/testdata/class_set_property_union.a", "internal/oracle/testdata/class_set_property_union_views.a"} {
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
	}{"internal/oracle/testdata/class_set_property_union_narrow.a", true, true})
}
