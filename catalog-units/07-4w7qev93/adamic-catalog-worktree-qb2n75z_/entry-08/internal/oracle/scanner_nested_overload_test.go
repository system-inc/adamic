package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"path/filepath"
	"testing"
)

func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/scanner_nested_overload.a", true, false})
}

func TestScannerNestedOverloadImplementationMutantIsCaught(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/scanner_nested_overload.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	changed := 0
	empty := len(program.Strings)
	program.Strings = append(program.Strings, "")
	for i := range program.Functions {
		function := &program.Functions[i]
		if function.Name != "inner" {
			continue
		}
		for j, statement := range function.Body {
			result, ok := statement.(ir.Return)
			if !ok {
				continue
			}
			result.Value = ir.StringConstant{Index: empty}
			function.Body[j] = result
			changed++
		}
	}
	if changed != 1 {
		t.Fatalf("want exactly one implementation return, changed %d", changed)
	}
	expected := onNode(t, path)
	actual, binary := nativelyUncached(t, program)
	if actual.exitCode != 0 || len(actual.stderr) != 0 {
		t.Fatalf("want clean mutant execution, got %+v", actual)
	}
	if report := leaksUncached(t, program, binary); report != "" {
		t.Fatal(report)
	}
	if difference := disagreement(expected, actual); difference != "stdout differs" {
		t.Fatalf("want Node stdout catcher, got %q", difference)
	}
	t.Logf("implementation mutant caught by Node: native %q, Node %q", actual.stdout, expected.stdout)
}
