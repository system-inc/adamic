package oracle

func init() {
	for _, name := range []string{"uint16", "class_value"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/new_expression_" + name + ".a", true, false})
	}
}
