package checker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStreamFacts(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "entry.ts")
	dependency := filepath.Join(dir, "dependency.ts")
	config := filepath.Join(dir, "tsconfig.json")
	source := "import { show } from './dependency';\nshow();\nconst [bound] = [1];\nbound;\nexport {};\n"
	for path, text := range map[string]string{file: source, dependency: "export function show(){ return 1; }\n", config: `{"compilerOptions":{"strict":true,"target":"ES2022","module":"ESNext","moduleResolution":"bundler","lib":["ES2022"]}}`} {
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	p, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	start := uint64(strings.Index(source, "\nshow"))
	end := start + uint64(len("\nshow"))
	for _, q := range []string{"stream-symbol", "stream-signature", "stream-program", "stream-file"} {
		kind := "SourceFile"
		pos, finish := uint64(0), uint64(len(source))
		if q == "stream-symbol" {
			kind = "Identifier"
			pos, finish = start, end
		}
		if q == "stream-signature" {
			kind = "CallExpression"
			pos, finish = start, end+2
		}
		wire, err := p.Inspect(file, pos, finish, kind, q)
		if err != nil {
			t.Fatalf("%s: %v", q, err)
		}
		fields := decodedFields(t, wire)
		if len(fields) < 3 || fields[0] != "1" || fields[1] != q {
			t.Fatalf("bad %s header: %q", q, fields)
		}
		if q == "stream-file" && (len(fields) != 4 || fields[2] != "1" || fields[3] != "0") {
			t.Fatalf("bad source mode %q", fields)
		}
		if q == "stream-signature" && (len(fields) < 12 || fields[2] != "1" || fields[3] != dependency) {
			t.Fatalf("wrong resolved declaration %q", fields)
		}
		if q == "stream-program" && (!strings.Contains(wire, dependency) || !strings.Contains(wire, "ImportDeclaration")) {
			t.Fatalf("missing import resolution %q", fields)
		}
		if q != "stream-symbol" {
			if _, err := p.Inspect(file, start, end, "Identifier", q); err == nil {
				t.Fatalf("accepted wrong node for %s", q)
			}
		}
	}
	bindingStart := uint64(strings.Index(source, "\nbound"))
	bindingEnd := bindingStart + uint64(len("\nbound"))
	if _, err := p.Inspect(file, bindingStart, bindingEnd, "Identifier", "stream-symbol"); err != nil {
		t.Fatalf("destructuring symbol facts: %v", err)
	}
	if _, err := p.Inspect(file, start, end, "Identifier", "stream-unknown"); err == nil {
		t.Fatal("accepted unknown question")
	}
}
