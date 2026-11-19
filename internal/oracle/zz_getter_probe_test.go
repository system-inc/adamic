package oracle

import (
	"path/filepath"
	"testing"
)

// Throwaway diagnostic, never committed to the pins branch: what the class getter twin does on each backend.
func TestZZGetterProbe(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/reland_refused/view_widened_held_class_getter.ts"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	t.Logf("PROBE node: exit %d stdout %q stderr %q", truth.exitCode, truth.stdout, truth.stderr)
	program, err := lowered(t, path)
	if err != nil {
		t.Logf("PROBE lowered: %v", err)
		return
	}
	native, _ := nativelyUncached(t, program)
	for name, got := range map[string]run{"sanitized": native, "release": releasedUncached(t, program), "javascript": onJavaScriptBackend(t, program)} {
		t.Logf("PROBE %s: exit %d stdout %q stderr %.600q", name, got.exitCode, got.stdout, got.stderr)
	}
}
