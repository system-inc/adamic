package oracle

func init() {
	for _, name := range []string{"scanner", "composition", "value"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/records_type_only_" + name + ".a", true, false})
	}
}
