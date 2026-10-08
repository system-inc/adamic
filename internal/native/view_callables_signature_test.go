package native

import (
	"bytes"
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestViewCallableProducerCertificateNative(t *testing.T) {
	header, err := runtime.ReadFile("runtime/view_callables_contract.h")
	if err != nil {
		t.Fatal(err)
	}
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
			expression := e.emitViewCallableCertificate(ir.Property{View: "node.run", ViewContract: 1, Absent: probe.absent}, "value")
			value := "closure"
			if probe.absent {
				value = "NULL"
			}
			source := string(header) + strings.Join(e.declarations, "\n") + `
static adamic_value adamic_function_0_callback(adamic_closure *self, adamic_value *arguments) {
 (void)self; (void)arguments; return (adamic_value){.number=8};
}
int main(void) {
 adamic_closure *closure=adamic_closure_new(adamic_function_0_callback,0);
 adamic_closure *value=` + value + ";\n" + e.out.String() + "adamic_closure *checked=" + expression + `;
 if (checked == NULL) puts("undefined"); else printf("%.0f\n", adamic_closure_call(checked,NULL,0).number);
 adamic_release(closure); return 0;
}
`
			for _, sanitize := range []bool{false, true} {
				binary := filepath.Join(t.TempDir(), "certificate")
				if err := Build(source, binary, Options{Sanitize: sanitize}); err != nil {
					t.Fatal(err)
				}
				command := exec.Command(binary)
				var stdout, stderr bytes.Buffer
				command.Stdout = &stdout
				command.Stderr = &stderr
				err := command.Run()
				exit := 0
				if err != nil {
					if failure, ok := err.(*exec.ExitError); ok {
						exit = failure.ExitCode()
					} else {
						t.Fatal(err)
					}
				}
				wantOut := "8\n"
				if probe.absent {
					wantOut = "undefined\n"
				}
				wantErr := ""
				wantExit := 0
				if probe.found != "" {
					wantOut = ""
					wantExit = 70
					wantErr = fmt.Sprintf("adamic: panic: field read failed: node.run expected (value: number) => number, found %s\n", probe.found)
				}
				if exit != wantExit || stdout.String() != wantOut || stderr.String() != wantErr {
					t.Fatalf("exit%d out%q err%q", exit, stdout.String(), stderr.String())
				}
			}
		})
	}
}
