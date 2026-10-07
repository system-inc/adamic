package checker

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

func TestJSXStructureRoles(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "input.tsx")
	config := filepath.Join(dir, "tsconfig.json")
	source := "const x=<button type='世界🌍'/>; function C(){return <input checked/>;}"
	for path, text := range map[string]string{file: source, config: `{"compilerOptions":{"strict":true,"jsx":"preserve"},"files":["input.tsx"]}`} {
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	p, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	wire, err := p.Inspect(file, 0, uint64(len(source)), "SourceFile", "jsx-structure")
	if err != nil {
		t.Fatal(err)
	}
	fields := decodedFields(t, wire)
	var nodes []*ast.Node
	ids := map[*ast.Node]int{}
	var walk func(*ast.Node)
	walk = func(n *ast.Node) {
		ids[n] = len(nodes) + 1
		nodes = append(nodes, n)
		n.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(p.Compiler.GetSourceFile(file).AsNode())
	if fields[2] != strconv.Itoa(len(nodes)) || len(fields) != 3+10*len(nodes) {
		t.Fatal("wrong syntax frame count")
	}
	for i, n := range nodes {
		if n.Kind == ast.KindJsxSelfClosingElement {
			opening := n.AsJsxSelfClosingElement()
			offset := 3 + 10*i
			if fields[offset+2] != strconv.Itoa(ids[opening.TagName]) || fields[offset+3] != strconv.Itoa(ids[opening.Attributes]) {
				t.Fatal("tag and attributes role mismatch")
			}
		}
	}
	if _, err := p.Inspect(file, 0, uint64(len(source)), "SourceFile", "jsx-structure\nextra"); err == nil {
		t.Fatal("accepted malformed syntax question")
	}
}
