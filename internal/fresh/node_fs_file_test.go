package fresh_test

import (
	"testing"

	"github.com/system-inc/adamic/internal/fresh"
	"github.com/system-inc/adamic/internal/ir"
)

func TestNodeFSFileOperationsAreKnown(t *testing.T) {
	operations := map[string]ir.Type{
		"read_sync": ir.Number, "read_buffer": ir.Array, "read_buffer_fd": ir.Array, "read_file": ir.String, "read_fd": ir.String, "open": ir.Number,
		"write_buffer": 0, "write_buffer_fd": 0, "write": ir.Number, "close": 0, "write_file": 0, "write_fd": 0,
		"exists": ir.Boolean, "stat": ir.Object, "mkdir": ir.String,
		"unlink": 0, "utimes": 0, "utimes_dates": 0,
		"utimes_atime_date": 0, "utimes_mtime_date": 0,
		"is_file": ir.Boolean, "is_directory": ir.Boolean,
		"is_symbolic_link": ir.Boolean, "date_new": ir.Object,
		"date_time": ir.Number,
	}
	for operation, result := range operations {
		t.Run(operation, func(t *testing.T) {
			// A nested host expression also checks that arguments are walked.
			program := &ir.Program{Main: []ir.Statement{ir.Evaluate{Value: ir.NodeFSFile{
				Operation: operation, Of: result,
				Arguments: []ir.Expression{ir.NodeFSFile{Operation: "date_new", Of: ir.Object}},
			}}}}
			for _, write := range fresh.ProveWrites(program) {
				if write.Kind == fresh.WriteUnknown {
					t.Fatalf("unhandled host expression: %s", write.Why)
				}
			}
		})
	}
}

func TestNodeFSFileOperandsStillJudgeWrites(t *testing.T) {
	closes := ir.ArrayPush{Array: ir.Property{Object: ir.Read{Local: 0, Of: ir.Object}, Name: "nodes", Of: ir.Array},
		Value: ir.Read{Local: 0, Of: ir.Object}, Element: ir.Object, Site: 2}
	program := regexTreeProgram(ir.NodeFSFile{Operation: "write", Of: ir.Number, Arguments: []ir.Expression{closes}})
	program.Main = []ir.Statement{
		ir.Declare{Local: 1, Value: ir.ObjectLiteral{Fields: []ir.Field{{Name: "nodes", Value: ir.ArrayLiteral{Element: ir.Object}}}}},
		ir.Evaluate{Value: ir.Call{Function: 0, Arguments: []ir.Expression{ir.Read{Local: 1, Of: ir.Object}}}},
	}
	program.Locals = append(program.Locals, ir.Local{Type: ir.Object, Function: -1})
	for _, write := range fresh.ProveWrites(program) {
		if write.Site == 2 && !write.Proven && write.Kind == fresh.WriteElement {
			return
		}
	}
	t.Fatal("fs argument evaluation lost the cycle-closing write")
}

func TestNodeFSFileDoesNotEscapeBorrowedObjects(t *testing.T) {
	for _, operation := range []string{"is_file", "date_time", "utimes_dates"} {
		program := regexTreeProgram(ir.NodeFSFile{Operation: operation, Of: ir.Number, Arguments: []ir.Expression{ir.Read{Local: 0, Of: ir.Object}}})
		writes := fresh.ProveWrites(program)
		if len(writes) != 1 || !writes[0].Proven {
			t.Errorf("%s: borrowed argument poisoned a fresh tree write: %+v", operation, writes)
		}
	}
}

func TestNodeFSFileResultsAreFresh(t *testing.T) {
	for _, operation := range []string{"stat", "date_new"} {
		program := &ir.Program{
			Locals: []ir.Local{{Type: ir.Object, Function: 0}, {Type: ir.Object, Function: 0}},
			Functions: []ir.Function{{Name: "hostResult", Closure: true, Parameters: []int{0}, Body: []ir.Statement{
				ir.Declare{Local: 1, Value: ir.NodeFSFile{Operation: operation, Of: ir.Object}},
				ir.SetProperty{Object: ir.Read{Local: 1, Of: ir.Object}, Name: "saved", Value: ir.Read{Local: 0, Of: ir.Object}, Site: 1},
				ir.Return{},
			}}},
		}
		writes := fresh.ProveWrites(program)
		if len(writes) != 1 || !writes[0].Proven {
			t.Errorf("%s: result must be confined even when its stored value is outside: %+v", operation, writes)
		}
	}
}
