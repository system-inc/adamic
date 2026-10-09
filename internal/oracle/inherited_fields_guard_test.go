package oracle

func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/inherited_fields_guard.a", true, false})
}
