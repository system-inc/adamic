package oracle

func init() {
	for _, name := range []string{"scanner", "module_specifiers", "chain", "while", "order", "proofs", "throw"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/assignment_" + name + ".a", true, false})
	}
}
