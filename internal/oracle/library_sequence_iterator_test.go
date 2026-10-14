package oracle

func init() {
	added := []struct {
		path            string
		lowers, checked bool
	}{
		{"internal/oracle/testdata/sequence_iterators/library_iterator_string_next.a", true, false},
		{"internal/oracle/testdata/sequence_iterators/library_iterator_map_set_live.a", true, false},
		{"internal/oracle/testdata/sequence_iterators/library_iterator_tsc_helpers.a", true, false},
		{"internal/oracle/testdata/sequence_iterators/library_iterator_identity_close.a", true, false},
		{"internal/oracle/testdata/sequence_iterators/library_iterator_to_array.a", true, false},
		{"internal/oracle/testdata/sequence_iterators/library_iterator_tags.a", true, false},
		{"internal/oracle/testdata/sequence_iterators/903f25b_iterator_tag.a", true, false},
		{"internal/oracle/testdata/sequence_iterators/library_overlay_iterable.a", true, false},
		{"internal/oracle/testdata/sequence_iterators/library_overlay_callback.a", true, false},
		{"internal/oracle/testdata/sequence_iterators/library_overlay_entries.a", true, false},
		{"internal/oracle/testdata/sequence_iterators/library_overlay_nested_entries.a", true, false},
		{"internal/oracle/testdata/sequence_iterators/library_overlay_done.a", true, false},
		{"internal/oracle/testdata/sequence_iterators/runtime_tuple_map_entries.a", true, false},
		{"internal/oracle/testdata/sequence_iterators/runtime_tuple_array_entries.a", true, false},
		{"internal/oracle/testdata/sequence_iterators/runtime_tuple_object_entries.a", true, false},
		{"internal/oracle/testdata/sequence_iterators/runtime_tuple_object_entries_checked.a", true, false},
		{"internal/oracle/testdata/sequence_iterators/library_iterator_array_checked.a", true, false},
	}
	for i, f := range fixtures {
		if f.path == "internal/oracle/testdata/library_array_join.a" {
			fixtures = append(fixtures[:i+1:i+1], append(added, fixtures[i+1:]...)...)
			return
		}
	}
	panic("missing Array join registration")
}
