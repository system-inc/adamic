package oracle

func init() {
	for _, name := range []string{
		"nullable_references_strings.a",
		"nullable_references_generics.a",
		"nullable_references_json.a",
		"nullable_references_slots.a",
		"nullable_references_typeof_match.a",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/" + name, true, false})
	}
}
