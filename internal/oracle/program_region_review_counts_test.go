package oracle

// b31444fa accepted these slice 6 graph programs into review/agree.
// Register them with the ordinary fixture inventory so Linux counts cover them too.
func init() {
	for _, name := range []string{
		"cwd.a",
		"fxspptb_aee89b2_alias_fill.a",
		"fxspptb_aee89b2_alias_mapset.a",
		"fxspptb_aee89b2_alias_reverse.a",
		"fxspptb_aee89b2_alias_setadd.a",
		"fxspptb_aee89b2_alias_sort.a",
		"fxspptb_aee89b2_clobber_link_loop.a",
		"fxspptb_aee89b2_escaped_before.a",
		"fxspptb_aee89b2_map_entries_pattern.a",
		"fxspptb_aee89b2_spread_outside.a",
		"fxspptb_aee89b2_throw_keeps_old.a",
		"fxspptb_devirt_fresh_method_keeps_argument.a",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{
			"internal/oracle/testdata/review/agree/" + name, true, false,
		})
	}
}
