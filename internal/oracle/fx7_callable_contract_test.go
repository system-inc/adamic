package oracle

func init() {
	for _, path := range []string{
		"internal/oracle/testdata/review/agree/fxspptb_oct9_native_p18_cast_kind_no_field_types.a",
		"internal/oracle/testdata/review/agree/fxspptb_oct9_native_p20_callable_member_call_good.a",
		"internal/lower/testdata/fx7_wrong_aborts/wide.a",
		"internal/oracle/testdata/review/agree/fxspptb_oct9_native_p17_callable_call_first_half.a",
		"internal/oracle/testdata/review/agree/fxspptb_oct9_native_p26_callable_or_undefined.a",
		"internal/oracle/testdata/review/agree/fxspptb_oct9_native_p33_map_in_union.a",
		"internal/oracle/testdata/review/agree/fxspptb_oct9_native_p34_typed_array_in_union.a",
		"internal/oracle/testdata/review/agree/fxspptb_oct9_native_p35_class_in_union.a",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
}
