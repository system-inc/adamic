package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestOptionalFunctionKeepsRequiredRefusal(t *testing.T) {
	_, err := lowerSource(t, "function test(callback: () => number): boolean { if (callback) { return callback() > 0; } return false; }")
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(err.Error(), "as a condition") {
		t.Fatalf("got %v, want required function condition refusal", err)
	}
}

func TestOptionalFunctionMixedAbsenceIsNotYet(t *testing.T) {
	_, err := lowerSource(t, "function test(callback: (() => number) | undefined | null): boolean { if (callback) { return true; } return false; }")
	var notYet *NotYet
	if !errors.As(err, &notYet) {
		t.Fatalf("got %v, want mixed null and undefined representation refusal", err)
	}
}

func TestOptionalFunctionNullableObjectIsNotYet(t *testing.T) {
	_, err := lowerSource(t, "function test(value: { size: number } | null): boolean { if (value) { return true; } return false; }")
	var notYet *NotYet
	if !errors.As(err, &notYet) {
		t.Fatalf("got %v, want unchanged nullable object representation refusal", err)
	}
}
