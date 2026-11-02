package oracle

// An object literal in a conditional's arm or an arrow's expression body is stored as objectLiteral builds it, held to
// Node (#6c4swvt G4). Mutants: the conditional, arrow or optionalValue arrow case dropped, and the fixture is refused.
func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/records_literal_branches.a", true, false})
}
