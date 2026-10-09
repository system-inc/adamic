package oracle

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

const refusalFixtureRoot = "internal/oracle/testdata/refusal_rulings/"

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{refusalFixtureRoot + "function_signature.a", true, false})
	additionalFixtureCounts = append(additionalFixtureCounts, functionRulingCounts)
}

// Source witnesses remain .a in the repository. A physical temporary .ts copy
// exercises the compiler-source frontend without weakening .a's promises.
func rulingProgram(t *testing.T, file string, typescript bool) (*ir.Program, error) {
	t.Helper()
	path := filepath.Join(repository, refusalFixtureRoot, file)
	if typescript {
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		path = filepath.Join(t.TempDir(), strings.TrimSuffix(file, ".a")+".ts")
		if err := os.WriteFile(path, source, 0600); err != nil {
			t.Fatal(err)
		}
	}
	return lowered(t, path)
}

func rulingRefusal(t *testing.T, file string, typescript bool, diagnostic string) {
	t.Helper()
	_, err := rulingProgram(t, file, typescript)
	var refused *lower.Refused
	if !errors.As(err, &refused) {
		t.Fatalf("want Refused, got %v", err)
	}
	name := file
	if typescript {
		name = strings.TrimSuffix(file, ".a") + ".ts"
	}
	if !strings.Contains(refused.Where, name+":") || refused.What+"; "+refused.Fix != diagnostic {
		t.Fatalf("un-pinned refusal: %v", err)
	}
	result := rulingNode(t, file)
	if result.exitCode != 0 || len(result.stderr) != 0 {
		t.Fatalf("source Node: %#v", result)
	}
}

func rulingAgrees(t *testing.T, file string, typescript bool) {
	t.Helper()
	program, err := rulingProgram(t, file, typescript)
	if err != nil {
		t.Fatal(err)
	}
	expected := rulingNode(t, file)
	js := onJavaScriptBackend(t, program)
	actual, binary := natively(t, program)
	for _, result := range []run{js, actual, released(t, program)} {
		if result.exitCode != expected.exitCode || !bytes.Equal(result.stdout, expected.stdout) || !bytes.Equal(result.stderr, expected.stderr) {
			t.Fatalf("Node %#v; backend %#v", expected, result)
		}
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
}

func TestFunctionAdamicUnusedAnnotationRefused(t *testing.T) {
	t.Parallel()
	rulingRefusal(t, "function_unused.a", false, "the Function annotation in .a; write a call signature with its parameters and result, like (value: number) => number")
}
func TestFunctionAdamicLengthAnnotationRefused(t *testing.T) {
	t.Parallel()
	rulingRefusal(t, "function_length.a", false, "the Function annotation in .a; write a call signature with its parameters and result, like (value: number) => number")
}
func TestFunctionTypeScriptUnusedAccepted(t *testing.T) {
	t.Parallel()
	rulingAgrees(t, "function_unused.a", true)
}
func TestFunctionTypeScriptLengthAgreesWithNode(t *testing.T) {
	t.Parallel()
	rulingAgrees(t, "function_length.a", true)
}
func TestFunctionTypeScriptCallRefused(t *testing.T) {
	t.Parallel()
	rulingRefusal(t, "function_call.a", true, "calling through the unchecked Function type; write a call signature with the actual parameters and result before calling")
}
func TestFunctionCallSignatureAgreesWithNode(t *testing.T) {
	t.Parallel()
	rulingAgrees(t, "function_signature.a", false)
}

func rulingCount(t *testing.T, file string, typescript bool) string {
	t.Helper()
	program, err := rulingProgram(t, file, typescript)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "counted")
	if err := native.Build(native.C(program), binary, native.Options{Count: true}); err != nil {
		t.Fatal(err)
	}
	name, args := pinnedStack(binary)
	result := execute(t, name, args...)
	match := countsLine.FindSubmatch(result.stderr)
	if match == nil {
		t.Fatalf("no counts: %#v", result)
	}
	mode := ""
	if typescript {
		mode = " (TypeScript mode)"
	}
	return fmt.Sprintf("| %s%s | %s | %s | %s | %s | %s | %s |", refusalFixtureRoot+file, mode, match[1], match[2], match[3], match[4], match[5], match[6])
}
func functionRulingCounts(t *testing.T) []string {
	return []string{rulingCount(t, "function_unused.a", true), rulingCount(t, "function_length.a", true)}
}

func rulingNode(t *testing.T, file string) run {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, refusalFixtureRoot, file))
	if err != nil {
		t.Fatal(err)
	}
	return onNode(t, path)
}
