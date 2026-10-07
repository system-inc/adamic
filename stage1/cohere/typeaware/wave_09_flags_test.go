package typeaware

import (
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"testing"
)

func TestWave09PinnedTypeParameterFlag(t *testing.T) {
	t.Parallel()
	if checker.TypeFlagsTypeParameter != 524288 {
		t.Fatalf("type parameter flag changed: %d", checker.TypeFlagsTypeParameter)
	}
}
