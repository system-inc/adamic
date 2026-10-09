package main

import (
	"bufio"
	"flag"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"
)

var updateSerialBaseline = flag.Bool("update-serial-baseline", false, "rewrite the repository's serial-test baseline")

// enforceParallelRule was off for an hour on Oct 9: hundreds of serial shard tests landed through integration's
// test-only lane, which ran no checks then, after the baseline, and turned main red. Since cloud/merge-tree 81cde1e6
// the lane runs this test on the merged tree (lane-checks.py), so a violator is refused at its own landing.
const enforceParallelRule = true

func TestEveryTestIsParallelOrSaysWhy(t *testing.T) {
	t.Parallel()
	root := parallelCheckRoot(t)
	failures := make(map[string]bool)
	tests := 0
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if entry.Name() == ".git" || entry.Name() == "vendor" || entry.Name() == "testdata" || path == filepath.Join(root, "cohere") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(entry.Name(), "_test.go") {
			return nil
		}
		source, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		positions := token.NewFileSet()
		file, err := parser.ParseFile(positions, path, source, parser.ParseComments)
		if err != nil {
			return err
		}
		testingNames := make(map[string]bool)
		for _, imp := range file.Imports {
			if imp.Path.Value == `"testing"` {
				name := "testing"
				if imp.Name != nil {
					name = imp.Name.Name
				}
				testingNames[name] = true
			}
		}
		lines := strings.Split(string(source), "\n")
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			param, ok := parallelCheckTestParameter(fn, testingNames)
			if !ok {
				continue
			}
			tests++
			if parallelCheckFirstStatement(fn, param) || parallelCheckReason(fn, positions, lines) {
				continue
			}
			failures[filepath.ToSlash(rel)+"\t"+fn.Name.Name] = true
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	keys := make([]string, 0, len(failures))
	for key := range failures {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	t.Logf("checked %d top-level tests; %d fail the parallel rule", tests, len(keys))
	baselinePath := filepath.Join(root, "cmd", "adamic-gate", "testdata", "serial-baseline.txt")
	if *updateSerialBaseline {
		if err := os.MkdirAll(filepath.Dir(baselinePath), 0755); err != nil {
			t.Fatal(err)
		}
		content := strings.Join(keys, "\n")
		if len(keys) > 0 {
			content += "\n"
		}
		if err := os.WriteFile(baselinePath, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
		t.Logf("wrote %d serial baseline entries", len(keys))
		return
	}
	baseline, err := os.Open(baselinePath)
	if err != nil {
		t.Fatalf("read serial baseline: %v (regenerate with -update-serial-baseline)", err)
	}
	defer baseline.Close()
	allowed := make(map[string]bool)
	scanner := bufio.NewScanner(baseline)
	for scanner.Scan() {
		key := scanner.Text()
		parts := strings.Split(key, "\t")
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" || allowed[key] {
			t.Fatalf("invalid or duplicate serial baseline entry: %q", key)
		}
		allowed[key] = true
		if !failures[key] {
			t.Logf("stale serial baseline entry (now passes or no longer exists): %s", key)
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	t.Logf("serial baseline contains %d entries", len(allowed))
	for _, key := range keys {
		if !allowed[key] {
			report := t.Errorf
			if !enforceParallelRule {
				report = t.Logf
			}
			report("%s: call the test parameter's Parallel() as the first statement, or add // Not parallel: <shared state it touches> directly above the test with a non-empty reason", key)
		}
	}
}

func parallelCheckRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
		if err == nil && strings.Contains(string(data), "module github.com/system-inc/adamic\n") {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("cannot find Adamic repository root from test working directory")
		}
		dir = parent
	}
}

func parallelCheckTestParameter(fn *ast.FuncDecl, testingNames map[string]bool) (string, bool) {
	name := fn.Name.Name
	if !strings.HasPrefix(name, "Test") {
		return "", false
	}
	if suffix := strings.TrimPrefix(name, "Test"); suffix != "" {
		r, _ := utf8.DecodeRuneInString(suffix)
		if unicode.IsLower(r) {
			return "", false
		}
	}
	if fn.Recv != nil || fn.Body == nil || fn.Type.TypeParams != nil || fn.Type.Results != nil || fn.Type.Params.NumFields() != 1 || len(fn.Type.Params.List) != 1 {
		return "", false
	}
	param := fn.Type.Params.List[0]
	if len(param.Names) != 1 {
		return "", false
	}
	ptr, ok := param.Type.(*ast.StarExpr)
	if !ok {
		return "", false
	}
	if typ, ok := ptr.X.(*ast.SelectorExpr); ok {
		pkg, ok := typ.X.(*ast.Ident)
		if ok && testingNames[pkg.Name] && typ.Sel.Name == "T" {
			return param.Names[0].Name, true
		}
	}
	if typ, ok := ptr.X.(*ast.Ident); ok && testingNames["."] && typ.Name == "T" {
		return param.Names[0].Name, true
	}
	return "", false
}

func parallelCheckFirstStatement(fn *ast.FuncDecl, param string) bool {
	if len(fn.Body.List) == 0 {
		return false
	}
	stmt, ok := fn.Body.List[0].(*ast.ExprStmt)
	if !ok {
		return false
	}
	call, ok := stmt.X.(*ast.CallExpr)
	if !ok || len(call.Args) != 0 {
		return false
	}
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "Parallel" {
		return false
	}
	receiver, ok := selector.X.(*ast.Ident)
	return ok && receiver.Name == param
}

func parallelCheckReason(fn *ast.FuncDecl, positions *token.FileSet, lines []string) bool {
	hasReason := func(line string) bool {
		_, reason, found := strings.Cut(line, "Not parallel:")
		// Strip block-comment delimiters before checking for a non-empty reason.
		reason = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(reason), "*/"))
		return found && reason != ""
	}
	if fn.Doc != nil {
		for _, comment := range fn.Doc.List {
			for _, line := range strings.Split(comment.Text, "\n") {
				if hasReason(line) {
					return true
				}
			}
		}
	}
	line := positions.Position(fn.Pos()).Line
	if line > 1 {
		above := strings.TrimSpace(lines[line-2])
		return strings.HasPrefix(above, "//") && hasReason(above)
	}
	return false
}
