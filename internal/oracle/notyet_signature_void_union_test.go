package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"path/filepath"
	"testing"
)

func init() {
	for _, path := range []string{
		"internal/oracle/testdata/notyet_signature_void_union.a",
		"internal/oracle/testdata/notyet_signature_void_references.a",
	} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path, true, false})
	}
}

func TestVoidUnionSignatureMutant(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/notyet_signature_void_union.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	skip := -1
	for index, text := range program.Strings {
		if text == "skip" {
			skip = index
		}
	}
	if skip < 0 {
		t.Fatal("missing skip constant")
	}
	changed := false
	for index := range program.Functions {
		function := &program.Functions[index]
		if function.Name == "visit" {
			for statement := range function.Body {
				if returned, ok := function.Body[statement].(ir.Return); ok {
					if _, absent := returned.Value.(ir.Undefined); absent {
						function.Body[statement] = ir.Return{Value: ir.StringConstant{Index: skip}}
						changed = true
					}
				}
			}
		}
	}
	if !changed {
		t.Fatal("implicit return mutant target absent")
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
	t.Log("implicit void union result replaced by skip caught by Node stdout in both backends")
}

func TestVoidUnionRepresentationMutants(t *testing.T) {
	for _, name := range []string{"builder", "mixed", "absent"} {
		t.Run(name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/notyet_signature_void_references.a"))
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
				if function.Name != name {
					continue
				}
				var value ir.Expression = ir.Undefined{Of: function.Returns}
				if name == "mixed" {
					value = ir.Box{Value: ir.BooleanConstant{Value: true}}
				}
				if name == "absent" {
					value = ir.ObjectLiteral{}
				}
				function.Body = []ir.Statement{ir.Return{Value: value}}
				changed = true
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
			t.Logf("%s representation mutant caught by Node stdout in both backends", name)
		})
	}
}
