package oracle

// Keep this unit's registrations outside the shared oracle driver.
func init() {
	for _, path := range []string{
		"stage3/drivers/scanner/probes/postfix-call-argument.a",
		"stage3/drivers/scanner/probes/prefix-call-argument.a",
		"internal/oracle/testdata/increment_values.a",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
}
