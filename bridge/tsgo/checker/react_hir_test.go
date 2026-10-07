package checker

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReactHIRQuestionContract(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "input.a")
	config := filepath.Join(dir, "tsconfig.json")
	source := "function Component(){const C=()=> '世界🌍';return C;}"
	for path, text := range map[string]string{file: source, config: `{"compilerOptions":{"strict":true},"files":["input.a"]}`} {
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	p, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	for _, question := range []string{"react-hir\nplain", "react-hir\nmemo"} {
		wire, err := p.Inspect(file, 0, uint64(len(source)), "SourceFile", question)
		if err != nil {
			t.Fatal(err)
		}
		fields := decodedFields(t, wire)
		if len(fields) != 3 || fields[0] != "1" || fields[1] != "react-hir" {
			t.Fatalf("unexpected frame header: %v", fields[:2])
		}
		var graphs []any
		if err := json.Unmarshal([]byte(fields[2]), &graphs); err != nil {
			t.Fatal(err)
		}
		if len(graphs) != 2 || !strings.Contains(fields[2], "TypeAliasName") || !strings.Contains(fields[2], "世界🌍") {
			t.Fatal("missing independent functions, raw types or Unicode")
		}
	}
	for _, question := range []string{"react-hir", "react-hir\n", "react-hir\nunknown", "react-hir\nplain\nextra"} {
		if _, err := p.Inspect(file, 0, uint64(len(source)), "SourceFile", question); err == nil {
			t.Fatalf("accepted malformed question %q", question)
		}
	}
	start := strings.Index(source, "Component") - 1
	if _, err := p.Inspect(file, uint64(start), uint64(start+len("Component")+1), "Identifier", "react-hir\nplain"); err == nil || !strings.Contains(err.Error(), "source file or function") {
		t.Fatalf("wrong non-function refusal: %v", err)
	}
}
