package oracle

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

const step18Directory = "docs/step-18/fixtures/"

var step18Supported = map[string]bool{"runtime-chain-parentheses-write.a": true, "runtime-chain-parentheses.a": true, "runtime-chain-method-continuation.a": true, "runtime-chain-index.a": true, "runtime-chain-intrinsics.a": true, "runtime-chain-callable.a": true, "runtime-chain-present-undefined.a": true, "element.a": true, "call-result.a": true, "cross-call.a": true, "receiver-call.a": true, "two-guards.a": true, "size.a": true, "runtime-cross-chain.a": true, "function.a": true, "method.a": true, "runtime-values.a": true, "runtime-order.a": true, "runtime-arguments.a": true, "runtime-return-descriptor.a": true, "runtime-methods.a": true, "runtime-bound-methods.a": true, "runtime-bound-method-arguments.a": true, "runtime-bound-method-selection.a": true, "runtime-discarded-reference.a": true}

func init() {
	for _, name := range []string{"function.a", "method.a", "element.a", "call-result.a", "cross-call.a", "receiver-call.a", "two-guards.a", "size.a", "number.a"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{step18Directory + name, step18Supported[name], false})
	}
	for _, name := range []string{"runtime-chain-parentheses-write.a", "runtime-chain-parentheses.a", "runtime-chain-method-continuation.a", "runtime-chain-index.a", "runtime-chain-intrinsics.a", "runtime-chain-callable.a", "runtime-chain-present-undefined.a", "runtime-values.a", "runtime-order.a", "runtime-methods.a", "runtime-arguments.a", "runtime-cross-chain.a", "runtime-return-descriptor.a", "runtime-bound-methods.a", "runtime-bound-method-arguments.a", "runtime-bound-method-selection.a", "runtime-discarded-reference.a"} {
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
	t.Parallel()
	for _, observation := range step18Observations(t) {
		t.Run(observation.File, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, step18Directory+observation.File))
			if err != nil {
				t.Fatal(err)
			}
			got := onNode(t, path)
			// Historical source observations keep the pre-step-21 exit-70 renderer.
			// The same language TypeError now has Node's ordinary uncaught exit 1.
			if observation.Node.Exit == 70 && got.exitCode != 70 && strings.HasPrefix(observation.Node.Stderr, "adamic: panic: TypeError:") {
				message := strings.TrimSpace(strings.TrimPrefix(observation.Node.Stderr, "adamic: panic: "))
				if got.exitCode != 1 || string(got.stdout) != observation.Node.Stdout || !strings.Contains(string(got.stderr), message) {
					t.Fatalf("Node language exception differs: got %#v, want stdout %q and %s", got, observation.Node.Stdout, message)
				}
				return
			}
			if got.exitCode != observation.Node.Exit || string(got.stdout) != observation.Node.Stdout || string(got.stderr) != observation.Node.Stderr {
				t.Fatalf("Node differs from source baseline: exit %d, stdout %q, stderr %q; want exit %d, stdout %q, stderr %q", got.exitCode, got.stdout, got.stderr, observation.Node.Exit, observation.Node.Stdout, observation.Node.Stderr)
			}
		})
	}
}

func TestStep18RecordedGaps(t *testing.T) {
	t.Parallel()
	for _, observation := range step18Observations(t) {
		if step18Supported[observation.File] || observation.File == "runtime-bound-method-replacement.a" {
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

func TestStep18MethodWritesStayRefused(t *testing.T) {
	t.Parallel()
	path := filepath.Join(repository, step18Directory, "runtime-bound-method-replacement.a")
	_, err := lowered(t, path)
	var refused *lower.Refused
	if !errors.As(err, &refused) || !strings.Contains(refused.What, "method read as a value") {
		t.Fatalf("got %v, want the existing unbound-method refusal", err)
	}
}

// LeakSanitizer can keep a discarded pointer reachable on a native stack. The
// counted build also checks that every heap value's ownership was returned.
func TestStep18DiscardedReferencesBalance(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"runtime-chain-intrinsics.a", "runtime-discarded-reference.a"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			program, err := lowered(t, filepath.Join(repository, step18Directory, name))
			if err != nil {
				t.Fatal(err)
			}
			binary := filepath.Join(t.TempDir(), "counted")
			if err := native.Build(native.C(program), binary, native.Options{Count: true}); err != nil {
				t.Fatal(err)
			}
			if report := unbalanced(t, execute(t, binary)); report != "" {
				t.Fatal(report)
			}
		})
	}
}
