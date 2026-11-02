package oracle

// A record enters a union naming its own type as that member, and a read on Options | Printer dispatches on the value, held
// to Node (#6c4swvt G2). Mutants: either union branch's member check dropped, and the fixture is refused at its entry.
func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/records_union_member_entry.a", true, false})
}
