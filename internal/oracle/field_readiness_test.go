package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"testing"
)

// Pin the inserted check independently of either backend. Zero is a present value only after a write.
func TestFieldReadinessRepresentation(test *testing.T) {
	test.Parallel()
	for _, assigned := range []bool{false, true} {
		program := &ir.Program{
			Source: "field-readiness",
			Locals: []ir.Local{{Name: "object", Type: ir.Object, Function: -1}},
		}
		object := ir.Read{Local: 0, Of: ir.Object}
		program.Main = []ir.Statement{ir.Declare{Local: 0, Value: ir.ObjectLiteral{
			Fields: []ir.Field{{Name: "value", Value: ir.NumberConstant{}, Uninitialized: true}},
		}}}
		if assigned {
			program.Main = append(
				program.Main,
				ir.SetProperty{Object: object, Name: "value", Value: ir.NumberConstant{}},
			)
		}
		program.Main = append(program.Main, ir.WriteLine{Stream: ir.Stdout, Value: ir.NumberToString{
			Value: ir.Property{Object: object, Name: "value", Of: ir.Number, Readiness: "object.value"},
		}})
		want := run{
			exitCode: 70,
			stderr:   []byte("adamic: panic: read before assignment: field 'value' in object.value\n"),
		}
		if assigned {
			want = run{stdout: []byte("0\n")}
		}
		javascript := onJavaScriptBackend(test, program)
		actual, binary := nativelyUncached(test, program)
		for _, got := range []run{javascript, actual} {
			if difference := disagreement(want, got); difference != "" {
				test.Fatal(difference)
			}
		}
		if assigned {
			if report := leaksUncached(test, program, binary); report != "" {
				test.Fatal(report)
			}
		}
		if !assigned {
			line := program.Main[1].(ir.WriteLine)
			line.Value = ir.NumberToString{Value: ir.Property{Object: object, Name: "value", Of: ir.Number}}
			program.Main[1] = line
			mutant, _ := nativelyUncached(test, program)
			if difference := disagreement(want, mutant); difference == "" {
				test.Fatal("dropping readiness check escaped pinned assertion")
			}
			test.Log("drop field check mutant caught by pinned exit and message")
		}
	}
}
