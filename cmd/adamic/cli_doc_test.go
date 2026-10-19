package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// docs/cli.md is the command's inventory, so its usage block is the usage the binary prints, and every word the
// command matches (a command name, a flag, a target) is a word that usage names. A command or flag added without
// usage, or usage changed without the page, fails here. Mutants: add `case "--quiet":` to buildWithNative's switch,
// or change a word of usage, or of the page's block.
func TestDocumentedUsageIsTheCommands(t *testing.T) {
	t.Parallel()
	page, err := os.ReadFile(filepath.Join("..", "..", "docs", "cli.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(page), "```text\n"+usage+"\n```\n") {
		t.Errorf("docs/cli.md's usage block is not the usage cmd/adamic prints:\n%s", usage)
	}
	words := strings.FieldsFunc(usage, func(character rune) bool {
		return character == ' ' || character == '\n' || character == '[' || character == ']'
	})
	for _, matched := range matchedWords(t) {
		if !slices.Contains(words, matched) {
			t.Errorf("cmd/adamic matches %q, which usage doesn't name", matched)
		}
	}
}

// matchedWords are the string literals run compares an argument with and the cases buildWithNative switches on.
func matchedWords(t *testing.T) []string {
	t.Helper()
	files := token.NewFileSet()
	var matched []string
	for _, name := range []string{"main.go", "tsgo.go"} {
		file, err := parser.ParseFile(files, name, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, declaration := range file.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok || (function.Name.Name != "run" && function.Name.Name != "buildWithNative") {
				continue
			}
			ast.Inspect(function.Body, func(node ast.Node) bool {
				var literals []ast.Expr
				switch node := node.(type) {
				case *ast.BinaryExpr:
					if node.Op == token.EQL || node.Op == token.NEQ {
						literals = []ast.Expr{node.X, node.Y}
					}
				case *ast.CaseClause:
					literals = node.List
				}
				for _, expression := range literals {
					if literal, ok := expression.(*ast.BasicLit); ok && literal.Kind == token.STRING {
						word, err := strconv.Unquote(literal.Value)
						if err != nil {
							t.Fatal(err)
						}
						// An empty string is a value not yet given (options.Target != ""), never a word typed.
						if word != "" {
							matched = append(matched, word)
						}
					}
				}
				return true
			})
		}
	}
	if len(matched) == 0 {
		t.Fatal("found no words run or buildWithNative match; the census looks at the wrong functions")
	}
	return matched
}
