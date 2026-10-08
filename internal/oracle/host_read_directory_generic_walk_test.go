package oracle

// The pinned library host fixture is the integration oracle for the generic
// function value and subsequent lowering walk. Keep its source unchanged.
func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"stage3/fixtures/host/25_readDirectory.a", true, false})
}
