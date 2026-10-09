package lower

import (
	"os"
	"testing"
)

func fx7Source(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile("testdata/fx7/" + name + ".a")
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestFX7P04(t *testing.T) {
	t.Parallel()
	source := fx7Source(t, "p04")
	lowersAndAgreesWithNode(t, source)
	// Native exposes field storage checks and semantic nullish tags.
	lowersAndAgreesWithNodeNative(t, source)
}

func TestFX7P06(t *testing.T) {
	t.Parallel()
	source := fx7Source(t, "p06")
	lowersAndAgreesWithNode(t, source)
	// Native exposes field storage checks and semantic nullish tags.
	lowersAndAgreesWithNodeNative(t, source)
}

func TestFX7P08(t *testing.T) {
	t.Parallel()
	source := fx7Source(t, "p08")
	lowersAndAgreesWithNode(t, source)
	// Native exposes field storage checks and semantic nullish tags.
	lowersAndAgreesWithNodeNative(t, source)
}

func TestFX7P59(t *testing.T) {
	t.Parallel()
	source := fx7Source(t, "p59")
	lowersAndAgreesWithNode(t, source)
	// Native exposes field storage checks and semantic nullish tags.
	lowersAndAgreesWithNodeNative(t, source)
}

func TestFX7P75(t *testing.T) {
	t.Parallel()
	source := fx7Source(t, "p75")
	lowersAndAgreesWithNode(t, source)
	// Native exposes field storage checks and semantic nullish tags.
	lowersAndAgreesWithNodeNative(t, source)
}

func TestFX7P77(t *testing.T) {
	t.Parallel()
	source := fx7Source(t, "p77")
	lowersAndAgreesWithNode(t, source)
	// Native exposes field storage checks and semantic nullish tags.
	lowersAndAgreesWithNodeNative(t, source)
}
