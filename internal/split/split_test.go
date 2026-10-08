package split_test

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/split"
)

func lowered(t *testing.T, path string) *ir.Program {
	t.Helper()
	source, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), source)
	if err != nil {
		t.Fatal(err)
	}
	return program
}

func TestFixtures(t *testing.T) {
	paths, err := filepath.Glob("testdata/*.a")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range paths {
		t.Run(filepath.Base(path), func(t *testing.T) {
			if filepath.Base(path) == "compute_json.a" {
				// The encoder lands independently. This fixture runs in the
				// compute integration witness until its implementation lands.
				if _, err := os.Stat("../ir/library_json_encode.go"); os.IsNotExist(err) {
					t.Skip("encodeJson dependency absent; run in compute integration checkout")
				} else if err != nil {
					t.Fatal(err)
				}
			}
			program := lowered(t, path)
			var output bytes.Buffer
			if err := split.Print(&output, split.Analyze(program)); err != nil {
				t.Fatal(err)
			}
			expected, err := os.ReadFile(path[:len(path)-2] + ".txt")
			if err != nil {
				t.Fatal(err)
			}
			if output.String() != string(expected) {
				t.Fatalf("decision table:\n%s\nwant:\n%s", output.String(), expected)
			}
			ordinaryOracle(t, path, program)
		})
	}
}
func ordinaryOracle(t *testing.T, path string, program *ir.Program) {
	t.Helper()
	absolute, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}
	runner, err := filepath.Abs("../../oracle/node.mjs")
	if err != nil {
		t.Fatal(err)
	}
	truth, err := exec.Command("node", "--disable-warning=ExperimentalWarning", runner, absolute).CombinedOutput()
	if err != nil {
		t.Fatalf("source Node: %v: %s", err, truth)
	}
	directory := t.TempDir()
	module, binary := filepath.Join(directory, "program.mjs"), filepath.Join(directory, "program")
	if err := os.WriteFile(module, []byte(javascript.JavaScript(program)), 0644); err != nil {
		t.Fatal(err)
	}
	if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	for _, command := range []*exec.Cmd{exec.Command("node", "--disable-warning=ExperimentalWarning", runner, module), exec.Command(binary)} {
		observed, err := command.CombinedOutput()
		if err != nil || !bytes.Equal(observed, truth) {
			t.Fatalf("%v: %v: observed %q, source Node %q", command.Args, err, observed, truth)
		}
	}
	t.Logf("Node, JavaScript and sanitized native agree: %q", truth)
}

func TestBoundaryParameterOrderAndPositions(t *testing.T) {
	program := lowered(t, "testdata/decisions.a")
	for _, function := range program.Functions {
		if function.Name != "mixed" {
			continue
		}
		if len(function.Boundary.Parameters) != 3 {
			t.Fatal("boundary parameters missing")
		}
		want := []string{"number", "string", "array"}
		for i, schema := range function.Boundary.Parameters {
			if schema == nil || schema.Nodes[schema.Root].Kind != want[i] {
				t.Fatalf("parameter %d schema: %#v", i+1, schema)
			}
		}
		array := function.Boundary.Parameters[2]
		if array.Nodes[array.Nodes[array.Root].Children[0]].Kind != "number" {
			t.Fatal("array element schema")
		}
		if function.Boundary.Return == nil || function.Boundary.Return.Nodes[function.Boundary.Return.Root].Kind != "number" {
			t.Fatal("return schema")
		}
		if filepath.Base(function.Position.File) != "decisions.a" || function.Position.Line != 112 || function.Position.Column != 10 {
			t.Fatalf("position: %+v", function.Position)
		}
		return
	}
	t.Fatal("mixed missing")
}

func TestBoundaryUnsupportedTypesAreNil(t *testing.T) {
	program := lowered(t, "testdata/decisions.a")
	for _, function := range program.Functions {
		if function.Name == "takesMap" || function.Name == "takesClosure" {
			if !reflect.DeepEqual(function.Boundary.Parameters, []*ir.JSONDecodeSchema{nil}) {
				t.Fatalf("%s: unsupported boundary %#v", function.Name, function.Boundary)
			}
		}
	}
}

func TestBoundaryMoreUnsupportedTypesAreNil(t *testing.T) {
	for _, path := range []string{"testdata/effects.a", "testdata/shapes.a"} {
		program := lowered(t, path)
		for _, function := range program.Functions {
			switch function.Name {
			case "takesSet", "takesWeak", "classParameter":
				if len(function.Boundary.Parameters) != 1 || function.Boundary.Parameters[0] != nil {
					t.Fatalf("%s schema was accepted", function.Name)
				}
			case "returnsClosure":
				if function.Boundary.Return != nil {
					t.Fatal("closure return schema was accepted")
				}
			}
		}
	}
}

func TestConstructorBoundary(t *testing.T) {
	program := lowered(t, "testdata/shapes.a")
	for _, function := range program.Functions {
		if function.Name == "Box_new" {
			if function.Position.Line != 11 || len(function.Boundary.Parameters) != 1 || function.Boundary.Parameters[0] == nil || function.Boundary.Return != nil {
				t.Fatalf("constructor metadata: %+v", function)
			}
			return
		}
	}
	t.Fatal("constructor missing")
}

type futureOperation struct{}

func (futureOperation) Type() ir.Type { return ir.Number }
func TestUnknownOperationIsImpure(t *testing.T) {
	number := &ir.JSONDecodeSchema{Root: 0, Nodes: []ir.JSONDecodeNode{{Kind: "number", Of: ir.Number}}}
	program := &ir.Program{Functions: []ir.Function{{Name: "future", Returns: ir.Number, Boundary: ir.FunctionBoundary{Return: number}, Body: []ir.Statement{ir.Evaluate{Value: futureOperation{}}, ir.Loop{Condition: ir.BooleanConstant{Value: false}}}}}}
	decisions := split.Analyze(program)
	if len(decisions) != 1 || decisions[0].Eligible || decisions[0].Reason != "impure: unknown operation futureOperation" {
		t.Fatalf("unknown operation: %+v", decisions)
	}
}

func TestBoundaryUnionFences(t *testing.T) {
	program := lowered(t, "testdata/unions.a")
	seen := 0
	for _, decision := range split.Analyze(program) {
		if decision.Name == "objectUnion" || decision.Name == "broadUnion" {
			schema := program.Functions[decision.Index].Boundary.Parameters[0]
			if schema == nil || schema.Nodes[schema.Root].Kind != "union" {
				t.Fatalf("boundary union missing: %s", decision.Name)
			}
			if decision.Eligible || decision.Reason != "signature: parameter 1 not crossable" {
				t.Fatalf("union crossed: %+v", decision)
			}
			seen++
		}
	}
	if seen != 2 {
		t.Fatalf("union fixtures missing: %d", seen)
	}
}
func TestBoundaryConstructorAlignment(t *testing.T) {
	program := lowered(t, "testdata/unions.a")
	initializers := 0
	for _, function := range program.Functions {
		if function.Position.File == "" {
			continue
		}
		if len(function.Parameters) != len(function.Boundary.Parameters) {
			t.Fatalf("misaligned %s: %d parameters, %d schemas", function.Name, len(function.Parameters), len(function.Boundary.Parameters))
		}
		if function.Name == "Parent_new_initialize" || function.Name == "Child_new_initialize" {
			if function.Boundary.Parameters[0] != nil || function.Boundary.Parameters[1] == nil {
				t.Fatalf("receiver schema: %+v", function)
			}
			initializers++
		}
	}
	if initializers != 2 {
		t.Fatalf("live initializers: %d", initializers)
	}
}

func TestBoundaryStringTable(t *testing.T) {
	program := lowered(t, "testdata/unions.a")
	if !reflect.DeepEqual(program.BoundaryStrings.Values, []string{"boundary-only"}) {
		t.Fatalf("metadata strings: %q", program.BoundaryStrings.Values)
	}
	if program.BoundaryStrings.Intern("boundary-only") != 0 || len(program.BoundaryStrings.Values) != 1 {
		t.Fatal("metadata table did not reuse literal")
	}
	for _, text := range program.Strings {
		if text == "boundary-only" {
			t.Fatal("boundary literal entered codegen table")
		}
	}
	if strings.Contains(native.C(program), "boundary-only") || strings.Contains(javascript.JavaScript(program), "boundary-only") {
		t.Fatal("metadata literal emitted")
	}
}
