package fixtures

import (
	"encoding/json"
	"github.com/system-inc/adamic/internal/hostconfiguration"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// Inspect the actual top-level declarations, rather than a second registry that
// could stay green after a test is removed or renamed out of go test's list.
func TestFixtureDirectoriesHaveTopLevelTests(test *testing.T) {
	test.Parallel()
	files, globError := filepath.Glob("*_test.go")
	if globError != nil {
		test.Fatal(globError)
	}
	units := map[string]string{}
	selectedFiles := map[string]map[string]string{}
	for _, path := range files {
		active, matchFileError := build.Default.MatchFile(".", path)
		if matchFileError != nil {
			test.Fatal(matchFileError)
		}
		if !active {
			continue
		}
		source, parseFileError := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if parseFileError != nil {
			test.Fatal(parseFileError)
		}
		for _, declaration := range source.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok ||
				function.Recv != nil ||
				!strings.HasPrefix(function.Name.Name, "TestFixtures") ||
				function.Body == nil {
				continue
			}
			// Each unit makes its parallel contract visible to the lane analyzer,
			// then calls the helper directly with its own parameter and directory.
			if len(function.Type.Params.List) != 1 ||
				len(function.Type.Params.List[0].Names) != 1 ||
				len(function.Body.List) != 2 {
				test.Errorf(
					"%s must call its test parameter's Parallel() then testFixtureDirectory directly",
					function.Name.Name,
				)
				continue
			}
			parameter := function.Type.Params.List[0].Names[0].Name
			parallelStatement, ok := function.Body.List[0].(*ast.ExprStmt)
			if !ok {
				test.Errorf("%s must call its test parameter's Parallel() first", function.Name.Name)
				continue
			}
			parallelCall, ok := parallelStatement.X.(*ast.CallExpr)
			if !ok || len(parallelCall.Args) != 0 || parallelCall.Ellipsis.IsValid() {
				test.Errorf("%s must call its test parameter's Parallel() first", function.Name.Name)
				continue
			}
			parallel, ok := parallelCall.Fun.(*ast.SelectorExpr)
			if !ok || parallel.Sel.Name != "Parallel" {
				test.Errorf("%s must call its test parameter's Parallel() first", function.Name.Name)
				continue
			}
			receiver, ok := parallel.X.(*ast.Ident)
			if !ok || receiver.Name != parameter {
				test.Errorf("%s must call its test parameter's Parallel() first", function.Name.Name)
				continue
			}
			statement, ok := function.Body.List[1].(*ast.ExprStmt)
			if !ok {
				test.Errorf("%s must call testFixtureDirectory directly", function.Name.Name)
				continue
			}
			call, ok := statement.X.(*ast.CallExpr)
			if !ok || len(call.Args) < 2 || call.Ellipsis.IsValid() {
				test.Errorf("%s must call testFixtureDirectory directly", function.Name.Name)
				continue
			}
			argument, ok := call.Args[0].(*ast.Ident)
			if !ok || argument.Name != parameter {
				test.Errorf("%s must pass its own test parameter to testFixtureDirectory", function.Name.Name)
				continue
			}
			helper, ok := call.Fun.(*ast.Ident)
			if !ok || helper.Name != "testFixtureDirectory" {
				test.Errorf("%s must call testFixtureDirectory directly", function.Name.Name)
				continue
			}
			literal, ok := call.Args[1].(*ast.BasicLit)
			if !ok || literal.Kind != token.STRING {
				test.Errorf("%s must name a literal fixture directory", function.Name.Name)
				continue
			}
			directory, unquoteError := strconv.Unquote(literal.Value)
			if unquoteError != nil {
				test.Fatal(unquoteError)
			}
			if len(call.Args) == 2 {
				if previous, exists := units[directory]; exists {
					test.Errorf(
						"fixture directory %s has duplicate units %s and %s",
						directory,
						previous,
						function.Name.Name,
					)
				}
				if len(selectedFiles[directory]) != 0 {
					test.Errorf("fixture directory %s mixes directory and fixture units", directory)
				}
				units[directory] = function.Name.Name
			} else {
				if _, exists := units[directory]; exists {
					test.Errorf("fixture directory %s mixes directory and fixture units", directory)
				}
				if selectedFiles[directory] == nil {
					selectedFiles[directory] = map[string]string{}
				}
				for _, argument := range call.Args[2:] {
					literal, ok := argument.(*ast.BasicLit)
					if !ok || literal.Kind != token.STRING {
						test.Errorf("%s must name literal fixture files", function.Name.Name)
						continue
					}
					file, unquoteError := strconv.Unquote(literal.Value)
					if unquoteError != nil {
						test.Fatal(unquoteError)
					}
					if !validFixturePath(file) {
						test.Errorf("invalid fixture unit %q", file)
					}
					if previous, exists := selectedFiles[directory][file]; exists {
						test.Errorf(
							"fixture %s/%s has duplicate units %s and %s",
							directory,
							file,
							previous,
							function.Name.Name,
						)
					}
					selectedFiles[directory][file] = function.Name.Name
				}
			}
		}
	}
	entries, readDirectoryError := os.ReadDir(*fixtureRoot)
	if readDirectoryError != nil {
		test.Fatal(readDirectoryError)
	}
	var directories, fixtures int
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		directory := entry.Name()
		// This manifestless directory belongs to internal/oracle, through
		// enum_initialization_reach_test.go; it was never in TestFixtures.
		if directory == "enum-init-reach" {
			continue
		}
		isFixture := false
		walkError := filepath.WalkDir(filepath.Join(
			*fixtureRoot,
			directory,
		), func(path string, entry fs.DirEntry, pathError error) error {
			if pathError != nil {
				return pathError
			}
			if !entry.IsDir() &&
				(entry.Name() == "status.json" || filepath.Ext(path) == ".a" || filepath.Ext(path) == ".ts") {
				isFixture = true
			}
			return nil
		})
		if walkError != nil {
			test.Fatal(walkError)
		}
		if !isFixture {
			continue
		}
		directories++
		unit, exists := units[directory]
		if !exists && len(selectedFiles[directory]) == 0 {
			test.Errorf("fixture directory %s has no top-level test", directory)
			continue
		}
		data, readError := os.ReadFile(filepath.Join(*fixtureRoot, directory, "status.json"))
		if readError != nil {
			test.Fatal(readError)
		}
		// Count the exact original status rows; the directory tests use the
		// same manifests and retain all existing per-fixture checks.
		var rows []fixture
		if decodeError := json.Unmarshal(data, &rows); decodeError != nil {
			test.Fatal(decodeError)
		}
		if len(rows) == 0 {
			test.Errorf("fixture directory %s is empty", directory)
		}
		if !exists {
			for _, row := range rows {
				if _, owned := selectedFiles[directory][row.File]; !owned {
					test.Errorf("fixture %s/%s has no top-level test", directory, row.File)
				}
				delete(selectedFiles[directory], row.File)
			}
			for file := range selectedFiles[directory] {
				test.Errorf("fixture unit names nonexistent fixture %s/%s", directory, file)
			}
			delete(selectedFiles, directory)
		}
		for index, row := range rows {
			_, skip, resolveError := hostconfiguration.Resolve(index+1, row.Configuration, os.Getenv)
			if resolveError != nil {
				test.Fatal(resolveError)
			}
			if skip != "" {
				test.Log(skip)
			}
		}
		fixtures += len(rows)
		test.Logf("%s directory=%s fixtures=%d", unit, directory, len(rows))
		delete(units, directory)
	}
	for directory, unit := range units {
		test.Errorf("top-level test %s names nonexistent fixture directory %s", unit, directory)
	}
	for directory := range selectedFiles {
		test.Errorf("fixture units name nonexistent directory %s", directory)
	}
	if directories == 0 || fixtures == 0 {
		test.Fatal("empty fixture directory union")
	}
	test.Logf("union: %d directories, %d fixtures", directories, fixtures)
}
