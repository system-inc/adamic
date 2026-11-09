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
// naming fd in its last cast. Both stop at the read, at this rule. n3 (an annotation naming fd over an
// exact cast) and n4 (the handle left as any) are stopped before the rule, by the hiding rule
// (adamic/no-optional-widening) and by lowering's any stop, and their .stop files pin those stops, so
// the rule's annotation and any arms are unreached today and stand as fail-closed defense. Past the
// chain, a parameter or variable that adds an optional member is the hiding rule's too (Codex 01a1bf0d).
// node_process_blocking.a and node_process_terminal.a (TestNodeProcess*) are the controls that still
// lower and agree. Mutant (internal/lower/stdout-handle-view.mutants.md): stdoutHandleView answers "",
// and n1 and n2 lower and print undefined natively.
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
