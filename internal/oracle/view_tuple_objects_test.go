package oracle

func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/lower/testdata/tuple_object_views/p70_copy.a", true, false})
}
