package checker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProcessFacts(t *testing.T) {
	directory := t.TempDir()
	config := filepath.Join(directory, "tsconfig.json")
	entry := filepath.Join(directory, "Entry.ts")
	helper := filepath.Join(directory, "Help.ts")
	source := "import {write as show} from './Help';\nshow();\n"
	for path, text := range map[string]string{config: `{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"],"module":"NodeNext"}}`, entry: source, helper: "export function write():void { console.log('世界🌍'); }\n"} {
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	p, err := Open(config, []string{entry})
	if err != nil {
		t.Fatal(err)
	}
	start := uint64(strings.Index(source, "\nshow"))
	end := start + uint64(len("\nshow()"))
	query := func(kind, question string, finish uint64) []string {
		t.Helper()
		wire, e := p.Inspect(entry, start, finish, kind, question)
		if e != nil {
			t.Fatal(e)
		}
		return decodedFields(t, wire)
	}
	alias := query("Identifier", "node-symbol-context", start+5)
	followed := query("Identifier", "node-symbol-context\nfollow-alias", start+5)
	if alias[2] != "1" || alias[5] != "show" || followed[5] != "write" || !containsProcessFact(followed, helper) || !containsProcessFact(followed, "FunctionDeclaration") {
		t.Fatalf("alias identity or declaration absent: %q %q", alias, followed)
	}
	declaration := query("CallExpression", "resolved-declaration", end)
	if declaration[2] != "1" || declaration[3] != helper || declaration[4] != "FunctionDeclaration" || declaration[len(declaration)-1] != "export function write():void { console.log('世界🌍'); }\n" {
		t.Fatalf("wrong resolved declaration: %q", declaration)
	}
	modules := query("CallExpression", "program-modules", end)
	if !containsProcessFact(modules, helper) || !containsProcessFact(modules, "ImportDeclaration") || !containsProcessFact(modules, source) {
		t.Fatalf("missing module graph facts: %q", modules)
	}
	for _, question := range []string{"node-symbol-context\nunknown", "program-modules\nunknown", "resolved-declaration\nunknown"} {
		if _, e := p.Inspect(entry, start, end, "CallExpression", question); e == nil {
			t.Fatal("malformed question accepted", question)
		}
	}
	if _, e := p.Inspect(entry, start, start+5, "Identifier", "resolved-declaration"); e == nil {
		t.Fatal("signature query accepted noncall")
	}
}
func containsProcessFact(fields []string, want string) bool {
	for _, f := range fields {
		if f == want {
			return true
		}
	}
	return false
}
