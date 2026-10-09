package load

import (
	"os"
	"strings"
	"testing"
)

func TestErrorRuntimePreludeDeclarations(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("testdata/error-facts/debug.a")
	if err != nil {
		t.Fatal(err)
	}
	paths := projectProgram(t, `{"strict":true,"target":"es2024","lib":["es2024"],"types":[]}`, string(source))
	if _, err := Load(paths); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(projectProgram(t, `{"strict":true,"target":"es2024","lib":["es2024"],"types":["node"]}`, string(source))); err != nil {
		t.Fatalf("explicit Node type library without node: imports: %v", err)
	}

	limit := `const defaultLimit: number = Error.stackTraceLimit; Error.stackTraceLimit = 3; const capture = Error.captureStackTrace; capture({}, undefined);`
	if _, err := Load(projectProgram(t, `{"strict":true,"lib":["es2024"],"types":[]}`, limit)); err != nil {
		t.Fatal(err)
	}
}

func TestErrorRuntimePreludeTypesRejectLies(t *testing.T) {
	t.Parallel()
	for _, source := range []string{`Error.captureStackTrace(3);`, `Error.captureStackTrace({}, 3);`, `Error.stackTraceLimit = 'three';`} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			paths := writeProgram(t, [2]string{"main.a", source})
			diagnostics := checkErrors(t, paths)
			if len(diagnostics) != 1 || !(strings.Contains(diagnostics[0], "TS2345") || strings.Contains(diagnostics[0], "TS2322")) {
				t.Fatalf("wrong runtime fact types: %v", diagnostics)
			}
		})
	}
}
