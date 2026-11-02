package oracle

// A Pick of a record is stored as that record, compared as it both ways, held to Node (#6c4swvt G1). Mutants:
// sameRecordStorage's or optionalWidened's substitution of a Pick by its base dropped, and the fixture is refused at a view.
func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/records_pick_views.a", true, false})
}
