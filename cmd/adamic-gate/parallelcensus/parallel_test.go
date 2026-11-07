package parallelcensus

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

func TestEveryTestIsParallelOrSaysWhy(test *testing.T) {
	test.Parallel()
	// Reads: stage1/cohere/typeaware/volume_test.go
	// Reads: internal/native/normalize_test.go
	// Reads: internal/native/string_index_test.go
	// Reads: internal/oracle/input_test.go
	// Reads: internal/oracle/oracle_test.go
	// Reads: stage1/cohere/css/print_test.go
	// Reads: stage1/cohere/lint/lint_test.go
	// Reads: *_test.go
	root := parallelCheckRoot(test)
	failures := make(map[string]bool)
	tests := 0
	parallelCheckEachTestError := parallelCheckEachTest(root, func(
		rel string,
		fn *ast.FuncDecl,
		parameter string,
		positions *token.FileSet,
		lines []string,
	) {
		tests++
		if parallelCheckFirstStatement(fn, parameter) || parallelCheckReason(fn, positions, lines) {
			return
		}
		failures[rel+"\t"+fn.Name.Name] = true
	})
	if parallelCheckEachTestError != nil {
		test.Fatal(parallelCheckEachTestError)
	}
	keys := make([]string, 0, len(failures))
	for key := range failures {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	test.Logf("checked %d top-level tests; %d fail the parallel rule", tests, len(keys))
	baselinePath := filepath.Join(root, "cmd", "adamic-gate", "parallelcensus", "testdata", "serial-baseline.txt")
	if *updateSerialBaseline {
		if createDirectoryError := os.MkdirAll(filepath.Dir(baselinePath), 0755); createDirectoryError != nil {
			test.Fatal(createDirectoryError)
		}
		content := strings.Join(keys, "\n")
		if len(keys) > 0 {
			content += "\n"
		}
		if writeError := os.WriteFile(baselinePath, []byte(content), 0644); writeError != nil {
			test.Fatal(writeError)
		}
		test.Logf("wrote %d serial baseline entries", len(keys))
		return
	}
	baseline, openError := os.Open(baselinePath)
	if openError != nil {
		test.Fatalf("read serial baseline: %v (regenerate with -update-serial-baseline)", openError)
	}
	defer baseline.Close()
	allowed := make(map[string]bool)
	scanner := bufio.NewScanner(baseline)
	for scanner.Scan() {
		key := scanner.Text()
		parts := strings.Split(key, "\t")
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" || allowed[key] {
			test.Fatalf("invalid or duplicate serial baseline entry: %q", key)
		}
		allowed[key] = true
		if !failures[key] {
			test.Logf("stale serial baseline entry (now passes or no longer exists): %s", key)
		}
	}
	if scanError := scanner.Err(); scanError != nil {
		test.Fatal(scanError)
	}
	test.Logf("serial baseline contains %d entries", len(allowed))
	for _, key := range keys {
		if !allowed[key] {
			report := test.Errorf
			if !enforceParallelRule {
				report = test.Logf
			}
			report(
				"%s: call the test parameter's Parallel() as the first statement, or add // Not parallel: <shared state it touches> directly above the test with a non-empty reason",
				key,
			)
		}
	}
}

func parallelCheckRoot(test *testing.T) string {
	test.Helper()
	dir, workingDirectoryError := os.Getwd()
	if workingDirectoryError != nil {
		test.Fatal(workingDirectoryError)
	}
	for {
		data, readError := os.ReadFile(filepath.Join(dir, "go.mod"))
		if readError == nil && strings.Contains(string(data), "module github.com/system-inc/adamic\n") {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			test.Fatal("cannot find Adamic repository root from test working directory")
		}
		dir = parent
	}
}

// parallelCheckEachTest walks every _test.go under root, skipping .git, vendor, testdata and the cohere submodule,
// and visits each top-level test with its slash-separated path from root and the name of its *testing.T parameter.
func parallelCheckEachTest(
	root string,
	visit func(rel string, fn *ast.FuncDecl, parameter string, positions *token.FileSet, lines []string),
) error {
	return filepath.WalkDir(root, func(path string, entry fs.DirEntry, pathError error) error {
		if pathError != nil {
			return pathError
		}
		if entry.IsDir() {
			if entry.Name() == ".git" ||
				entry.Name() == "vendor" ||
				entry.Name() == "testdata" ||
				path == filepath.Join(root, "cohere") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(entry.Name(), "_test.go") {
			return nil
		}
		source, readError := os.ReadFile(path)
		if readError != nil {
			return readError
		}
		positions := token.NewFileSet()
		file, parseFileError := parser.ParseFile(positions, path, source, parser.ParseComments)
		if parseFileError != nil {
			return parseFileError
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
		rel, relativePathError := filepath.Rel(root, path)
		if relativePathError != nil {
			return relativePathError
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			if parameter, ok := parallelCheckTestParameter(fn, testingNames); ok {
				visit(filepath.ToSlash(rel), fn, parameter, positions, lines)
			}
		}
		return nil
	})
}

func parallelCheckTestParameter(fn *ast.FuncDecl, testingNames map[string]bool) (string, bool) {
	name := fn.Name.Name
	if !strings.HasPrefix(name, "Test") {
		return "", false
	}
	if suffix := strings.TrimPrefix(name, "Test"); suffix != "" {
		character, _ := utf8.DecodeRuneInString(suffix)
		if unicode.IsLower(character) {
			return "", false
		}
	}
	if fn.Recv != nil ||
		fn.Body == nil ||
		fn.Type.TypeParams != nil ||
		fn.Type.Results != nil ||
		fn.Type.Params.NumFields() != 1 ||
		len(fn.Type.Params.List) != 1 {
		return "", false
	}
	parameter := fn.Type.Params.List[0]
	if len(parameter.Names) != 1 {
		return "", false
	}
	ptr, ok := parameter.Type.(*ast.StarExpr)
	if !ok {
		return "", false
	}
	if typ, ok := ptr.X.(*ast.SelectorExpr); ok {
		ident, ok := typ.X.(*ast.Ident)
		if ok && testingNames[ident.Name] && typ.Sel.Name == "T" {
			return parameter.Names[0].Name, true
		}
	}
	if typ, ok := ptr.X.(*ast.Ident); ok && testingNames["."] && typ.Name == "T" {
		return parameter.Names[0].Name, true
	}
	return "", false
}

func parallelCheckFirstStatement(fn *ast.FuncDecl, parameter string) bool {
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
	return ok && receiver.Name == parameter
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
