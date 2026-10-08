package oracle

func init() {
	for _, name := range []string{"binder", "checker", "builtin"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/void_value_" + name + ".a", true, false})
	}
}
