package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// These are backend primitives, not a claim that checked-view lowering is complete.
// The failures are pinned independently; all mutants produce valid C and harmless scalar values.
func TestRequiredViewFieldPrimitive(t *testing.T) {
	for _, test := range []struct {
		name   string
		fields []ir.Field
		wanted ir.Type
		want   run
	}{
		{"number", []ir.Field{{Name: "value", Value: ir.NumberConstant{Value: 7}}}, ir.Number, run{stdout: []byte("7\n")}},
		{"boolean", []ir.Field{{Name: "value", Value: ir.BooleanConstant{Value: true}}}, ir.Boolean, run{stdout: []byte("true\n")}},
		{"string", []ir.Field{{Name: "value", Value: ir.StringConstant{}}}, ir.String, run{stdout: []byte("name\n")}},
		{"wrong string", []ir.Field{{Name: "value", Value: ir.NumberConstant{}}}, ir.String, run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: node.value is not a string; expected string, found number\n")}},
		{"missing", nil, ir.Number, run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: node.value is not initialized; expected number, found missing\n")}},
		{"uninitialized", []ir.Field{{Name: "value", Value: ir.NumberConstant{}, Uninitialized: true}}, ir.Number, run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: node.value is not initialized; expected number, found uninitialized\n")}},
		{"wrong type", []ir.Field{{Name: "value", Value: ir.NumberConstant{}}}, ir.Boolean, run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: node.value is not a boolean; expected boolean, found number\n")}},
	} {
		t.Run(test.name, func(t *testing.T) {
			program := &ir.Program{Source: "view-field", Strings: []string{"name"}, Locals: []ir.Local{{Name: "node", Type: ir.Object, Function: -1}}}
			property := ir.Property{Object: ir.Read{Local: 0, Of: ir.Object}, Name: "value", Of: test.wanted, View: "node.value"}
			output := func(property ir.Property) ir.Statement {
				var value ir.Expression = ir.NumberToString{Value: property}
				if test.wanted == ir.Boolean {
					value = ir.BooleanToString{Value: property}
				}
				if test.wanted == ir.String {
					value = property
				}
				return ir.WriteLine{Stream: ir.Stdout, Value: value}
			}
			program.Main = []ir.Statement{ir.Declare{Local: 0, Value: ir.ObjectLiteral{Fields: test.fields}}, output(property)}
			if test.want.exitCode == 0 {
				source := "const node = {value: 7}; console.log(String(node.value));\n"
				if test.wanted == ir.Boolean {
					source = "const node = {value: true}; console.log(String(node.value));\n"
				}
				if test.wanted == ir.String {
					source = "const node = {value: \"name\"}; console.log(String(node.value));\n"
				}
				path := filepath.Join(t.TempDir(), "field.a")
				if err := os.WriteFile(path, []byte(source), 0600); err != nil {
					t.Fatal(err)
				}
				if difference := disagreement(test.want, onNode(t, path)); difference != "" {
					t.Fatal("Node source oracle: " + difference)
				}
			}
			javascript := onJavaScriptBackend(t, program)
			native, _ := nativelyUncached(t, program)
			for _, got := range []run{javascript, native, releasedUncached(t, program)} {
				if difference := disagreement(test.want, got); difference != "" {
					t.Fatalf("%s: want %#v, got %#v", difference, test.want, got)
				}
			}
			if test.name == "wrong type" || test.name == "uninitialized" {
				property.View = ""
				program.Main[1] = output(property)
				for _, got := range []run{onJavaScriptBackend(t, program), releasedUncached(t, program)} {
					if disagreement(test.want, got) == "" {
						t.Fatal("drop field check mutant escaped independent assertion")
					}
				}
				t.Log("drop field check mutant caught by pinned exit/message, with valid C")
			}
			if test.name == "uninitialized" {
				program.Main[1] = output(ir.Property{Object: property.Object, Name: "value", Of: ir.Number, View: "node.value"})
				fields := append([]ir.Field(nil), test.fields...)
				fields[0].Uninitialized = false
				program.Main[0] = ir.Declare{Local: 0, Value: ir.ObjectLiteral{Fields: fields}}
				for _, got := range []run{onJavaScriptBackend(t, program), releasedUncached(t, program)} {
					if disagreement(test.want, got) == "" {
						t.Fatal("drop initialization tracking mutant escaped independent assertion")
					}
				}
				t.Log("drop initialization tracking mutant caught by pinned exit/message")
			}
		})
	}
}

func TestRequiredViewFieldOperandOnce(t *testing.T) {
	counter := ir.Read{Local: 0, Of: ir.Number}
	program := &ir.Program{Source: "view-once", Locals: []ir.Local{{Name: "calls", Type: ir.Number, Global: true, Function: -1}}}
	program.Functions = []ir.Function{{Name: "make", Returns: ir.Object, Body: []ir.Statement{
		ir.Assign{Local: 0, Value: ir.Binary{Operator: ir.Add, Left: counter, Right: ir.NumberConstant{Value: 1}}},
		ir.Return{Value: ir.ObjectLiteral{Fields: []ir.Field{{Name: "value", Value: ir.NumberConstant{Value: 7}}}}},
	}}}
	program.Main = []ir.Statement{
		ir.Declare{Local: 0, Value: ir.NumberConstant{}},
		ir.WriteLine{Stream: ir.Stdout, Value: ir.NumberToString{Value: ir.Property{Object: ir.Call{Function: 0, Returns: ir.Object}, Name: "value", Of: ir.Number, View: "make().value"}}},
		ir.WriteLine{Stream: ir.Stdout, Value: ir.NumberToString{Value: counter}},
	}
	source := "let calls = 0; function make() { calls++; return {value: 7}; } console.log(String(make().value)); console.log(String(calls));\n"
	path := filepath.Join(t.TempDir(), "once.a")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	want := onNode(t, path)
	if want.exitCode != 0 || string(want.stdout) != "7\n1\n" {
		t.Fatalf("Node: %#v", want)
	}
	native, _ := nativelyUncached(t, program)
	for _, got := range []run{onJavaScriptBackend(t, program), native, releasedUncached(t, program)} {
		if difference := disagreement(want, got); difference != "" {
			t.Fatalf("operand evaluated more than once: %s; got %#v", difference, got)
		}
	}
}

// Unchanged storage sources now run as checked TypeScript; their initializers
// stop eagerly. The shared field readiness primitives remain exercised above.
func TestNarrowedFieldUsesSharedReadiness(t *testing.T) {
	for _, probe := range []struct {
		name, field, stdout string
		initialized         bool
	}{
		{"identifier", "escapedText", "identifier\nokok\n", true},
		{"identifier-uninitialized", "escapedText", "identifier\n", false},
		{"number", "value", "number\n0\n", true},
		{"number-uninitialized", "value", "number\n", false},
	} {
		t.Run(probe.name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "stage3/interface-downcasts/readiness-"+probe.name+".ts"))
			if err != nil {
				t.Fatal(err)
			}
			expression := "undefined!"
			if strings.HasPrefix(probe.name, "number") {
				expression = "null!"
			}
			assertMigratedNonNullCheck(t, path, expression, "", false)
		})
	}
}

func TestViewFieldInheritedStaticReadiness(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/non_null_static_initialized.ts"))
	if err != nil {
		t.Fatal(err)
	}
	assertMigratedNonNullCheck(t, path, "undefined!", "", false)
}

// These exercise real source casts and writes, without replacing any lowered IR.
func TestDefaultTaggedSourceViews(t *testing.T) {
	for _, probe := range []struct{ name, stdout, diagnostic string }{
		{"default-staged", "okok\n", ""},
		{"default-boxed-string", "okok\n", ""},
		{"default-boxed-write", "42\ntrue\n", ""},
		{"default-destructure", "okok\n", ""},
		{"default-wrong-type", "", "field read failed: (node as Identifier).name is not a string; expected string, found number"},
		{"default-wrong-boolean", "", "field read failed: (node as Identifier).ready is not a boolean; expected boolean, found number"},
		{"default-literal", "", "field read failed: (node as Identifier).name expected \"wanted\", found string other"},
		{"default-read-before-set", "", "field read failed: (held as Identifier).escapedText is not initialized; expected string, found uninitialized"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			if probe.name == "default-staged" || probe.name == "default-boxed-write" || probe.name == "default-read-before-set" {
				path, err := filepath.Abs(filepath.Join(repository, interfaceFixturePath(probe.name)))
				if err != nil {
					t.Fatal(err)
				}
				assertMigratedNonNullCheck(t, path, "undefined!", "", false)
				return
			}
			program, path := interfaceFixture(t, probe.name)
			want := run{stdout: []byte(probe.stdout)}
			if probe.diagnostic == "" {
				if difference := disagreement(want, onNode(t, path)); difference != "" {
					t.Fatal("Node: " + difference)
				}
			} else {
				want.exitCode = 70
				want.stderr = []byte("adamic: panic: " + probe.diagnostic + "\n")
			}
			actual, _ := nativelyUncached(t, program)
			for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Fatalf("%s; got %#v", difference, got)
				}
			}
		})
	}
}
