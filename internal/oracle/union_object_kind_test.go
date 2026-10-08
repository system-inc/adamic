package oracle

func init() {
	for _, fixture := range []struct {
		path            string
		lowers, checked bool
	}{
		{"internal/oracle/testdata/union_object_kind.a", true, false},
		{"internal/oracle/testdata/reland_refused/narrowed_union_object_tag.a", true, false},
		{"internal/oracle/testdata/union_object_kind_stale.a", true, true},
	} {
		fixtures = append(fixtures, fixture)
	}
}
