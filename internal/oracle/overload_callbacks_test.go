package oracle

func init() {
	for _, path := range []string{
		"internal/oracle/testdata/overload_ancestor_directory.a",
		"internal/oracle/testdata/overload_leading_comment_range.a",
		"internal/oracle/testdata/overload_trailing_comment_range.a",
		"internal/oracle/testdata/overload_original_node.a",
	} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path, true, false})
	}
}
