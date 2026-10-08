package oracle

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{path: "internal/oracle/testdata/for_of_tuple.a", lowers: true})
}
