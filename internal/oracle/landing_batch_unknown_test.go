package oracle

import (
	"path/filepath"
	"testing"
)

func TestLandingUnknownOptionalBoolean(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "docs/verification/landing-batch-3/unknown-optional-boolean.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	node := onNode(t, path)
	native, binary := nativelyUncached(t, program)
	release := released(t, program)
	backend := onJavaScriptBackend(t, program)
	if report := leaksUncached(t, program, binary); report != "" {
		t.Errorf("leaks: %s", report)
	}
	for _, observed := range []struct {
		name   string
		result run
	}{{"native", native}, {"release", release}, {"backend", backend}} {
		t.Logf("%s: exit %d stdout %q stderr %q", observed.name, observed.result.exitCode, observed.result.stdout, observed.result.stderr)
		if difference := disagreement(node, observed.result); difference != "" {
			t.Errorf("%s: %s; Node exit %d stdout %q stderr %q", observed.name, difference, node.exitCode, node.stdout, node.stderr)
		}
	}
}
