package oracle

// These are the executable string-console profiles of the project-reference
// fixtures. The stage3 verifier copies their sources to project-owned .ts roots
// and holds that loading path to stock tsc before testing both backends.
func init() {
	for _, path := range []string{
		"stage3/project-references-source/fixture/two/app/main.a",
		"stage3/project-references-source/fixture/chain/app/main-print.a",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
}
