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

func checkCallableUnionResult(t *testing.T, name, output string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/callable_union_result_"+name+".a"))
	if err != nil {
		t.Fatal(err)
	}
	want := onNode(t, path)
	if string(want.stdout) != output || want.exitCode != 0 {
		t.Fatalf("source Node: %#v", want)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	actual, _ := nativelyUncached(t, program)
	for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, got); difference != "" {
			t.Fatalf("%s: want %#v, got %#v", difference, want, got)
		}
	}
}

func TestCallableUnionResultNumber(t *testing.T) {
	t.Parallel()
	checkCallableUnionResult(t, "number", "6\n")
}

func TestCallableUnionResultString(t *testing.T) {
	t.Parallel()
	checkCallableUnionResult(t, "string", "value:5\n")
}

func TestCallableUnionResultZero(t *testing.T) {
	t.Parallel()
	checkCallableUnionResult(t, "zero", "0\n")
}

func TestCallableUnionResultBoth(t *testing.T) {
	t.Parallel()
	checkCallableUnionResult(t, "both", "6\nvalue:5\n6\n6\n")
}
