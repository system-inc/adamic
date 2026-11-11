package oracle

// A for-of's binding reassigned in its body reads its getter, held to Node (Language's must on
// #p2n17c4; internal/lower/forof_writes_test.go pins the flow graph's side).
func init() {
	for _, name := range []string{"r1_reassigned", "r2_pattern_reassigned"} {
		for _, extension := range []string{".a", ".ts"} {
			fixtures = append(fixtures, struct {
				path    string
				lowers  bool
				checked bool
			}{"internal/oracle/testdata/forof_writes/" + name + extension, true, false})
		}
	}
}
