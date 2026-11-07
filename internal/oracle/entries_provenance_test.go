package oracle

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func init() {
	for _, name := range []string{
		"entries_scanner",
		"entries_proven",
		"entries_checked_fit",
		"entries_checked_misfit",
		"entries_proven_import",
		"entries_checked_boolean",
		"entries_checked_literal",
		"entries_checked_boxed",
		"entries_record_alias",
		"entries_record_misfit",
	} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{
			"internal/oracle/testdata/" + name + ".a",
			true,
			name == "entries_checked_misfit" || name == "entries_checked_literal" || name == "entries_record_misfit",
		})
	}
}

func TestEntriesProvenance(test *testing.T) {
	test.Parallel()
	for _, name := range []string{
		"entries_scanner",
		"entries_proven",
		"entries_checked_fit",
		"entries_checked_misfit",
		"entries_proven_import",
		"entries_checked_boolean",
		"entries_checked_literal",
		"entries_checked_boxed",
		"entries_record_alias",
		"entries_record_misfit",
	} {
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			path, absolutePathError := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/"+name+".a"))
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			program, loweredError := lowered(test, path)
			if loweredError != nil {
				test.Fatal(loweredError)
			}
			node := onNode(test, path)
			if node.exitCode != 0 {
				test.Fatalf("source Node: %#v", node)
			}
			want := node

			if name == "entries_checked_literal" {
				want = run{exitCode: 70, stderr: []byte(
					"adamic: panic: Object enumeration key 'hidden': actual string, declared \"good\"\n",
				)}
			}
			if name == "entries_record_misfit" {
				want = run{
					exitCode: 70,
					stdout:   []byte("before entries\n"),
					stderr:   []byte("adamic: panic: Object enumeration key 'late': actual string, declared number\n"),
				}
			}
			if name == "entries_checked_misfit" {
				want = run{
					exitCode: 70,
					stderr: []byte(
						"adamic: panic: Object enumeration key 'hidden': actual string, declared number\n",
					),
				}
			}
			actual, _ := nativelyUncached(test, program)
			for _, got := range []run{actual, releasedUncached(test, program), onJavaScriptBackend(test, program)} {
				if difference := disagreement(want, got); difference != "" {
					test.Fatalf("%s: want %#v, got %#v", difference, want, got)
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
				test.Fatalf("proven=%t, checks=%d", proven, checks)
			}
			if name == "entries_checked_misfit" ||
				name == "entries_checked_literal" ||
				name == "entries_record_misfit" {
				for _, got := range []run{releasedUncached(test, program), onJavaScriptBackend(test, program)} {
					if difference := disagreement(node, got); difference != "" {
						test.Fatalf("unchecked mutant must match Node: %s, %#v", difference, got)
					}
					if disagreement(want, got) == "" {
						test.Fatal("unproven-as-proven mutant survived pinned exit/message")
					}
				}
				test.Log(
					"unproven-as-proven mutant prints Node's completed and exits 0; pinned exit 70 catches both backends",
				)
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
					for _, got := range []run{releasedUncached(test, program), onJavaScriptBackend(test, program)} {
						if difference := disagreement(node, got); difference != "" {
							test.Fatalf("literal contract mutant must match Node: %s, %#v", difference, got)
						}
						if disagreement(want, got) == "" {
							test.Fatal("drop literal membership mutant survived")
						}
					}
					test.Log(
						"drop literal membership mutant caught by pinned literal-contract exit/message in both backends",
					)
				}

			}
		})
	}
}

// Backend primitive coverage: the base refusal pass does not yet admit this
// staged initializer in a standalone module. Do not claim source admission here.
func TestEntriesRuntimeReadiness(test *testing.T) {
	test.Parallel()
	program := &ir.Program{
		Source:  "entries-runtime-readiness",
		Strings: []string{"completed"},
		Locals:  []ir.Local{{Name: "source", Type: ir.Object, Function: -1}},
	}
	program.Main = []ir.Statement{
		ir.Declare{
			Local: 0,
			Value: ir.ObjectLiteral{Fields: []ir.Field{{Name: "z", Value: ir.NumberConstant{}, Uninitialized: true}}},
		},
		ir.Evaluate{Value: ir.ObjectCall{
			Method:      "entries",
			Arguments:   []ir.Expression{ir.Read{Local: 0, Of: ir.Object}},
			Element:     ir.Number,
			Returns:     ir.Array,
			Checked:     true,
			ElementName: "number",
		}},
		ir.WriteLine{Stream: ir.Stdout, Value: ir.StringConstant{}},
	}
	path := filepath.Join(test.TempDir(), "source.a")
	if writeError := os.WriteFile(
		path,
		[]byte("const source = {z: undefined}; Object.entries(source); console.log('completed');\n"),
		0600,
	); writeError != nil {
		test.Fatal(writeError)
	}
	node := onNode(test, path)
	want := run{
		exitCode: 70,
		stderr:   []byte("adamic: panic: Object enumeration key 'z': actual undefined, declared number\n"),
	}
	actual, _ := nativelyUncached(test, program)
	for _, got := range []run{actual, releasedUncached(test, program), onJavaScriptBackend(test, program)} {
		if difference := disagreement(want, got); difference != "" {
			test.Fatalf("%s: want %#v, got %#v", difference, want, got)
		}
	}
	statement := program.Main[1].(ir.Evaluate)
	call := statement.Value.(ir.ObjectCall)
	call.Checked = false
	statement.Value = call
	program.Main[1] = statement
	for _, got := range []run{releasedUncached(test, program), onJavaScriptBackend(test, program)} {
		if difference := disagreement(node, got); difference != "" {
			test.Fatalf("unchecked readiness primitive mutant: %s, %#v", difference, got)
		}
		if disagreement(want, got) == "" {
			test.Fatal("unchecked readiness primitive mutant survived")
		}
	}
	test.Log("unchecked readiness primitive mutant caught by pinned exit/message in both backends")
}
