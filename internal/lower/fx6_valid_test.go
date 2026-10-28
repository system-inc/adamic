package lower

import (
	"os"
	"path/filepath"
	"testing"
)

// fx6Source returns the fixture file's source and its extension, the mode it lowers in.
func fx6Source(t *testing.T, file string) (string, string) {
	t.Helper()
	source, err := os.ReadFile("testdata/fx6/" + file)
	if err != nil {
		t.Fatal(err)
	}
	return string(source), filepath.Ext(file)
}

func TestLowerViewsViewedProducerMutation(t *testing.T) {
	t.Parallel()
	source, extension := fx6Source(t, "p35.ts")
	lowersAndAgreesWithNodeAs(t, source, extension)
}

func TestLowerViewsUnrelatedMutableField(t *testing.T) {
	t.Parallel()
	source, extension := fx6Source(t, "p36.ts")
	lowersAndAgreesWithNodeAs(t, source, extension)
}

func TestLowerViewsUnrelatedUnionField(t *testing.T) {
	t.Parallel()
	source, extension := fx6Source(t, "p37.ts")
	lowersAndAgreesWithNodeAs(t, source, extension)
}

func TestLowerViewsUnrelatedOptionalStringField(t *testing.T) {
	t.Parallel()
	source, extension := fx6Source(t, "p38.ts")
	lowersAndAgreesWithNodeAs(t, source, extension)
}

func TestLowerViewsDestructuredScalarUnion(t *testing.T) {
	t.Parallel()
	source, extension := fx6Source(t, "p51.ts")
	lowersAndAgreesWithNodeAs(t, source, extension)
}

func TestLowerViewsDestructuredCallableUnion(t *testing.T) {
	t.Parallel()
	source, extension := fx6Source(t, "p53.ts")
	lowersAndAgreesWithNodeAs(t, source, extension)
}

func TestLowerViewsUnrelatedOptionalNumberField(t *testing.T) {
	t.Parallel()
	source, extension := fx6Source(t, "p38_number.ts")
	lowersAndAgreesWithNodeAs(t, source, extension)
}

func TestLowerViewsDestructuredObjectUnionLiteral(t *testing.T) {
	t.Parallel()
	// .ts, where the cast stays a checked view: in .a its tag test every Root passes is refused (#q97a1rr).
	lowersAndAgreesWithNodeAs(t, `interface Left { readonly text: string; }
interface Right { readonly count: number; }
interface Root { readonly kind: 'Holder'; }
interface Holder extends Root { readonly value: Left | Right; }
const raw = { kind: 'Holder' as const, value: { text: 'valid' } };
const base: Root = raw;
const view = base as Holder;
const { value } = view;
console.log(String(value !== undefined));`, ".ts")
}
