package oracle

func init() {
	for _, name := range []string{"first", "middle", "last"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/accessor_spread_throw_" + name + ".a", true, false})
	}
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/plain_data_spread.a", true, false})
}
