package oracle

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/generic_function_value.a", true, false})
}

// A specialized numeric body returning the wrong value must fail only Node's
// output comparison, independently in both backends.
func TestGenericFunctionValueWrongResultMutant(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/generic_function_value.a"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	changed := false
	for index, function := range program.Functions {
		if strings.HasPrefix(function.Name, "identity_") && !function.Closure && function.Returns == ir.Number {
			program.Functions[index].Body = []ir.Statement{ir.Return{Value: ir.NumberConstant{Value: -1}}}
			changed = true
		}
	}
	if !changed {
		test.Fatal("missing numeric instance")
	}
	want := onNode(test, path)
	native, _ := natively(test, program)
	for name, got := range map[string]run{"native": native, "JavaScript": onJavaScriptBackend(test, program)} {
		if got.exitCode != 0 || len(got.stderr) != 0 {
			test.Fatalf("%s mutant must finish cleanly: %+v", name, got)
		}
		if difference := disagreement(want, got); difference != "stdout differs" {
			test.Fatalf("%s mutant survived: %q", name, difference)
		}
		test.Logf("%s caught by stdout alone: Node %q; mutant %q", name, want.stdout, got.stdout)
	}
}
