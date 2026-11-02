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

// tsc's Debug.fail tests Error.captureStackTrace and calls it on the error it throws, and its sys.ts
// tests global.gc before calling it. node_globals_capture_gc.a makes both and reads each a second way,
// and node_globals_gc_call.ts calls global.gc where Node has none. Both are registered fixtures held
// to Node on native, the JavaScript backend and wasm32 by the families: Error.captureStackTrace is one
// function whose call evaluates its arguments, global.gc reads as undefined, and its call throws
// Node's TypeError, which a catch sees.
func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/node_globals_capture_gc.a", true, false}, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/node_globals_gc_call.ts", true, false})
}

// Every use of Error.captureStackTrace but a read and a call as a statement through the global Error
// identifier stops with a named NotYet: an alias of Error, an optional call, its result used as a
// value, a write, and its value in a slot that sees its parameters otherwise. So does every use of
// global.gc but a read and a call through the global identifier (gc by its bare name, through
// globalThis, an optional call, a write), and global anywhere but as global.gc's receiver. Each
// program runs on Node, and its whole first diagnostic is pinned in its .stopped file.
func TestNodeGlobalsCaptureGCRefusals(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/node_globals_capture_gc_refused"
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
			truth := onNode(t, absolute)
			if truth.exitCode != 0 || len(truth.stderr) != 0 || len(truth.stdout) == 0 {
				t.Fatalf("Node: %+v", truth)
			}
			program, err := load.Load([]string{absolute})
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			_, stopped := lower.Lower(context.Background(), program)
			var notYet *lower.NotYet
			if !errors.As(stopped, &notYet) {
				t.Fatalf("want a NotYet, got %v", stopped)
			}
			got := strings.ReplaceAll(stopped.Error(), strings.TrimSuffix(absolute, path), "")
			if got != strings.TrimSpace(string(expected)) {
				t.Fatalf("diagnostic: got %q, want %q", got, strings.TrimSpace(string(expected)))
			}
		})
	}
	if programs != 11 {
		t.Fatalf("%d programs in %s, want 11: a program added or removed needs this count changed with it", programs, directory)
	}
}
