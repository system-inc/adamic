package oracle

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func TestCheckedViewTupleForEachTransfer(t *testing.T) {
	verifyOriginalTupleCases(t, []originalTupleCase{
		{"transfer-tuple", "object\n", "", ""},
		{"transfer-scalar", "number\n", "", ""},
		{"transfer-throw", "number\ncaught\ndone\n", "", ""},
		{"transfer-store", "object\n", "", ""},
	})
}

// Both mutations affect exactly the transferred scalar box's normal release.
// Skip remains semantically correct until the shared leak check observes it;
// double frees a real allocation and must reach AddressSanitizer.
func TestCheckedViewTupleForEachTransferMutants(t *testing.T) {
	mode := os.Getenv("ADAMIC_TUPLE_TRANSFER_MUTANT")
	if mode == "" {
		t.Skip("set skip or double to run the native ownership mutant")
	}
	root, _ := tupleOriginalInputs(t)
	_, program := tupleOriginalProgram(t, root, "transfer-scalar")
	code := native.C(program)
	pattern := regexp.MustCompile(`/\* tuple forEach transfer release \*/\s*(adamic_release\([^\n]+\);)`)
	if len(pattern.FindAllString(code, -1)) != 1 {
		t.Fatal("wanted exactly one transfer release")
	}
	switch mode {
	case "skip":
		code = pattern.ReplaceAllString(code, "/* mutant skips transferred release */")
	case "double":
		code = pattern.ReplaceAllString(code, "$1\n$1")
	default:
		t.Fatal("unknown transfer mutant")
	}
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	actual := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=0"}, binary)
	if mode == "skip" {
		if diff := disagreement(run{stdout: []byte("number\n")}, actual); diff != "" {
			t.Fatalf("skip must finish with unchanged output: %s; %#v", diff, actual)
		}
		if report := leakChecked(t, code, binary); report != "" {
			t.Fatalf("transfer mutant caught by leak check: %s", report)
		}
		return
	}
	if actual.exitCode != 0 && strings.Contains(string(actual.stderr), "AddressSanitizer") {
		t.Fatalf("double-release mutant caught by sanitizer: %s", actual.stderr)
	}
	return
}
