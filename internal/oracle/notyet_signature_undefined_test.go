package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"path/filepath"
	"testing"
)

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{
		"internal/oracle/testdata/notyet_signature_undefined.a", true, false,
	})
}

func TestUndefinedSignatureMutant(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/notyet_signature_undefined.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	changed := false
	for index := range program.Functions {
		function := &program.Functions[index]
		if function.Name == "explicit" {
			for statement := range function.Body {
				if _, ok := function.Body[statement].(ir.Return); ok {
					function.Body[statement] = ir.Return{Value: ir.ObjectLiteral{}}
					changed = true
				}
			}
		}
	}
	if !changed {
		t.Fatal("mutant target absent")
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
	t.Log("undefined replaced by a present object caught by Node stdout in both backends")
}
