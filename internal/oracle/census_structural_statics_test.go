package oracle

import (
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/census_structural_statics.a", true, false})
}

func TestCensusStructuralMutants(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/census_structural_statics.a"))
	if err != nil {
		t.Fatal(err)
	}
	for _, mutation := range []string{"receiver twice", "static instead of instance", "lookup after arguments"} {
		t.Run(mutation, func(t *testing.T) {
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			source := native.C(program)
			mutant := source
			if mutation == "receiver twice" {
				pattern := regexp.MustCompile(`(?m)^(\s*)(adamic_object \* adamic_temporary_\d+ = )(adamic_function_\d+_getFactory)\(\);$`)
				mutant = pattern.ReplaceAllString(source, "${1}adamic_release(${3}());\n${1}${2}${3}();")
			} else if mutation == "lookup after arguments" {
				for index := range program.Functions {
					function := &program.Functions[index]
					if function.Name == "structural_receiver" {
						returned := function.Body[0].(ir.Return)
						literal := returned.Value.(ir.ObjectLiteral)
						literal.Fields[1].Value = ir.Undefined{Of: ir.Closure}
						returned.Value = literal
						function.Body[0] = returned
					}
				}
				mutant = native.C(program)
			} else {
				target := -1
				for index, function := range program.Functions {
					if function.Name == "Factory_static_createExportDeclaration" {
						target = index
					}
				}
				if target < 0 {
					t.Fatal("static mutant target absent")
				}
				changed := false
				for index := range program.Functions {
					function := &program.Functions[index]
					if function.Name != "structural_method" || function.Returns != ir.Object {
						continue
					}
					last := len(function.Body) - 1
					conditional, ok := function.Body[last].(ir.If)
					if !ok {
						continue
					}
					returned := conditional.Then[0].(ir.Return)
					fallback, ok := returned.Value.(ir.CallClosure)
					if !ok || fallback.Closure.(ir.Property).Name != "createExportDeclaration" {
						continue
					}
					arguments := []ir.Expression{fallback.Closure.(ir.Property).Object}
					arguments = append(arguments, fallback.Arguments...)
					class := 0
					for index, metadata := range program.Classes {
						if metadata.Name == "Factory" && !metadata.Static {
							class = index + 1
						}
					}
					if class == 0 {
						t.Fatal("instance class mutant target absent")
					}
					wrong := ir.If{Condition: ir.InstanceOf{Value: arguments[0], Class: class}, Then: []ir.Statement{ir.Return{Value: ir.Call{Function: target, Arguments: arguments, Returns: ir.Object}}}}
					function.Body = append(function.Body[:last], wrong, conditional)
					changed = true
				}
				if !changed {
					t.Fatal("instance dispatch mutant target absent")
				}
				mutant = native.C(program)
			}
			if mutant == source {
				t.Fatal("mutant did not change emitted C")
			}
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(mutant, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			result := execute(t, binary)
			if result.exitCode != 0 || len(result.stderr) != 0 {
				t.Fatalf("mutant did not finish cleanly: %d %s", result.exitCode, result.stderr)
			}
			if difference := disagreement(onNode(t, path), result); difference != "stdout differs" {
				t.Fatalf("want stdout differs, got %q", difference)
			}
			t.Logf("Node caught %s: %q", mutation, strings.TrimSpace(string(result.stdout)))
		})
	}
}
