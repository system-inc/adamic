package configuration

import "testing"

// A severity that renders as a control character is the failure this method exists to prevent.
// Severity is an iota, so SeverityOff is 0, and any diagnostic that formats one with %s or %q
// without a String method prints '\x00' — illegible in exactly the output a reader consults to
// find out whether their config is being honored.
func TestSeverityRendersAsTheConfigSpellsIt(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		severity Severity
		expected string
	}{
		{SeverityOff, "off"},
		{SeverityWarn, "warn"},
		{SeverityError, "error"},
		{Severity(99), "unknown"},
	} {
		if actual := testCase.severity.String(); actual != testCase.expected {
			t.Errorf("severity %d rendered as %q, want %q", int(testCase.severity), actual, testCase.expected)
		}
	}
}
