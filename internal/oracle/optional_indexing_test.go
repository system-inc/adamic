package oracle

func init() {
	for _, path := range []string{"internal/oracle/testdata/optional_indexing_array.a", "internal/oracle/testdata/optional_indexing_string.a", "internal/oracle/testdata/optional_indexing_string_gap.a", "internal/oracle/testdata/optional_indexing_typed_array.a", "internal/oracle/testdata/optional_indexing_map.a", "internal/oracle/testdata/optional_indexing_chain.a", "internal/oracle/testdata/optional_aside_throw.a"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path, true, false})
	}
}
