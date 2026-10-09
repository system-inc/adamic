package oracle

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// The node library is staged separately from the compiler refusal on main.
// Always check source Node; enable backend and count checks when that dependency lands.
func TestOptionalWideningNodeParentheses(t *testing.T) {
	t.Parallel()
	rows := []string{}
	missingLibrary := false
	for _, name := range []string{"argument", "nested"} {
		path := "internal/oracle/testdata/optional_widening_node_parens_" + name + ".a"
		t.Run(name, func(t *testing.T) {
			absolute, err := filepath.Abs(filepath.Join(repository, path))
			if err != nil {
				t.Fatal(err)
			}
			source := execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle", "node.mjs"), absolute)
			if source.exitCode != 0 || string(source.stdout) != "directory\nabsent=true\n" || len(source.stderr) != 0 {
				t.Fatalf("source Node: %+v", source)
			}
			program, err := load.Load([]string{absolute})
			if err != nil && (strings.Contains(err.Error(), "Cannot find module 'node:fs'") || strings.Contains(err.Error(), "Cannot find name 'node:fs'")) {
				missingLibrary = true
				t.Skip("backend requires the node library staged in 23ea5d92; source Node passed")
			}
			if err != nil {
				t.Fatal(err)
			}
			lowered, err := lower.Lower(context.Background(), program)
			if err != nil {
				t.Fatal(err)
			}
			backend := onJavaScriptBackend(t, lowered)
			compiled, sanitized := natively(t, lowered)
			for name, result := range map[string]run{"JavaScript": backend, "sanitized native": compiled, "release native": released(t, lowered)} {
				if difference := disagreement(source, result); difference != "" {
					t.Errorf("%s: %s", name, difference)
				}
			}
			if leaked := leaks(t, lowered, sanitized); leaked != "" {
				t.Errorf("leaks: %s", leaked)
			}
			rows = append(rows, counted(t, path, false, nil, false, false))
		})
	}
	if t.Failed() {
		return
	}
	if missingLibrary {
		t.Skip("node library is staged separately; both source Node variants passed")
	}
	// These dependency-specific rows live separately so main's table does not claim
	// that its placeholder node hook can compile them. Measure with -update-counts.
	measured := strings.Join(rows, "\n") + "\n"
	const counts = "optional_widening_parens_counts.txt"
	if *updateCounts {
		if err := os.WriteFile(counts, []byte(measured), 0644); err != nil {
			t.Fatal(err)
		}
		return
	}
	recorded, err := os.ReadFile(counts)
	if err != nil {
		t.Fatal(err)
	}
	if string(recorded) != measured {
		t.Fatalf("parenthesized node counts: recorded %q, measured %q", recorded, measured)
	}
}
