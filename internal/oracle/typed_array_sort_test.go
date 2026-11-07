package oracle

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
)

func typedArraySortFixture(test *testing.T) (*ir.Program, run) {
	test.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/typed_arrays_sort_uint16array.a"))
	if err != nil {
		test.Fatal(err)
	}
	program, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
	}
	expected := onNode(test, path)
	if expected.exitCode != 0 || len(expected.stderr) != 0 {
		test.Fatalf("Node: %+v", expected)
	}
	return program, expected
}

func TestTypedArraySortNativeCallMutant(test *testing.T) {
	test.Parallel()
	program, expected := typedArraySortFixture(test)
	control := released(test, program)
	if difference := disagreement(expected, control); difference != "" {
		test.Fatal("control: " + difference)
	}
	call := regexp.MustCompile(`adamic_typed_array_sort\(([^;\n]+)\);`)
	source := native.C(program)
	if !call.MatchString(source) {
		test.Fatal("missing native sort call")
	}
	source = call.ReplaceAllString(source, "(void)($1);")
	binary := filepath.Join(test.TempDir(), "without-sort")
	if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
		test.Fatal("mutant must compile:", err)
	}
	leaks := "0"
	if runtime.GOOS == "linux" {
		leaks = "1"
	}
	actual := executeWith(
		test,
		[]string{"ASAN_OPTIONS=detect_leaks=" + leaks + ":halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1"},
		binary,
	)
	if actual.exitCode != 0 || len(actual.stderr) != 0 || bytes.Equal(actual.stdout, expected.stdout) {
		test.Fatalf("native call mutant must finish and disagree with Node: %+v", actual)
	}
	test.Log("omitted native sort caught by Node stdout difference")
}

func TestTypedArraySortJavaScriptCallMutant(test *testing.T) {
	test.Parallel()
	program, expected := typedArraySortFixture(test)
	control := onJavaScriptBackend(test, program)
	if difference := disagreement(expected, control); difference != "" {
		test.Fatal("control: " + difference)
	}
	source := javascript.JavaScript(program)
	if !strings.Contains(source, ".sort()") {
		test.Fatal("missing JavaScript sort call")
	}
	source = strings.ReplaceAll(source, ".sort()", "")
	path := filepath.Join(test.TempDir(), "without-sort.mjs")
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		test.Fatal(err)
	}
	actual := onNode(test, path)
	if actual.exitCode != 0 || len(actual.stderr) != 0 || bytes.Equal(actual.stdout, expected.stdout) {
		test.Fatalf("JavaScript call mutant must finish and disagree with Node: %+v", actual)
	}
	test.Log("omitted JavaScript sort caught by Node stdout difference")
}
