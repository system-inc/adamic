package oracle

func init() {
	for _, name := range []string{"notyet_destructuring_lengths.a", "notyet_map_iterable_pairs.a", "notyet_iterated_optional_arrays.a", "notyet_tuple_member_assignment.a"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/" + name, true, false})
	}
}
