package oracle

func init() {
	for _, name := range []string{"function", "class"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/anonymous_default_refused/" + name + ".a", false, false})
	}
}
