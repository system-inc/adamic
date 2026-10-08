package oracle

func init() {
	for _, fixture := range []struct {
		path    string
		lowers  bool
		checked bool
	}{
		{"internal/oracle/testdata/class_interface_literal.a", true, false},
		{"internal/oracle/testdata/class_interface_presence.a", true, false},
		{"internal/oracle/testdata/class_interface_checked.a", true, true},
		{"internal/oracle/testdata/class_interface_cast.a", true, true},
	} {
		fixtures = append(fixtures, fixture)
	}
}
