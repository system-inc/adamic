package oracle

func init() {
	for _, path := range []string{
		"internal/oracle/testdata/library_array_alias_is_array.a",
		"internal/oracle/testdata/library_array_alias_every.a",
		"internal/oracle/testdata/library_array_alias_join.a",
		"internal/oracle/testdata/library_array_alias_includes.a",
		"internal/oracle/testdata/library_array_alias_search.a",
		"internal/oracle/testdata/library_array_map_oversized.a",
		"internal/oracle/testdata/library_array_shift_function.a",
		"internal/oracle/testdata/library_array_sort_never.a",
		"internal/oracle/testdata/library_array_to_sorted_never.a",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
}
