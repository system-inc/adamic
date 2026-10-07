package oracle

func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/census_small_rest.a", true, false}, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/census_small_boolean.a", true, false}, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/census_small_overload.a", true, false})
}
