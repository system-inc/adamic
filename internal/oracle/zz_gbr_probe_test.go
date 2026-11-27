package oracle

import "testing"

// Probe for #gbr6vs2, never merged: the honest program on Node, natively (sanitized), released and on the
// JavaScript backend, and the lie's output on each, logged whole so the job reads them.
func TestGbrProbe(t *testing.T) {
	for _, name := range []string{"honest", "lie"} {
		program, path := castFixture(t, "gbr_probe/"+name+".ts")
		node := onNode(t, path)
		t.Logf("%s node: exit %d stdout %q stderr %q", name, node.exitCode, node.stdout, node.stderr)
		native, _ := natively(t, program)
		for backend, result := range map[string]run{"native": native, "release": released(t, program), "JavaScript": onJavaScriptBackend(t, program)} {
			t.Logf("%s %s: exit %d stdout %q stderr %q disagreement %q", name, backend, result.exitCode, result.stdout, result.stderr, disagreement(node, result))
		}
	}
}
