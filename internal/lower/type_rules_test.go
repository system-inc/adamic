package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestCohereNominalInvocation(t *testing.T) {
	_, err := lowerSource(t, `class A { value(): number { return 1; } } const value: A = { value: () => 2 }; console.log('accepted');`)
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(err.Error(), "adamic/nominal-class") {
		t.Fatalf("want cohere nominal refusal, got %v", err)
	}
}
