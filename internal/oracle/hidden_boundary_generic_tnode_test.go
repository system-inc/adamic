package oracle

func init() {
	for _, path := range []string{"internal/oracle/testdata/hidden_boundary_generic_tnode.a", "internal/oracle/testdata/hidden_boundary_generic_tnode_constraints.a"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/hidden_boundary_generic_tnode_value.a", false, false})
}

func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/hidden_boundary_generic_tnode_mutation.a", false, false})
}
