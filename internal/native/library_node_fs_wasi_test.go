package native

import (
	"errors"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestWASIFSArgumentProofs(t *testing.T) {
	// Not parallel: the sysroot and compiler lookup environment are process-wide.
	t.Setenv("WASI_SYSROOT", t.TempDir())
	t.Setenv("PATH", t.TempDir())
	number := func(value float64) ir.Expression { return ir.NumberConstant{Value: value} }
	dynamic := ir.Read{Local: 0, Of: ir.Number}
	path := ir.StringConstant{Index: 0}
	flag := func(index int) ir.Expression { return ir.StringConstant{Index: index} }
	mode := func(operation string, value ir.Expression, flagIndex int) ir.NodeFSFile {
		args := []ir.Expression{path, flag(flagIndex), value}
		if operation == "write_file" || operation == "write_buffer" || operation == "write_fd" || operation == "write_buffer_fd" {
			args = []ir.Expression{path, path, flag(flagIndex), value, ir.BooleanConstant{Value: false}}
		}
		if operation == "mkdir" {
			args[1] = ir.BooleanConstant{Value: false}
		}
		return ir.NodeFSFile{Operation: operation, Arguments: args}
	}
	times := func(operation string, a, b ir.Expression) ir.NodeFSFile {
		return ir.NodeFSFile{Operation: operation, Arguments: []ir.Expression{path, a, b}}
	}
	for _, probe := range []struct {
		name   string
		call   ir.NodeFSFile
		member string
	}{
		{"open custom creation", mode("open", number(0600), 1), "fs.openSync"},
		{"open dynamic creation", mode("open", dynamic, 1), "fs.openSync"},
		{"open default creation", mode("open", number(0666), 1), ""},
		{"open custom read", mode("open", number(0600), 2), ""},
		{"open dynamic mode read", mode("open", dynamic, 2), ""},
		{"open custom rs+", mode("open", number(0600), 3), ""},
		{"string write custom", mode("write_file", number(0600), 1), "fs.writeFileSync"},
		{"buffer write dynamic", mode("write_buffer", dynamic, 1), "fs.writeFileSync"},
		{"string write default", mode("write_file", number(0666), 1), ""},
		{"buffer write default", mode("write_buffer", number(0666), 1), ""},
		{"string fd ignored mode", mode("write_fd", dynamic, 1), ""},
		{"buffer fd ignored mode", mode("write_buffer_fd", dynamic, 1), ""},
		{"mkdir default", mode("mkdir", number(0777), 1), ""},
		{"mkdir custom", mode("mkdir", number(0700), 1), "fs.mkdirSync"},
		{"mkdir dynamic", mode("mkdir", dynamic, 1), "fs.mkdirSync"},
		{"utimes whole", times("utimes", number(0), number(1700000000)), ""},
		{"utimes maximum", times("utimes", number(18446744073), number(0)), ""},
		{"utimes range", times("utimes", number(18446744074), number(0)), "fs.utimesSync"},
		{"utimes fractional atime", times("utimes", number(0.5), number(0)), "fs.utimesSync"},
		{"utimes fractional mtime", times("utimes", number(0), number(0.5)), "fs.utimesSync"},
		{"utimes negative", times("utimes", number(-1), number(0)), "fs.utimesSync"},
		{"utimes NaN", times("utimes", number(math.NaN()), number(0)), "fs.utimesSync"},
		{"utimes Infinity", times("utimes", number(math.Inf(1)), number(0)), "fs.utimesSync"},
		{"utimes dynamic", times("utimes", dynamic, number(0)), "fs.utimesSync"},
		{"utimes dates", times("utimes_dates", dynamic, dynamic), "fs.utimesSync"},
		{"utimes atime date", times("utimes_atime_date", dynamic, number(0)), "fs.utimesSync"},
		{"utimes mtime date", times("utimes_mtime_date", number(0), dynamic), "fs.utimesSync"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			emitter := &emitter{program: &ir.Program{Strings: []string{"path", "w", "r", "rs+"}}}
			emitter.nodeFSFileWASIMarker(probe.call)
			emitter.nodeFSFileWASIMarker(probe.call)
			if probe.member == "" {
				if len(emitter.declarations) != 0 {
					t.Fatalf("supported call marked: %v", emitter.declarations)
				}
				return
			}
			if len(emitter.declarations) != 1 {
				t.Fatalf("duplicate or absent marker: %v", emitter.declarations)
			}
			binary := filepath.Join(t.TempDir(), "refused.wasm")
			err := Build(emitter.declarations[0]+"\nint main(void){return 0;}\n", binary, Options{Target: "wasm32-wasi"})
			var refused *TargetRefused
			if !errors.As(err, &refused) || refused.Member != probe.member {
				t.Fatalf("want %s before compiler lookup, got %v", probe.member, err)
			}
			if _, err := os.Stat(binary); !os.IsNotExist(err) {
				t.Fatalf("refusal created artifact: %v", err)
			}
		})
	}
}
