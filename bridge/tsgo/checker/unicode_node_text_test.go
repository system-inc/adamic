package checker

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"unicode"
)

func TestUnicodeNodeText(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "input.ts")
	config := filepath.Join(dir, "tsconfig.json")
	source := `const ßName=1, İName=2, _ÜName=3, 𐐀Name=4, σName=5;export {};`
	for path, text := range map[string]string{file: source, config: `{"compilerOptions":{"target":"ES2022"}}`} {
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	p, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	checked := 0
	var walk func(*ast.Node)
	walk = func(n *ast.Node) {
		if n.Kind == ast.KindIdentifier {
			wire, err := p.Inspect(file, uint64(n.Pos()), uint64(n.End()), "Identifier", "unicode-node-text")
			if err != nil {
				t.Fatal(err)
			}
			f := decodedFields(t, wire)
			runes := []rune(n.Text())
			if len(f) != 3+3*len(runes) || f[0] != "1" || f[1] != "unicode-node-text" || f[2] != strconv.Itoa(len(runes)) {
				t.Fatal("wrong scalar frame", f)
			}
			for i, r := range runes {
				for j, want := range []rune{r, unicode.ToUpper(r), unicode.ToLower(r)} {
					if f[3+3*i+j] != strconv.Itoa(int(want)) {
						t.Fatalf("lost Unicode simple case: %q scalar %d field %d", n.Text(), i, j)
					}
				}
			}
			checked++
			if _, err := p.Inspect(file, uint64(n.Pos()), uint64(n.End()), "Identifier", "unicode-node-text\nextra"); err == nil {
				t.Fatal("accepted question suffix")
			}
		}
		n.ForEachChild(func(c *ast.Node) bool { walk(c); return false })
	}
	walk(p.Compiler.GetSourceFile(tspath.RootedFilePathFromAbsolute(file)).AsNode())
	if checked != 5 {
		t.Fatal("missing case controls", checked)
	}
}
