package oracle

// flag ? {} : undefined in a record slot is the record its literal is built as, or undefined, held to Node (#6c4swvt G4).
// Mutant: the record case dropped, and the fixture stops at a conditional whose branches have different types.
func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/records_literal_branch_undefined.a", true, false})
}
