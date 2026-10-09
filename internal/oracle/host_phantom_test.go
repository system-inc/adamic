package oracle

func init() {
	for _, name := range []string{"phantom_brands.a", "phantom_undefined.a", "phantom_catch.a", "host_phantom_overload_primitive.a", "host_phantom_optional_array.a"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/" + name, true, false})
	}
}
