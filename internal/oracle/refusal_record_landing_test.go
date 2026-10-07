package oracle

// The refusal audit predates own-key record storage. Execute its formerly blocked
// programs against Node rather than retaining a blanket refusal expectation.
func init() {
	for _, name := range []string{
		"record.a", "record-alias.a", "record-generic-alias.a",
		"record-dynamic-read.a", "record-dynamic-write.a", "record-element-read.a",
		"record-values.a", "record-entries.a", "observe-record.a",
	} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/lower/testdata/refusal_pass/" + name, true, false})
	}
}
