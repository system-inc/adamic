package checker

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAdamicRootKeepsConfigDeclarations(test *testing.T) {
	test.Parallel()
	directory := test.TempDir()
	config := filepath.Join(directory, "tsconfig.json")
	declaration := filepath.Join(directory, "prelude.d.ts")
	source := filepath.Join(directory, "program.a")
	for path, text := range map[string]string{
		config:      `{"compilerOptions":{"strict":true},"files":["prelude.d.ts"]}`,
		declaration: `declare const ambient: "世界🌍";`,
		source:      "ambient;\n",
	} {
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			test.Fatal(err)
		}
	}
	program, err := Open(config, []string{source})
	if err != nil {
		test.Fatal(err)
	}
	facts, err := program.Inspect(source, 0, 7, "Identifier", "raw-type")
	if err != nil {
		test.Fatal(err)
	}
	values := decodedFields(test, facts)
	if len(values) != 19 || values[10] != `"世界🌍"` {
		test.Fatalf("explicit .a root lost its configured declaration: %q", values)
	}
}
