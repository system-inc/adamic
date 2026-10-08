package oracle

func init() {
	for _, shape := range []string{"instance", "static", "mixed", "inherited", "optional", "generic"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/structural_statics_" + shape + ".a", true, false})
	}
}
