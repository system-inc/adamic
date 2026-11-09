package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// process.stdout._handle is the runtime's {setBlocking} object, and @types declares no _handle, so the
// type a program reads it as is its own cast (#0asf98m). Every type it's seen as on its way out of the
// read (the read's own type, each enclosing as, satisfies or !, an annotated variable it initializes)
// may name only setBlocking. n1 is the census probe: a cast naming readonly fd, whose read printed
// undefined natively where Node prints the descriptor. n2 is node_process_terminal.a's as-any chain
// naming fd in its last cast. n3 names fd in a variable's annotation over an exact cast. n4 lets the
// handle leave the chain as any. Each stops at the read in its .stop file's words; node_process_blocking.a
// and node_process_terminal.a (TestNodeProcess*) are the controls that still lower and agree. Mutant
// (internal/lower/stdout-handle-view.mutants.md): stdoutHandleView answers "", and n1 lowers and prints
// undefined natively.
func TestStdoutHandleViews(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/stdout_handle_view"
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		t.Fatal(err)
	}
	programs := 0
	for _, entry := range entries {
		if extension := filepath.Ext(entry.Name()); extension != ".a" && extension != ".ts" {
			continue
		}
		programs++
		t.Run(entry.Name(), func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(directory, entry.Name())
			absolute, err := filepath.Abs(filepath.Join(repository, path))
			if err != nil {
				t.Fatal(err)
			}
			expected, err := os.ReadFile(absolute + ".stop")
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, absolute)
			if truth.exitCode != 0 || len(truth.stderr) != 0 {
				t.Fatalf("source Node: exit %d stderr %q", truth.exitCode, truth.stderr)
			}
			program, err := lowered(t, absolute)
			if stop := strings.TrimSpace(string(expected)); stop != "lowers" {
				if err == nil {
					t.Fatalf("lowered, want %q", stop)
				}
				if got := strings.ReplaceAll(err.Error(), strings.TrimSuffix(absolute, path), ""); got != stop {
					t.Fatalf("stop: got %q, want %q", got, stop)
				}
				return
			}
			if err != nil {
				t.Fatalf("Lower: %v", err)
			}
			sanitized, _ := natively(t, program)
			for _, backend := range []struct {
				name string
				got  run
			}{{"release", released(t, program)}, {"sanitized", sanitized}, {"JavaScript", onJavaScriptBackend(t, program)}} {
				if difference := disagreement(truth, backend.got); difference != "" {
					t.Fatalf("%s: %s: Node %q, %s %q", backend.name, difference, truth.stdout, backend.name, backend.got.stdout)
				}
			}
		})
	}
	// A program lost from the directory must fail here, not shrink the walk.
	if programs != 8 {
		t.Fatalf("found %d programs, want 8", programs)
	}
}
