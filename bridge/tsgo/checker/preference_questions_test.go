package checker

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

func TestPreferenceQuestionContracts(t *testing.T) {
	directory := t.TempDir()
	file := filepath.Join(directory, "input.a")
	config := filepath.Join(directory, "tsconfig.json")
	source := "RegExp('x');function f(){arguments;arguments.length;}new Promise((resolve,reject)=>reject(5));/* 世界 */"
	for path, text := range map[string]string{file: source, config: `{"compilerOptions":{"strict":true,"target":"ES2022"},"files":["input.a"]}`} {
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	program, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	sf := program.Compiler.GetSourceFile(file)
	checker, release := program.Compiler.GetTypeCheckerForFile(context.Background(), sf)
	defer release()
	count := 0
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		count++
		if node.Kind == ast.KindIdentifier {
			wire, err := program.Inspect(file, uint64(node.Pos()), uint64(node.End()), "Identifier", "preference-binding")
			if err != nil {
				t.Fatal(err)
			}
			got := decodedFields(t, wire)
			plain := checker.GetSymbolAtLocation(node)
			expected := []string{"1", "preference-binding"}
			for _, symbol := range []*ast.Symbol{plain, plain} {
				declarations := []*ast.Node{}
				if symbol != nil {
					declarations = symbol.Declarations
				}
				expected = append(expected, strconv.FormatUint(program.symbolID(symbol), 10), strconv.Itoa(len(declarations)))
				for _, declaration := range declarations {
					sourceFile := ast.GetSourceFileOfNode(declaration)
					flag := "0"
					if sourceFile != nil && sourceFile.IsDeclarationFile {
						flag = "1"
					}
					expected = append(expected, flag)
				}
			}
			if strings.Join(got, "|") != strings.Join(expected, "|") {
				t.Fatalf("binding: %v want %v", got, expected)
			}
			if _, err := program.Inspect(file, uint64(node.Pos()), uint64(node.End()), "Identifier", "preference-binding\nextra"); err == nil {
				t.Fatal("accepted extra binding fields")
			}
		}
		node.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(sf.AsNode())
	wire, err := program.Inspect(file, 0, uint64(len(source)), "SourceFile", "preference-structure")
	if err != nil {
		t.Fatal(err)
	}
	got := decodedFields(t, wire)
	if got[3] != strconv.Itoa(count) {
		t.Fatalf("syntax population %s want %d", got[3], count)
	}
	wire, err = program.Inspect(file, 0, 11, "CallExpression", "regex-pattern\ng\n/\n")
	if err != nil {
		t.Fatal(err)
	}
	got = decodedFields(t, wire)
	expected := []string{"1", "regex-pattern", "1", "2", "0", "1", "/", "1", "2", "\n", "0", "0"}
	if strings.Join(got, "|") != strings.Join(expected, "|") {
		t.Fatalf("pattern: %v", got)
	}
	wire, err = program.Inspect(file, 0, uint64(len(source)), "SourceFile", "source-comments")
	if err != nil {
		t.Fatal(err)
	}
	got = decodedFields(t, wire)
	start := strings.Index(source, "/*")
	expected = []string{"1", "source-comments", "1", strconv.Itoa(start), strconv.Itoa(len(source))}
	if strings.Join(got, "|") != strings.Join(expected, "|") {
		t.Fatalf("comments: %v", got)
	}
	for _, question := range []string{"preference-structure\nextra", "source-comments\nextra", "preference-binding", "regex-pattern"} {
		if _, err := program.Inspect(file, 0, uint64(len(source)), "SourceFile", question); err == nil {
			t.Fatalf("accepted malformed or wrong kind: %s", question)
		}
	}
}
