package checker

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAdamicRootKeepsConfigDeclarations(test *testing.T) {
	test.Parallel()
	directory := test.TempDir()
	configuration := filepath.Join(directory, "tsconfig.json")
	declaration := filepath.Join(directory, "prelude.d.ts")
	source := filepath.Join(directory, "program.a")
	for path, text := range map[string]string{
		configuration: `{"compilerOptions":{"strict":true},"files":["prelude.d.ts"]}`,
		declaration:   `declare const ambient: "世界🌍";`,
		source:        "ambient;\n",
	} {
		if writeError := os.WriteFile(path, []byte(text), 0600); writeError != nil {
			test.Fatal(writeError)
		}
	}
	program, openError := Open(configuration, []string{source})
	if openError != nil {
		test.Fatal(openError)
	}
	facts, inspectError := program.Inspect(source, 0, 7, "Identifier", "raw-type")
	if inspectError != nil {
		test.Fatal(inspectError)
	}
	values := decodedFields(test, facts)
	if len(values) != 19 || values[10] != `"世界🌍"` {
		test.Fatalf("explicit .a root lost its configured declaration: %q", values)
	}
}
