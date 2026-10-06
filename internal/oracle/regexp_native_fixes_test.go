package oracle

func init() {
	for _, name := range []string{"class_octal", "class_escapes", "quantifier_bounds"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{
			"internal/oracle/testdata/regexp_native_" + name + ".a", true, false,
		})
	}
}
