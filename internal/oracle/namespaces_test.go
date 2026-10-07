package oracle

import (
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

func init() {
	for _, path := range []string{"internal/oracle/testdata/namespaces_nested.a", "internal/oracle/testdata/namespaces_pair/main.a", "internal/oracle/testdata/namespaces_parser_enums.a", "internal/oracle/testdata/namespaces_parser_state.a", "internal/oracle/testdata/namespaces.a", "internal/oracle/testdata/namespaces_modules/main.a", "internal/oracle/testdata/namespaces_parser_body.a"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/namespaces_unready.a", true, true})

}

func TestNamespaceSemanticMutants(t *testing.T) {
	for _, family := range []string{"wrong scoped function", "wrong scoped constant", "wrong namespace enum", "wrong body order"} {
		t.Run(family, func(t *testing.T) {
			t.Parallel()
			fixture := "namespaces.a"
			if family == "wrong body order" {
				fixture = "namespaces_parser_body.a"
			}
			if family == "wrong namespace enum" {
				fixture = "namespaces_parser_enums.a"
			}
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", fixture))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			changed := false
			if family == "wrong body order" {
				positions := []int{}
				for index, statement := range program.Main {
					if _, ok := statement.(ir.WriteLine); ok {
						positions = append(positions, index)
					}
				}
				if len(positions) >= 2 {
					first, second := positions[len(positions)-2], positions[len(positions)-1]
					program.Main[first], program.Main[second] = program.Main[second], program.Main[first]
					changed = true
				}
			} else if family == "wrong namespace enum" {
				for _, statement := range program.Main {
					declaration, ok := statement.(ir.Declare)
					if !ok {
						continue
					}
					literal, ok := declaration.Value.(ir.ObjectLiteral)
					if !ok {
						continue
					}
					if program.Locals[declaration.Local].Name == "ParsingContext" {
						for index := range literal.Fields {
							if literal.Fields[index].Name == "SourceElements" {
								literal.Fields[index].Value = ir.NumberConstant{Value: 99}
								changed = true
							}
						}
					}
				}
			} else if family == "wrong scoped function" {
				for index := range program.Functions {
					if program.Functions[index].Name == "left" {
						program.Functions[index].Body = []ir.Statement{ir.Return{Value: ir.NumberConstant{Value: 90}}}
						changed = true
					}
				}
			} else {
				for index, statement := range program.Main {
					declaration, ok := statement.(ir.Declare)
					if ok && program.Locals[declaration.Local].Name == "offset" {
						if constant, ok := declaration.Value.(ir.NumberConstant); ok && constant.Value == 10 {
							declaration.Value = ir.NumberConstant{Value: 99}
							program.Main[index] = declaration
							changed = true
							break
						}
					}
				}
			}
			if !changed {
				t.Fatal("mutant changed no namespace binding")
			}
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			result := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=0", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
			if result.exitCode != 0 || len(result.stderr) != 0 {
				t.Fatalf("mutant must finish cleanly: %+v", result)
			}
			if difference := disagreement(onNode(t, path), result); difference != "stdout differs" {
				t.Fatalf("not caught by Node: %q", difference)
			}
			t.Log("caught by Node stdout; exit 0 with clean sanitizers")
		})
	}
}

func TestNamespaceStateMutants(t *testing.T) {
	for _, test := range []struct{ name, path string }{
		{"lose assignment", "namespaces_parser_state.a"},
		{"skip ready check", "namespaces_unready.a"},
	} {
		t.Run(test.name, func(t *testing.T) {
			path, pathErr := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", test.path))
			if pathErr != nil {
				t.Fatal(pathErr)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			changed := false
			for index := range program.Functions {
				body := program.Functions[index].Body
				for at, statement := range body {
					if test.name == "lose assignment" {
						if assign, ok := statement.(ir.Assign); ok && program.Locals[assign.Local].Name == "text" {
							body[at] = ir.Evaluate{Value: assign.Value}
							changed = true
						}
					} else if returned, ok := statement.(ir.Return); ok {
						if read, ok := returned.Value.(ir.Read); ok && read.Checked {
							read.Checked = false
							returned.Value = read
							body[at] = returned
							changed = true
						}
					}
				}
			}
			if !changed {
				t.Fatal("mutant changed nothing")
			}
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			observed := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=0", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
			if test.name == "skip ready check" {
				if expected := onJavaScriptBackend(t, mustLowerNamespace(t, path)); observed.exitCode == expected.exitCode {
					t.Fatal("ready check mutant survived")
				}
				t.Log("caught by checked JavaScript exit comparison")
			} else {
				if disagreement(onNode(t, path), observed) == "" {
					t.Fatal("lost assignment survived")
				}
				t.Log("caught by Node comparison")
			}
		})
	}
}
func mustLowerNamespace(t *testing.T, path string) *ir.Program {
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	return program
}
