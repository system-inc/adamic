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
	for _, file := range []string{"function_unused.a", "function_length.a"} {
		additionalFixtureCounts = append(
			additionalFixtureCounts,
			countCase{refusalFixtureRoot + file, func(test *testing.T) (string, func() string) {
				return rulingCount(test, file, true)
			}},
		)
	}
}

// Source witnesses remain .a in the repository. A physical temporary .ts copy
// exercises the compiler-source frontend without weakening .a's promises.
func rulingProgram(test *testing.T, file string, typescript bool) (*ir.Program, error) {
	test.Helper()
	program, _, rulingProgramAtError := rulingProgramAt(test, file, typescript)
	return program, rulingProgramAtError
}

// Keep the physical frontend copy's path available for main's portable count digest.
func rulingProgramAt(test *testing.T, file string, typescript bool) (*ir.Program, string, error) {
	test.Helper()
	path := filepath.Join(repository, refusalFixtureRoot, file)
	if typescript {
		source, readError := os.ReadFile(path)
		if readError != nil {
			test.Fatal(readError)
		}
		path = filepath.Join(test.TempDir(), strings.TrimSuffix(file, ".a")+".ts")
		if writeError := os.WriteFile(path, source, 0600); writeError != nil {
			test.Fatal(writeError)
		}
	}
	program, loweredError := lowered(test, path)
	return program, path, loweredError
}

func rulingRefusal(test *testing.T, file string, typescript bool, diagnostic string) {
	test.Helper()
	_, rulingProgramError := rulingProgram(test, file, typescript)
	var refused *lower.Refused
	if !errors.As(rulingProgramError, &refused) {
		test.Fatalf("want Refused, got %v", rulingProgramError)
	}
	name := file
	if typescript {
		name = strings.TrimSuffix(file, ".a") + ".ts"
	}

	positions := map[string]string{
		"function_unused.a": "2:22", "function_length.a": "2:23", "function_call.a": "3:16",
		"merge_unused.a": "2:1", "merge_missing.a": "2:1", "merge_missing_computed.a": "2:1",
	}
	if file == "function_call.a" && !typescript {
		positions[file] = "2:10"
	}
	if !strings.HasSuffix(refused.Where, name+":"+positions[file]) || refused.What+"; "+refused.Fix != diagnostic {
		test.Fatalf("un-pinned refusal: %v", rulingProgramError)
	}
	result := rulingNode(test, file)
	if result.exitCode != 0 || len(result.stderr) != 0 {
		test.Fatalf("source Node: %#v", result)
	}
}

func rulingAgrees(test *testing.T, file string, typescript bool) {
	test.Helper()
	program, rulingProgramError := rulingProgram(test, file, typescript)
	if rulingProgramError != nil {
		test.Fatal(rulingProgramError)
	}
	expected := rulingNode(test, file)
	js := onJavaScriptBackend(test, program)
	actual, binary := natively(test, program)
	for _, result := range []run{js, actual, released(test, program)} {
		if result.exitCode != expected.exitCode ||
			!bytes.Equal(result.stdout, expected.stdout) ||
			!bytes.Equal(result.stderr, expected.stderr) {
			test.Fatalf("Node %#v; backend %#v", expected, result)
		}
	}
	if report := leaks(test, program, binary); report != "" {
		test.Fatal(report)
	}
}

func TestFunctionAdamicUnusedAnnotationRefused(test *testing.T) {
	test.Parallel()
	rulingRefusal(
		test,
		"function_unused.a",
		false,
		"the Function annotation in .a; write a call signature with its parameters and result, like (value: number) => number",
	)
}
func TestFunctionAdamicLengthAnnotationRefused(test *testing.T) {
	test.Parallel()
	rulingRefusal(
		test,
		"function_length.a",
		false,
		"the Function annotation in .a; write a call signature with its parameters and result, like (value: number) => number",
	)
}
func TestFunctionTypeScriptUnusedAccepted(test *testing.T) {
	test.Parallel()
	rulingAgrees(test, "function_unused.a", true)
}
func TestFunctionTypeScriptLengthAgreesWithNode(test *testing.T) {
	test.Parallel()
	rulingAgrees(test, "function_length.a", true)
}
func TestFunctionTypeScriptCallRefused(test *testing.T) {
	test.Parallel()
	rulingRefusal(
		test,
		"function_call.a",
		true,
		"calling through the unchecked Function type; write a call signature with the actual parameters and result before calling",
	)
}
func TestFunctionCallSignatureAgreesWithNode(test *testing.T) {
	test.Parallel()
	rulingAgrees(test, "function_signature.a", false)
}

func rulingCount(test *testing.T, file string, typescript bool) (string, func() string) {
	test.Helper()
	program, path, rulingProgramAtError := rulingProgramAt(test, file, typescript)
	if rulingProgramAtError != nil {
		test.Fatal(rulingProgramAtError)
	}
	source := native.C(program)
	return fmt.Sprintf("typescript %v\n%s", typescript, countPortable(test, source, path)), func() string {
		return rulingCountRun(test, file, typescript, source)
	}
}

func rulingCountRun(test *testing.T, file string, typescript bool, source string) string {
	test.Helper()
	binary := filepath.Join(test.TempDir(), "counted")
	if buildError := native.Build(source, binary, native.Options{Count: true}); buildError != nil {
		test.Fatal(buildError)
	}
	name, args := pinnedStack(binary)
	result := execute(test, name, args...)
	match := countsLine.FindSubmatch(result.stderr)
	if match == nil {
		test.Fatalf("no counts: %#v", result)
	}
	return fmt.Sprintf(
		"| %s | %s | %s | %s | %s | %s | %s |",
		refusalFixtureRoot+file,
		match[1],
		match[2],
		match[3],
		match[4],
		match[5],
		match[6],
	)
}

func rulingNode(test *testing.T, file string) run {
	test.Helper()
	path, absolutePathError := filepath.Abs(filepath.Join(repository, refusalFixtureRoot, file))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	return onNode(test, path)
}

func TestFunctionAdamicCallAnnotationRefused(test *testing.T) {
	test.Parallel()
	rulingRefusal(
		test,
		"function_call.a",
		false,
		"the Function annotation in .a; write a call signature with its parameters and result, like (value: number) => number",
	)
}
