package oracle

func init() {
	for _, path := range []string{"internal/oracle/testdata/enums_flags.a", "internal/oracle/testdata/enums_flags_never_default.a", "internal/oracle/testdata/enums_flags_modules/main.a"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
}
