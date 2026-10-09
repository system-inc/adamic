package lower

import (
	"os"
	"testing"
)

func scalarUnionViewSource(t *testing.T, name string) string {
	t.Helper()
	source, err := os.ReadFile("testdata/scalar_union_views/" + name + ".a")
	if err != nil {
		t.Fatal(err)
	}
	return string(source)
}

func TestScalarUnionViewMaybeNumber(t *testing.T) {
	t.Parallel()
	// Only native has a packed presence bit separate from the numeric payload.
	lowersAndAgreesWithNodeNative(t, scalarUnionViewSource(t, "p54_n"))
}

func TestScalarUnionViewMaybeBoolean(t *testing.T) {
	t.Parallel()
	// Only native has a packed presence bit separate from the boolean payload.
	lowersAndAgreesWithNodeNative(t, scalarUnionViewSource(t, "p54_b"))
}

func TestScalarUnionViewMaybeString(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNode(t, scalarUnionViewSource(t, "p54_s"))
}

func TestScalarUnionViewMaybeNumberTransitions(t *testing.T) {
	t.Parallel()
	// Native must snapshot a present payload before the next source mutation.
	lowersAndAgreesWithNodeNative(t, scalarUnionViewSource(t, "p54_n")+"raw.value = 7; console.log(`${viewed.value}`); raw.value = undefined; console.log(`${viewed.value}`);")
}

func TestScalarUnionViewMaybeBooleanTransitions(t *testing.T) {
	t.Parallel()
	// Native must preserve true, false and absence as three distinct states.
	lowersAndAgreesWithNodeNative(t, scalarUnionViewSource(t, "p54_b")+"raw.value = true; console.log(`${viewed.value}`); raw.value = undefined; console.log(`${viewed.value}`);")
}
