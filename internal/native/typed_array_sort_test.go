package native

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func typedArraySortInputs(t *testing.T, kind string) (string, []byte) {
	t.Helper()
	source, err := os.ReadFile("testdata/typed-arrays/sort.c")
	if err != nil {
		t.Fatal(err)
	}
	expected := inlineArrayRun(t, "node", "testdata/typed-arrays/sort-node.a", kind)
	if expected.exit != 0 || len(expected.stderr) != 0 {
		t.Fatalf("Node: %+v", expected)
	}
	return string(source), expected.stdout
}

func typedArraySortRun(t *testing.T, source, kind string, options Options) inlineArrayResult {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "sort")
	if err := Build(source, binary, options); err != nil {
		t.Fatal("sort must compile:", err)
	}
	return inlineArrayRun(t, binary, kind)
}

func typedArraySortControl(t *testing.T, kind string) {
	t.Helper()
	source, expected := typedArraySortInputs(t, kind)
	for _, options := range []Options{{Count: true}, {Count: true, Sanitize: true}} {
		actual := typedArraySortRun(t, source, kind, options)
		if actual.exit != 0 || !bytes.Equal(actual.stdout, expected) || !strings.HasPrefix(string(actual.stderr), "adamic: counts:") {
			t.Fatalf("runtime differs from Node: sanitize=%t, exit=%d stdout=%q stderr=%q", options.Sanitize, actual.exit, actual.stdout, actual.stderr)
		}
	}
}

// Each mutant compiles the shipping implementation into its C unit under a
// distinct entry name. Only the named ordering rule changes; Node catches it.
func typedArraySortMutation(t *testing.T, kind, before, after string) {
	t.Helper()
	source, expected := typedArraySortInputs(t, kind)
	control := typedArraySortRun(t, source, kind, Options{Count: true, Sanitize: true})
	if control.exit != 0 || !bytes.Equal(control.stdout, expected) {
		t.Fatalf("control differs from Node: %+v", control)
	}
	implementation, err := runtime.ReadFile("runtime/typed_array_sort.c")
	if err != nil {
		t.Fatal(err)
	}
	code := string(implementation)
	if strings.Count(code, before) != 1 {
		t.Fatal("stale or ambiguous sort mutation", before)
	}
	code = strings.Replace(code, before, after, 1)
	code = strings.ReplaceAll(code, "adamic_typed_array_sort(", "mutant_adamic_typed_array_sort(")
	source = strings.ReplaceAll(source, "adamic_typed_array_sort(", "mutant_adamic_typed_array_sort(")
	source = strings.Replace(source, "#include \"adamic.h\"", "#include \"adamic.h\"\n"+code, 1)
	actual := typedArraySortRun(t, source, kind, Options{Count: true, Sanitize: true})
	if actual.exit != 0 || bytes.Equal(actual.stdout, expected) || !strings.HasPrefix(string(actual.stderr), "adamic: counts:") {
		t.Fatalf("mutant must finish and differ only in Node's output: %+v", actual)
	}
	t.Log("shipping sort mutant caught by Node stdout difference")
}

func TestTypedArraySortUint8(t *testing.T)   { t.Parallel(); typedArraySortControl(t, "1") }
func TestTypedArraySortUint16(t *testing.T)  { t.Parallel(); typedArraySortControl(t, "4") }
func TestTypedArraySortInt32(t *testing.T)   { t.Parallel(); typedArraySortControl(t, "2") }
func TestTypedArraySortFloat64(t *testing.T) { t.Parallel(); typedArraySortControl(t, "3") }

func TestTypedArraySortSignedZeroMutant(t *testing.T) {
	t.Parallel()
	typedArraySortMutation(t, "3", "return signbit(left) && !signbit(right);", "return false;")
}
func TestTypedArraySortNaNMutant(t *testing.T) {
	t.Parallel()
	typedArraySortMutation(t, "3", "if (isnan(left)) { return false; }\n\tif (isnan(right)) { return true; }", "if (isnan(left)) { return !isnan(right); }\n\tif (isnan(right)) { return false; }")
}
func TestTypedArraySortViewMutant(t *testing.T) {
	t.Parallel()
	typedArraySortMutation(t, "4", "adamic_typed_array *range = array;", "adamic_typed_array *range = array->owner != NULL ? array->owner : array;")
}
func TestTypedArraySortStringOrderMutant(t *testing.T) {
	t.Parallel()
	typedArraySortMutation(t, "1", "#define NUMERIC_BEFORE(left, right) ((left) < (right))", `static bool before_string(double left, double right) {
 char a[ADAMIC_NUMBER_FORMAT_MAX + 1], b[ADAMIC_NUMBER_FORMAT_MAX + 1];
 a[adamic_number_format(left, a)] = 0; b[adamic_number_format(right, b)] = 0;
 return strcmp(a, b) < 0;
}
#define NUMERIC_BEFORE(left, right) before_string(left, right)`)
}
func TestTypedArraySortStableNaNsMutant(t *testing.T) {
	t.Parallel()
	typedArraySortMutation(t, "3", "if (before(source[right], source[left]))", "if (!before(source[left], source[right]))")
}
