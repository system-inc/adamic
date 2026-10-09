package lower

import (
	"os"
	"testing"
)

func TestFX7CheckedCastFieldTypes(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("testdata/fx7_wrong_aborts/p18.a")
	if err != nil {
		t.Fatal(err)
	}
	lowersAndAgreesWithNode(t, string(source))
	// Only native exposes the field representation metadata used by CheckedCast.
	lowersAndAgreesWithNodeNative(t, string(source))
}
