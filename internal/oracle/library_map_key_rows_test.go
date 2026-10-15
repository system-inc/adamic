package oracle

func init() {
	added := []struct {
		path            string
		lowers, checked bool
	}{
		{"internal/oracle/testdata/map_set_keys/library_map_set_keys_objects.a", true, false},
		{"internal/oracle/testdata/map_set_keys/library_map_set_keys_classes.a", true, false},
		{"internal/oracle/testdata/map_set_keys/library_map_set_keys_numbers.a", true, false},
		{"internal/oracle/testdata/map_set_keys/library_map_set_keys_strings.a", true, false},
		{"internal/oracle/testdata/map_set_keys/library_map_set_keys_mixed.a", true, false},
		{"internal/oracle/testdata/map_set_keys/library_map_small_numbers.a", true, false},
		{"internal/oracle/testdata/map_set_keys/library_map_small_order.a", true, false},
		{"internal/oracle/testdata/map_set_keys/library_map_small_live.a", true, false},
		{"internal/oracle/testdata/map_set_keys/library_map_small_set_live.a", true, false},
		{"internal/oracle/testdata/map_set_keys/library_map_inferred.a", true, false},
		{"internal/oracle/testdata/map_set_keys/library_set_inferred.a", true, false},
	}
	for i, f := range fixtures {
		if f.path == "internal/oracle/testdata/library_map_set_group_by.a" {
			fixtures = append(fixtures[:i+1:i+1], append(added, fixtures[i+1:]...)...)
			return
		}
	}
	panic("missing Map groupBy registration")
}
