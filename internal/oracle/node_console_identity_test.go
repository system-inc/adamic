package oracle

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"stage3/fixtures/host/console_identity.a", true, false})
}
