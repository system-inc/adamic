package oracle

func init() {
	for _, path := range []string{
		"internal/oracle/testdata/optional_call_locals.a",
		"internal/oracle/testdata/optional_call_exports.a",
		"internal/oracle/testdata/optional_call_values.a",
		"internal/oracle/testdata/optional_call_receiver.a",
		"internal/oracle/testdata/optional_call_collections.a",
		"internal/oracle/testdata/optional_call_chain.a",
		"internal/oracle/testdata/optional_call_result_chain.a",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
}
