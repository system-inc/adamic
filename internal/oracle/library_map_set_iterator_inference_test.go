package oracle

// A Set's or Map's iterator passed where Iterable<T> is wanted infers T as the element, so the
// copies and steps read present values (#onediag).
func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/library_map_set_iterator_inference.a", true, false})
}
