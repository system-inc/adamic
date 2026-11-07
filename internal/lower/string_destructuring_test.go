package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestStringDestructuringMemberDiagnostic(test *testing.T) {
	test.Parallel()
	_, lowerSourceError := lowerSource(test, "const { charCodeAt } = 'text'; console.log('unreachable');")
	var stopped *NotYet
	if !errors.As(lowerSourceError, &stopped) ||
		!strings.Contains(stopped.Where, "main.a:1:9") ||
		!strings.Contains(stopped.What, "string member 'charCodeAt'") ||
		!strings.Contains(stopped.What, "call methods on the saved string receiver") {
		test.Fatalf("want string member refusal with path and fix, got %v", lowerSourceError)
	}
}
