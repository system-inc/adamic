package oracle

func init() {
	for _, name := range []string{
		"binary_logical_value_value.a", "binary_logical_value_boolean.a",
		"binary_logical_boolean_value.a", "binary_logical_number_number.a",
		"binary_logical_number_boolean.a", "binary_logical_union.a",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/" + name, true, false})
	}
}
