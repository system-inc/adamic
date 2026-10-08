package oracle

func init() {
	for _, name := range []string{"first", "last", "typed", "union"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/array_literal_empty_" + name + ".a", true, false})
	}
}
