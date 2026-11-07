package oracle

import (
	"github.com/system-inc/adamic/internal/lower"
	"os"
	"path/filepath"
	"testing"
)

func checkCallbackWidening(test *testing.T, name, want string, backends ...string) {
	test.Helper()
	path, _ := filepath.Abs("testdata/callback_widening/" + name + ".a")
	node := onNode(test, path)
	if diff := disagreement(run{stdout: []byte(want)}, node); diff != "" {
		test.Fatal("Node: " + diff)
	}
	program, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
	}
	if len(backends) == 0 {
		backends = []string{"native", "native-sanitized", "javascript"}
	}
	for _, backend := range backends {
		var got run
		switch backend {
		case "native":
			got = releasedUncached(test, program)
		case "native-sanitized":
			var binary string
			got, binary = nativelyUncached(test, program)
			if report := leaksUncached(test, program, binary); report != "" {
				test.Fatal(report)
			}
		case "javascript":
			got = onJavaScriptBackend(test, program)
		default:
			test.Fatalf("unknown backend %q", backend)
		}
		if diff := disagreement(node, got); diff != "" {
			test.Fatalf("%s: %s; stdout %q stderr %q exit %d", backend, diff, got.stdout, got.stderr, got.exitCode)
		}
	}
}
func TestCallbackWideningNumberEveryRelease(test *testing.T) {
	test.Parallel()
	checkCallbackWidening(test, "number-every", "true\n", "native")
}

func TestCallbackWideningNumberEverySanitized(test *testing.T) {
	test.Parallel()
	checkCallbackWidening(test, "number-every", "true\n", "native-sanitized")
}

func TestCallbackWideningNumberEveryJavaScript(test *testing.T) {
	test.Parallel()
	checkCallbackWidening(test, "number-every", "true\n", "javascript")
}
func TestCallbackWideningNumberFilter(test *testing.T) {
	test.Parallel()
	checkCallbackWidening(test, "number-filter", "1,3\n")
}
func TestCallbackWideningNumberFind(test *testing.T) {
	test.Parallel()
	checkCallbackWidening(test, "number-find", "2\n")
}
func TestCallbackWideningNumberForEach(test *testing.T) {
	test.Parallel()
	checkCallbackWidening(test, "number-forEach", "1\n2\n3\n")
}
func TestCallbackWideningNumberGeneric(test *testing.T) {
	test.Parallel()
	checkCallbackWidening(test, "number-generic", "1,2,3\n")
}
func TestCallbackWideningNumberMap(test *testing.T) {
	test.Parallel()
	checkCallbackWidening(test, "number-map", "1,2,3\n")
}
func TestCallbackWideningNumberReduce(test *testing.T) {
	test.Parallel()
	checkCallbackWidening(test, "number-reduce", "123\n")
}
func TestCallbackWideningNumberSome(test *testing.T) {
	test.Parallel()
	checkCallbackWidening(test, "number-some", "true\n")
}
func TestCallbackWideningNumberSort(test *testing.T) {
	test.Parallel()
	checkCallbackWidening(test, "number-sort", "1,2,3\n")
}
func TestCallbackWideningObjectEvery(test *testing.T) {
	test.Parallel()
	checkCallbackWidening(test, "object-every", "true\n")
}
func TestCallbackWideningObjectFilter(test *testing.T) {
	test.Parallel()
	checkCallbackWidening(test, "object-filter", "1,3\n")
}
func TestCallbackWideningObjectFind(test *testing.T) {
	test.Parallel()
	checkCallbackWidening(test, "object-find", "2\n")
}
func TestCallbackWideningObjectForEach(test *testing.T) {
	test.Parallel()
	checkCallbackWidening(test, "object-forEach", "1\n2\n3\n")
}
func TestCallbackWideningObjectGeneric(test *testing.T) {
	test.Parallel()
	checkCallbackWidening(test, "object-generic", "1,2,3\n")
}
func TestCallbackWideningObjectMap(test *testing.T) {
	test.Parallel()
	checkCallbackWidening(test, "object-map", "1,2,3\n")
}
func TestCallbackWideningObjectReduce(test *testing.T) {
	test.Parallel()
	checkCallbackWidening(test, "object-reduce", "123\n")
}
func TestCallbackWideningObjectSome(test *testing.T) {
	test.Parallel()
	checkCallbackWidening(test, "object-some", "true\n")
}
func TestCallbackWideningObjectSort(test *testing.T) {
	test.Parallel()
	checkCallbackWidening(test, "object-sort", "1,2,3\n")
}
func TestCallbackWideningNumberNamedSort(test *testing.T) {
	test.Parallel()
	checkCallbackWidening(test, "number-named-sort", "1,2,3\n")
}
func TestCallbackWideningNumberIndex(test *testing.T) {
	test.Parallel()
	checkCallbackWidening(test, "number-index", "4:0,5:1\n")
}
func TestCallbackWideningNumberReduceAccumulator(test *testing.T) {
	test.Parallel()
	checkCallbackWidening(test, "number-reduce-accumulator", "6\n")
}
func TestCallbackWideningNumberCapture(test *testing.T) {
	test.Parallel()
	checkCallbackWidening(test, "number-capture", "v1,v2\nw3\n2\n")
}

func TestCallbackWideningNumberThrow(test *testing.T) {
	test.Parallel()
	checkCallbackWidening(test, "number-throw", "caught\n2\n")
}
func TestCallbackWideningLibraryMap(test *testing.T) {
	test.Parallel()
	checkCallbackWidening(test, "library-map", "1,2,3\n11,NaN,3\n")
}
func TestCallbackWideningNumberFindLast(test *testing.T) {
	test.Parallel()
	checkCallbackWidening(test, "number-findLast", "2\n")
}
func TestCallbackWideningNumberFindIndex(test *testing.T) {
	test.Parallel()
	checkCallbackWidening(test, "number-findIndex", "1\n")
}
func TestCallbackWideningNumberToSorted(test *testing.T) {
	test.Parallel()
	checkCallbackWidening(test, "number-toSorted", "1,2,3\n")
}

func TestCallbackWideningNumberGenericMap(test *testing.T) {
	test.Parallel()
	checkCallbackWidening(test, "number-generic-map", "1,2,3\n")
}

func TestCallbackWideningObjectGenericMap(test *testing.T) {
	test.Parallel()
	checkCallbackWidening(test, "object-generic-map", "1,2,3\n")
}

func TestCallbackWideningNumberArguments(test *testing.T) {
	test.Parallel()
	checkCallbackWidening(test, "number-arguments", "1:3,2:3\n")
}

func TestCallbackWideningBooleanMap(test *testing.T) {
	test.Parallel()
	checkCallbackWidening(test, "boolean-map", "true,false\n")
}

func TestCallbackWideningStringMap(test *testing.T) {
	test.Parallel()
	checkCallbackWidening(test, "string-map", "a,b\n")
}

func TestCallbackWideningOptionalNumberMap(test *testing.T) {
	test.Parallel()
	checkCallbackWidening(test, "optional-number-map", "1,undefined\n")
}

func TestCallbackWideningNullableObjectMapRelease(test *testing.T) {
	test.Parallel()
	checkCallbackWidening(test, "nullable-object-map", "1,null\n", "native")
}

func TestCallbackWideningNullableObjectMapSanitized(test *testing.T) {
	test.Parallel()
	checkCallbackWidening(test, "nullable-object-map", "1,null\n", "native-sanitized")
}

func TestCallbackWideningNullableObjectMapJavaScript(test *testing.T) {
	test.Parallel()
	checkCallbackWidening(test, "nullable-object-map", "1,null\n", "javascript")
}

func TestCallbackWideningOptionalObjectMap(test *testing.T) {
	test.Parallel()
	checkCallbackWidening(test, "optional-object-map", "1,undefined\n")
}

func TestCallbackWideningUnrepresentedParameter(test *testing.T) {
	test.Parallel()
	path, _ := filepath.Abs("../lower/testdata/callback_widening/array_structural.a")
	node := onNode(test, path)
	if diff := disagreement(run{stdout: []byte("1,2|3,4\n")}, node); diff != "" {
		test.Fatal("Node: " + diff)
	}
	// The callback's parameter takes each element, an array, where Slice, a plain object type, is held, so
	// it stops at the callback (lower's shape_fit.go, #nw6d7ws) before the intrinsic's adapter could refuse it.
	_, err := lowered(test, path)
	if _, ok := err.(*lower.NotYet); !ok ||
		err.Error() != path+":3:31: stage 0 can't lower an array going where a plain object type is held yet" {
		test.Fatalf("expected the array's NotYet at the callback, got %v", err)
	}
}

// The adapter's refusal still has a witness: a Promise element going to a callback parameter held as an
// object (PromiseLike<number>), a pair the container checks don't judge, since only a map or an array goes
// to a plain object type there.
func TestCallbackWideningUnrepresentedPromiseParameter(test *testing.T) {
	test.Parallel()
	path := filepath.Join(test.TempDir(), "promise_like.a")
	if err := os.WriteFile(path, []byte(
		"const tasks = [Promise.resolve(1)];\nconsole.log(tasks.map((p: PromiseLike<number>): string => typeof p).join(','));\n",
	), 0o644); err != nil {
		test.Fatal(err)
	}
	node := onNode(test, path)
	if diff := disagreement(run{stdout: []byte("object\n")}, node); diff != "" {
		test.Fatal("Node: " + diff)
	}
	_, err := lowered(test, path)
	refused, ok := err.(*lower.Refused)
	if !ok {
		test.Fatalf("expected unrepresented parameter refusal, got %v", err)
	}
	if refused.What != "an intrinsic callback parameter p of type PromiseLike<number> receiving array element type Promise<number> without a representation adapter" ||
		refused.Fix != "annotate parameter p as the element type Promise<number>" {
		test.Fatalf("wrong refusal: %#v", refused)
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
