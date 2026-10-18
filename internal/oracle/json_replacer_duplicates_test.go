package oracle

func init() {
	for _, name := range []string{"records_json_replacer_duplicates", "json_replacer_duplicates_object"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/" + name + ".a", true, false})
	}
}
