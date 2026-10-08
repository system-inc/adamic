package oracle

func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/object_small_tuple_pop.a", true, false})
}
