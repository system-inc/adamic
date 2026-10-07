package oracle

// Keep the callback alias probes in the differential oracle and the counts table.
func init() {
	for _, name := range []string{"callback", "closure", "elements", "foreach", "map", "method"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/range_alias_" + name + ".a", true, false})
	}
}
