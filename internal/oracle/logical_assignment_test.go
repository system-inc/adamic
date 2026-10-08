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
	}{"internal/oracle/testdata/logical_or_assignment.a", true, false})
}

// Mutants preserve valid IR and ownership. Each must compile and finish cleanly, then disagree
// with source Node in both backends on the side-effect counters or expression value.
func TestLogicalAssignmentMutants(t *testing.T) {
	for _, operator := range []string{"or"} {
		for _, mutation := range []string{"target_twice", "right_always"} {
			t.Run(operator+"/"+mutation, func(t *testing.T) {
				path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/logical_"+operator+"_assignment.a"))
				if err != nil {
					t.Fatal(err)
				}
				program, err := lowered(t, path)
				if err != nil {
					t.Fatal(err)
				}
				changed := 0
				for index := range program.Functions {
					function := &program.Functions[index]
					if function.Name != "logical_"+operator+"_assignment" {
						continue
					}
					body := function.Body
					for at, statement := range body {
						if mutation == "target_twice" {
							declaration, ok := statement.(ir.Declare)
							if !ok || (program.Locals[declaration.Local].Name != "assignment_object" && program.Locals[declaration.Local].Name != "assignment_index") {
								continue
							}
							function.Body = append(append(append([]ir.Statement{}, body[:at]...), ir.Evaluate{Value: declaration.Value}), body[at:]...)
							changed++
							break
						}
						if mutation == "right_always" {
							branch, ok := statement.(ir.If)
							if !ok {
								continue
							}
							first, ok := branch.Then[0].(ir.Declare)
							if !ok {
								t.Fatal("no right-hand declaration")
							}
							branch.Then = branch.Then[1:]
							function.Body = append(append(append([]ir.Statement{}, body[:at]...), first, branch), body[at+1:]...)
							changed++
							break
						}
					}
				}
				if changed == 0 {
					t.Fatal("mutant changed nothing")
				}
				truth := onNode(t, path)
				got, binary := natively(t, program)
				if got.exitCode != 0 || len(got.stderr) != 0 {
					t.Fatalf("mutant must finish cleanly: exit %d stderr %q", got.exitCode, got.stderr)
				}
				if report := leaks(t, program, binary); report != "" {
					t.Fatal(report)
				}
				if difference := disagreement(truth, got); difference != "stdout differs" {
					t.Fatalf("native mutant caught by %q", difference)
				}
				js := onJavaScriptBackend(t, program)
				if js.exitCode != 0 || len(js.stderr) != 0 {
					t.Fatalf("JavaScript mutant must finish cleanly: %+v", js)
				}
				if difference := disagreement(truth, js); difference != "stdout differs" {
					t.Fatalf("JavaScript mutant caught by %q", difference)
				}
				t.Logf("%d mutations caught by Node stdout in native and JavaScript", changed)
			})
		}
	}
}
