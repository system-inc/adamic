package oracle

import (
	"errors"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/lower"
	"testing"
)

func init() {
	for _, name := range []string{"scanner.a", "shared.a", "neighbors.a", "fallthrough.a", "dead_zone.a", "dead_zone_direct.a", "dead_zone_initializer.a", "dead_zone_write.a"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/switch_case_declarations/" + name, true, false})
	}
}

func TestSwitchCaseDeadZoneStop(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/switch_case_declarations/dead_zone.a"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 1 || string(truth.stdout) != "value1\nbefore\n" || !strings.Contains(string(truth.stderr), "ReferenceError: Cannot access 'value' before initialization") {
		t.Fatalf("Node dead-zone stop: %+v", truth)
	}
}

func TestSwitchCaseDeclarationMutants(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"unready_read", "unready_write", "no_fallthrough", "function_not_hoisted"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			fixture := map[string]string{
				"unready_read":         "dead_zone_direct.a",
				"unready_write":        "dead_zone_write.a",
				"no_fallthrough":       "fallthrough.a",
				"function_not_hoisted": "shared.a",
			}[name]
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/switch_case_declarations", fixture))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			changed := false
			if name == "unready_read" || name == "no_fallthrough" {
				mutate := func(value ir.Expression) ir.Expression {
					if read, ok := value.(ir.Read); name == "unready_read" && ok && read.Checked && program.Locals[read.Local].Ready != 0 {
						read.Checked = false
						changed = true
						return read
					}
					if binary, ok := value.(ir.Binary); name == "no_fallthrough" && ok && binary.Operator == ir.LessOrEqual {
						if read, ok := binary.Left.(ir.Read); ok && program.Locals[read.Local].Name == "switch_entry" {
							binary.Operator = ir.Equal
							changed = true
							return binary
						}
					}
					return value
				}
				mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), mutate)
				mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), mutate)
			} else {
				ready := -1
				for _, local := range program.Locals {
					if local.Name == "double" && local.Ready != 0 {
						ready = local.Ready - 1
					}
				}
				mutate := func(statement ir.Statement) ir.Statement {
					assign, ok := statement.(ir.Assign)
					if !ok {
						return statement
					}
					if name == "unready_write" && assign.Checked && program.Locals[assign.Local].Ready != 0 {
						assign.Checked = false
						changed = true
						return assign
					}
					if name == "function_not_hoisted" {
						if closure, ok := assign.Value.(ir.MakeClosure); ok && program.Functions[closure.Function].Name == "double" || assign.Local == ready {
							changed = true
							return ir.Block{}
						}
					}
					return statement
				}
				mutateFallthroughStatements(reflect.ValueOf(&program.Main).Elem(), mutate)
				mutateFallthroughStatements(reflect.ValueOf(&program.Functions).Elem(), mutate)
			}
			if !changed {
				t.Fatal("mutant changed nothing")
			}
			truth := onNode(t, path)
			got, binary := natively(t, program)
			if name == "function_not_hoisted" {
				if got.exitCode != 1 {
					t.Fatalf("mutant failed outside its initialization guard: %+v", got)
				}
			} else {
				if got.exitCode != 0 || len(got.stderr) != 0 {
					t.Fatalf("mutant failed outside the oracle comparison: %+v", got)
				}
				if report := leaks(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
			js := onJavaScriptBackend(t, program)
			if disagreement(truth, got) == "" || disagreement(truth, js) == "" {
				t.Fatal("mutant survived the Node oracle")
			}
			t.Logf("Node stdout %q exit %d; native stdout %q exit %d; JavaScript stdout %q exit %d", truth.stdout, truth.exitCode, got.stdout, got.exitCode, js.stdout, js.exitCode)
		})
	}
}

// Overload signatures have no body. The existing closure convention does not lower them;
// encountering one must diagnose it before visiting a nonexistent body.
func TestSwitchCaseOverloadIsNotYet(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/switch_case_declarations/overload.a"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || len(truth.stderr) != 0 || string(truth.stdout) != "2\n" {
		t.Fatalf("Node overload: %+v", truth)
	}
	_, err = lowered(t, path)
	var notYet *lower.NotYet
	want := path + ":3:1: stage 0 can't lower a function without a body yet"
	if !errors.As(err, &notYet) || err.Error() != want {
		t.Fatalf("got %v; want %s", err, want)
	}
}
