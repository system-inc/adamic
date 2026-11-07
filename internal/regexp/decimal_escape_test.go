package regexp

import (
	"errors"
	"testing"
)

func TestUnicodeDecimalEscape(test *testing.T) {
	test.Parallel()
	// Node rejects the missing second capture in Unicode mode; legacy mode
	// accepts the same escape as an octal character escape.
	const pattern = `(a)\2`
	_, parseError := Parse(pattern, "u")
	var syntax *SyntaxError
	if !errors.As(parseError, &syntax) || syntax.Message != "invalid decimal escape" {
		test.Errorf("Parse(%q, %q) error=%v, want SyntaxError for invalid decimal escape", pattern, "u", parseError)
	}
	if _, parseError := Parse(pattern, ""); parseError != nil {
		test.Errorf("Parse(%q, %q) error=%v, want legacy escape accepted", pattern, "", parseError)
	}
}
