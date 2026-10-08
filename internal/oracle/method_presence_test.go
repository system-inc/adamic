package oracle

func init() {
	for _, path := range []string{"internal/oracle/testdata/method_presence.a", "internal/oracle/testdata/method_presence_wide_union.a"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
}
