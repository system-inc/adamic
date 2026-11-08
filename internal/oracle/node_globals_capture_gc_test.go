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
// global.gc but a read and a call as a statement through the global identifier (gc by its bare name,
// through globalThis, an optional call, a write, the call's result used as a value), and global anywhere
// but as global.gc's receiver. Each
// program runs on Node, and its whole first diagnostic is pinned in its .stopped file.
func TestNodeGlobalsCaptureGcRefusals(test *testing.T) {
	test.Parallel()
	directory := "internal/oracle/testdata/node_globals_capture_gc_refused"
	entries, readDirectoryError := os.ReadDir(filepath.Join(repository, directory))
	if readDirectoryError != nil {
		test.Fatal(readDirectoryError)
	}
	programs := 0
	for _, entry := range entries {
		name := entry.Name()
		// A .ts program is one the checker lets through only past a non-null assertion, which .a refuses.
		extension := filepath.Ext(name)
		if extension != ".a" && extension != ".ts" {
			continue
		}
		programs++
		path := directory + "/" + name
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			absolute, absolutePathError := filepath.Abs(filepath.Join(repository, path))
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			expected, readError := os.ReadFile(strings.TrimSuffix(absolute, extension) + ".stopped")
			if readError != nil {
				test.Fatal(readError)
			}
			truth := onNode(test, absolute)
			if truth.exitCode != 0 || len(truth.stderr) != 0 || len(truth.stdout) == 0 {
				test.Fatalf("Node: %+v", truth)
			}
			program, loadError := load.Load([]string{absolute})
			if loadError != nil {
				test.Fatalf("Load: %v", loadError)
			}
			_, stopped := lower.Lower(context.Background(), program)
			var notYet *lower.NotYet
			if !errors.As(stopped, &notYet) {
				test.Fatalf("want a NotYet, got %v", stopped)
			}
			got := strings.ReplaceAll(stopped.Error(), strings.TrimSuffix(absolute, path), "")
			if got != strings.TrimSpace(string(expected)) {
				test.Fatalf("diagnostic: got %q, want %q", got, strings.TrimSpace(string(expected)))
			}
		})
	}
	if programs != 12 {
		test.Fatalf(
			"%d programs in %s, want 12: a program added or removed needs this count changed with it",
			programs,
			directory,
		)
	}
}
