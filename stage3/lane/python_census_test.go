package lane

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// A Python suite under stage3 that no gate runs can't fail, so it isn't true (#mpesmv7): eleven ran nowhere until
// they were wrapped. Every stage3 test_*.py is either run by one of python_test.go's wrappers or listed here with
// why it isn't, and a new one with neither is red. stage3/typescript is the upstream harness's own, never ours.

// unwrappedPythonSuites are the stage3 test_*.py no wrapper runs, each with why.
var unwrappedPythonSuites = map[string]string{
	"stage3/census/tsc-closure/test_fixture.py": "the known-answer check of a one-off measurement (rerun-2026-10-09): it takes the census binary " +
		"and hidden-tools tree that prepare.py builds from f1502d13's tools against main with a full cohere checkout, as " +
		"arguments, so it has nothing to run on alone",
}

func TestEveryStage3PythonSuiteRunsOrSaysWhy(t *testing.T) {
	t.Parallel()
	root := repositoryRoot(t)
	suites := stage3PythonSuites(t, root)
	wrapped := wrappedPythonSuites(t, filepath.Join(root, "stage3/lane/python_test.go"))
	if gaps := pythonSuiteGaps(suites, wrapped, unwrappedPythonSuites); len(gaps) > 0 {
		t.Fatalf("stage3 Python suites no wrapper runs and none lists: %s", strings.Join(gaps, ", "))
	}
	for suite := range unwrappedPythonSuites {
		if wrapped[suite] {
			t.Errorf("%s is listed as unwrapped, but a wrapper runs it", suite)
		}
	}
	t.Logf("%d stage3 Python suites: %d wrapped, %d listed", len(suites), len(suites)-len(unwrappedPythonSuites), len(unwrappedPythonSuites))
}

// A suite the census can't place is named, and so is a wrapped or listed one that no longer exists. Mutants: a new
// test_x.py with neither a wrapper nor a reason, and a wrapper naming a suite that's gone.
func TestPythonSuiteGapsNamesWhatNoWrapperRuns(t *testing.T) {
	t.Parallel()
	suites := []string{"stage3/test_a.py", "stage3/lane/test_b.py", "stage3/census/x/test_c.py", "stage3/test_x.py"}
	wrapped := map[string]bool{"stage3/test_a.py": true, "stage3/lane/test_b.py": true, "stage3/lane/test_gone.py": true}
	listed := map[string]string{"stage3/census/x/test_c.py": "a script"}
	gaps := pythonSuiteGaps(suites, wrapped, listed)
	want := []string{"stage3/lane/test_gone.py (wrapped, but not in the tree)", "stage3/test_x.py"}
	if !slices.Equal(gaps, want) {
		t.Fatalf("gaps = %q, want %q", gaps, want)
	}
}

// pythonSuiteGaps is every suite neither wrapped nor listed, and every wrapped or listed suite not among suites, sorted.
func pythonSuiteGaps(suites []string, wrapped map[string]bool, listed map[string]string) []string {
	present := map[string]bool{}
	gaps := []string{}
	for _, suite := range suites {
		present[suite] = true
		if _, ok := listed[suite]; !wrapped[suite] && !ok {
			gaps = append(gaps, suite)
		}
	}
	for suite := range wrapped {
		if !present[suite] {
			gaps = append(gaps, suite+" (wrapped, but not in the tree)")
		}
	}
	for suite := range listed {
		if !present[suite] {
			gaps = append(gaps, suite+" (listed, but not in the tree)")
		}
	}
	slices.Sort(gaps)
	return gaps
}

// stage3PythonSuites is every test_*.py under root's stage3, slash-separated from root, leaving out stage3/typescript
// and any node_modules. It walks the files, since a runner's tree has no .git to list.
func stage3PythonSuites(t *testing.T, root string) []string {
	t.Helper()
	suites := []string{}
	err := filepath.WalkDir(filepath.Join(root, "stage3"), func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, name)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		if entry.IsDir() {
			if relative == "stage3/typescript" || entry.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if base := path.Base(relative); strings.HasPrefix(base, "test_") && strings.HasSuffix(base, ".py") {
			suites = append(suites, relative)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return suites
}

// wrappedPythonSuites is every suite the wrappers' calls name, read from the source: a runPython call's
// directory and modules, and a runPythonCommand call's directory and the .py script among its arguments. Only string
// literals count, so a wrapper is read as written.
func wrappedPythonSuites(t *testing.T, file string) map[string]bool {
	t.Helper()
	parsed, err := parser.ParseFile(token.NewFileSet(), file, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	literal := func(expression ast.Expr) (string, bool) {
		basic, ok := expression.(*ast.BasicLit)
		if !ok || basic.Kind != token.STRING {
			return "", false
		}
		value, err := strconv.Unquote(basic.Value)
		return value, err == nil
	}
	wrapped := map[string]bool{}
	ast.Inspect(parsed, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		function, ok := call.Fun.(*ast.Ident)
		if !ok || len(call.Args) < 3 {
			return true
		}
		directory, ok := literal(call.Args[1])
		if !ok {
			return true
		}
		switch function.Name {
		case "runPython":
			for _, argument := range call.Args[2:] {
				if module, ok := literal(argument); ok {
					wrapped[directory+"/"+module+".py"] = true
				}
			}
		case "runPythonCommand":
			for _, argument := range call.Args[3:] {
				if script, ok := literal(argument); ok && strings.HasSuffix(script, ".py") {
					wrapped[directory+"/"+script] = true
				}
			}
		}
		return true
	})
	return wrapped
}
