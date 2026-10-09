package lower

import (
	"os"
	"path/filepath"
	"testing"
)

func arrayNarrowingSource(t *testing.T, name string) string {
	t.Helper()
	source, err := os.ReadFile(filepath.Join("testdata", "array_narrowing", name+".a"))
	if err != nil {
		t.Fatal(err)
	}
	return string(source)
}

func TestArrayNarrowingEveryAgrees(t *testing.T) {
	t.Parallel()
	source := arrayNarrowingSource(t, "every")
	lowersAndAgreesWithNode(t, source)
	// Only native exposes the stored union slots and checked scalar unboxing.
	lowersAndAgreesWithNodeNative(t, source)
}

func TestArrayNarrowingInferredAgrees(t *testing.T) {
	t.Parallel()
	source := arrayNarrowingSource(t, "inferred")
	lowersAndAgreesWithNode(t, source)
	// Only native exposes the stored union slots and checked scalar unboxing.
	lowersAndAgreesWithNodeNative(t, source)
}

func TestArrayNarrowingReadonlyAgrees(t *testing.T) {
	t.Parallel()
	source := arrayNarrowingSource(t, "readonly")
	lowersAndAgreesWithNode(t, source)
	// Only native exposes the stored union slots and checked scalar unboxing.
	lowersAndAgreesWithNodeNative(t, source)
}

func TestArrayNarrowingEarlyReturnAgrees(t *testing.T) {
	t.Parallel()
	source := arrayNarrowingSource(t, "early_return")
	lowersAndAgreesWithNode(t, source)
	// Only native exposes the stored union slots and checked scalar unboxing.
	lowersAndAgreesWithNodeNative(t, source)
}

func TestArrayNarrowingPropertyAgrees(t *testing.T) {
	t.Parallel()
	source := arrayNarrowingSource(t, "property")
	lowersAndAgreesWithNode(t, source)
	// Only native exposes the stored union slots and checked scalar unboxing.
	lowersAndAgreesWithNodeNative(t, source)
}

func TestArrayNarrowingForwardAgrees(t *testing.T) {
	t.Parallel()
	source := arrayNarrowingSource(t, "forward")
	lowersAndAgreesWithNode(t, source)
	// Only native exposes the stored union slots and checked scalar unboxing.
	lowersAndAgreesWithNodeNative(t, source)
}

func TestArrayNarrowingCatAgrees(t *testing.T) {
	t.Parallel()
	source := arrayNarrowingSource(t, "cat")
	lowersAndAgreesWithNode(t, source)
	// Only native exposes the stored union slots and checked scalar unboxing.
	lowersAndAgreesWithNodeNative(t, source)
}

func TestArrayNarrowingFindAgrees(t *testing.T) {
	t.Parallel()
	source := arrayNarrowingSource(t, "find")
	lowersAndAgreesWithNode(t, source)
	// Only native exposes the stored union slots and checked scalar unboxing.
	lowersAndAgreesWithNodeNative(t, source)
}

func TestArrayNarrowingFindLastAgrees(t *testing.T) {
	t.Parallel()
	source := arrayNarrowingSource(t, "find_last")
	lowersAndAgreesWithNode(t, source)
	// Only native exposes the stored union slots and checked scalar unboxing.
	lowersAndAgreesWithNodeNative(t, source)
}

func TestArrayNarrowingOptionalNumberAgrees(t *testing.T) {
	t.Parallel()
	source := arrayNarrowingSource(t, "optional_number")
	lowersAndAgreesWithNode(t, source)
	// Only native exposes whether narrowed elements use their stored layout.
	lowersAndAgreesWithNodeNative(t, source)
}

func TestArrayNarrowingOptionalStringAgrees(t *testing.T) {
	t.Parallel()
	source := arrayNarrowingSource(t, "optional_string")
	lowersAndAgreesWithNode(t, source)
	// Only native exposes whether narrowed elements use their stored layout.
	lowersAndAgreesWithNodeNative(t, source)
}

func TestArrayNarrowingSomeNegatedAgrees(t *testing.T) {
	t.Parallel()
	source := arrayNarrowingSource(t, "some_negated")
	lowersAndAgreesWithNode(t, source)
	// Only native exposes whether narrowed elements use their stored layout.
	lowersAndAgreesWithNodeNative(t, source)
}

func TestArrayNarrowingFindLastIndexAgrees(t *testing.T) {
	t.Parallel()
	source := arrayNarrowingSource(t, "find_last_index")
	lowersAndAgreesWithNode(t, source)
	// Only native exposes whether narrowed elements use their stored layout.
	lowersAndAgreesWithNodeNative(t, source)
}

func TestArrayNarrowingPreservedUnionAgrees(t *testing.T) {
	t.Parallel()
	source := arrayNarrowingSource(t, "preserved_union")
	lowersAndAgreesWithNode(t, source)
	// Only native exposes whether the union layout is retained.
	lowersAndAgreesWithNodeNative(t, source)
}

func TestArrayNarrowingBooleanCallbackAgrees(t *testing.T) {
	t.Parallel()
	source := arrayNarrowingSource(t, "boolean_callback")
	lowersAndAgreesWithNode(t, source)
	// Only native exposes whether the union layout is retained.
	lowersAndAgreesWithNodeNative(t, source)
}

func TestArrayNarrowingStoredReadsAgrees(t *testing.T) {
	t.Parallel()
	source := `const items: (string | number)[] = [1, 2];
 if (items.every((value): value is number => typeof value === 'number')) {
 const first = items[0];
 if (first !== undefined) console.log('first ' + (first + 1));
 console.log(items.map(value => value * 2).join(','));
 for (const value of items) console.log('next ' + (value + 1));
 }`
	lowersAndAgreesWithNode(t, source)
	// Native exposes boxed storage in indexed, mapped and loop reads.
	lowersAndAgreesWithNodeNative(t, source)
}
