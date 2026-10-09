package oracle

import (
	"github.com/system-inc/adamic/internal/lower"
	"path/filepath"
	"testing"
)

func checkCallbackWidening(t *testing.T, name, want string) {
	t.Helper()
	path, _ := filepath.Abs("testdata/callback_widening/" + name + ".a")
	node := onNode(t, path)
	if diff := disagreement(run{stdout: []byte(want)}, node); diff != "" {
		t.Fatal("Node: " + diff)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	sanitized, binary := nativelyUncached(t, program)
	for backend, got := range map[string]run{"native": releasedUncached(t, program), "native-sanitized": sanitized, "javascript": onJavaScriptBackend(t, program)} {
		if diff := disagreement(node, got); diff != "" {
			t.Fatalf("%s: %s; stdout %q stderr %q exit %d", backend, diff, got.stdout, got.stderr, got.exitCode)
		}
	}
	if report := leaksUncached(t, program, binary); report != "" {
		t.Fatal(report)
	}
}
func TestCallbackWideningNumberEvery(t *testing.T) {
	t.Parallel()
	checkCallbackWidening(t, "number-every", "true\n")
}
func TestCallbackWideningNumberFilter(t *testing.T) {
	t.Parallel()
	checkCallbackWidening(t, "number-filter", "1,3\n")
}
func TestCallbackWideningNumberFind(t *testing.T) {
	t.Parallel()
	checkCallbackWidening(t, "number-find", "2\n")
}
func TestCallbackWideningNumberForEach(t *testing.T) {
	t.Parallel()
	checkCallbackWidening(t, "number-forEach", "1\n2\n3\n")
}
func TestCallbackWideningNumberGeneric(t *testing.T) {
	t.Parallel()
	checkCallbackWidening(t, "number-generic", "1,2,3\n")
}
func TestCallbackWideningNumberMap(t *testing.T) {
	t.Parallel()
	checkCallbackWidening(t, "number-map", "1,2,3\n")
}
func TestCallbackWideningNumberReduce(t *testing.T) {
	t.Parallel()
	checkCallbackWidening(t, "number-reduce", "123\n")
}
func TestCallbackWideningNumberSome(t *testing.T) {
	t.Parallel()
	checkCallbackWidening(t, "number-some", "true\n")
}
func TestCallbackWideningNumberSort(t *testing.T) {
	t.Parallel()
	checkCallbackWidening(t, "number-sort", "1,2,3\n")
}
func TestCallbackWideningObjectEvery(t *testing.T) {
	t.Parallel()
	checkCallbackWidening(t, "object-every", "true\n")
}
func TestCallbackWideningObjectFilter(t *testing.T) {
	t.Parallel()
	checkCallbackWidening(t, "object-filter", "1,3\n")
}
func TestCallbackWideningObjectFind(t *testing.T) {
	t.Parallel()
	checkCallbackWidening(t, "object-find", "2\n")
}
func TestCallbackWideningObjectForEach(t *testing.T) {
	t.Parallel()
	checkCallbackWidening(t, "object-forEach", "1\n2\n3\n")
}
func TestCallbackWideningObjectGeneric(t *testing.T) {
	t.Parallel()
	checkCallbackWidening(t, "object-generic", "1,2,3\n")
}
func TestCallbackWideningObjectMap(t *testing.T) {
	t.Parallel()
	checkCallbackWidening(t, "object-map", "1,2,3\n")
}
func TestCallbackWideningObjectReduce(t *testing.T) {
	t.Parallel()
	checkCallbackWidening(t, "object-reduce", "123\n")
}
func TestCallbackWideningObjectSome(t *testing.T) {
	t.Parallel()
	checkCallbackWidening(t, "object-some", "true\n")
}
func TestCallbackWideningObjectSort(t *testing.T) {
	t.Parallel()
	checkCallbackWidening(t, "object-sort", "1,2,3\n")
}
func TestCallbackWideningNumberNamedSort(t *testing.T) {
	t.Parallel()
	checkCallbackWidening(t, "number-named-sort", "1,2,3\n")
}
func TestCallbackWideningNumberIndex(t *testing.T) {
	t.Parallel()
	checkCallbackWidening(t, "number-index", "4:0,5:1\n")
}
func TestCallbackWideningNumberReduceAccumulator(t *testing.T) {
	t.Parallel()
	checkCallbackWidening(t, "number-reduce-accumulator", "6\n")
}
func TestCallbackWideningNumberCapture(t *testing.T) {
	t.Parallel()
	checkCallbackWidening(t, "number-capture", "v1,v2\nw3\n2\n")
}

func TestCallbackWideningNumberThrow(t *testing.T) {
	t.Parallel()
	checkCallbackWidening(t, "number-throw", "caught\n2\n")
}
func TestCallbackWideningLibraryMap(t *testing.T) {
	t.Parallel()
	checkCallbackWidening(t, "library-map", "1,2,3\n11,NaN,3\n")
}
func TestCallbackWideningNumberFindLast(t *testing.T) {
	t.Parallel()
	checkCallbackWidening(t, "number-findLast", "2\n")
}
func TestCallbackWideningNumberFindIndex(t *testing.T) {
	t.Parallel()
	checkCallbackWidening(t, "number-findIndex", "1\n")
}
func TestCallbackWideningNumberToSorted(t *testing.T) {
	t.Parallel()
	checkCallbackWidening(t, "number-toSorted", "1,2,3\n")
}

func TestCallbackWideningNumberGenericMap(t *testing.T) {
	t.Parallel()
	checkCallbackWidening(t, "number-generic-map", "1,2,3\n")
}

func TestCallbackWideningObjectGenericMap(t *testing.T) {
	t.Parallel()
	checkCallbackWidening(t, "object-generic-map", "1,2,3\n")
}

func TestCallbackWideningNumberArguments(t *testing.T) {
	t.Parallel()
	checkCallbackWidening(t, "number-arguments", "1:3,2:3\n")
}

func TestCallbackWideningBooleanMap(t *testing.T) {
	t.Parallel()
	checkCallbackWidening(t, "boolean-map", "true,false\n")
}

func TestCallbackWideningStringMap(t *testing.T) {
	t.Parallel()
	checkCallbackWidening(t, "string-map", "a,b\n")
}

func TestCallbackWideningOptionalNumberMap(t *testing.T) {
	t.Parallel()
	checkCallbackWidening(t, "optional-number-map", "1,undefined\n")
}

func TestCallbackWideningNullableObjectMap(t *testing.T) {
	t.Parallel()
	checkCallbackWidening(t, "nullable-object-map", "1,null\n")
}

func TestCallbackWideningOptionalObjectMap(t *testing.T) {
	t.Parallel()
	checkCallbackWidening(t, "optional-object-map", "1,undefined\n")
}

func TestCallbackWideningUnrepresentedParameter(t *testing.T) {
	t.Parallel()
	path, _ := filepath.Abs("../lower/testdata/callback_widening/array_structural.a")
	node := onNode(t, path)
	if diff := disagreement(run{stdout: []byte("1,2|3,4\n")}, node); diff != "" {
		t.Fatal("Node: " + diff)
	}
	_, err := lowered(t, path)
	refused, ok := err.(*lower.Refused)
	if !ok {
		t.Fatalf("expected unrepresented parameter refusal, got %v", err)
	}
	if refused.What != "an intrinsic callback parameter x of type Slice receiving array element type number[] without a representation adapter" || refused.Fix != "annotate parameter x as the element type number[]" {
		t.Fatalf("wrong refusal: %#v", refused)
	}
}

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/callback_widening/boolean-map.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/callback_widening/library-map.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/callback_widening/nullable-object-map.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/callback_widening/number-arguments.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/callback_widening/number-capture.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/callback_widening/number-every.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/callback_widening/number-filter.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/callback_widening/number-find.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/callback_widening/number-findIndex.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/callback_widening/number-findLast.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/callback_widening/number-forEach.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/callback_widening/number-generic-map.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/callback_widening/number-generic.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/callback_widening/number-index.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/callback_widening/number-map.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/callback_widening/number-named-sort.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/callback_widening/number-reduce-accumulator.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/callback_widening/number-reduce.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/callback_widening/number-some.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/callback_widening/number-sort.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/callback_widening/number-throw.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/callback_widening/number-toSorted.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/callback_widening/object-every.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/callback_widening/object-filter.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/callback_widening/object-find.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/callback_widening/object-forEach.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/callback_widening/object-generic-map.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/callback_widening/object-generic.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/callback_widening/object-map.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/callback_widening/object-reduce.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/callback_widening/object-some.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/callback_widening/object-sort.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/callback_widening/optional-number-map.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/callback_widening/optional-object-map.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/callback_widening/string-map.a", true, false})
}
