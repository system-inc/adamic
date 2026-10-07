package oracle

// Register these forms without changing the shared oracle harness.
func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/taste_optional_join.a", true, false})
	for _, path := range []string{"taste_truthiness.a", "taste_logical_assignment.a", "taste_comma.a", "taste_labels.a", "taste_void.a", "taste_exports/main.a", "taste_stage3_representations.a"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/" + path, true, false})
	}
}
