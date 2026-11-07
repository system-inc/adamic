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

func lowered(test *testing.T, path string) *ir.Program {
	test.Helper()
	source, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	program, lowerError := lower.Lower(context.Background(), source)
	if lowerError != nil {
		test.Fatal(lowerError)
	}
	return program
}

func TestFixtures(test *testing.T) {
	test.Parallel()
	paths, globError := filepath.Glob("testdata/*.a")
	if globError != nil {
		test.Fatal(globError)
	}
	for _, path := range paths {
		test.Run(filepath.Base(path), func(test *testing.T) {
			program := lowered(test, path)
			var output bytes.Buffer
			if printError := split.Print(&output, split.Analyze(program)); printError != nil {
				test.Fatal(printError)
			}
			expected, readError := os.ReadFile(path[:len(path)-2] + ".txt")
			if readError != nil {
				test.Fatal(readError)
			}
			if output.String() != string(expected) {
				test.Fatalf("decision table:\n%s\nwant:\n%s", output.String(), expected)
			}
			ordinaryOracle(test, path, program)
		})
	}
}
func ordinaryOracle(test *testing.T, path string, program *ir.Program) {
	test.Helper()
	absolute, absolutePathError := filepath.Abs(path)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	runner, absolutePathError := filepath.Abs("../../oracle/node.mjs")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	truthCommand := exec.Command("node", "--disable-warning=ExperimentalWarning", runner, absolute)
	truthCommand.Env = append(os.Environ(), jsonTypesEnvironment(test, absolute)...)
	truth, combinedOutputError := truthCommand.CombinedOutput()
	if combinedOutputError != nil {
		test.Fatalf("source Node: %v: %s", combinedOutputError, truth)
	}
	directory := test.TempDir()
	module, binary := filepath.Join(directory, "program.mjs"), filepath.Join(directory, "program")
	if writeError := os.WriteFile(module, []byte(javascript.JavaScript(program)), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	if buildError := native.Build(native.C(program), binary, native.Options{Sanitize: true}); buildError != nil {
		test.Fatal(buildError)
	}
	for _, command := range []*exec.Cmd{
		exec.Command("node", "--disable-warning=ExperimentalWarning", runner, module),
		exec.Command(binary),
	} {
		observed, combinedOutputError := command.CombinedOutput()
		if combinedOutputError != nil || !bytes.Equal(observed, truth) {
			test.Fatalf("%v: %v: observed %q, source Node %q", command.Args, combinedOutputError, observed, truth)
		}
	}
	test.Logf("Node, JavaScript and sanitized native agree: %q", truth)
}

func TestBoundaryParameterOrderAndPositions(test *testing.T) {
	test.Parallel()
	program := lowered(test, "testdata/decisions.a")
	for _, function := range program.Functions {
		if function.Name != "mixed" {
			continue
		}
		if len(function.Boundary.Parameters) != 3 {
			test.Fatal("boundary parameters missing")
		}
		want := []string{"number", "string", "array"}
		for i, parameter := range function.Boundary.Parameters {
			schema := parameter.Schema
			if schema == nil || schema.Nodes[schema.Root].Kind != want[i] {
				test.Fatalf("parameter %d schema: %#v", i+1, schema)
			}
		}
		array := function.Boundary.Parameters[2].Schema
		if array.Nodes[array.Nodes[array.Root].Children[0]].Kind != "number" {
			test.Fatal("array element schema")
		}
		if function.Boundary.Return == nil ||
			function.Boundary.Return.Nodes[function.Boundary.Return.Root].Kind != "number" {
			test.Fatal("return schema")
		}
		if filepath.Base(function.Position.File) != "decisions.a" ||
			function.Position.Line != 112 ||
			function.Position.Column != 10 {
			test.Fatalf("position: %+v", function.Position)
		}
		return
	}
	test.Fatal("mixed missing")
}

func TestBoundaryUnsupportedTypesHaveNoSchema(test *testing.T) {
	test.Parallel()
	program := lowered(test, "testdata/decisions.a")
	for _, function := range program.Functions {
		if function.Name == "takesMap" || function.Name == "takesClosure" {
			if !reflect.DeepEqual(function.Boundary.Parameters, []ir.BoundaryParameter{{}}) {
				test.Fatalf("%s: unsupported boundary %#v", function.Name, function.Boundary)
			}
		}
	}
}

func TestBoundaryMoreUnsupportedTypesHaveNoSchema(test *testing.T) {
	test.Parallel()
	for _, path := range []string{"testdata/effects.a", "testdata/shapes.a"} {
		program := lowered(test, path)
		for _, function := range program.Functions {
			switch function.Name {
			case "takesSet", "takesWeak", "classParameter":
				if len(function.Boundary.Parameters) != 1 || function.Boundary.Parameters[0].Schema != nil {
					test.Fatalf("%s schema was accepted", function.Name)
				}
			case "returnsClosure":
				if function.Boundary.Return != nil {
					test.Fatal("closure return schema was accepted")
				}
			}
		}
	}
}

func TestConstructorBoundary(test *testing.T) {
	test.Parallel()
	program := lowered(test, "testdata/shapes.a")
	for _, function := range program.Functions {
		if function.Name == "Box_new" {
			if function.Position.Line != 11 ||
				len(function.Boundary.Parameters) != 1 ||
				function.Boundary.Parameters[0].Schema == nil ||
				function.Boundary.Return != nil {
				test.Fatalf("constructor metadata: %+v", function)
			}
			return
		}
	}
	test.Fatal("constructor missing")
}

type futureOperation struct{}

func (futureOperation) Type() ir.Type { return ir.Number }
func TestUnknownOperationIsImpure(test *testing.T) {
	test.Parallel()
	number := &ir.JSONDecodeSchema{Root: 0, Nodes: []ir.JSONDecodeNode{{Kind: "number", Of: ir.Number}}}
	program := &ir.Program{Functions: []ir.Function{{
		Name:     "future",
		Returns:  ir.Number,
		Boundary: ir.FunctionBoundary{Return: number},
		Body: []ir.Statement{
			ir.Evaluate{Value: futureOperation{}},
			ir.Loop{Condition: ir.BooleanConstant{Value: false}},
		},
	}}}
	decisions := split.Analyze(program)
	if len(decisions) != 1 ||
		decisions[0].Eligible ||
		decisions[0].Reason != "impure: unknown operation futureOperation" {
		test.Fatalf("unknown operation: %+v", decisions)
	}
}

func TestBoundaryUnionFences(test *testing.T) {
	test.Parallel()
	program := lowered(test, "testdata/unions.a")
	seen := 0
	for _, decision := range split.Analyze(program) {
		if decision.Name == "objectUnion" || decision.Name == "broadUnion" {
			schema := program.Functions[decision.Index].Boundary.Parameters[0].Schema
			if schema == nil || schema.Nodes[schema.Root].Kind != "union" {
				test.Fatalf("boundary union missing: %s", decision.Name)
			}
			if decision.Eligible || decision.Reason != "signature: parameter 1 not crossable" {
				test.Fatalf("union crossed: %+v", decision)
			}
			seen++
		}
	}
	if seen != 2 {
		test.Fatalf("union fixtures missing: %d", seen)
	}
}
func TestBoundaryConstructorAlignment(test *testing.T) {
	test.Parallel()
	program := lowered(test, "testdata/unions.a")
	initializers := 0
	for _, function := range program.Functions {
		if function.Position.File == "" {
			continue
		}
		if len(function.Parameters) != len(function.Boundary.Parameters) {
			test.Fatalf(
				"misaligned %s: %d parameters, %d schemas",
				function.Name,
				len(function.Parameters),
				len(function.Boundary.Parameters),
			)
		}
		if function.Name == "Parent_new_initialize" || function.Name == "Child_new_initialize" {
			if function.Boundary.Parameters[0].Schema != nil || function.Boundary.Parameters[1].Schema == nil {
				test.Fatalf("receiver schema: %+v", function)
			}
			initializers++
		}
	}
	if initializers != 2 {
		test.Fatalf("live initializers: %d", initializers)
	}
}

func TestBoundaryStringTable(test *testing.T) {
	test.Parallel()
	program := lowered(test, "testdata/unions.a")
	if !reflect.DeepEqual(program.BoundaryStrings.Values, []string{"boundary-only"}) {
		test.Fatalf("metadata strings: %q", program.BoundaryStrings.Values)
	}
	if program.BoundaryStrings.Intern("boundary-only") != 0 || len(program.BoundaryStrings.Values) != 1 {
		test.Fatal("metadata table did not reuse literal")
	}
	for _, text := range program.Strings {
		if text == "boundary-only" {
			test.Fatal("boundary literal entered codegen table")
		}
	}
	if strings.Contains(native.C(program), "boundary-only") ||
		strings.Contains(javascript.JavaScript(program), "boundary-only") {
		test.Fatal("metadata literal emitted")
	}
}
