package checker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReferenceSymbol(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "source.ts")
	config := filepath.Join(dir, "tsconfig.json")
	source := "const origin = RegExp;\norigin;\nconst value = { origin };\nexport { origin };\n"
	for path, text := range map[string]string{file: source, config: `{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"]}}`} {
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	p, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	var identity string
	for _, needle := range []string{"\norigin", "{ origin", "export { origin"} {
		end := uint64(strings.Index(source, needle) + len(needle))
		start := end - 7
		wire, err := p.Inspect(file, start, end, "Identifier", "reference-symbol")
		if err != nil {
			t.Fatal(err)
		}
		fields := decodedFields(t, wire)
		if len(fields) != 9 || fields[0] != "1" || fields[1] != "reference-symbol" || fields[2] == "0" || fields[3] != "1" || fields[4] != file || fields[5] != "0" || fields[6] != "VariableDeclaration" {
			t.Fatalf("wrong reference origin: %q", fields)
		}
		if identity == "" {
			identity = fields[2]
		} else if identity != fields[2] {
			t.Fatalf("value identity differs: %s != %s", identity, fields[2])
		}
	}
	if _, err := p.Inspect(file, 0, uint64(len(source)), "SourceFile", "reference-symbol"); err == nil {
		t.Fatal("accepted non-Identifier reference")
	}
}
