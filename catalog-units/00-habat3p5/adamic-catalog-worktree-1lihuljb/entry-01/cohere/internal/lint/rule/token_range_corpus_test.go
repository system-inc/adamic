package rule

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
)

// tokenRangeCorpusVariable names more roots for the proof below, separated by the path list separator: a real
// tree, such as ahra, read the way the fixtures are. Unset, the proof reads the repository's fixtures alone.
const tokenRangeCorpusVariable = "COHERE_TOKEN_RANGE_CORPUS"

// TokenRange finds a node's first token with the compiler's trivia skip, where it used to build a scanner per
// call (scanner.GetRangeOfTokenAtPosition): 252 MB and 1.5M objects of a cold ahra run (#9xfg09f).
// It keeps only the start, so the two agree exactly when they agree on every node's start. Findings identical
// on four consumers would only show that no current rule reached a shape where they differ, so this compares
// them on every node of every file: the repository's fixtures, JSX, templates and JSDoc among them, and a real
// tree when the variable names one. A start one past the scanner's is the mutant it must catch.
func TestTokenRangeStartsWhereTheScannerDoes(t *testing.T) {
	t.Parallel()
	roots := []string{repositoryRoot(t)}
	if extra := os.Getenv(tokenRangeCorpusVariable); extra != "" {
		roots = append(roots, filepath.SplitList(extra)...)
	}
	files := corpusFiles(t, roots)
	if len(files) < 50 {
		t.Fatalf("the corpus holds %d files, too few for agreement to mean anything", len(files))
	}

	scannerStart := func(sourceFile *ast.SourceFile, node *ast.Node) int {
		return scanner.GetRangeOfTokenAtPosition(sourceFile, node.Pos()).Pos()
	}
	nodes, disagreements := compareStarts(files, scannerStart, func(sourceFile *ast.SourceFile, node *ast.Node) int {
		return TokenRange(sourceFile, node).Pos()
	})
	if len(disagreements) > 0 {
		t.Errorf("TokenRange and the scanner disagree on %d of %d nodes, by kind: %s", total(disagreements), nodes, summary(disagreements))
	}
	t.Logf("%d nodes in %d files compared, %d disagreeing", nodes, len(files), total(disagreements))

	_, caught := compareStarts(files, scannerStart, func(sourceFile *ast.SourceFile, node *ast.Node) int {
		return TokenRange(sourceFile, node).Pos() + 1
	})
	if total(caught) == 0 {
		t.Error("a start one past the scanner's agreed on every node, so this comparison could not have failed")
	}
}

// repositoryRoot is the module's root, where the fixtures are.
func repositoryRoot(t *testing.T) string {
	t.Helper()
	directory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(directory, "go.mod")); err == nil {
			return directory
		}
		parent := filepath.Dir(directory)
		if parent == directory {
			t.Fatal("no go.mod above the test's directory")
		}
		directory = parent
	}
}

// corpusFiles is every TypeScript and JavaScript source under roots, outside dependency, build and cache
// directories and the TypeScript checkout, each parsed.
func corpusFiles(t *testing.T, roots []string) []*ast.SourceFile {
	t.Helper()
	kinds := map[string]core.ScriptKind{".ts": core.ScriptKindTS, ".mts": core.ScriptKindTS, ".cts": core.ScriptKindTS,
		".tsx": core.ScriptKindTSX, ".js": core.ScriptKindJS, ".jsx": core.ScriptKindJSX, ".mjs": core.ScriptKindJS}
	skipped := map[string]bool{"node_modules": true, ".git": true, ".cache": true, ".next": true, "dist": true, "TypeScript": true}
	var files []*ast.SourceFile
	for _, root := range roots {
		filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if entry.IsDir() {
				if path != root && skipped[entry.Name()] {
					return filepath.SkipDir
				}
				return nil
			}
			kind, handled := kinds[strings.ToLower(filepath.Ext(path))]
			if !handled || entry.Type()&fs.ModeSymlink != 0 {
				return nil
			}
			contents, err := os.ReadFile(path)
			if err != nil {
				return nil
			}
			fileName := tspath.NormalizePath(path)
			if sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: tspath.RootedFilePath(fileName), PathKey: tspath.PathKey(fileName)},
				string(contents), kind); sourceFile != nil {
				files = append(files, sourceFile)
			}
			return nil
		})
	}
	return files
}

// compareStarts counts the nodes of files, JSDoc included, and the kinds of those whose starts differ.
func compareStarts(files []*ast.SourceFile, expected func(*ast.SourceFile, *ast.Node) int,
	actual func(*ast.SourceFile, *ast.Node) int) (int, map[string]int) {
	nodes := 0
	disagreements := map[string]int{}
	for _, sourceFile := range files {
		var visit func(node *ast.Node) bool
		visit = func(node *ast.Node) bool {
			nodes++
			if expected(sourceFile, node) != actual(sourceFile, node) {
				disagreements[node.Kind.String()]++
			}
			for _, documentation := range node.JSDoc(sourceFile) {
				visit(documentation)
			}
			node.ForEachChild(visit)
			return false
		}
		sourceFile.AsNode().ForEachChild(visit)
	}
	return nodes, disagreements
}

func total(counts map[string]int) int {
	sum := 0
	for _, count := range counts {
		sum += count
	}
	return sum
}

func summary(counts map[string]int) string {
	kinds := make([]string, 0, len(counts))
	for kind, count := range counts {
		kinds = append(kinds, fmt.Sprintf("%s %d", kind, count))
	}
	sort.Strings(kinds)
	return strings.Join(kinds, ", ")
}
