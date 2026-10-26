package lower

import (
	"os"
	"testing"
)

func fx6Source(t *testing.T, name string) string {
	t.Helper()
	source, err := os.ReadFile("testdata/fx6/" + name + ".ts")
	if err != nil {
		t.Fatal(err)
	}
	return string(source)
}

func TestLowerViewsViewedProducerMutation(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNodeChecked(t, fx6Source(t, "p35"))
}

func TestLowerViewsUnrelatedMutableField(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNodeChecked(t, fx6Source(t, "p36"))
}

func TestLowerViewsUnrelatedUnionField(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNodeChecked(t, fx6Source(t, "p37"))
}

func TestLowerViewsUnrelatedOptionalStringField(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNodeChecked(t, fx6Source(t, "p38"))
}

func TestLowerViewsDestructuredScalarUnion(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNodeChecked(t, fx6Source(t, "p51"))
}

func TestLowerViewsDestructuredCallableUnion(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNodeChecked(t, fx6Source(t, "p53"))
}

func TestLowerViewsUnrelatedOptionalNumberField(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNodeChecked(t, fx6Source(t, "p38_number"))
}

func TestLowerViewsDestructuredObjectUnionLiteral(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNode(t, `interface Left { readonly text: string; }
interface Right { readonly count: number; }
interface Root { readonly kind: 'Holder'; }
interface Holder extends Root { readonly value: Left | Right; }
const raw = { kind: 'Holder' as const, value: { text: 'valid' } };
const base: Root = raw;
const view = base as Holder;
const { value } = view;
console.log(String(value !== undefined));`)
}
