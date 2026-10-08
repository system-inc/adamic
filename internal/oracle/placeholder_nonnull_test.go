package oracle

func init() {
	for _, name := range []string{"scanner", "local", "field"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/placeholder_nonnull_" + name + ".a", true, false})
	}
}
