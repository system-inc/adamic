package oracle

import (
	"path/filepath"
	"strings"
	"testing"
)

// Register with the existing Node, both-backend, sanitizer, leak and counts harness.
// This keeps the unit out of the shared oracle_test.go registry.
func init() {
	for _, name := range []string{"factories", "nested", "hooks", "throw", "stage3_roots", "null_hook_refused"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/json_objects_" + name + ".a", name != "null_hook_refused", false})
	}
}

// The unit gate uses -run JSON, so exercise registered .a fixtures here as well.
func TestJSONObjects(t *testing.T) {
	for _, fixture := range fixtures {
		if !strings.Contains(fixture.path, "/json_objects_") {
			continue
		}
		t.Run(filepath.Base(fixture.path), func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, fixture.path))
			if err != nil {
				t.Fatal(err)
			}
			want := onNode(t, path)
			t.Logf("Node 24.19.0: exit %d stdout %s", want.exitCode, want.stdout)
			program, err := lowered(t, path)
			if !fixture.lowers {
				if err == nil || !strings.Contains(err.Error(), "JSON.stringify a value of type null") {
					t.Fatalf("want existing null-return refusal, got %v", err)
				}
				t.Logf("not yet: %v", err)
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			got, binary := natively(t, program)
			for name, observation := range map[string]run{"sanitized": got, "release": released(t, program), "backend": onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, observation); difference != "" {
					t.Errorf("%s: %s\nNode %s\nactual %s\nstderr %s", name, difference, want.stdout, observation.stdout, observation.stderr)
				}
			}
			if report := leaks(t, program, binary); report != "" {
				t.Errorf("leaks: %s", report)
			}
		})
	}
}
