package oracle

func init() {
	for _, name := range []string{
		"nullable_references_strings.a",
		"nullable_coverage_string.a",
		"nullable_coverage_calls.a",
		"nullable_coverage_object.a",
		"nullable_coverage_array.a",
		"nullable_coverage_map.a",
		"nullable_coverage_class.a",
		"nullable_coverage_nested_generics.a",
		"nullable_coverage_json_arguments.a",

		"nullable_references_generics.a",
		"nullable_references_json.a",
		"nullable_references_slots.a",
		"nullable_references_typeof_match.a",
		"nullable_references_map_narrowed.a",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/" + name, true, false})
	}
}
