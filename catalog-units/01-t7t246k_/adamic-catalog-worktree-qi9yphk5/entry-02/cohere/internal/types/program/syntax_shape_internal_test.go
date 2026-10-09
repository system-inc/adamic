package program

import (
	"os"
	"path/filepath"
	"testing"
)

// syntaxOf parses text as lib.ts in a fresh project and returns its syntax hash.
func syntaxOf(t *testing.T, text string) string {
	t.Helper()
	root := t.TempDir()
	config := `{"compilerOptions": {"target": "ES2022", "module": "esnext", "moduleResolution": "bundler", "strict": true, "noEmit": true}}`
	if err := os.WriteFile(filepath.Join(root, "tsconfig.json"), []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "lib.ts"), []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	graph, err := Build(Options{ConfigFileName: filepath.Join(root, "tsconfig.json")})
	if err != nil {
		t.Fatal(err)
	}
	for _, sourceFile := range graph.ProjectFiles() {
		if filepath.Base(sourceFile.FileName().AsString()) == "lib.ts" {
			return elidedBodiesVersion(sourceFile)
		}
	}
	t.Fatal("lib.ts is not in the program")
	return ""
}

// The syntax half of a shape leaves out what no rule can read off an imported declaration without reading
// the file's text, plain comments and whitespace, and keeps what a rule or the compiler reads without it:
// JSDoc, directives, pragmas, and every token, a regular expression or string holding "//" included
// (#zqsdzbq: a comment appended to a file moved every importer's shape key).
func TestTheSyntaxShapeIgnoresPlainCommentsAndKeepsWhatCanBeRead(t *testing.T) {
	t.Parallel()
	base := "/** The value. */\nexport function value(a: string): string {\n  return a;\n}\nexport const pattern = /a\\/\\/b/;\nexport const url = \"http://x\";\n"
	original := syntaxOf(t, base)

	same := map[string]string{
		"a comment appended":           base + "// bench edit 1\n",
		"a block comment appended":     base + "/* note */\n",
		"a comment between exports":    "/** The value. */\nexport function value(a: string): string {\n  return a;\n}\n// between\nexport const pattern = /a\\/\\/b/;\nexport const url = \"http://x\";\n",
		"a comment inside a parameter": "/** The value. */\nexport function value(a: /* the input */ string): string {\n  return a;\n}\nexport const pattern = /a\\/\\/b/;\nexport const url = \"http://x\";\n",
		"blank lines and indentation":  "\n\n/** The value. */\nexport function value(a: string):  string {\n  return a;\n}\n\nexport const pattern = /a\\/\\/b/;\n  export const url = \"http://x\";\n",
		"an edit inside the body":      "/** The value. */\nexport function value(a: string): string {\n  return a + \"\";\n}\nexport const pattern = /a\\/\\/b/;\nexport const url = \"http://x\";\n",
	}
	for name, text := range same {
		if got := syntaxOf(t, text); got != original {
			t.Errorf("%s moved the syntax shape", name)
		}
	}

	moved := map[string]string{
		"the JSDoc edited":            "/** The old value. */\nexport function value(a: string): string {\n  return a;\n}\nexport const pattern = /a\\/\\/b/;\nexport const url = \"http://x\";\n",
		"a directive added":           "/// <reference lib=\"dom\" />\n" + base,
		"a pragma added":              "// @ts-nocheck\n" + base,
		"a regular expression edited": "/** The value. */\nexport function value(a: string): string {\n  return a;\n}\nexport const pattern = /a\\/\\/c/;\nexport const url = \"http://x\";\n",
		"a string holding // edited":  "/** The value. */\nexport function value(a: string): string {\n  return a;\n}\nexport const pattern = /a\\/\\/b/;\nexport const url = \"http://y\";\n",
		"a parameter renamed":         "/** The value. */\nexport function value(b: string): string {\n  return b;\n}\nexport const pattern = /a\\/\\/b/;\nexport const url = \"http://x\";\n",
		// Trivia appearing where there was none is a change: the separator says the tokens were apart.
		"a comment where there was no space": "/** The value. */\nexport function value(/* the input */a: string): string {\n  return a;\n}\nexport const pattern = /a\\/\\/b/;\nexport const url = \"http://x\";\n",
		"two tokens a space apart joined":    "/** The value. */\nexport function value(a: string): string {\n  return a;\n}\nexport const pattern = /a\\/\\/b/;\nexport const url = \"http://x\";\nexport type T = typeof url;\n",
	}
	for name, text := range moved {
		if got := syntaxOf(t, text); got == original {
			t.Errorf("%s left the syntax shape unmoved", name)
		}
	}
}
