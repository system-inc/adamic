package oracle

func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/notyet_element_access/fields.a", false, false})
}
