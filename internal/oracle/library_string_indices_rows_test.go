package oracle

func init() {
	added := []struct {
		path            string
		lowers, checked bool
	}{
		{"internal/oracle/testdata/string_indices/library_string_primitive_identity.a", true, false},
		{"internal/oracle/testdata/string_indices/library_string_raw_empty.a", true, false},
		{"internal/oracle/testdata/string_indices/library_string_affixes.a", true, false},
		{"internal/oracle/testdata/string_indices/library_string_generic_pad_fill.a", true, false},
		{"internal/oracle/testdata/string_indices/library_string_generic_pad_fill_call.a", true, false},
		{"internal/oracle/testdata/string_indices/library_string_generic_receiver.a", true, false},
		{"internal/oracle/testdata/string_indices/library_string_generic_receiver_charat.a", true, false},
		{"internal/oracle/testdata/string_indices/library_string_generic_receiver_number.a", true, false},
		{"internal/oracle/testdata/string_indices/library_string_null_receiver.a", true, false},
		{"internal/oracle/testdata/string_indices/library_string_tail_repeat.a", true, false},
		{"internal/oracle/testdata/string_indices/library_string_typeof_conversion.a", true, false},
		{"internal/oracle/testdata/string_indices/library_tsc_census_substr.a", true, false},
		{"internal/oracle/testdata/string_indices/string_lastindexof_position.a", true, false},
	}
	for i, f := range fixtures {
		if f.path == "internal/oracle/testdata/library_string_raw.a" {
			fixtures = append(fixtures[:i+1:i+1], append(added, fixtures[i+1:]...)...)
			return
		}
	}
	panic("missing String raw registration")
}
