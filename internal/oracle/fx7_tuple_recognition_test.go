package oracle

func init() {
	for _, path := range []string{
		"internal/oracle/testdata/review/agree/fxspptb_oct9_views_p68_object_view_over_tuple.a",
		"internal/oracle/testdata/review/agree/fxspptb_oct9_views_p69_tuple_view_over_tuple_union.a",
		"internal/lower/testdata/fx7_tuple_recognition/p69.a",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
}
