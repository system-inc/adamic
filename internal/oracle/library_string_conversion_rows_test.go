package oracle

func init() {
	added := []struct {
		path            string
		lowers, checked bool
	}{
		{"internal/oracle/testdata/string_conversion/collections.a", true, false},
		{"internal/oracle/testdata/string_conversion/template.a", true, false},
	}
	for i, f := range fixtures {
		if f.path == "internal/oracle/testdata/library_string_existing.a" {
			fixtures = append(fixtures[:i+1:i+1], append(added, fixtures[i+1:]...)...)
			return
		}
	}
	panic("missing String existing registration")
}
