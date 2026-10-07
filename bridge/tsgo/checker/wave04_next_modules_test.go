package checker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWave04NextModuleAndSourceContext(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	source := filepath.Join(directory, "source.a")
	target := filepath.Join(directory, "target.ts")
	config := filepath.Join(directory, "tsconfig.json")
	text := "import type { Input } from './target.ts';\nimport { value } from './target.ts';\nasync function nested(){await value();}\nimport(specifier);\nawait import('./target.ts');\n"
	for path, data := range map[string]string{source: text, target: "export type Input=number; export function value():void{}", config: `{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"]},"files":["prelude.d.ts"]}`, filepath.Join(directory, "prelude.d.ts"): "declare const specifier: string;"} {
		if err := os.WriteFile(path, []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	p, err := Open(config, []string{source, target})
	if err != nil {
		t.Fatal(err)
	}
	wire, err := p.InspectWave04Next(source, 0, uint64(len(text)), "SourceFile", "wave04-next-program-modules")
	if err != nil {
		t.Fatal(err)
	}
	got := decodedFields(t, wire)
	if len(got) < 3 || got[0] != "1" || got[1] != "wave04-next-program-modules" {
		t.Fatalf("module header: %v", got)
	}
	joined := strings.Join(got, "\n")
	for _, needle := range []string{"ImportDeclaration\n1", "ImportDeclaration\n0", "ImportKeyword\n\n1\nIdentifier\n\n", "StringLiteral\n./target.ts\n" + target} {
		if !strings.Contains(joined, needle) {
			t.Fatalf("raw module fact absent: %q", needle)
		}
	}
	wire, err = p.InspectWave04Next(source, 0, uint64(len(text)), "SourceFile", "wave04-next-source-context")
	if err != nil {
		t.Fatal(err)
	}
	got = decodedFields(t, wire)
	if len(got) != 3+5*4 || got[2] != "5" {
		t.Fatalf("raw statement count/shape: %v", got)
	}
	// A nested async body's await does not change its declaration's top-level context.
	if got[3+2*4+3] != "0" || got[3+4*4+3] != "1" {
		t.Fatalf("top-level versus nested await: %v", got)
	}
}
