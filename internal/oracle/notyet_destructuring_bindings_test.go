package oracle

func init() {
	for _, name := range []string{"notyet_destructuring_arrays.a", "notyet_destructuring_defaults.a", "notyet_destructuring_nested.a", "notyet_destructuring_slots.a"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/" + name, true, false})
	}
}
