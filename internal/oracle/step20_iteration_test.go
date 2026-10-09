package oracle

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

type step20Outcome struct {
	File   string `json:"file"`
	Origin string `json:"origin"`
	Kind   string `json:"kind"`
	Reason string `json:"reason"`
	Stdout string `json:"stdout"`
}

func init() {
	for _, file := range []string{"collections.a", "strings.a", "object_iteration.a", "array_view_stress.a", "test262_array_views.a", "array_view_weak.a", "user_forwarding.a", "generator.a", "delegated_generator.a"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path: "stage3/fixtures/iteration/" + file, lowers: true})
	}
}

// Recording observes current outcomes; ordinary runs require the committed exact snapshot.
func TestStep20IterationOutcomes(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile(filepath.Join(repository, "stage3/fixtures/iteration/outcomes.json"))
	if err != nil {
		t.Fatal(err)
	}
	var probes []step20Outcome
	if err := json.Unmarshal(data, &probes); err != nil {
		t.Fatal(err)
	}
	record := os.Getenv("ADAMIC_STEP20_RECORD")
	observed := make([]step20Outcome, len(probes))
	for index, probe := range probes {
		t.Run(probe.File, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "stage3/fixtures/iteration", probe.File))
			if err != nil {
				t.Fatal(err)
			}
			node := onNode(t, path)
			if node.exitCode != 0 || len(node.stderr) != 0 || string(node.stdout) != probe.Stdout {
				t.Fatalf("source on Node: exit %d stdout %q stderr %q; want stdout %q", node.exitCode, node.stdout, node.stderr, probe.Stdout)
			}
			program, failure := lowered(t, path)
			actual := probe
			actual.Kind, actual.Reason = "accepted", ""
			var notYet *lower.NotYet
			var refused *lower.Refused
			if errors.As(failure, &notYet) {
				actual.Kind, actual.Reason = "NotYet", notYet.What
			} else if errors.As(failure, &refused) {
				actual.Kind, actual.Reason = "Refused", refused.What
			} else if failure != nil {
				t.Fatalf("unexpected checker/lowering error: %v", failure)
			}
			if actual.Kind == "accepted" {
				native, binary := natively(t, program)
				for backend, got := range map[string]run{"native": native, "JavaScript": onJavaScriptBackend(t, program), "release": released(t, program)} {
					if difference := disagreement(node, got); difference != "" {
						t.Errorf("%s: %s", backend, difference)
					}
				}
				if report := leaks(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
			t.Logf("%s: %s %s", probe.File, actual.Kind, actual.Reason)
			observed[index] = actual
			if record == "" && (actual.Kind != probe.Kind || actual.Reason != probe.Reason) {
				t.Errorf("outcome changed: want %s %q, got %s %q", probe.Kind, probe.Reason, actual.Kind, actual.Reason)
			}
		})
	}
	if record != "" && !t.Failed() {
		data, err := json.MarshalIndent(observed, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(record, append(data, '\n'), 0644); err != nil {
			t.Fatal(err)
		}
	}
}
