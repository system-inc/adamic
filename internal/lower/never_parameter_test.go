package lower

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/load"
)

func TestNeverFrameMarkerRoads(t *testing.T) {
	t.Parallel()
	for _, pin := range []struct {
		name      string
		holds     bool
		nodeTypes bool
	}{
		{"f_debug", true, false}, {"f_nullish", true, false}, {"f_cycle", true, false},
		{"f_function", true, false}, {"f_cycle_only", false, false}, {"callback", false, false},
		{"f_alias", false, false}, {"f_namespace_alias", false, false},
		{"f_destructured", false, false}, {"f_call", false, false}, {"f_apply", false, false},
		{"f_typeof", false, false}, {"f_name", false, false}, {"f_length", false, false},
		{"f_compare", false, false}, {"f_stored", false, false}, {"f_default", false, false},
		{"f_called_marker", false, false},
		{"f_arguments", false, false}, {"c01", false, false},
		// These roads resolve to the library member. The write refusals
		// are what keep these programs from replacing the stack hooks.
		{"c02", true, true}, {"c02_node", true, false},
		{"c03", true, true}, {"c03_node", true, false},
	} {
		t.Run(pin.name, func(t *testing.T) {
			t.Parallel()
			inputs := []string{filepath.Join("..", "oracle", "testdata", "never_parameter", pin.name+".ts")}
			if pin.nodeTypes {
				// Supply the library for the proof without changing the original bytes.
				inputs = append(inputs, writeAgreementSource(t, "node-types.ts", "import 'node:process';\n"))
			}
			program, err := load.Load(inputs)
			if err != nil {
				t.Fatal(err)
			}
			entry := program.Files()[0]
			checked, release := program.Checker(context.Background(), entry)
			defer release()
			l := &lowering{program: program, checker: checked}
			name := "assertNever"
			if pin.name == "callback" {
				name = "never1"
			}
			for _, declaration := range entry.Statements.Nodes {
				if declaration.Kind == ast.KindFunctionDeclaration && declaration.Name().Text() == name {
					if got := l.directOnlyNeverFunction(declaration); got != pin.holds {
						t.Fatalf("directOnlyNeverFunction = %t, want %t", got, pin.holds)
					}
					return
				}
			}
			t.Fatalf("%s declaration missing", name)
		})
	}
}
