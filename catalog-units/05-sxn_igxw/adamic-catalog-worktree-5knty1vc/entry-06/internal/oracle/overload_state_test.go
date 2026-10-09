package oracle

func init() {
	for _, path := range []string{
		"internal/oracle/testdata/overload_mutate_map.a",
		"internal/oracle/testdata/overload_mutate_map_skipping_new.a",
		"internal/oracle/testdata/overload_resolve_type_names.a",
		"internal/oracle/testdata/overload_sort_deduplicate.a",
	} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path, true, false})
	}
}
