package oracle

func init() {
	for _, path := range []string{
		"internal/fresh/testdata/regexp_tree.ts",
		"internal/oracle/testdata/regexp_cycle_fields.a",
		"internal/oracle/testdata/regexp_cycle_collections.a",
		"internal/oracle/testdata/regexp_cycle_closures.a",
		"internal/oracle/testdata/regexp_cycle_weak.a",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
}
