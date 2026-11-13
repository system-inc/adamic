package oracle

import (
	"path/filepath"
	"testing"
)

var unhandledRejectionFixtures = []struct {
	name, stdout, message string
}{
	{"void_view", "", "rejected"},
	{"chained", "", "chained"},
	{"handled_later", "handled\n", ""},
	{"timer", "timer\n", "timer rejection"},
	{"two", "", "first"},
}

func unhandledRejectionPath(name string) string {
	return "internal/oracle/testdata/unhandled_rejection/" + name + ".a"
}

func init() {
	// Keep the dedicated runtime-stop harness, as json_errors_named does, while
	// registering the counted executions in the ordinary count family.
	for _, fixture := range unhandledRejectionFixtures {
		path := unhandledRejectionPath(fixture.name)
		additionalFixtureCounts = append(additionalFixtureCounts, countCase{path, func(t *testing.T) (string, func() string) {
			return counted(t, path, false, nil, false, false)
		}})
	}
}

// Skipping the checkpoint reds void_view and chained. Omitting await's handled
// bit reds handled_later; omitting the queue's retain reds void_view under sanitizers.
func TestUnhandledRejectionAgreesWithNode(t *testing.T) {
	t.Parallel()
	for _, fixture := range unhandledRejectionFixtures {
		t.Run(fixture.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, unhandledRejectionPath(fixture.name)))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			want := run{stdout: []byte(fixture.stdout)}
			if fixture.message != "" {
				want.exitCode = 70
				want.stderr = []byte("adamic: panic: Error: " + fixture.message + "\n")
			}
			truth := onNode(t, path)
			if difference := disagreement(want, truth); difference != "" {
				t.Fatalf("source Node: %s: exit %d stdout %q stderr %q", difference, truth.exitCode, truth.stdout, truth.stderr)
			}
			for _, backend := range []string{"sanitized", "release", "javascript"} {
				t.Run(backend, func(t *testing.T) {
					var got run
					switch backend {
					case "sanitized":
						var binary string
						got, binary = natively(t, program)
						if want.exitCode == 0 {
							if report := leaks(t, program, binary); report != "" {
								t.Fatal(report)
							}
						}
					case "release":
						got = released(t, program)
					case "javascript":
						got = onJavaScriptBackend(t, program)
					}
					if difference := disagreement(truth, got); difference != "" {
						t.Fatalf("%s: exit %d stdout %q stderr %q", difference, got.exitCode, got.stdout, got.stderr)
					}
				})
			}
		})
	}
}
