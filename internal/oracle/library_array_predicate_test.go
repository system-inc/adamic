package oracle

// Insert beside the existing findLast control so sibling count additions stay
// on distinct rows and merge without changing their registration order.
func init() {
	added := []struct {
		path            string
		lowers, checked bool
	}{
		{"internal/oracle/testdata/array_predicates/library_array_predicate.a", true, false},
		{"internal/oracle/testdata/array_predicates/047cb0d_f_maybe.a", true, false},
		{"internal/oracle/testdata/array_predicates/047cb0d_f_maybe_ref.a", true, false},
		{"internal/oracle/testdata/array_predicates/57f2d04_find_last_shrinks.a", true, true},
		{"internal/oracle/testdata/array_predicates/57f2d04_find_last_shrinks2.a", true, true},
		{"internal/oracle/testdata/array_predicates/search_shrink_boolean.a", true, false},
		{"internal/oracle/testdata/array_predicates/search_shrink_default.a", true, false},
		{"internal/oracle/testdata/array_predicates/search_shrink_findIndex_checked.a", true, true},
		{"internal/oracle/testdata/array_predicates/search_shrink_findLastIndex_maybe.a", true, false},
		{"internal/oracle/testdata/array_predicates/search_shrink_findLast_maybe.a", true, false},
		{"internal/oracle/testdata/array_predicates/search_shrink_find_checked.a", true, true},
		{"internal/oracle/testdata/array_predicates/search_shrink_find_maybe.a", true, false},
		{"internal/oracle/testdata/array_predicates/search_shrink_found.a", true, false},
		{"internal/oracle/testdata/array_predicates/search_shrink_omitted.a", true, false},
		{"internal/oracle/testdata/array_predicates/search_shrink_unknown.a", true, false},
	}
	for i, fixture := range fixtures {
		if fixture.path == "internal/oracle/testdata/library_array_find_last.a" {
			fixtures = append(fixtures[:i+1:i+1], append(added, fixtures[i+1:]...)...)
			return
		}
	}
	panic("missing findLast fixture registration")
}
