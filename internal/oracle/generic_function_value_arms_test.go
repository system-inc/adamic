package oracle

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/generic_function_value_arms.a", true, false})
}
