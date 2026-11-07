package oracle

import (
	"path/filepath"
	"testing"
)

func init() {
	for _, name := range []string{"number", "string", "zero", "both"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{
			"internal/oracle/testdata/callable_union_result_" + name + ".a", true, false,
		})
	}
}

func checkCallableUnionResult(test *testing.T, name, output string) {
	test.Helper()
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/callable_union_result_"+name+".a"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	want := onNode(test, path)
	if string(want.stdout) != output || want.exitCode != 0 {
		test.Fatalf("source Node: %#v", want)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	actual, _ := nativelyUncached(test, program)
	for _, got := range []run{actual, releasedUncached(test, program), onJavaScriptBackend(test, program)} {
		if difference := disagreement(want, got); difference != "" {
			test.Fatalf("%s: want %#v, got %#v", difference, want, got)
		}
	}
}

func TestCallableUnionResultNumber(test *testing.T) {
	test.Parallel()
	checkCallableUnionResult(test, "number", "6\n")
}

func TestCallableUnionResultString(test *testing.T) {
	test.Parallel()
	checkCallableUnionResult(test, "string", "value:5\n")
}

func TestCallableUnionResultZero(test *testing.T) {
	test.Parallel()
	checkCallableUnionResult(test, "zero", "0\n")
}

func TestCallableUnionResultBoth(test *testing.T) {
	test.Parallel()
	checkCallableUnionResult(test, "both", "6\nvalue:5\n6\n6\n")
}
