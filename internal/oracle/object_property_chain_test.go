package oracle

func init() {
	for _, path := range []string{"internal/oracle/testdata/object_property_chain.a", "internal/oracle/testdata/object_property_chain_missing.a"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path, true, false})
	}
}
