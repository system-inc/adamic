package oracle

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func init() {
	for _, name := range []string{"fallthrough_probe.a", "fallthrough_ownership.a", "fallthrough_nested.a", "fallthrough_labels.a", "implicit_returns.a", "fallthrough_exceptions.a"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/" + name, true, false})
	}
}

// These restore the three erroneous control-flow behaviors without breaking compilation.
func TestFallthroughMutants(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"isolated_cases", "missing_implicit_return", "continue_as_break"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/fallthrough_probe.a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			changed := false
			switch name {
			case "isolated_cases":
				mutate := func(value ir.Expression) ir.Expression {
					binary, ok := value.(ir.Binary)
					if !ok || binary.Operator != ir.LessOrEqual {
						return value
					}
					read, ok := binary.Left.(ir.Read)
					if !ok || program.Locals[read.Local].Name != "switch_entry" {
						return value
					}
					binary.Operator = ir.Equal
					changed = true
					return binary
				}
				mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), mutate)
				mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), mutate)
			case "missing_implicit_return":
				for index := range program.Functions {
					function := &program.Functions[index]
					if function.Name != "maybe" {
						continue
					}
					if _, ok := function.Body[len(function.Body)-1].(ir.Return); !ok {
						t.Fatal("missing implicit return to mutate")
					}
					function.Body = function.Body[:len(function.Body)-1]
					changed = true
				}
			case "continue_as_break":
				mutateFallthroughStatements(reflect.ValueOf(&program.Main).Elem(), func(statement ir.Statement) ir.Statement {
					if _, ok := statement.(ir.Continue); ok {
						changed = true
						return ir.Break{}
					}
					return statement
				})
			}
			if !changed {
				t.Fatal("mutant changed nothing")
			}
			truth := onNode(t, path)
			got, binary := natively(t, program)
			if name == "missing_implicit_return" {
				if got.exitCode != 70 || string(got.stderr) != "adamic: panic: compiler bug: a function ended without returning\n" {
					t.Fatalf("missing return survived: %+v", got)
				}
				t.Logf("Node %q; native exit %d stderr %q", truth.stdout, got.exitCode, got.stderr)
				return
			}
			if got.exitCode != 0 || len(got.stderr) != 0 {
				t.Fatalf("mutant failed outside stdout: %+v", got)
			}
			if report := leaks(t, program, binary); report != "" {
				t.Fatal(report)
			}
			if difference := disagreement(truth, got); difference != "stdout differs" {
				t.Fatalf("native caught by %q", difference)
			}
			js := onJavaScriptBackend(t, program)
			if difference := disagreement(truth, js); difference != "stdout differs" {
				t.Fatalf("JavaScript caught by %q", difference)
			}
			t.Logf("Node %q; native %q; JavaScript %q; caught only by stdout", truth.stdout, got.stdout, js.stdout)
		})
	}
}

func mutateFallthroughStatements(value reflect.Value, mutate func(ir.Statement) ir.Statement) {
	switch value.Kind() {
	case reflect.Interface:
		if value.IsNil() {
			return
		}
		if statement, ok := value.Interface().(ir.Statement); ok {
			replacement := mutate(statement)
			if !reflect.DeepEqual(statement, replacement) {
				value.Set(reflect.ValueOf(replacement))
				return
			}
		}
		copy := reflect.New(value.Elem().Type()).Elem()
		copy.Set(value.Elem())
		mutateFallthroughStatements(copy, mutate)
		value.Set(copy)
	case reflect.Struct:
		for index := 0; index < value.NumField(); index++ {
			mutateFallthroughStatements(value.Field(index), mutate)
		}
	case reflect.Slice:
		for index := 0; index < value.Len(); index++ {
			mutateFallthroughStatements(value.Index(index), mutate)
		}
	}
}
