package checker

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestSymbolLocationsMergedAndDestructured(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "input.ts")
	config := filepath.Join(dir, "tsconfig.json")
	source := "interface C{}; const C=1; const {value}= {value:2}; C; value;"
	for path, text := range map[string]string{file: source, config: `{"compilerOptions":{"strict":true},"files":["input.ts"]}`} {
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	p, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"C", "value"} {
		start := strings.LastIndex(source, name) - 1
		wire, err := p.Inspect(file, uint64(start), uint64(start+len(name)+1), "Identifier", "symbol-locations")
		if err != nil {
			t.Fatal(err)
		}
		fields := decodedFields(t, wire)
		count, err := strconv.Atoi(fields[2])
		if err != nil {
			t.Fatal(err)
		}
		expected := 1
		if name == "C" {
			expected = 2
		}
		if count != expected || len(fields) != 3+4*count {
			t.Fatalf("missing declarations: %v", fields)
		}
		if name == "value" && fields[4] != "BindingElement" {
			t.Fatal("lost destructuring declaration")
		}
	}
}
