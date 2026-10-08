package oracle

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/notyet_scanner_generic_return.a", true, false})
}

// The scanner witness must preserve the callback result in its concrete instance.
func TestScannerGenericReturnMutant(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/notyet_scanner_generic_return.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	wrong := len(program.Strings)
	program.Strings = append(program.Strings, "wrong")
	changed := false
	for i := range program.Functions {
		function := &program.Functions[i]
		if strings.HasPrefix(function.Name, "each_") && function.Returns == ir.String {
			function.Body = []ir.Statement{ir.Return{Value: ir.StringConstant{Index: wrong}}}
			changed = true
		}
	}
	if !changed {
		t.Fatal("missing concrete scanner return mutant target")
	}
	want := onNode(t, path)
	got, binary := natively(t, program)
	if got.exitCode != 0 || len(got.stderr) != 0 {
		t.Fatalf("mutant must finish cleanly: %+v", got)
	}
	if difference := disagreement(want, got); difference != "stdout differs" {
		t.Fatalf("native mutant survived: %q", difference)
	}
	if difference := disagreement(want, onJavaScriptBackend(t, program)); difference != "stdout differs" {
		t.Fatalf("JavaScript mutant survived: %q", difference)
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
	t.Log("concrete scanner return mutant caught by Node stdout in both backends")
}
