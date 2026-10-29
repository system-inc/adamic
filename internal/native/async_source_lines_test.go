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

func TestDebugAsyncTerminals(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "fixture.a")
	if err := os.WriteFile(path, []byte(debugLineFixture), 0644); err != nil {
		t.Fatal(err)
	}
	program, err := debugProgram(t, path, true)
	if err != nil {
		t.Fatal(err)
	}
	index := -1
	for i, function := range program.Functions {
		if function.Name == "later" {
			index = i
			break
		}
	}
	if index < 0 {
		t.Fatal("missing later function")
	}
	before := flow.Build(program, index)
	locations := asyncSourceLines(program.Functions[index].Body)
	after := flow.Build(program, index)
	if !reflect.DeepEqual(before, after) {
		t.Fatal("source mapping changed the flow graph")
	}
	awaits, throws := 0, 0
	for _, block := range after.Blocks {
		for _, id := range block.Instructions {
			if _, marker := (*after.Instructions[id].At).(ir.SourceLine); marker {
				t.Fatal("source marker became an instruction")
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
					t.Fatal("suspend lost its final await")
				}
			case ir.Evaluate:
				if _, ok := statement.Value.(ir.Await); !ok {
					t.Fatal("suspend lost its final await")
				}
			default:
				t.Fatalf("suspend ends with %T", statement)
			}
			if source != (ir.SourceLine{File: "fixture.a", Line: 8}) {
				t.Fatalf("await marker=%+v", source)
			}
			awaits++
		}
		if _, throwing := (*last.At).(ir.Throw); throwing {
			if _, ok := block.Terminal.(*flow.Goto); !ok {
				t.Fatal("throw lost its handler edge")
			}
			if source != (ir.SourceLine{File: "fixture.a", Line: 12}) {
				t.Fatalf("throw marker=%+v", source)
			}
			throws++
		}
	}
	if awaits == 0 || throws == 0 {
		t.Fatalf("terminal witnesses: awaits=%d throws=%d", awaits, throws)
	}

	directory := debugLineProduct(t, false)
	release := debugRunBinary(t, filepath.Join(directory, "release"))
	debug := debugRunBinary(t, filepath.Join(directory, "program"))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	command := toolchain.CommandContextIn(ctx, "../..", "node", "--disable-warning=ExperimentalWarning", "oracle/node.mjs", path)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("source Node: %v: %s", err, output)
	}
	expected := mainPartsRun{Stdout: "3\n2\n1\n0\n6\n5\ncaught\n9\n", Exit: 0}
	if release != expected || debug != expected || string(output) != expected.Stdout {
		t.Fatalf("release=%+v debug=%+v Node=%q, want %+v", release, debug, output, expected)
	}
	t.Logf("%d await and %d throw terminals unchanged; debug/release/Node stdout, stderr and exit agree", awaits, throws)
}

func TestDebugAsyncSourceInheritance(t *testing.T) {
	t.Parallel()
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
			t.Fatalf("inherited marker at %p = %+v", slot, locations[slot])
		}
	}
	if locations[&nested[1]] != (ir.SourceLine{File: "nested.a", Line: 20}) {
		t.Fatal("nested marker did not override its inherited marker")
	}
}
