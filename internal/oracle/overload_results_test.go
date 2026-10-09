package oracle

func init() {
	for _, name := range []string{"transform", "evaluate", "binding", "parameter", "scalar"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/overload_results_" + name + ".a", true, false})
	}
}
