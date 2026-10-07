package oracle

func init() {
	for _, path := range []string{
		"internal/oracle/testdata/coverage_relation_fresh_copies.a",
		"internal/oracle/testdata/coverage_guard_imports.a",
		"internal/oracle/testdata/coverage_guard_imports_helper.a",
		"internal/oracle/testdata/coverage_guard_match_null.a",
		"internal/oracle/testdata/coverage_guard_tags.a",

		"internal/oracle/testdata/coverage_guard_nodes.a",
		"internal/oracle/testdata/coverage_guard_literals.a",
		"internal/oracle/testdata/coverage_assertion_forms.a",
		"internal/oracle/testdata/coverage_relation_containers.a",
		"internal/oracle/testdata/coverage_relation_views.a",
		"internal/oracle/testdata/coverage_relation_array_literals.a",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
}
