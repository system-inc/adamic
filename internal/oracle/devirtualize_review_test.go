package oracle

// Keep review fixtures in the ordinary three-way oracle without editing its registry.
func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/devirt_borrow_doc_claim.a", true, false})
}
