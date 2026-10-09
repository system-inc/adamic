package oracle

func init() {
	for _, path := range []string{
		"internal/oracle/testdata/overload_array_to_map.a",
		"internal/oracle/testdata/overload_inference_contracts.a",
		"internal/oracle/testdata/overload_array_to_multimap.a",
		"internal/oracle/testdata/overload_array_to_numeric_map.a",
	} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path, true, false})
	}
}
