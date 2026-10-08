package lower

import (
	"errors"
	"testing"
)

func TestClockGenericReturnsT06UnsupportedIntersections(t *testing.T) {
	for _, extra := range []string{
		`readonly children: readonly number[]`,
		`readonly children: readonly string[]; [key: string]: unknown`,
		`readonly children: readonly string[]; (): string`,
		`readonly children: readonly string[]; new(): object`,
		`readonly children: Map<string, string>`,
	} {
		t.Run(extra, func(t *testing.T) {
			source := `interface E { readonly text: string; ` + extra + `; } interface N { readonly kind: "number"; readonly value: number; } function member(): E & N { throw "unused"; }`
			_, err := lowerSource(t, source)
			var stop *NotYet
			var refused *Refused
			if !errors.As(err, &stop) && !errors.As(err, &refused) {
				t.Fatalf("want unsupported signature to be rejected, got %v", err)
			}
		})
	}
}
