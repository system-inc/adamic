package oracle

func init() {
	for _, path := range []string{
		"internal/oracle/testdata/047cb0d_n_element_plain.a",
		"internal/oracle/testdata/047cb0d_n_element.a",
		"internal/oracle/testdata/047cb0d_n_element_method.a",
		"internal/oracle/testdata/047cb0d_n_param.a",
		"internal/oracle/testdata/047cb0d_n_numparam.a",
		"internal/oracle/testdata/047cb0d_n_coalesce.a",
		"internal/oracle/testdata/047cb0d_n_typeof.a",
		"internal/oracle/testdata/047cb0d_n_template.a",
		"internal/oracle/testdata/047cb0d_n_optional.a",
		"internal/oracle/testdata/047cb0d_n_arrayindex.a",
		"internal/oracle/testdata/047cb0d_n_paren.a",
		"internal/oracle/testdata/047cb0d_n_conditional.a",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
}
