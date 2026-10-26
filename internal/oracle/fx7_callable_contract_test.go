package oracle

func init() {
	for _, path := range []string{
		"internal/oracle/testdata/review/agree/fxspptb_oct9_native_p18_cast_kind_no_field_types.ts",
		"internal/oracle/testdata/review/agree/fxspptb_oct9_native_p20_callable_member_call_good.ts",
		"internal/lower/testdata/fx7_wrong_aborts/wide.a",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
}
