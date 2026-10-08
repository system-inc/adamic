package oracle

func init() {
	for _, name := range []string{"block", "evaluator"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/overload_structural_" + name + ".a", true, false})
	}
}
