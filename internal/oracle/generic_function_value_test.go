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
func TestGenericFunctionValueWrongResultMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/generic_function_value.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	changed := false
	for index, function := range program.Functions {
		if strings.HasPrefix(function.Name, "identity_") && !function.Closure && function.Returns == ir.Number {
			program.Functions[index].Body = []ir.Statement{ir.Return{Value: ir.NumberConstant{Value: -1}}}
			changed = true
		}
	}
	if !changed {
		t.Fatal("missing numeric instance")
	}
	want := onNode(t, path)
	native, _ := natively(t, program)
	for name, got := range map[string]run{"native": native, "JavaScript": onJavaScriptBackend(t, program)} {
		if got.exitCode != 0 || len(got.stderr) != 0 {
			t.Fatalf("%s mutant must finish cleanly: %+v", name, got)
		}
		if difference := disagreement(want, got); difference != "stdout differs" {
			t.Fatalf("%s mutant survived: %q", name, difference)
		}
		t.Logf("%s caught by stdout alone: Node %q; mutant %q", name, want.stdout, got.stdout)
	}
}
