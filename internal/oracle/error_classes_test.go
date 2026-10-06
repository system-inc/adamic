package oracle

func init() {
	for _, path := range []string{"internal/oracle/testdata/error_classes.a", "internal/oracle/testdata/error_classes_uncaught.a", "internal/oracle/testdata/error_classes_uncaught_name.a", "internal/oracle/testdata/error_classes_uncaught_empty.a"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
}
