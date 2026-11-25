package oracle

func init() {
	for _, path := range []string{
		"internal/oracle/testdata/generic_void_readback/config.ts",
		"internal/oracle/testdata/generic_void_readback/contextual.ts",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
}
