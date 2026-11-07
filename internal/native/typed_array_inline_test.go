package native

import (
	"bytes"
	"os"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/boundedrun"
)

type inlineArrayResult struct {
	stdout, stderr []byte
	exit           int
}

func inlineArrayRun(test *testing.T, command string, arguments ...string) inlineArrayResult {
	test.Helper()
	process, cancel := boundedrun.Command(10*time.Second, command, arguments...)
	defer cancel()
	leaks := "0"
	if goruntime.GOOS == "linux" {
		leaks = "1"
	}
	process.Env = append(
		os.Environ(),
		"ASAN_OPTIONS=detect_leaks="+leaks+":halt_on_error=1",
		"UBSAN_OPTIONS=halt_on_error=1",
	)
	var stdout, stderr bytes.Buffer
	process.Stdout, process.Stderr = &stdout, &stderr
	err := process.Run()
	exit := 0
	if err != nil {
		if process.ProcessState == nil {
			test.Fatal(err)
		}
		exit = process.ProcessState.ExitCode()
	}
	return inlineArrayResult{stdout.Bytes(), stderr.Bytes(), exit}
}

// Copy the shipping helper, rather than a second implementation, into the C
// unit under a new name. Only its check changes; all other runtime code is shared.
func typedArrayInlineCheck(test *testing.T, mode, helper, before, after, diagnostic string) {
	test.Helper()
	source, err := os.ReadFile("testdata/typed-arrays/runtime.c")
	if err != nil {
		test.Fatal(err)
	}
	node := inlineArrayRun(test, "node", "testdata/typed-arrays/node.a", mode)
	if node.exit != 0 || len(node.stderr) != 0 {
		test.Fatalf("Node: %+v", node)
	}
	options := Options{Sanitize: true, Count: true}
	control := filepath.Join(test.TempDir(), "control")
	if err := Build(string(source), control, options); err != nil {
		test.Fatal(err)
	}
	actual := inlineArrayRun(test, control, mode)
	if strings.HasSuffix(mode, "write") {
		if actual.exit != 70 ||
			!strings.HasPrefix(string(actual.stderr), "adamic: panic: index 3 is outside an array of length 3\n") ||
			len(actual.stdout) != 0 ||
			string(node.stdout) != strings.Repeat("undefined\n", 4) {
			test.Fatalf("write stop: Node %+v, native %+v", node, actual)
		}
	} else if actual.exit != 0 ||
		!bytes.Equal(actual.stdout, node.stdout) ||
		!strings.HasPrefix(string(actual.stderr), "adamic: counts:") {
		test.Fatalf("control differs from Node: Node %+v native %+v", node, actual)
	}
	header, err := runtime.ReadFile("runtime/adamic.h")
	if err != nil {
		test.Fatal(err)
	}
	start := strings.Index(string(header), " "+helper+"(")
	if start < 0 {
		test.Fatal("missing shipping helper", helper)
	}
	start = strings.LastIndex(string(header)[:start], "\n") + 1
	body := string(header)[start:]
	end := strings.Index(body, "\n}\n")
	if end < 0 {
		test.Fatal("missing helper end")
	}
	body = body[:end+3]
	if strings.Count(body, before) != 1 {
		test.Fatal("missing or ambiguous mutation", helper, before)
	}
	body = strings.Replace(body, before, after, 1)
	body = strings.ReplaceAll(body, helper+"(", "mutant_"+helper+"(")
	code := strings.ReplaceAll(string(source), helper+"(", "mutant_"+helper+"(")
	code = strings.Replace(code, "#include \"adamic.h\"", "#include \"adamic.h\"\n"+body, 1)
	mutant := filepath.Join(test.TempDir(), "mutant")
	if err := Build(code, mutant, options); err != nil {
		test.Fatal("mutant must compile:", err)
	}
	changed := inlineArrayRun(test, mutant, mode)
	if diagnostic == "Node" {
		if changed.exit != 0 ||
			bytes.Equal(changed.stdout, node.stdout) ||
			!strings.HasPrefix(string(changed.stderr), "adamic: counts:") {
			test.Fatalf("only Node's output must catch this mutant: %+v", changed)
		}
	} else if changed.exit == 0 || !strings.Contains(string(changed.stderr), diagnostic) {
		test.Fatalf("mutant survived or failed differently, want %s: %+v", diagnostic, changed)
	}
	test.Logf("%s agrees with Node; %s mutant caught by %s", mode, helper, diagnostic)
}

func TestTypedArrayInlineUint8Conversion(test *testing.T) {
	test.Parallel()
	typedArrayInlineCheck(
		test,
		"conversion",
		"adamic_typed_array_set",
		"value < 256.0",
		"value < 257.0",
		"outside the range of representable values",
	)
}
func TestTypedArrayInlineUint16Conversion(test *testing.T) {
	test.Parallel()
	typedArrayInlineCheck(
		test,
		"conversion",
		"adamic_typed_array_set",
		"value < 65536.0",
		"value < 65537.0",
		"outside the range of representable values",
	)
}
func TestTypedArrayInlineInt32Conversion(test *testing.T) {
	test.Parallel()
	typedArrayInlineCheck(
		test,
		"conversion",
		"adamic_typed_array_set",
		"value < 2147483648.0",
		"value < 2147483649.0",
		"outside the range of representable values",
	)
}
func TestTypedArrayInlineFloat64Conversion(test *testing.T) {
	test.Parallel()
	typedArrayInlineCheck(
		test,
		"conversion",
		"adamic_typed_array_set",
		"((double *)array->data)[whole] = value;",
		"((double *)array->data)[whole] = fabs(value);",
		"Node",
	)
}
func TestTypedArrayInlineFractionalIndex(test *testing.T) {
	test.Parallel()
	typedArrayInlineCheck(
		test,
		"fractional",
		"adamic_typed_array_get",
		"if ((double)whole != index)",
		"if (false)",
		"Node",
	)
}
func TestTypedArrayInlineNaNIndex(test *testing.T) {
	test.Parallel()
	typedArrayInlineCheck(
		test,
		"nan",
		"adamic_typed_array_get",
		"!(index >= 0)",
		"index < 0",
		"outside the range of representable values",
	)
}
func TestTypedArrayInlineNegativeZeroIndex(test *testing.T) {
	test.Parallel()
	typedArrayInlineCheck(test, "negative-zero", "adamic_typed_array_get", "!(index >= 0)", "!(index > 0)", "Node")
}
func TestTypedArrayInlineBoundsRead(test *testing.T) {
	test.Parallel()
	typedArrayInlineCheck(
		test,
		"bounds",
		"adamic_typed_array_get",
		"index >= (double)array->length",
		"index > (double)array->length",
		"heap-buffer-overflow",
	)
}
func TestTypedArrayInlineBoundsWrite(test *testing.T) {
	test.Parallel()
	typedArrayInlineCheck(
		test,
		"bounds-write",
		"adamic_typed_array_set",
		"index < (double)array->length",
		"index <= (double)array->length",
		"heap-buffer-overflow",
	)
}
func TestTypedArrayInlineIntegerBoundsRead(test *testing.T) {
	test.Parallel()
	typedArrayInlineCheck(
		test,
		"integer-bounds",
		"adamic_typed_array_get_integer",
		"(uint64_t)index >= array->length",
		"(uint64_t)index > array->length",
		"heap-buffer-overflow",
	)
}
func TestTypedArrayInlineIntegerBoundsWrite(test *testing.T) {
	test.Parallel()
	typedArrayInlineCheck(
		test,
		"integer-bounds-write",
		"adamic_typed_array_set_integer",
		"(uint64_t)index >= array->length",
		"(uint64_t)index > array->length",
		"heap-buffer-overflow",
	)
}
