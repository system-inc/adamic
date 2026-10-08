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
	}{"internal/oracle/testdata/comma_operator.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/logical_nullish_assignment.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/logical_and_assignment.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/logical_or_assignment.a", true, false})
}

// Mutants preserve valid IR and ownership. Each must compile and finish cleanly, then disagree
// with source Node in both backends on the side-effect counters or expression value.
func TestLogicalAssignmentMutants(t *testing.T) {
	for _, operator := range []string{"or", "and", "nullish"} {
		mutations := []string{"target_twice", "right_always"}
		if operator == "nullish" {
			mutations = append(mutations, "zero_nullish", "empty_nullish")
		}
		for _, mutation := range mutations {
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
						if mutation == "zero_nullish" || mutation == "empty_nullish" {
							branch, ok := statement.(ir.If)
							if !ok {
								continue
							}
							declaration, ok := body[at-1].(ir.Declare)
							if !ok || program.Locals[declaration.Local].Name != "assignment_current" {
								t.Fatal("no saved current value")
							}
							current := ir.Read{Local: declaration.Local, Of: declaration.Value.Type()}
							if mutation == "zero_nullish" && current.Type() == ir.Number {
								branch.Condition = ir.Binary{Operator: ir.Equal, Left: current, Right: ir.NumberConstant{Value: 0}}
							} else if mutation == "zero_nullish" && current.Type() == ir.MaybeNumber {
								branch.Condition = ir.Conditional{Condition: ir.IsUndefined{Value: current}, WhenTrue: ir.BooleanConstant{Value: true}, WhenNot: ir.Binary{Operator: ir.Equal, Left: ir.Unwrap{Value: current}, Right: ir.NumberConstant{Value: 0}}}
							} else if mutation == "empty_nullish" && current.Type() == ir.String {
								branch.Condition = ir.Binary{Operator: ir.Or, Left: ir.IsUndefined{Value: current}, Right: ir.Binary{Operator: ir.Equal, Left: ir.StringLength{Value: current}, Right: ir.NumberConstant{Value: 0}}}
							} else {
								continue
							}
							body[at] = branch
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

func TestCommaMutants(t *testing.T) {
	for _, mutation := range []string{"drop_left", "reverse_order", "return_left"} {
		t.Run(mutation, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/comma_operator.a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			changed := 0
			for i := range program.Functions {
				f := &program.Functions[i]
				if f.Name != "comma_expression" || len(f.Body) != 2 {
					continue
				}
				left, ok := f.Body[0].(ir.Evaluate)
				ret, returns := f.Body[1].(ir.Return)
				if !ok || !returns || left.Value.Type() != ir.Number || ret.Value.Type() != ir.Number {
					continue
				}
				switch mutation {
				case "drop_left":
					f.Body = f.Body[1:]
				case "return_left":
					f.Body = []ir.Statement{ir.Evaluate{Value: ret.Value}, ir.Return{Value: left.Value}}
				case "reverse_order":
					local := len(program.Locals)
					program.Locals = append(program.Locals, ir.Local{Name: "comma_mutant_right", Type: ir.Number, Function: i})
					f.Body = []ir.Statement{ir.Declare{Local: local, Value: ret.Value}, left, ir.Return{Value: ir.Read{Local: local, Of: ir.Number}}}
				}
				changed++
			}
			if changed == 0 {
				t.Fatal("mutant changed nothing")
			}
			truth := onNode(t, path)
			got, binary := natively(t, program)
			if got.exitCode != 0 || len(got.stderr) != 0 {
				t.Fatalf("mutant failed execution: %+v", got)
			}
			if report := leaks(t, program, binary); report != "" {
				t.Fatal(report)
			}
			if diff := disagreement(truth, got); diff != "stdout differs" {
				t.Fatalf("native mutant caught by %q", diff)
			}
			js := onJavaScriptBackend(t, program)
			if js.exitCode != 0 || len(js.stderr) != 0 {
				t.Fatalf("JavaScript mutant failed execution: %+v", js)
			}
			if diff := disagreement(truth, js); diff != "stdout differs" {
				t.Fatalf("JavaScript mutant caught by %q", diff)
			}
			t.Logf("%d mutations caught by Node stdout in both backends", changed)
		})
	}
}
