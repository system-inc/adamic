package lower

import (
	"strings"
	"testing"
)

func TestStep21NullishNarrowingAfterWritingCall(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNodeNative(t, "Step21NullishNarrowingAfterWritingCall")
}

func TestStep21NullishGuardDoesNotRestoreOtherNarrowings(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"let value: Error | undefined = new TypeError('first'); function change(): void { value=new Error('second'); } if (value instanceof TypeError) { change(); if (value !== undefined) console.log(value.name); }",
		"interface Box { v: string; } let value: Box | null | undefined={v:'before'}; function change(): void { value=null; } if (value !== undefined && value !== null) { change(); if (value !== undefined) console.log(value.v); }",
	} {
		_, err := lowerSource(t, source)
		if err == nil || !strings.Contains(err.Error(), "Adamic 0.1 refuses a narrowed read of value after change() can write it") {
			t.Fatalf("nullish guard restored an unsupported narrowing: %v", err)
		}
	}
}

func TestStep21FreshNumberObservation(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNodeNative(t, "Step21FreshNumberObservation")
}
