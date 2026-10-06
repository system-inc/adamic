package oracle

// Register this slice separately from the shared fixture list.
func init() {
	for _, fixture := range []struct {
		path    string
		lowers  bool
		checked bool
	}{
		{"internal/oracle/testdata/library_object_iterator_own.a", true, false},
		{"internal/oracle/testdata/library_object_iterator_tag.a", true, false},
		{"internal/oracle/testdata/library_object_private_mangled.a", true, false},
	} {
		fixtures = append(fixtures, fixture)
	}
}
