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
type requiredViewProbe struct {
	name   string
	fields []ir.Field
	wanted ir.Type
	want   run
}

func TestRequiredViewFieldPrimitiveNumber(test *testing.T) {
	test.Parallel()
	assertRequiredViewFieldPrimitive(test, requiredViewProbe{
		"number",
		[]ir.Field{{Name: "value", Value: ir.NumberConstant{Value: 7}}},
		ir.Number,
		run{stdout: []byte("7\n")},
	})
}

func TestRequiredViewFieldPrimitiveBoolean(test *testing.T) {
	test.Parallel()
	assertRequiredViewFieldPrimitive(test, requiredViewProbe{
		"boolean",
		[]ir.Field{{Name: "value", Value: ir.BooleanConstant{Value: true}}},
		ir.Boolean,
		run{stdout: []byte("true\n")},
	})
}

func TestRequiredViewFieldPrimitiveString(test *testing.T) {
	test.Parallel()
	assertRequiredViewFieldPrimitive(test, requiredViewProbe{
		"string",
		[]ir.Field{{Name: "value", Value: ir.StringConstant{}}},
		ir.String,
		run{stdout: []byte("name\n")},
	})
}

func TestRequiredViewFieldPrimitiveWrongString(test *testing.T) {
	test.Parallel()
	assertRequiredViewFieldPrimitive(test, requiredViewProbe{
		"wrong string",
		[]ir.Field{{Name: "value", Value: ir.NumberConstant{}}},
		ir.String,
		run{exitCode: 70, stderr: []byte(
			"adamic: panic: field read failed: node.value is not a string; expected string, found number\n",
		)},
	})
}

func TestRequiredViewFieldPrimitiveMissing(test *testing.T) {
	test.Parallel()
	assertRequiredViewFieldPrimitive(
		test,
		requiredViewProbe{"missing", nil, ir.Number, run{exitCode: 70, stderr: []byte(
			"adamic: panic: field read failed: node.value is not initialized; expected number, found missing\n",
		)}},
	)
}

func TestRequiredViewFieldPrimitiveUninitialized(test *testing.T) {
	test.Parallel()
	assertRequiredViewFieldPrimitive(test, requiredViewProbe{
		"uninitialized",
		[]ir.Field{{Name: "value", Value: ir.NumberConstant{}, Uninitialized: true}},
		ir.Number,
		run{exitCode: 70, stderr: []byte(
			"adamic: panic: field read failed: node.value is not initialized; expected number, found uninitialized\n",
		)},
	})
}

func TestRequiredViewFieldPrimitiveWrongType(test *testing.T) {
	test.Parallel()
	assertRequiredViewFieldPrimitive(test, requiredViewProbe{
		"wrong type",
		[]ir.Field{{Name: "value", Value: ir.NumberConstant{}}},
		ir.Boolean,
		run{exitCode: 70, stderr: []byte(
			"adamic: panic: field read failed: node.value is not a boolean; expected boolean, found number\n",
		)},
	})
}

func assertRequiredViewFieldPrimitive(t *testing.T, test requiredViewProbe) {
	t.Helper()
	program := &ir.Program{
		Source:  "view-field",
		Strings: []string{"name"},
		Locals:  []ir.Local{{Name: "node", Type: ir.Object, Function: -1}},
	}
	property := ir.Property{
		Object: ir.Read{Local: 0, Of: ir.Object},
		Name:   "value",
		Of:     test.wanted,
		View:   "node.value",
	}
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
		if writeError := os.WriteFile(path, []byte(source), 0600); writeError != nil {
			t.Fatal(writeError)
		}
		if difference := disagreement(test.want, onSliceNode(t, path)); difference != "" {
			t.Fatal("Node source oracle: " + difference)
		}
	}
	javascript := onSliceJavaScript(t, program)
	native, _ := nativelyUncached(t, program)
	for _, got := range []run{javascript, native, releasedUncached(t, program)} {
		if difference := disagreement(test.want, got); difference != "" {
			t.Fatalf("%s: want %#v, got %#v", difference, test.want, got)
		}
	}
	if test.name == "wrong type" || test.name == "uninitialized" {
		property.View = ""
		program.Main[1] = output(property)
		for _, got := range []run{onSliceJavaScript(t, program), releasedUncached(t, program)} {
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
		for _, got := range []run{onSliceJavaScript(t, program), releasedUncached(t, program)} {
			if disagreement(test.want, got) == "" {
				t.Fatal("drop initialization tracking mutant escaped independent assertion")
			}
		}
		t.Log("drop initialization tracking mutant caught by pinned exit/message")
	}
}

func TestRequiredViewFieldOperandOnce(test *testing.T) {
	test.Parallel()
	counter := ir.Read{Local: 0, Of: ir.Number}
	program := &ir.Program{
		Source: "view-once",
		Locals: []ir.Local{{Name: "calls", Type: ir.Number, Global: true, Function: -1}},
	}
	program.Functions = []ir.Function{{Name: "make", Returns: ir.Object, Body: []ir.Statement{
		ir.Assign{Local: 0, Value: ir.Binary{Operator: ir.Add, Left: counter, Right: ir.NumberConstant{Value: 1}}},
		ir.Return{Value: ir.ObjectLiteral{Fields: []ir.Field{{Name: "value", Value: ir.NumberConstant{Value: 7}}}}},
	}}}
	program.Main = []ir.Statement{
		ir.Declare{Local: 0, Value: ir.NumberConstant{}},
		ir.WriteLine{Stream: ir.Stdout, Value: ir.NumberToString{Value: ir.Property{
			Object: ir.Call{Function: 0, Returns: ir.Object},
			Name:   "value",
			Of:     ir.Number,
			View:   "make().value",
		}}},
		ir.WriteLine{Stream: ir.Stdout, Value: ir.NumberToString{Value: counter}},
	}
	source := "let calls = 0; function make() { calls++; return {value: 7}; } console.log(String(make().value)); console.log(String(calls));\n"
	path := filepath.Join(test.TempDir(), "once.a")
	if writeError := os.WriteFile(path, []byte(source), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	want := onSliceNode(test, path)
	if want.exitCode != 0 || string(want.stdout) != "7\n1\n" {
		test.Fatalf("Node: %#v", want)
	}
	native, _ := nativelyUncached(test, program)
	for _, got := range []run{onSliceJavaScript(test, program), native, releasedUncached(test, program)} {
		if difference := disagreement(want, got); difference != "" {
			test.Fatalf("operand evaluated more than once: %s; got %#v", difference, got)
		}
	}
}

// Placeholder-backed fields share tagged storage across their narrowed views.
type narrowedReadinessProbe struct {
	name, field, stdout string
	initialized         bool
}

func TestNarrowedFieldUsesSharedReadinessIdentifier(test *testing.T) {
	test.Parallel()
	assertNarrowedFieldUsesSharedReadiness(
		test,
		narrowedReadinessProbe{"identifier", "escapedText", "identifier\nokok\n", true},
	)
}

func TestNarrowedFieldUsesSharedReadinessIdentifierUninitialized(test *testing.T) {
	test.Parallel()
	assertNarrowedFieldUsesSharedReadiness(
		test,
		narrowedReadinessProbe{"identifier-uninitialized", "escapedText", "identifier\n", false},
	)
}

func TestNarrowedFieldUsesSharedReadinessNumber(test *testing.T) {
	test.Parallel()
	assertNarrowedFieldUsesSharedReadiness(test, narrowedReadinessProbe{"number", "value", "number\n0\n", true})
}

func TestNarrowedFieldUsesSharedReadinessNumberUninitialized(test *testing.T) {
	test.Parallel()
	assertNarrowedFieldUsesSharedReadiness(
		test,
		narrowedReadinessProbe{"number-uninitialized", "value", "number\n", false},
	)
}

func assertNarrowedFieldUsesSharedReadiness(test *testing.T, probe narrowedReadinessProbe) {
	test.Helper()
	program, path := interfaceFixture(test, "readiness-"+probe.name)
	want := run{stdout: []byte(probe.stdout)}
	if probe.initialized {
		if difference := disagreement(want, onSliceNode(test, path)); difference != "" {
			test.Fatal(difference)
		}
	} else {
		want.exitCode = 70
		use := "argument"
		if strings.HasPrefix(probe.name, "number") {
			use = "use as T"
		}
		want.stderr = []byte(
			"adamic: panic: placeholder '" + probe.field + "' is unset at " + use + " via node." + probe.field + "\n",
		)
	}
	native, _ := nativelyUncached(test, program)
	for _, got := range []run{native, onSliceJavaScript(test, program)} {
		if difference := disagreement(want, got); difference != "" {
			test.Fatal(difference)
		}
	}
}

func TestViewFieldInheritedStaticReadiness(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/non_null_static_initialized.ts"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	want := onSliceNode(test, path)
	native, _ := nativelyUncached(test, program)
	for _, got := range []run{native, onSliceJavaScript(test, program)} {
		if difference := disagreement(want, got); difference != "" {
			test.Fatal(difference)
		}
	}
}

// These exercise real source casts and writes, without replacing any lowered IR.
type defaultTaggedProbe struct{ name, stdout, diagnostic string }

func TestDefaultTaggedSourceViewsDefaultStaged(test *testing.T) {
	test.Parallel()
	assertDefaultTaggedSourceViews(test, defaultTaggedProbe{"default-staged", "okok\n", ""})
}

func TestDefaultTaggedSourceViewsDefaultBoxedString(test *testing.T) {
	test.Parallel()
	assertDefaultTaggedSourceViews(test, defaultTaggedProbe{"default-boxed-string", "okok\n", ""})
}

func TestDefaultTaggedSourceViewsDefaultBoxedWrite(test *testing.T) {
	test.Parallel()
	assertDefaultTaggedSourceViews(test, defaultTaggedProbe{"default-boxed-write", "42\ntrue\n", ""})
}

func TestDefaultTaggedSourceViewsDefaultDestructure(test *testing.T) {
	test.Parallel()
	assertDefaultTaggedSourceViews(test, defaultTaggedProbe{"default-destructure", "okok\n", ""})
}

func TestDefaultTaggedSourceViewsDefaultWrongType(test *testing.T) {
	test.Parallel()
	assertDefaultTaggedSourceViews(test, defaultTaggedProbe{
		"default-wrong-type",
		"",
		"field read failed: (node as Identifier).name is not a string; expected string, found number",
	})
}

func TestDefaultTaggedSourceViewsDefaultWrongBoolean(test *testing.T) {
	test.Parallel()
	assertDefaultTaggedSourceViews(test, defaultTaggedProbe{
		"default-wrong-boolean",
		"",
		"field read failed: (node as Identifier).ready is not a boolean; expected boolean, found number",
	})
}

func TestDefaultTaggedSourceViewsDefaultLiteral(test *testing.T) {
	test.Parallel()
	assertDefaultTaggedSourceViews(test, defaultTaggedProbe{
		"default-literal",
		"",
		"field read failed: (node as Identifier).name expected \"wanted\", found string other",
	})
}

func TestDefaultTaggedSourceViewsDefaultReadBeforeSet(test *testing.T) {
	test.Parallel()
	assertDefaultTaggedSourceViews(test, defaultTaggedProbe{
		"default-read-before-set",
		"",
		"placeholder 'escapedText' is unset at argument via (held as Identifier).escapedText",
	})
}

func assertDefaultTaggedSourceViews(test *testing.T, probe defaultTaggedProbe) {
	test.Helper()

	program, path := interfaceFixture(test, probe.name)
	want := run{stdout: []byte(probe.stdout)}
	if probe.diagnostic == "" {
		if difference := disagreement(want, onSliceNode(test, path)); difference != "" {
			test.Fatal("Node: " + difference)
		}
	} else {
		want.exitCode = 70
		want.stderr = []byte("adamic: panic: " + probe.diagnostic + "\n")
	}
	actual, _ := nativelyUncached(test, program)
	for _, got := range []run{actual, releasedUncached(test, program), onSliceJavaScript(test, program)} {
		if difference := disagreement(want, got); difference != "" {
			test.Fatalf("%s; got %#v", difference, got)
		}
	}
}
