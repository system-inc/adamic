package oracle

func init() {
	for _, name := range []string{"scanner", "module_specifiers", "chain", "while", "order", "proofs", "throw"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/assignment_" + name + ".a", true, false})
	}
	// Acceptance programs from codex/step12-assign-fixtures e98e49ad.
	for _, name := range []string{"01_scanner_keyword", "02_if_narrowing", "03_return_defined", "04_chained", "05_compound", "06_wider_target"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"stage3/fixtures/assignment-proofs/" + name + ".a", true, false})
	}
}
