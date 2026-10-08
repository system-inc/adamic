package oracle

import "testing"

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/notyet_signature_optional_escaped_string.a", true, false})
}

func TestOptionalEscapedStringReturnMutant(t *testing.T) {
	escapedStringReturnMutant(t, "notyet_signature_optional_escaped_string.a", "escapedLookup", true)
}
