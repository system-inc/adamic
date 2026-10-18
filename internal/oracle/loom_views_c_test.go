package oracle

import (
	"path/filepath"
	"testing"
)

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/loom_views_c/snapshot-original.a", true, false})
}
func TestLoomViewsCOriginal(t *testing.T) {
	t.Parallel()
	path := filepath.Join(repository, "internal/oracle/testdata/loom_views_c/snapshot-original.a")
	path, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}
	want := onNode(t, path)
	if want.exitCode != 0 || string(want.stdout) != "true\nundefined\n" {
		t.Fatalf("Node: %#v", want)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	sanitized, _ := nativelyUncached(t, program)
	for backend, got := range map[string]run{"native": releasedUncached(t, program), "sanitized": sanitized, "javascript": onJavaScriptBackend(t, program)} {
		if diff := disagreement(want, got); diff != "" {
			t.Errorf("%s: %s, stderr %q", backend, diff, got.stderr)
		}
	}
}
