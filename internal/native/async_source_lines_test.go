package native

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/flow"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/toolchain"
)

func TestDebugAsyncTerminals(test *testing.T) {
	test.Parallel()
	path := filepath.Join(test.TempDir(), "fixture.a")
	if err := os.WriteFile(path, []byte(debugLineFixture), 0644); err != nil {
		test.Fatal(err)
	}
	program, err := debugProgram(test, path, true)
	if err != nil {
		test.Fatal(err)
	}
	index := -1
	for i, function := range program.Functions {
		if function.Name == "later" {
			index = i
			break
		}
	}
	if index < 0 {
		test.Fatal("missing later function")
	}
	before := flow.Build(program, index)
	locations := asyncSourceLines(program.Functions[index].Body)
	after := flow.Build(program, index)
	if !reflect.DeepEqual(before, after) {
		test.Fatal("source mapping changed the flow graph")
	}
	awaits, throws := 0, 0
	for _, block := range after.Blocks {
		for _, id := range block.Instructions {
			if _, marker := (*after.Instructions[id].At).(ir.SourceLine); marker {
				test.Fatal("source marker became an instruction")
			}
		}
		if len(block.Instructions) == 0 {
			continue
		}
		last := after.Instructions[block.Instructions[len(block.Instructions)-1]]
		source := locations[last.At]
		if _, suspends := block.Terminal.(*flow.Suspend); suspends {
			switch statement := (*last.At).(type) {
			case ir.Declare:
				if _, ok := statement.Value.(ir.Await); !ok {
					test.Fatal("suspend lost its final await")
				}
			case ir.Evaluate:
				if _, ok := statement.Value.(ir.Await); !ok {
					test.Fatal("suspend lost its final await")
				}
			default:
				test.Fatalf("suspend ends with %T", statement)
			}
			if source != (ir.SourceLine{File: "fixture.a", Line: 8}) {
				test.Fatalf("await marker=%+v", source)
			}
			awaits++
		}
		if _, throwing := (*last.At).(ir.Throw); throwing {
			if _, ok := block.Terminal.(*flow.Goto); !ok {
				test.Fatal("throw lost its handler edge")
			}
			if source != (ir.SourceLine{File: "fixture.a", Line: 12}) {
				test.Fatalf("throw marker=%+v", source)
			}
			throws++
		}
	}
	if awaits == 0 || throws == 0 {
		test.Fatalf("terminal witnesses: awaits=%d throws=%d", awaits, throws)
	}

	directory := debugLineProduct(test, false)
	release := debugRunBinary(test, filepath.Join(directory, "release"))
	debug := debugRunBinary(test, filepath.Join(directory, "program"))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	command := toolchain.CommandContextIn(
		ctx,
		"../..",
		"node",
		"--disable-warning=ExperimentalWarning",
		"oracle/node.mjs",
		path,
	)
	output, err := command.CombinedOutput()
	if err != nil {
		test.Fatalf("source Node: %v: %s", err, output)
	}
	expected := mainPartsRun{Stdout: "3\n2\n1\n0\n6\n5\ncaught\n9\n", Exit: 0}
	if release != expected || debug != expected || string(output) != expected.Stdout {
		test.Fatalf("release=%+v debug=%+v Node=%q, want %+v", release, debug, output, expected)
	}
	test.Logf(
		"%d await and %d throw terminals unchanged; debug/release/Node stdout, stderr and exit agree",
		awaits,
		throws,
	)
}

func TestDebugAsyncSourceInheritance(test *testing.T) {
	test.Parallel()
	nested := []ir.Statement{ir.SourceLine{File: "nested.a", Line: 20}, ir.Debugger{}}
	inherited := []ir.Statement{ir.Debugger{}}
	aside := []ir.Statement{ir.Debugger{}}
	body := []ir.Statement{
		ir.SourceLine{File: "outer.a", Line: 3},
		ir.If{Condition: ir.BooleanConstant{Value: true}, Then: nested, Else: inherited},
		ir.Evaluate{Value: ir.Effects{Body: aside, Result: ir.NumberConstant{Value: 1}}},
		ir.Debugger{},
	}
	locations := asyncSourceLines(body)
	for _, slot := range []*ir.Statement{&body[1], &inherited[0], &body[2], &aside[0], &body[3]} {
		if locations[slot] != (ir.SourceLine{File: "outer.a", Line: 3}) {
			test.Fatalf("inherited marker at %p = %+v", slot, locations[slot])
		}
	}
	if locations[&nested[1]] != (ir.SourceLine{File: "nested.a", Line: 20}) {
		test.Fatal("nested marker did not override its inherited marker")
	}
}

func TestDebugAsyncUnexportedFields(test *testing.T) {
	test.Parallel()
	hidden := []ir.Statement{ir.SourceLine{File: "hidden.a", Line: 99}, ir.Debugger{}}
	visible := []ir.Statement{ir.SourceLine{File: "visible.a", Line: 3}, ir.Debugger{}}
	value := reflect.ValueOf(struct {
		hidden  []ir.Statement
		Visible []ir.Statement
	}{hidden: hidden, Visible: visible})
	test.Run("struct", func(test *testing.T) {
		locations := asyncSourceLineValues(value)
		if _, found := locations[&hidden[1]]; found {
			test.Fatal("walked an unexported field")
		}
		if locations[&visible[1]] != (ir.SourceLine{File: "visible.a", Line: 3}) {
			test.Fatal("lost the exported field's marker")
		}
	})
	test.Run("field", func(test *testing.T) {
		if locations := asyncSourceLineValues(value.Field(0)); len(locations) != 0 {
			test.Fatal("walked a non-interfaceable value")
		}
	})
}
