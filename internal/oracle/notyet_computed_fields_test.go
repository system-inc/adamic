package oracle

func init() {
	for _, name := range []string{"notyet_computed_fields.a", "notyet_computed_numeric_names.a", "notyet_computed_key_order.a"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/" + name, true, false})
	}
}
