package lower

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

// Lower's parameter roads must stop before a callee can install browser behind its constant read.
func TestHostBindingProcessParameterMutationRoadsStop(t *testing.T) {
	t.Parallel()
	for _, parameter := range []struct{ name, generic, of string }{
		{"generic", "<T extends object>", "T"},
		{"unknown", "", "unknown"},
		{"object", "", "object"},
	} {
		for _, operation := range []string{
			"Reflect.set(o, 'browser', 1);",
			"Object.defineProperty(o, 'browser', { value: 1 });",
			"Object.assign(o, { browser: 1 });",
			"(o as { browser?: number }).browser = 1;",
		} {
			t.Run(parameter.name+"/"+operation, func(t *testing.T) {
				t.Parallel()
				directory := t.TempDir()
				core := processDeclaration + "function mutate" + parameter.generic + "(o: " + parameter.of + "): void {\n" + operation + "\n}\nexport function run(): void {\nif (typeof process !== 'undefined') { mutate(process); console.log(String(process.browser)); }\n}\n"
				for name, source := range map[string]string{
					"core.ts": core,
					"main.ts": "import 'node:process';\nimport { run } from './core.ts';\nrun();\n",
				} {
					if err := os.WriteFile(filepath.Join(directory, name), []byte(source), 0o644); err != nil {
						t.Fatal(err)
					}
				}
				program, err := load.Load([]string{filepath.Join(directory, "main.ts")})
				if err != nil {
					if parameter.name != "unknown" || !strings.Contains(err.Error(), "error TS2769:") || (!strings.HasPrefix(operation, "Reflect.set") && !strings.HasPrefix(operation, "Object.assign")) {
						t.Fatalf("unexpected Load stop: %v", err)
					}
					t.Logf("Load: %v", err)
					return
				}
				_, err = Lower(t.Context(), program)
				var notYet *NotYet
				var refused *Refused
				if !errors.As(err, &notYet) && !errors.As(err, &refused) {
					t.Fatalf("mutation road did not stop; compare its run with Node before changing lowering: %v", err)
				}
				t.Log(err)
			})
		}
	}
}
