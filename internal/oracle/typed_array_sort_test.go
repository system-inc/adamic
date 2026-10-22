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

func typedArraySortFixture(t *testing.T) (*ir.Program, run) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/typed_arrays_sort_uint16array.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	expected := onNode(t, path)
	if expected.exitCode != 0 || len(expected.stderr) != 0 {
		t.Fatalf("Node: %+v", expected)
	}
	return program, expected
}

func TestTypedArraySortNativeCallMutant(t *testing.T) {
	t.Parallel()
	program, expected := typedArraySortFixture(t)
	control := released(t, program)
	if difference := disagreement(expected, control); difference != "" {
		t.Fatal("control: " + difference)
	}
	call := regexp.MustCompile(`adamic_typed_array_sort\(([^;\n]+)\);`)
	source := native.C(program)
	if !call.MatchString(source) {
		t.Fatal("missing native sort call")
	}
	source = call.ReplaceAllString(source, "(void)($1);")
	binary := filepath.Join(t.TempDir(), "without-sort")
	if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal("mutant must compile:", err)
	}
	leaks := "0"
	if runtime.GOOS == "linux" {
		leaks = "1"
	}
	actual := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=" + leaks + ":halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
	if actual.exitCode != 0 || len(actual.stderr) != 0 || bytes.Equal(actual.stdout, expected.stdout) {
		t.Fatalf("native call mutant must finish and disagree with Node: %+v", actual)
	}
	t.Log("omitted native sort caught by Node stdout difference")
}

func TestTypedArraySortJavaScriptCallMutant(t *testing.T) {
	t.Parallel()
	program, expected := typedArraySortFixture(t)
	control := onJavaScriptBackend(t, program)
	if difference := disagreement(expected, control); difference != "" {
		t.Fatal("control: " + difference)
	}
	source := javascript.JavaScript(program)
	if !strings.Contains(source, ".sort()") {
		t.Fatal("missing JavaScript sort call")
	}
	source = strings.ReplaceAll(source, ".sort()", "")
	path := filepath.Join(t.TempDir(), "without-sort.mjs")
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	actual := onNode(t, path)
	if actual.exitCode != 0 || len(actual.stderr) != 0 || bytes.Equal(actual.stdout, expected.stdout) {
		t.Fatalf("JavaScript call mutant must finish and disagree with Node: %+v", actual)
	}
	t.Log("omitted JavaScript sort caught by Node stdout difference")
}
