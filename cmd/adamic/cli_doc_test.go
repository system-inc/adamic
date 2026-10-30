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

	"github.com/system-inc/adamic/cmd/adamic/treewalk"
)

// docs/cli.md is the command's inventory, so its usage block is the usage the binary prints, and every word the
// command matches (a command name, a flag, a target) is a word that usage names. A command or flag added without
// usage, or usage changed without the page, fails here. Mutants: add `case "--quiet":` to build's flag switch,
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

// matchedWords are the string literals the command compares an argument with: anything compared with, or switched
// on as, `arguments[...]` or a name assigned from it (flag := arguments[index]), in every one of its source files.
// The census finds them by that shape, not by function name, since build's flags moved from buildWithNative to
// buildWithNativeExports without it noticing, and --debug was parsed with no usage to name it.
func matchedWords(t *testing.T) []string {
	t.Helper()
	sources, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	files := token.NewFileSet()
	var matched []string
	for _, name := range sources {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(files, name, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, declaration := range file.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok || function.Body == nil {
				continue
			}
			// The names a function binds to one argument, so `switch flag` reads as `switch arguments[index]`.
			taken := map[string]bool{}
			ast.Inspect(function.Body, func(node ast.Node) bool {
				if assign, ok := node.(*ast.AssignStmt); ok && len(assign.Lhs) == len(assign.Rhs) {
					for index, value := range assign.Rhs {
						if identifier, ok := assign.Lhs[index].(*ast.Ident); ok && isArgument(value, nil) {
							taken[identifier.Name] = true
						}
					}
				}
				return true
			})
			ast.Inspect(function.Body, func(node ast.Node) bool {
				var literals []ast.Expr
				switch node := node.(type) {
				case *ast.BinaryExpr:
					if (node.Op == token.EQL || node.Op == token.NEQ) && (isArgument(node.X, taken) || isArgument(node.Y, taken)) {
						literals = []ast.Expr{node.X, node.Y}
					}
				case *ast.SwitchStmt:
					if node.Tag != nil && isArgument(node.Tag, taken) {
						for _, statement := range node.Body.List {
							literals = append(literals, statement.(*ast.CaseClause).List...)
						}
					}
				}
				for _, expression := range literals {
					if literal, ok := expression.(*ast.BasicLit); ok && literal.Kind == token.STRING {
						word, err := strconv.Unquote(literal.Value)
						if err != nil {
							t.Fatal(err)
						}
						// An empty string is a value not yet given, never a word typed.
						if word != "" {
							matched = append(matched, word)
						}
					}
				}
				return true
			})
		}
	}
	// A census that finds no command or no build flag passes by matching nothing.
	for _, word := range []string{"build", "--count"} {
		if !slices.Contains(matched, word) {
			t.Fatalf("the census found no %q among the words the command compares an argument with; it reads the wrong shape: %q", word, matched)
		}
	}
	return matched
}

// isArgument is whether an expression is one of the command's arguments: `arguments[...]`, or a name bound to one.
func isArgument(expression ast.Expr, taken map[string]bool) bool {
	switch expression := expression.(type) {
	case *ast.IndexExpr:
		identifier, ok := expression.X.(*ast.Ident)
		return ok && identifier.Name == "arguments"
	case *ast.Ident:
		return taken[expression.Name]
	}
	return false
}

// docs/cli.md's citations of files other than Go must hold: every `path:line` a line the file has, and every
// backticked path one the tree has. treewalk checks the Go ones, whose files can sit anywhere in the tree, so a Go edit
// that breaks a citation reruns the check wherever the file is (#hbq28nr). Mutants: cite a line past a file's end, or
// name a path that isn't there.
func TestDocumentedCitationsResolve(t *testing.T) {
	t.Parallel()
	root := filepath.Join("..", "..")
	page, err := os.ReadFile(filepath.Join(root, "docs", "cli.md"))
	if err != nil {
		t.Fatal(err)
	}
	lineCitations := 0
	for _, match := range treewalk.Citation.FindAllStringSubmatch(string(page), -1) {
		path, anchor := match[1], match[2]
		if strings.HasSuffix(path, ".go") {
			continue
		}
		if anchor == "" {
			if _, err := os.Stat(filepath.Join(root, path)); err != nil {
				t.Errorf("docs/cli.md names %s, which isn't in the tree", path)
			}
			continue
		}
		lineCitations++
		text, err := os.ReadFile(filepath.Join(root, path))
		if err != nil {
			t.Errorf("docs/cli.md cites %s:%s, which isn't in the tree", path, anchor)
			continue
		}
		lines := strings.Count(string(text), "\n")
		for _, number := range strings.FieldsFunc(anchor, func(character rune) bool { return character == ',' || character == '-' }) {
			if line, err := strconv.Atoi(number); err != nil || line < 1 || line > lines {
				t.Errorf("docs/cli.md cites %s:%s, past its %d lines", path, anchor, lines)
			}
		}
	}
	// A page with no line cites would pass every check above, so it fails here instead.
	if lineCitations == 0 {
		t.Fatal("found no line citations in docs/cli.md; the pattern looks at the wrong text")
	}
}
