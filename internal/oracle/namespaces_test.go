package oracle

import (
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

func init() {
	for _, path := range []string{"internal/oracle/testdata/namespaces.a", "internal/oracle/testdata/namespaces_modules/main.a"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
}

func TestNamespaceSemanticMutants(t *testing.T) {
	for _, family := range []string{"wrong scoped function", "wrong scoped constant"} {
		t.Run(family, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/namespaces.a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			changed := false
			if family == "wrong scoped function" {
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
