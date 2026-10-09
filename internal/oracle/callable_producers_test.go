package oracle

func init() {
	for _, name := range []string{"p21", "p60", "p61", "p23", "p22_adapted", "p59_adapted", "p62_adapted", "p135_tagged_unread"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/lower/testdata/callable_producers/" + name + ".a", true, false})
	}
}
