package oracle

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func init() {
	for _, name := range []string{"entries_scanner", "entries_proven", "entries_checked_fit", "entries_checked_misfit", "entries_proven_import", "entries_checked_boolean", "entries_checked_literal", "entries_checked_boxed"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/" + name + ".a", true, name == "entries_checked_misfit" || name == "entries_checked_literal"})
	}
}

func TestEntriesProvenance(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"entries_scanner", "entries_proven", "entries_checked_fit", "entries_checked_misfit", "entries_proven_import", "entries_checked_boolean", "entries_checked_literal", "entries_checked_boxed"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/"+name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			node := onNode(t, path)
			if node.exitCode != 0 {
				t.Fatalf("source Node: %#v", node)
			}
			want := node

			if name == "entries_checked_literal" {
				want = run{exitCode: 70, stderr: []byte("adamic: panic: Object enumeration key 'hidden': actual string, declared \"good\"\n")}
			}
			if name == "entries_checked_misfit" {
				want = run{exitCode: 70, stderr: []byte("adamic: panic: Object enumeration key 'hidden': actual string, declared number\n")}
			}
			actual, _ := nativelyUncached(t, program)
			for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Fatalf("%s: want %#v, got %#v", difference, want, got)
				}
			}
			checks := 0
			inspect := func(value any) any {
				if call, ok := value.(ir.ObjectCall); ok && call.Checked {
					checks++
					call.Checked = false
					return call
				}
				return value
			}
			program.Main = mutateReadiness(program.Main, inspect)
			for index := range program.Functions {
				program.Functions[index].Body = mutateReadiness(program.Functions[index].Body, inspect)
			}
			proven := name == "entries_scanner" || name == "entries_proven" || name == "entries_proven_import"
			if proven && checks != 0 || !proven && checks == 0 {
				t.Fatalf("proven=%t, checks=%d", proven, checks)
			}
			if name == "entries_checked_misfit" || name == "entries_checked_literal" {
				for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if difference := disagreement(node, got); difference != "" {
						t.Fatalf("unchecked mutant must match Node: %s, %#v", difference, got)
					}
					if disagreement(want, got) == "" {
						t.Fatal("unproven-as-proven mutant survived pinned exit/message")
					}
				}
				t.Log("unproven-as-proven mutant prints Node's completed and exits 0; pinned exit 70 catches both backends")
				if name == "entries_checked_literal" {
					removeAllowed := func(value any) any {
						if call, ok := value.(ir.ObjectCall); ok && call.Method == "entries" {
							call.Checked = true
							call.Allowed = nil
							return call
						}
						return value
					}
					program.Main = mutateReadiness(program.Main, removeAllowed)
					for index := range program.Functions {
						program.Functions[index].Body = mutateReadiness(program.Functions[index].Body, removeAllowed)
					}
					for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
						if difference := disagreement(node, got); difference != "" {
							t.Fatalf("literal contract mutant must match Node: %s, %#v", difference, got)
						}
						if disagreement(want, got) == "" {
							t.Fatal("drop literal membership mutant survived")
						}
					}
					t.Log("drop literal membership mutant caught by pinned literal-contract exit/message in both backends")
				}

			}
		})
	}
}

// Backend primitive coverage: the base refusal pass does not yet admit this
// staged initializer in a standalone module. Do not claim source admission here.
func TestEntriesRuntimeReadiness(t *testing.T) {
	t.Parallel()
	program := &ir.Program{Source: "entries-runtime-readiness", Strings: []string{"completed"}, Locals: []ir.Local{{Name: "source", Type: ir.Object, Function: -1}}}
	program.Main = []ir.Statement{
		ir.Declare{Local: 0, Value: ir.ObjectLiteral{Fields: []ir.Field{{Name: "z", Value: ir.NumberConstant{}, Uninitialized: true}}}},
		ir.Evaluate{Value: ir.ObjectCall{Method: "entries", Arguments: []ir.Expression{ir.Read{Local: 0, Of: ir.Object}}, Element: ir.Number, Returns: ir.Array, Checked: true, ElementName: "number"}},
		ir.WriteLine{Stream: ir.Stdout, Value: ir.StringConstant{}},
	}
	path := filepath.Join(t.TempDir(), "source.a")
	if err := os.WriteFile(path, []byte("const source = {z: undefined}; Object.entries(source); console.log('completed');\n"), 0600); err != nil {
		t.Fatal(err)
	}
	node := onNode(t, path)
	want := run{exitCode: 70, stderr: []byte("adamic: panic: Object enumeration key 'z': actual undefined, declared number\n")}
	actual, _ := nativelyUncached(t, program)
	for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, got); difference != "" {
			t.Fatalf("%s: want %#v, got %#v", difference, want, got)
		}
	}
	statement := program.Main[1].(ir.Evaluate)
	call := statement.Value.(ir.ObjectCall)
	call.Checked = false
	statement.Value = call
	program.Main[1] = statement
	for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if difference := disagreement(node, got); difference != "" {
			t.Fatalf("unchecked readiness primitive mutant: %s, %#v", difference, got)
		}
		if disagreement(want, got) == "" {
			t.Fatal("unchecked readiness primitive mutant survived")
		}
	}
	t.Log("unchecked readiness primitive mutant caught by pinned exit/message in both backends")
}
