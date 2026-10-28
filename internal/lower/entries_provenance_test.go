package lower

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

func TestEntriesAllocationProof(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		source string
		proven bool
	}{
		{`const source={value:1, hidden:2}; const view:{readonly value:number}=source; Object.entries(view);`, true},
		{`let source={value:1, hidden:2}; const view:{readonly value:number}=source; Object.entries(view);`, true},
		{`let source:{readonly value:number}={value:1}; const other={value:2, hidden:'bad'}; source=other; Object.entries(source);`, false},
		{`let source:{readonly value:number}={value:1}; const other={value:2, hidden:'bad'}; [source]=[other]; Object.entries(source);`, false},
		{`function entries(source:{readonly value:number}):void {Object.entries(source);} entries({value:1});`, false},
		{`const source={value:1, hidden:'bad'}; const view:{readonly value:number}=source; Object.entries(view);`, false},
	} {
		t.Run(probe.source, func(t *testing.T) {
			t.Parallel()
			program, err := lowerSource(t, probe.source)
			if err != nil {
				t.Fatal(err)
			}
			found := 0
			inspect := func(value any) bool {
				if call, ok := value.(ir.ObjectCall); ok && call.Method == "entries" {
					found++
					if call.Checked == probe.proven {
						t.Errorf("proven=%t, checked=%t", probe.proven, call.Checked)
					}
				}
				return true
			}
			walk(program.Main, inspect)
			for _, function := range program.Functions {
				walk(function.Body, inspect)
			}
			if found != 1 {
				t.Fatalf("found %d entries calls", found)
			}
		})
	}
}

// Dictionary parameters and fields now lower through records; fixed-object spreads remain a boundary.
func TestEntriesRecordBoundaries(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		source, reason string
		admitted       bool
	}{
		{`const base={visible:1}; const source:{[key:string]:number;visible:number}={...base}; source['late']=2; Object.entries(source);`, "spread into record storage", false},
		{`function write(value:{[key:string]:number}):void {value['late']=2;} const source={visible:1}; Object.entries(source);`, "dictionary parameter write", true},
		{`const source:{[key:string]:number;visible:number}={visible:1}; source['late']=2; console.log(source.late.toString()); Object.entries(source);`, "dictionary field read", true}} {
		t.Run(probe.reason, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			if probe.admitted {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			var stopped *NotYet
			if !errors.As(err, &stopped) || !strings.Contains(stopped.What, probe.reason) {
				t.Fatalf("want explicit NotYet %q, got %v", probe.reason, err)
			}
		})
	}
}

// A lowering keeps entriesRecordLiteral's answer per literal and entriesProgramHasRecords' per program, and every ask
// answers as a fresh lowering does: two programs lowered in one process, one with a record and one without, each asked
// twice (#pv935yw). Mutants: the program's answer kept across lowerings (the second program reads the first's true), a
// kept answer read as false, or one answer kept for every literal, each go red here.
func TestEntriesRecordMemoAnswersAsAFreshLowering(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name, source string
		program      bool
		records      []bool
	}{
		{"records", `const source:{[key:string]:number;visible:number}={visible:1}; source['late']=2; const plain={value:1}; console.log(plain.value.toString()); console.log(source.late.toString());`, true, []bool{true, false}},
		{"no records", `const plain={value:1}; const other={value:2}; console.log((plain.value + other.value).toString());`, false, []bool{false, false}},
	} {
		path := filepath.Join(t.TempDir(), "main.a")
		if err := os.WriteFile(path, []byte(probe.source), 0644); err != nil {
			t.Fatal(err)
		}
		program, err := load.Load([]string{path})
		if err != nil {
			t.Fatal(err)
		}
		file := program.Files()[0]
		checked, release := program.Checker(context.Background(), file)
		var literals []*ast.Node
		var visit ast.Visitor
		visit = func(n *ast.Node) bool {
			if n.Kind == ast.KindObjectLiteralExpression {
				literals = append(literals, n)
			}
			n.ForEachChild(visit)
			return false
		}
		file.AsNode().ForEachChild(visit)
		if len(literals) != len(probe.records) {
			t.Fatalf("%s: %d object literals, want %d", probe.name, len(literals), len(probe.records))
		}
		fresh := func() *lowering { return &lowering{program: program, checker: checked, result: &ir.Program{}} }
		for index, literal := range literals {
			if got := fresh().entriesRecordLiteral(literal); got != probe.records[index] {
				t.Fatalf("%s: literal %d alone is a record=%t, want %t", probe.name, index, got, probe.records[index])
			}
		}
		if got := fresh().entriesProgramHasRecords(); got != probe.program {
			t.Fatalf("%s: the program alone has records=%t, want %t", probe.name, got, probe.program)
		}
		kept := fresh()
		for range 2 {
			if got := kept.entriesProgramHasRecords(); got != probe.program {
				t.Fatalf("%s: the program from the memo has records=%t, want %t", probe.name, got, probe.program)
			}
			for index, literal := range literals {
				if got := kept.entriesRecordLiteral(literal); got != probe.records[index] {
					t.Fatalf("%s: literal %d from the memo is a record=%t, want %t", probe.name, index, got, probe.records[index])
				}
			}
		}
		release()
	}
}
