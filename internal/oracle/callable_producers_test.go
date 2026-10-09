package oracle

func init() {
	for _, name := range []string{"p21", "p60", "p61"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/lower/testdata/callable_producers/" + name + ".a", true, false})
	}
}
