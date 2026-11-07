package native

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func typedArraySortInputs(test *testing.T, kind string) (string, []byte) {
	test.Helper()
	source, readError := os.ReadFile("testdata/typed-arrays/sort.c")
	if readError != nil {
		test.Fatal(readError)
	}
	expected := inlineArrayRun(test, "node", "testdata/typed-arrays/sort-node.a", kind)
	if expected.exit != 0 || len(expected.stderr) != 0 {
		test.Fatalf("Node: %+v", expected)
	}
	return string(source), expected.stdout
}

func typedArraySortRun(test *testing.T, source, kind string, options Options) inlineArrayResult {
	test.Helper()
	binary := filepath.Join(test.TempDir(), "sort")
	if buildError := Build(source, binary, options); buildError != nil {
		test.Fatal("sort must compile:", buildError)
	}
	return inlineArrayRun(test, binary, kind)
}

func typedArraySortControl(test *testing.T, kind string) {
	test.Helper()
	source, expected := typedArraySortInputs(test, kind)
	for _, options := range []Options{{Count: true}, {Count: true, Sanitize: true}} {
		actual := typedArraySortRun(test, source, kind, options)
		if actual.exit != 0 ||
			!bytes.Equal(actual.stdout, expected) ||
			!strings.HasPrefix(string(actual.stderr), "adamic: counts:") {
			test.Fatalf(
				"runtime differs from Node: sanitize=%t, exit=%d stdout=%q stderr=%q",
				options.Sanitize,
				actual.exit,
				actual.stdout,
				actual.stderr,
			)
		}
	}
}

// Each mutant compiles the shipping implementation into its C unit under a
// distinct entry name. Only the named ordering rule changes; Node catches it.
func typedArraySortMutation(test *testing.T, kind, before, after string) {
	test.Helper()
	source, expected := typedArraySortInputs(test, kind)
	control := typedArraySortRun(test, source, kind, Options{Count: true, Sanitize: true})
	if control.exit != 0 || !bytes.Equal(control.stdout, expected) {
		test.Fatalf("control differs from Node: %+v", control)
	}
	implementation, readFileError := runtime.ReadFile("runtime/typed_array_sort.c")
	if readFileError != nil {
		test.Fatal(readFileError)
	}
	code := string(implementation)
	if strings.Count(code, before) != 1 {
		test.Fatal("stale or ambiguous sort mutation", before)
	}
	code = strings.Replace(code, before, after, 1)
	code = strings.ReplaceAll(code, "adamic_typed_array_sort(", "mutant_adamic_typed_array_sort(")
	source = strings.ReplaceAll(source, "adamic_typed_array_sort(", "mutant_adamic_typed_array_sort(")
	source = strings.Replace(source, "#include \"adamic.h\"", "#include \"adamic.h\"\n"+code, 1)
	actual := typedArraySortRun(test, source, kind, Options{Count: true, Sanitize: true})
	if actual.exit != 0 ||
		bytes.Equal(actual.stdout, expected) ||
		!strings.HasPrefix(string(actual.stderr), "adamic: counts:") {
		test.Fatalf("mutant must finish and differ only in Node's output: %+v", actual)
	}
	test.Log("shipping sort mutant caught by Node stdout difference")
}

func TestTypedArraySortUint8(test *testing.T)   { test.Parallel(); typedArraySortControl(test, "1") }
func TestTypedArraySortUint16(test *testing.T)  { test.Parallel(); typedArraySortControl(test, "4") }
func TestTypedArraySortInt32(test *testing.T)   { test.Parallel(); typedArraySortControl(test, "2") }
func TestTypedArraySortFloat64(test *testing.T) { test.Parallel(); typedArraySortControl(test, "3") }

func TestTypedArraySortSignedZeroMutant(test *testing.T) {
	test.Parallel()
	typedArraySortMutation(test, "3", "return signbit(left) && !signbit(right);", "return false;")
}
func TestTypedArraySortNaNMutant(test *testing.T) {
	test.Parallel()
	typedArraySortMutation(
		test,
		"3",
		"if (isnan(left)) { return false; }\n\tif (isnan(right)) { return true; }",
		"if (isnan(left)) { return !isnan(right); }\n\tif (isnan(right)) { return false; }",
	)
}
func TestTypedArraySortViewMutant(test *testing.T) {
	test.Parallel()
	typedArraySortMutation(
		test,
		"4",
		"adamic_typed_array *range = array;",
		"adamic_typed_array *range = array->owner != NULL ? array->owner : array;",
	)
}
func TestTypedArraySortStringOrderMutant(test *testing.T) {
	test.Parallel()
	typedArraySortMutation(
		test,
		"1",
		"#define NUMERIC_BEFORE(left, right) ((left) < (right))",
		`static bool before_string(double left, double right) {
 char a[ADAMIC_NUMBER_FORMAT_MAX + 1], b[ADAMIC_NUMBER_FORMAT_MAX + 1];
 a[adamic_number_format(left, a)] = 0; b[adamic_number_format(right, b)] = 0;
 return strcmp(a, b) < 0;
}
#define NUMERIC_BEFORE(left, right) before_string(left, right)`,
	)
}
func TestTypedArraySortStableNaNsMutant(test *testing.T) {
	test.Parallel()
	typedArraySortMutation(
		test,
		"3",
		"if (before(source[right], source[left]))",
		"if (!before(source[left], source[right]))",
	)
}
