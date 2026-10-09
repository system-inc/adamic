package fixtures

import (
	"encoding/json"
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
func TestFixtureDirectoriesHaveTopLevelTests(t *testing.T) {
	t.Parallel()
	files, err := filepath.Glob("*_test.go")
	if err != nil {
		t.Fatal(err)
	}
	units := map[string]string{}
	for _, path := range files {
		active, err := build.Default.MatchFile(".", path)
		if err != nil {
			t.Fatal(err)
		}
		if !active {
			continue
		}
		source, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, declaration := range source.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok || function.Recv != nil || !strings.HasPrefix(function.Name.Name, "TestFixtures") || function.Body == nil {
				continue
			}
			// A unit is a direct call with a literal directory. It cannot hide
			// the directory in a subtest, conditional, or generated dispatch.
			if len(function.Body.List) != 1 {
				t.Errorf("%s must call testFixtureDirectory directly", function.Name.Name)
				continue
			}
			statement, ok := function.Body.List[0].(*ast.ExprStmt)
			if !ok {
				t.Errorf("%s must call testFixtureDirectory directly", function.Name.Name)
				continue
			}
			call, ok := statement.X.(*ast.CallExpr)
			if !ok || len(call.Args) != 2 {
				t.Errorf("%s must call testFixtureDirectory directly", function.Name.Name)
				continue
			}
			helper, ok := call.Fun.(*ast.Ident)
			if !ok || helper.Name != "testFixtureDirectory" {
				t.Errorf("%s must call testFixtureDirectory directly", function.Name.Name)
				continue
			}
			literal, ok := call.Args[1].(*ast.BasicLit)
			if !ok || literal.Kind != token.STRING {
				t.Errorf("%s must name a literal fixture directory", function.Name.Name)
				continue
			}
			directory, err := strconv.Unquote(literal.Value)
			if err != nil {
				t.Fatal(err)
			}
			if previous, exists := units[directory]; exists {
				t.Errorf("fixture directory %s has duplicate units %s and %s", directory, previous, function.Name.Name)
			}
			units[directory] = function.Name.Name
		}
	}
	entries, err := os.ReadDir(*fixtureRoot)
	if err != nil {
		t.Fatal(err)
	}
	var directories, fixtures int
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		directory := entry.Name()
		// Manifestless directories are owned by internal/oracle:
		// enum_initialization_reach_test.go, step16_generics_test.go,
		// step20_iteration_test.go and iteration_dispatch_test.go.
		// Their own tests pin outcomes and hold admitted programs to Node.
		if directory == "enum-init-reach" || directory == "generics" || directory == "iteration" || directory == "iteration-dispatch" {
			continue
		}
		isFixture := false
		err := filepath.WalkDir(filepath.Join(*fixtureRoot, directory), func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !entry.IsDir() && (entry.Name() == "status.json" || filepath.Ext(path) == ".a") {
				isFixture = true
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		if !isFixture {
			continue
		}
		directories++
		unit, exists := units[directory]
		if !exists {
			t.Errorf("fixture directory %s has no top-level test", directory)
			continue
		}
		data, err := os.ReadFile(filepath.Join(*fixtureRoot, directory, "status.json"))
		if err != nil {
			t.Fatal(err)
		}
		// Count the exact original status rows; the directory tests use the
		// same manifests and retain all existing per-fixture checks.
		var rows []fixture
		if err := json.Unmarshal(data, &rows); err != nil {
			t.Fatal(err)
		}
		if len(rows) == 0 {
			t.Errorf("fixture directory %s is empty", directory)
		}
		fixtures += len(rows)
		t.Logf("%s directory=%s fixtures=%d", unit, directory, len(rows))
		delete(units, directory)
	}
	for directory, unit := range units {
		t.Errorf("top-level test %s names nonexistent fixture directory %s", unit, directory)
	}
	if directories == 0 || fixtures == 0 {
		t.Fatal("empty fixture directory union")
	}
	t.Logf("union: %d directories, %d fixtures", directories, fixtures)
}
