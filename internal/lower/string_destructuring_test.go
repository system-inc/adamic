package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestStringDestructuringMemberDiagnostic(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, "const { charCodeAt } = 'text'; console.log('unreachable');")
	var stopped *NotYet
	if !errors.As(err, &stopped) || !strings.Contains(stopped.Where, "main.a:1:9") || !strings.Contains(stopped.What, "string member 'charCodeAt'") || !strings.Contains(stopped.What, "call methods on the saved string receiver") {
		t.Fatalf("want string member refusal with path and fix, got %v", err)
	}
}
