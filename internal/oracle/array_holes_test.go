package oracle

import (
	"path/filepath"
	"testing"
)

func TestArrayHolesMilestone(t *testing.T) {
	for _, name := range []string{"library_array_holes_scanner_probe.a", "library_array_holes_length.a", "library_array_holes_range.a"} {
		t.Run(name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", name))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			native, binary := natively(t, program)
			for _, got := range []run{native, released(t, program), onJavaScriptBackend(t, program)} {
				if difference := disagreement(truth, got); difference != "" {
					t.Fatalf("%s: Node %q %q; got %q %q", difference, truth.stdout, truth.stderr, got.stdout, got.stderr)
				}
			}
			if report := leaks(t, program, binary); report != "" {
				t.Fatal(report)
			}
			t.Logf("Node stdout: %s", truth.stdout)
		})
	}
}

func init() {
	for _, name := range []string{"library_array_holes_scanner_probe.a", "library_array_holes_length.a", "library_array_holes_range.a"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/" + name, true, false})
	}
}
