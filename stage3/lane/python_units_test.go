package lane

import (
	"bytes"
	"context"
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"
)

type pythonUnit struct {
	Name     string   `json:"name"`
	Family   string   `json:"family"`
	Selector string   `json:"selector"`
	Cases    []string `json:"cases"`
}

func pythonUnits(t testing.TB) []pythonUnit {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repositoryRoot(t), "stage3/lane/python-units.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Units []pythonUnit `json:"units"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	return manifest.Units
}

func TestPythonUnitsMatchCommands(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	command := boundedCommand(ctx, repositoryRoot(t), "python3", "-B", "stage3/lane/python_units.py", "inventory")
	actual, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("inventory: %v: %s", err, actual)
	}
	expected, err := os.ReadFile(filepath.Join(repositoryRoot(t), "stage3/lane/python-units.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(expected, actual) {
		t.Fatal("Python units differ from old command case discovery; regenerate python-units.json")
	}
	var expectedNames, actualNames []string
	for _, row := range pythonUnits(t) {
		expectedNames = append(expectedNames, row.Name)
	}
	files, err := filepath.Glob(filepath.Join(repositoryRoot(t), "stage3/lane/*_test.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range files {
		source, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, declaration := range source.Decls {
			if function, ok := declaration.(*ast.FuncDecl); ok && function.Recv == nil {
				name := function.Name.Name
				if strings.HasPrefix(name, "TestStage3ApplyShard") || name == "TestStage3LaneLandingLaneTests" || name == "TestStage3LaneLaneTableTests" {
					actualNames = append(actualNames, name)
				}
			}
		}
	}
	sort.Strings(expectedNames)
	sort.Strings(actualNames)
	if !reflect.DeepEqual(expectedNames, actualNames) {
		t.Fatalf("Python Go unit names differ: want %v, got %v", expectedNames, actualNames)
	}

}

func runPythonUnit(t *testing.T, name string) {
	t.Helper()
	var selected *pythonUnit
	for _, row := range pythonUnits(t) {
		if row.Name == name {
			if selected != nil {
				t.Fatal("duplicate Python unit")
			}
			copy := row
			selected = &copy
		}
	}
	if selected == nil {
		t.Fatal("missing Python unit " + name)
	}
	store := ""
	if selected.Family == "apply" {
		store = filepath.Join(adaptedTree(t), "products")
	}
	results := t.TempDir()
	report := filepath.Join(results, "cases.json")
	ctx, cancel := context.WithTimeout(context.Background(), 55*time.Second)
	defer cancel()
	command := boundedCommand(ctx, repositoryRoot(t), "python3", "-B", "stage3/lane/python_units.py", selected.Family, selected.Selector, report)
	command.Env = append(os.Environ(), "PYTHONDONTWRITEBYTECODE=1", "ADAMIC_BUILD_CACHE=read")
	if store != "" {
		command.Env = append(command.Env, "STAGE3_PRODUCT_STORE="+store)
	}
	output, err := command.CombinedOutput()
	t.Logf("%s", output)
	if err != nil {
		t.Fatalf("Python unit: %v", err)
	}
	data, err := os.ReadFile(report)
	if err != nil {
		t.Fatal(err)
	}
	var observed struct {
		Cases []struct {
			Case   string `json:"case"`
			Status string `json:"status"`
		} `json:"cases"`
		Successful bool `json:"successful"`
	}
	if err := json.Unmarshal(data, &observed); err != nil {
		t.Fatal(err)
	}
	actual := []string{}
	for _, row := range observed.Cases {
		if row.Status != "pass" {
			t.Fatalf("Python case %s: %s", row.Case, row.Status)
		}
		actual = append(actual, row.Case)
	}
	sort.Strings(actual)
	if !observed.Successful || !reflect.DeepEqual(actual, selected.Cases) {
		t.Fatalf("Python cases differ: got %v, want %v", actual, selected.Cases)
	}
	if directory := os.Getenv("STAGE3_PYTHON_RESULTS"); directory != "" {
		if err := os.WriteFile(filepath.Join(directory, name+".json"), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestStage3LaneLandingLaneTests(t *testing.T) { t.Parallel(); runPythonUnit(t, t.Name()) }
func TestStage3LaneLaneTableTests(t *testing.T)   { t.Parallel(); runPythonUnit(t, t.Name()) }

func TestStage3ApplyShard00(t *testing.T) { t.Parallel(); runPythonUnit(t, t.Name()) }

func TestStage3ApplyShard01(t *testing.T) { t.Parallel(); runPythonUnit(t, t.Name()) }

func TestStage3ApplyShard02(t *testing.T) { t.Parallel(); runPythonUnit(t, t.Name()) }

func TestStage3ApplyShard03(t *testing.T) { t.Parallel(); runPythonUnit(t, t.Name()) }

func TestStage3ApplyShard04(t *testing.T) { t.Parallel(); runPythonUnit(t, t.Name()) }

func TestStage3ApplyShard05(t *testing.T) { t.Parallel(); runPythonUnit(t, t.Name()) }

func TestStage3ApplyShard06(t *testing.T) { t.Parallel(); runPythonUnit(t, t.Name()) }

func TestStage3ApplyShard07(t *testing.T) { t.Parallel(); runPythonUnit(t, t.Name()) }

func TestStage3ApplyShard08(t *testing.T) { t.Parallel(); runPythonUnit(t, t.Name()) }

func TestStage3ApplyShard09(t *testing.T) { t.Parallel(); runPythonUnit(t, t.Name()) }

func TestStage3ApplyShard10(t *testing.T) { t.Parallel(); runPythonUnit(t, t.Name()) }

func TestStage3ApplyShard11(t *testing.T) { t.Parallel(); runPythonUnit(t, t.Name()) }

func TestStage3ApplyShard12(t *testing.T) { t.Parallel(); runPythonUnit(t, t.Name()) }

func TestStage3ApplyShard13(t *testing.T) { t.Parallel(); runPythonUnit(t, t.Name()) }

func TestStage3ApplyShard14(t *testing.T) { t.Parallel(); runPythonUnit(t, t.Name()) }

func TestStage3ApplyShard15(t *testing.T) { t.Parallel(); runPythonUnit(t, t.Name()) }

func TestStage3ApplyShard16(t *testing.T) { t.Parallel(); runPythonUnit(t, t.Name()) }
