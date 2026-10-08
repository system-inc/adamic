package oracle

func init() {
	for _, name := range []string{"sites", "names", "iterable"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path: "internal/oracle/testdata/for_of_object_destructure_" + name + ".a", lowers: true})
	}
}
