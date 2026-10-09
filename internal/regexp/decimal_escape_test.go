package regexp

import (
	"errors"
	"testing"
)

func TestUnicodeDecimalEscape(t *testing.T) {
	t.Parallel()
	// Node rejects the missing second capture in Unicode mode; legacy mode
	// accepts the same escape as an octal character escape.
	const pattern = `(a)\2`
	_, err := Parse(pattern, "u")
	var syntax *SyntaxError
	if !errors.As(err, &syntax) || syntax.Message != "invalid decimal escape" {
		t.Errorf("Parse(%q, %q) error=%v, want SyntaxError for invalid decimal escape", pattern, "u", err)
	}
	if _, err := Parse(pattern, ""); err != nil {
		t.Errorf("Parse(%q, %q) error=%v, want legacy escape accepted", pattern, "", err)
	}
}
