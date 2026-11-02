package oracle

// new Map from an array of pairs or undefined is new Map(pairs ?? []), held to Node (#assd1cf). Mutant:
// collections.go's optional branch dropped, and the fixture refuses at its first new Map, so it goes red.
func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/library_map_from_optional_pairs.a", true, false})
}
