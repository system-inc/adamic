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

func TestScannerNestedOverloadImplementationMutantIsCaught(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/scanner_nested_overload.a"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
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
		test.Fatalf("want exactly one implementation return, changed %d", changed)
	}
	expected := onNode(test, path)
	actual, binary := nativelyUncached(test, program)
	if actual.exitCode != 0 || len(actual.stderr) != 0 {
		test.Fatalf("want clean mutant execution, got %+v", actual)
	}
	if report := leaksUncached(test, program, binary); report != "" {
		test.Fatal(report)
	}
	if difference := disagreement(expected, actual); difference != "stdout differs" {
		test.Fatalf("want Node stdout catcher, got %q", difference)
	}
	test.Logf("implementation mutant caught by Node: native %q, Node %q", actual.stdout, expected.stdout)
}
