package oracle

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// Error.stackTraceLimit is Node's writable number on the global Error, which tsc's sys.ts raises to
// 100. node_error_stack_trace_limit.a reads, compares, writes and reads it back, a registered fixture
// held to Node on native, the JavaScript backend and wasm32 by the families; TestNodeProcessMutants
// holds it to a read that ignores the writes.
func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/node_error_stack_trace_limit.a", true, false})
}

// Every use of Error.stackTraceLimit but a plain read, or a plain = write of a number through the
// global Error identifier, stops with a named NotYet: an alias of Error, an optional chain, an element
// access, a compound assignment, an update and a write typed any. The Error constructor as a value
// stays NotYet, and a delete never reaches lowering, since the checker refuses deleting a member that
// isn't optional (TS2790). Each program but the delete runs on Node, and its whole first diagnostic is
// pinned in its .stopped file.
func TestNodeErrorStackTraceLimitRefusals(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/node_error_stack_trace_limit_refused"
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		t.Fatal(err)
	}
	programs := 0
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".a") {
			continue
		}
		programs++
		path := directory + "/" + name
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			absolute, err := filepath.Abs(filepath.Join(repository, path))
			if err != nil {
				t.Fatal(err)
			}
			expected, err := os.ReadFile(strings.TrimSuffix(absolute, ".a") + ".stopped")
			if err != nil {
				t.Fatal(err)
			}
			var stopped error
			if name == "delete.a" {
				// The oracle's Node run checks the source too, so this program stops there as well.
				_, stopped = load.Load([]string{absolute})
			} else {
				truth := onNode(t, absolute)
				if truth.exitCode != 0 || len(truth.stderr) != 0 || len(truth.stdout) == 0 {
					t.Fatalf("Node: %+v", truth)
				}
				program, loadError := load.Load([]string{absolute})
				if loadError != nil {
					t.Fatalf("Load: %v", loadError)
				}
				_, stopped = lower.Lower(context.Background(), program)
				var notYet *lower.NotYet
				if !errors.As(stopped, &notYet) {
					t.Fatalf("want a NotYet, got %v", stopped)
				}
			}
			if stopped == nil {
				t.Fatal("want the program stopped, but it went through")
			}
			got := strings.ReplaceAll(stopped.Error(), strings.TrimSuffix(absolute, path), "")
			if got != strings.TrimSpace(string(expected)) {
				t.Fatalf("diagnostic: got %q, want %q", got, strings.TrimSpace(string(expected)))
			}
		})
	}
	if programs != 8 {
		t.Fatalf("%d programs in %s, want 8: a program added or removed needs this count changed with it", programs, directory)
	}
}
