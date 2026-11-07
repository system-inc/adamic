package lower

import (
	"os"
	"path/filepath"
	"testing"
)

// fx6Source returns the fixture file's source and its extension, the mode it lowers in.
func fx6Source(test *testing.T, file string) (string, string) {
	test.Helper()
	source, err := os.ReadFile("testdata/fx6/" + file)
	if err != nil {
		test.Fatal(err)
	}
	return string(source), filepath.Ext(file)
}

func TestLowerViewsViewedProducerMutation(test *testing.T) {
	test.Parallel()
	source, extension := fx6Source(test, "p35.ts")
	lowersAndAgreesWithNodeAs(test, source, extension)
}

func TestLowerViewsUnrelatedMutableField(test *testing.T) {
	test.Parallel()
	source, extension := fx6Source(test, "p36.ts")
	lowersAndAgreesWithNodeAs(test, source, extension)
}

func TestLowerViewsUnrelatedUnionField(test *testing.T) {
	test.Parallel()
	source, extension := fx6Source(test, "p37.ts")
	lowersAndAgreesWithNodeAs(test, source, extension)
}

func TestLowerViewsUnrelatedOptionalStringField(test *testing.T) {
	test.Parallel()
	source, extension := fx6Source(test, "p38.ts")
	lowersAndAgreesWithNodeAs(test, source, extension)
}

func TestLowerViewsDestructuredScalarUnion(test *testing.T) {
	test.Parallel()
	source, extension := fx6Source(test, "p51.ts")
	lowersAndAgreesWithNodeAs(test, source, extension)
}

func TestLowerViewsDestructuredCallableUnion(test *testing.T) {
	test.Parallel()
	source, extension := fx6Source(test, "p53.ts")
	lowersAndAgreesWithNodeAs(test, source, extension)
}

func TestLowerViewsUnrelatedOptionalNumberField(test *testing.T) {
	test.Parallel()
	source, extension := fx6Source(test, "p38_number.ts")
	lowersAndAgreesWithNodeAs(test, source, extension)
}

func TestLowerViewsDestructuredObjectUnionLiteral(test *testing.T) {
	test.Parallel()
	// .ts, where the cast stays a checked view: in .a its tag test every Root passes is refused (#q97a1rr).
	lowersAndAgreesWithNodeAs(test, `interface Left { readonly text: string; }
interface Right { readonly count: number; }
interface Root { readonly kind: 'Holder'; }
interface Holder extends Root { readonly value: Left | Right; }
const raw = { kind: 'Holder' as const, value: { text: 'valid' } };
const base: Root = raw;
const view = base as Holder;
const { value } = view;
console.log(String(value !== undefined));`, ".ts")
}
