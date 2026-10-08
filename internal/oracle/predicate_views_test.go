package oracle

import (
	"errors"
	"github.com/system-inc/adamic/internal/lower"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPredicateStructuralViews(t *testing.T) {
	for _, name := range []string{"valid", "invalid", "direct_valid", "direct_invalid", "statement_valid", "statement_invalid", "intersection_valid", "intersection_invalid", "union_valid", "union_invalid"} {
		t.Run(name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/predicate_hatches/view_"+name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			provenSource := filepath.Join(t.TempDir(), "view.a")
			if err := os.WriteFile(provenSource, []byte(strings.ReplaceAll(string(source), "node!", "node")), 0600); err != nil {
				t.Fatal(err)
			}
			_, err = lowered(t, provenSource)
			var refusal *lower.Refused
			if !errors.As(err, &refusal) || !strings.Contains(refusal.What, "predicate") {
				t.Fatalf(".a must refuse unproven predicate: %v", err)
			}
			ts := filepath.Join(t.TempDir(), "view.ts")
			if err := os.WriteFile(ts, source, 0600); err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, ts)
			if err != nil {
				t.Fatal(err)
			}
			root, err := filepath.Abs(repository)
			if err != nil {
				t.Fatal(err)
			}
			relative, err := filepath.Rel(root, ts)
			if err != nil {
				t.Fatal(err)
			}
			node := onNode(t, path)
			if node.exitCode != 0 {
				t.Fatalf("Node: %d %s", node.exitCode, node.stderr)
			}
			native, _ := natively(t, program)
			for backend, got := range map[string]run{"javascript": onJavaScriptBackend(t, program), "sanitized": native, "release": released(t, program)} {
				if strings.HasSuffix(name, "valid") && !strings.HasSuffix(name, "invalid") {
					if d := disagreement(node, got); d != "" {
						t.Errorf("%s: %s", backend, d)
					}
				} else {
					if got.exitCode != 70 || string(got.stdout) != "1\n" || !strings.Contains(string(got.stderr), "label") {
						t.Errorf("%s must stop at label after kind: exit=%d stdout=%q stderr=%q", backend, got.exitCode, got.stdout, got.stderr)
					}
				}
			}
			row := counted(t, relative, false, nil, false, false)
			row = strings.Replace(row, relative, "internal/oracle/testdata/predicate_hatches/view_"+name+".a (.ts mode)", 1)
			t.Log(row)
			counts, err := os.ReadFile(countsPath)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(counts), row) {
				t.Errorf("counts.md missing measured row: %s", row)
			}

		})
	}
}

func TestPredicateWritableViewStaysRefused(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/predicate_hatches/view_writable_refused.a"))
	if err != nil {
		t.Fatal(err)
	}
	node := onNode(t, path)
	if node.exitCode != 0 || string(node.stdout) != "undefined\n" {
		t.Fatalf("Node: %d %q %q", node.exitCode, node.stdout, node.stderr)
	}
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, extension := range []string{".a", ".ts"} {
		candidate := filepath.Join(t.TempDir(), "view"+extension)
		if err := os.WriteFile(candidate, source, 0600); err != nil {
			t.Fatal(err)
		}
		_, err := lowered(t, candidate)
		var refusal *lower.Refused
		if !errors.As(err, &refusal) || !strings.Contains(refusal.What, "predicate") {
			t.Fatalf("%s must preserve writable-view refusal: %v", extension, err)
		}
	}
}
