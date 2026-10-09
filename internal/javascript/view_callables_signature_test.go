package javascript

import (
	"github.com/system-inc/adamic/internal/ir"
	"testing"
)

func TestViewCallableProducerCertificateNode(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name   string
		arity  int
		result ir.Type
		absent bool
		found  string
	}{
		{"valid", 1, ir.Number, false, ""}, {"wrong-arity", 0, ir.Number, false, "function with arity 0"},
		{"wrong-result", 1, ir.Boolean, false, "function with incompatible result representation"},
		{"optional-absent", 1, ir.Number, true, ""},
		{"unrecorded", 1, ir.Number, false, "function with unknown signature"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			program := &ir.Program{Locals: []ir.Local{{Type: ir.Number}}, Functions: []ir.Function{{Name: "callback", Closure: true, Returns: probe.result}}, ViewContracts: []ir.ViewContract{
				{Kind: ir.ViewCallable, Name: "(value: number) => number", Parameters: []ir.ViewContractID{2}, Result: 3}, {Kind: ir.ViewScalar, Of: ir.Number}, {Kind: ir.ViewScalar, Of: ir.Number},
			}}
			if probe.name == "unrecorded" {
				program.Functions[0].Closure = false
			}
			if probe.arity == 1 {
				program.Functions[0].Parameters = []int{0}
			}
			e := &emitter{program: program}
			value := "new AdamicClosure(function_0_callback)"
			if probe.absent {
				value = "undefined"
			}
			checked := e.emitViewCallableCertificate(ir.Property{View: "node.run", ViewContract: 1, Absent: probe.absent}, value)
			source := viewTestRuntime + viewCallableShapeRuntime + "\nfunction function_0_callback(self, values) { return 8; }\nconst checked=" + checked + ";\nconsole.log(checked === undefined ? 'undefined' : adamicCall(checked, []));\n"
			out, stderr, exit := "8\n", "", 0
			if probe.absent {
				out = "undefined\n"
			}
			if probe.found != "" {
				out = ""
				exit = 70
				stderr = "adamic: panic: field read failed: node.run expected (value: number) => number, found " + probe.found + "\n"
			}
			runViewNode(t, source, out, stderr, exit)
		})
	}
}
