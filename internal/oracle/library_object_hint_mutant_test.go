package oracle

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

func objectHintNodeMutant(t *testing.T, hint string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "object-hints.a")
	const source = `let calls = 0;
const value = {
 toString: (): string => { calls++; return "text" + calls; },
 valueOf: (): number => { calls++; return 42; }
};
function right(): string { calls++; return "suffix"; }
console.log(String(value));
console.log(` + "`${value}`" + `);
console.log(value + right());
console.log(` + "`${calls}`" + `);
try { console.log(String({toString: (): never => { throw new Error("conversion"); }})); }
catch (error) { if (error instanceof Error) { console.log(error.message); } }
try { console.log(String({toString: (): {x:number} => ({x:1}), valueOf: (): {x:number} => ({x:2})})); }
catch (error) { if (error instanceof Error) { console.log(error.name); console.log(error.message); } }
`
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	expected := onNode(t, path)
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	correct, binary := natively(t, program)
	if difference := disagreement(expected, correct); difference != "" {
		t.Fatalf("original native: %s", difference)
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
	if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
		if difference := disagreement(expected, onWASI(t, native.C(program))); difference != "" {
			t.Fatalf("original WASI: %s", difference)
		}
	}
	changed := false
	for index := range program.Functions {
		function := &program.Functions[index]
		if function.Name != "library_string_ordinary_"+hint || len(function.Body) != 1 {
			continue
		}
		returned, ok := function.Body[0].(ir.Return)
		if !ok {
			continue
		}
		expression := returned.Value
		if hint == "default" {
			conversion, ok := expression.(ir.NumberToString)
			if !ok {
				continue
			}
			expression = conversion.Value
		}
		call, ok := expression.(ir.CallClosure)
		if !ok {
			continue
		}
		property, ok := call.Closure.(ir.Property)
		if !ok {
			continue
		}
		if hint == "string" {
			property.Name = "valueOf"
			call.Returns = ir.Number
		} else {
			property.Name = "toString"
			call.Returns = ir.String
		}
		call.Closure = property
		returned.Value = call
		if hint == "string" {
			returned.Value = ir.NumberToString{Value: call}
		}
		function.Body[0] = returned
		changed = true
		break
	}
	if !changed {
		t.Fatal("mutant changed nothing")
	}
	actual, binary := natively(t, program)
	if actual.exitCode != 0 || len(actual.stderr) != 0 {
		t.Fatalf("mutant failed outside comparison: %+v", actual)
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
	if difference := disagreement(expected, actual); difference != "stdout differs" {
		t.Fatalf("native mutant: %q", difference)
	}
	if difference := disagreement(expected, onJavaScriptBackend(t, program)); difference != "stdout differs" {
		t.Fatalf("JavaScript mutant: %q", difference)
	}
	if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
		if difference := disagreement(expected, onWASI(t, native.C(program))); difference != "stdout differs" {
			t.Fatalf("WASI mutant: %q", difference)
		}
	}
	t.Log("Node stdout alone catches the hint mutant; clean exit, stderr and leak checks")
}

func TestObjectStringHintNodeMutant(t *testing.T)  { t.Parallel(); objectHintNodeMutant(t, "string") }
func TestObjectDefaultHintNodeMutant(t *testing.T) { t.Parallel(); objectHintNodeMutant(t, "default") }
