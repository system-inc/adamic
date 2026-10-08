package oracle

import (
	"errors"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/lower"
)

func TestScannerStringConstructorView(t *testing.T) {
	for _, test := range []struct{ name, output string }{{"string-any", "A\n"}, {"string-checked", "🙂\n1\n"}} {
		t.Run(test.name, func(t *testing.T) {
			program, path := scannerCastFixture(t, test.name)
			truth := onNode(t, path)
			if truth.exitCode != 0 || string(truth.stdout) != test.output {
				t.Fatalf("source Node: %#v", truth)
			}
			sanitized, binary := nativelyUncached(t, program)
			for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if diff := disagreement(truth, got); diff != "" {
					t.Fatalf("%s: %#v", diff, got)
				}
			}
			if report := leaks(t, program, binary); report != "" {
				t.Fatal(report)
			}
		})
	}
}

func TestScannerErrorLibraryCastNotYet(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "stage3/drivers/scanner/cast-checks/error-library-cast.a"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "ok\n" {
		t.Fatalf("source Node: %#v", truth)
	}
	_, err = lowered(t, path)
	if err == nil {
		t.Fatal("Error library cast gap closed; replace this NotYet pin with both-backend execution and counts")
	}
	var missing *lower.NotYet
	if !errors.As(err, &missing) || missing.Where != path+":3:1" || missing.What != "node:globals.ErrorConstructor.captureStackTrace" {
		t.Fatalf("want exact library NotYet, got %v", err)
	}
	t.Logf("library #ddwcejg dependency: %v", missing)
}

// Substitute a callable with the same ABI but without the intrinsic code identity.
// This tests the host producer boundary after lowering, independently of source
// restrictions preventing mutation of the intrinsic.
func TestScannerStringConstructorProducerMutant(t *testing.T) {
	program, _ := scannerCastFixture(t, "string-checked")
	function := len(program.Functions)
	local := len(program.Locals)
	program.Locals = append(program.Locals, ir.Local{Name: "point", Type: ir.Number, Function: function})
	constant := len(program.Strings)
	program.Strings = append(program.Strings, "liar")
	program.Functions = append(program.Functions, ir.Function{Name: "imposter", Closure: true, Parameters: []int{local}, Returns: ir.String, Body: []ir.Statement{ir.Return{Value: ir.StringConstant{Index: constant}}}})
	changed := 0
	substitute := func(value ir.Expression) ir.Expression {
		property, ok := value.(ir.Property)
		if ok && property.Name == "fromCodePoint" && property.View != "" {
			carrier, ok := property.Object.(ir.ObjectLiteral)
			if !ok || len(carrier.Fields) != 1 {
				t.Fatal("intrinsic carrier seam changed")
			}
			carrier.Fields[0].Value = ir.MakeClosure{Function: function}
			property.Object = carrier
			changed++
			return property
		}
		return value
	}
	mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), substitute)
	if changed != 1 {
		t.Fatalf("want one substituted host producer, got %d", changed)
	}
	expected := run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: (String as any).fromCodePoint expected (codePoint: number) => string, found function with unknown signature\n")}
	if got := onJavaScriptBackend(t, program); viewReadDisagreement(expected, got, program) != "" {
		t.Fatalf("imposter must match pinned stop: %#v", got)
	}
	t.Logf("producer liar pin: %q", expected.stderr)
	sanitized, _ := nativelyUncached(t, program)
	for _, got := range []run{sanitized, releasedUncached(t, program)} {
		if diff := viewReadDisagreement(expected, got, program); diff != "" {
			t.Fatalf("liar: %s: %#v", diff, got)
		}
	}
	omit := func(value ir.Expression) ir.Expression {
		if property, ok := value.(ir.Property); ok && property.Name == "fromCodePoint" {
			property.View = ""
			property.ViewContract = 0
			return property
		}
		return value
	}
	mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), omit)
	for backend, got := range map[string]run{"native release": releasedUncached(t, program), "JavaScript": onJavaScriptBackend(t, program)} {
		if got.exitCode != 0 || string(got.stdout) != "liar\n1\n" {
			t.Fatalf("%s unchecked producer must execute successfully: %#v", backend, got)
		}
		if viewReadDisagreement(expected, got, program) == "" {
			t.Fatal("producer-check omission survived")
		}
		t.Logf("%s omission caught: exit=%d stdout=%q", backend, got.exitCode, got.stdout)
	}
}
