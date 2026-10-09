package oracle

// Register the active review witnesses for ordinary oracle comparisons and counts.
func init() {
	for _, name := range []string{
		"fxspptb_oct9_native_p04_undefined_slot_write.a",
		"fxspptb_oct9_native_p06_undefined_other_name.a",
		"fxspptb_oct9_native_p08_null_array_write.a",
		"fxspptb_oct9_native_p59_map_null_write.a",
		"fxspptb_oct9_native_p75_unviewed_undefined_read.a",
		"fxspptb_oct9_native_p77_unviewed_undefined_read_scalar_view.a",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/review/agree/" + name, true, false})
	}
}
