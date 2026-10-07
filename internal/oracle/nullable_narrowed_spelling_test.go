package oracle

// A call can restore an empty case after narrowing. Spelling must keep which empty case it is.
func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{path: "internal/oracle/testdata/nullable_narrowed_spelling.a", lowers: true})
}
