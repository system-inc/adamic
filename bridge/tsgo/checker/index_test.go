package checker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

func TestExactIndexMatchesCompilerNodes(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	config := filepath.Join(directory, "tsconfig.json")
	if err := os.WriteFile(config, []byte(`{"compilerOptions":{"strict":true}}`), 0600); err != nil {
		t.Fatal(err)
	}
	var paths []string
	for name, text := range map[string]string{"first.ts": "type T = number;", "second.ts": "type T = string;", "nested.ts": "/* 世界🌍 */ a+b+c;", "empty.ts": ""} {
		path := filepath.Join(directory, name)
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
		paths = append(paths, path)
	}
	program, err := Open(config, paths)
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range paths {
		source := program.Compiler.GetSourceFile(path)
		var nodes []*ast.Node
		var walk func(*ast.Node)
		walk = func(node *ast.Node) {
			nodes = append(nodes, node)
			node.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
		}
		walk(source.AsNode())
		// The compiler's raw AST is the oracle. Exercise cold and warm queries,
		// reverse order, equal spans with different kinds and different end bounds.
		for round := 0; round < 2; round++ {
			for at := len(nodes) - 1; at >= 0; at-- {
				subject := nodes[at]
				if subject.Pos() >= subject.End() && subject.Kind != ast.KindSourceFile {
					continue
				}
				expected := subject
				for _, candidate := range nodes {
					if candidate.Pos() == subject.Pos() && candidate.End() == subject.End() && candidate.Kind == subject.Kind {
						expected = candidate
						break
					}
				}
				_, got, err := program.exact(path, uint64(subject.Pos()), uint64(subject.End()), strings.TrimPrefix(subject.Kind.String(), "Kind"))
				if err != nil || got != expected {
					t.Fatalf("exact index changed compiler node: %s %v %v", path, subject.Kind, err)
				}
			}
		}
		end := uint64(len(source.Text()))
		if _, _, err := program.exact(path, 0, end, "MissingKind"); err == nil {
			t.Fatal("warm index accepted wrong kind")
		}
		if _, _, err := program.exact(path, 0, end+1, "SourceFile"); err == nil {
			t.Fatal("warm index accepted wrong range")
		}
	}
}
