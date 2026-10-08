package oracle

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

const step18Directory = "docs/step-18/fixtures/"

var step18Supported = map[string]bool{"function.a": true, "method.a": true, "runtime-values.a": true, "runtime-order.a": true, "runtime-arguments.a": true, "runtime-return-descriptor.a": true}

func init() {
	for _, name := range []string{"function.a", "method.a", "element.a", "call-result.a", "cross-call.a", "receiver-call.a", "two-guards.a", "size.a", "number.a"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{step18Directory + name, step18Supported[name], false})
	}
	for _, name := range []string{"runtime-values.a", "runtime-order.a", "runtime-methods.a", "runtime-arguments.a", "runtime-cross-chain.a", "runtime-return-descriptor.a"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{step18Directory + name, step18Supported[name], false})
	}
}

type step18Observation struct {
	File   string
	Reason string
	Where  string
	Node   struct {
		Exit   int
		Stdout string
		Stderr string
	}
}

func step18Observations(t *testing.T) []step18Observation {
	t.Helper()
	var observations []step18Observation
	for _, name := range []string{"baseline.json", "runtime-observations.json"} {
		data, err := os.ReadFile(filepath.Join(repository, "docs/step-18", name))
		if err != nil {
			t.Fatal(err)
		}
		var rows []step18Observation
		if err := json.Unmarshal(data, &rows); err != nil {
			t.Fatal(err)
		}
		observations = append(observations, rows...)
	}
	return observations
}

// The source oracle still runs for a NotYet fixture, whose ordinary oracle row
// checks the compiler barrier without reaching either backend.
func TestStep18SourceBaselines(t *testing.T) {
	for _, observation := range step18Observations(t) {
		t.Run(observation.File, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, step18Directory+observation.File))
			if err != nil {
				t.Fatal(err)
			}
			got := onNode(t, path)
			if got.exitCode != observation.Node.Exit || string(got.stdout) != observation.Node.Stdout || string(got.stderr) != observation.Node.Stderr {
				t.Fatalf("Node differs from source baseline: exit %d, stdout %q, stderr %q; want exit %d, stdout %q, stderr %q", got.exitCode, got.stdout, got.stderr, observation.Node.Exit, observation.Node.Stdout, observation.Node.Stderr)
			}
		})
	}
}

func TestStep18RecordedGaps(t *testing.T) {
	for _, observation := range step18Observations(t) {
		if step18Supported[observation.File] {
			continue
		}
		t.Run(observation.File, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, step18Directory+observation.File))
			if err != nil {
				t.Fatal(err)
			}
			_, err = lowered(t, path)
			var notYet *lower.NotYet
			if !errors.As(err, &notYet) || notYet.What != observation.Reason || !strings.HasSuffix(notYet.Where, observation.Where) {
				t.Fatalf("got %v, want NotYet %q at %s", err, observation.Reason, observation.Where)
			}
		})
	}
}
