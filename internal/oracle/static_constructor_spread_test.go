package oracle

func init() {
	for _, name := range []string{"static_constructor_spread_gap", "static_constructor_spread_owned"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/" + name + ".a", true, false})
	}
}
