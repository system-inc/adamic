package oracle

import (
	"os"
	"path/filepath"
	"testing"
)

func init() {
	for _, name := range []string{"free", "bind"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/method_values/" + name + ".a", true, false})
	}
}

func TestMethodValuesTypeScript(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"free", "bind", "reading", "uncaught"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			source, err := os.ReadFile(filepath.Join(repository, "internal/oracle/testdata/method_values", name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(t.TempDir(), "main.ts")
			if err := os.WriteFile(path, source, 0600); err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			actual, sanitized := natively(t, program)
			for backend, result := range map[string]run{"native": actual, "javascript": onJavaScriptBackend(t, program), "release": released(t, program)} {
				if difference := disagreement(truth, result); difference != "" {
					t.Errorf("%s: %s; Node exit=%d stdout=%q stderr=%q; actual exit=%d stdout=%q stderr=%q", backend, difference, truth.exitCode, truth.stdout, truth.stderr, result.exitCode, result.stdout, result.stderr)
				}
			}
			if truth.exitCode == 0 {
				if report := leaks(t, program, sanitized); report != "" {
					t.Error(report)
				}
			}
		})
	}
}
