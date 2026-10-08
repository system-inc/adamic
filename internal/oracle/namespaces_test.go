package oracle

import (
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

func init() {
	for _, path := range []string{"stage3/namespaces/debug-groups/local_const_enums.a", "internal/oracle/testdata/namespaces_call_graph.a", "internal/oracle/testdata/namespaces_call_cycle.a", "internal/oracle/testdata/namespaces_map_before.a", "internal/oracle/testdata/namespaces_safe_initialization.a", "internal/oracle/testdata/namespaces_debug_probe.a", "internal/oracle/testdata/namespaces_observed_narrowing.a", "internal/oracle/testdata/namespaces_parser_factory.a", "stage3/fixtures/namespaces/08_parser_jsdoc_nested.a", "internal/oracle/testdata/namespaces_debug_modules/main.a", "internal/oracle/testdata/namespaces_debug_state.a", "internal/oracle/testdata/namespaces_parser_body.a", "internal/oracle/testdata/namespaces_parser_enums.a", "internal/oracle/testdata/namespaces_parser_state.a", "internal/oracle/testdata/namespaces.a", "internal/oracle/testdata/namespaces_modules/main.a"} {
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

	for _, path := range []string{"internal/oracle/testdata/namespaces_unknown_before.a", "internal/oracle/testdata/namespaces_unknown_function_before.a", "internal/oracle/testdata/namespaces_unknown_write_before.a", "internal/oracle/testdata/namespaces_unknown_void_before.a", "internal/oracle/testdata/namespaces_unknown_enum_before.a"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}

	for _, name := range []string{"BuilderState", "JsxNames", "ReactNames", "BinaryExpressionState", "Parser.JSDocParser"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"stage3/namespaces/shapes/" + name + ".a", true, false})
	}

}

func TestNamespaceSemanticMutants(t *testing.T) {
	for _, family := range []string{"wrong scoped function", "wrong scoped constant", "wrong namespace enum", "wrong body order", "wrong exported state", "wrong debug initialization", "drop returned assignment", "wrong factory binding"} {
		t.Run(family, func(t *testing.T) {
			t.Parallel()
			fixture := "namespaces.a"
			if family == "wrong debug initialization" {
				fixture = "namespaces_debug_probe.a"
			}
			if family == "wrong factory binding" {
				fixture = "namespaces_parser_factory.a"
			}
			if family == "drop returned assignment" {
				fixture = "../../../stage3/fixtures/namespaces/08_parser_jsdoc_nested.a"
			}
			if family == "wrong exported state" {
				fixture = "namespaces_debug_state.a"
			}
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
			if family == "wrong debug initialization" {
				for index, statement := range program.Main {
					if declared, ok := statement.(ir.Declare); ok && program.Locals[declared.Local].Name == "isDebugging" {
						declared.Value = ir.BooleanConstant{Value: true}
						program.Main[index] = declared
						changed = true
					}
				}
			} else if family == "wrong factory binding" {
				for index, statement := range program.Main {
					if assign, ok := statement.(ir.Assign); ok && program.Locals[assign.Local].Name == "factoryCreateNumericLiteral" {
						if property, ok := assign.Value.(ir.Property); ok {
							property.Name = "createLiteralLikeNode"
							assign.Value = property
							program.Main[index] = assign
							changed = true
						}
					}
				}
			} else if family == "drop returned assignment" {
				for index := range program.Functions {
					if program.Functions[index].Name != "nextTokenJSDoc" {
						continue
					}
					for at, statement := range program.Functions[index].Body {
						if assign, ok := statement.(ir.Assign); ok {
							program.Functions[index].Body[at] = ir.Evaluate{Value: assign.Value}
							changed = true
						}
					}
				}
			} else if family == "wrong exported state" {
				for index, statement := range program.Main {
					if assign, ok := statement.(ir.Assign); ok && program.Locals[assign.Local].Name == "currentLogLevel" {
						if constant, ok := assign.Value.(ir.NumberConstant); ok && constant.Value == 3 {
							assign.Value = ir.NumberConstant{Value: 30}
							program.Main[index] = assign
							changed = true
							break
						}
						// A qualified write snapshots its RHS before the write check.
						if read, ok := assign.Value.(ir.Read); ok {
							for at, statement := range program.Main {
								if declared, ok := statement.(ir.Declare); ok && declared.Local == read.Local {
									if number, ok := declared.Value.(ir.NumberConstant); ok && number.Value == 3 {
										declared.Value = ir.NumberConstant{Value: 30}
										program.Main[at] = declared
										changed = true
									}
								}
							}
						}
						if changed {
							break
						}

					}
				}
			} else if family == "wrong body order" {
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
		{"lose hoisting", "namespaces_parser_state.a"},
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
			if test.name == "lose hoisting" {
				kept := []ir.Statement{}
				for _, statement := range program.Main {
					if declaration, ok := statement.(ir.Declare); ok && program.Locals[declaration.Local].Name == "token" {
						changed = true
					} else {
						kept = append(kept, statement)
					}
				}
				program.Main = kept
			}
			for index := range program.Functions {
				body := program.Functions[index].Body
				for at, statement := range body {
					if test.name == "lose assignment" {
						if assign, ok := statement.(ir.Assign); ok && program.Locals[assign.Local].Name == "text" {
							body[at] = ir.Evaluate{Value: assign.Value}
							changed = true
						}
					} else if returned, ok := statement.(ir.Return); ok && test.name == "skip ready check" {
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
