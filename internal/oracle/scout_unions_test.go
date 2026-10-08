package oracle

import (
	"path/filepath"
	"strings"
	"testing"
)

// Historical port programs are remeasured without claiming to undo their corpus workarounds.
func TestScoutUnionSourceOutcomes(t *testing.T) {
	for _, probe := range []struct{ name, output, stop string }{
		{"selector-optional-boolean.a", "true\n", "a field of type boolean | undefined"},
		{"values-optional-boolean.a", "true\n", "a field of type boolean | undefined"},
		{"hir-optional-boolean.a", "true\n", ""},
		{"tsc-source-content.a", "source\nmissing\n", "a value of type string | null"},
		{"tsc-nullish-content.a", "string\nsource\nobject\nnull\nundefined\nundefined\n", "a value of type string | null | undefined"},
		{"presence.a", "false\n[]\nfalse\ntrue\n[first]\ntrue\n", "refuses delete"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "docs/step-17-unions/probes", probe.name))
			if err != nil {
				t.Fatal(err)
			}
			source := onNode(t, path)
			if difference := disagreement(run{stdout: []byte(probe.output)}, source); difference != "" {
				t.Fatal(difference)
			}
			program, err := lowered(t, path)
			if probe.stop != "" {
				if err == nil || !strings.Contains(err.Error(), probe.stop) {
					t.Fatalf("expected stop %q, got %v", probe.stop, err)
				}
				t.Logf("CURRENT STOP: %v", err)
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			backend := onJavaScriptBackend(t, program)
			native, sanitized := natively(t, program)
			for name, result := range map[string]run{"JavaScript": backend, "native": native, "release": released(t, program)} {
				if difference := disagreement(source, result); difference != "" {
					t.Errorf("%s: %s", name, difference)
				}
			}
			if leak := leaks(t, program, sanitized); leak != "" {
				t.Fatal(leak)
			}
			t.Log("CURRENT OUTCOME: both backends match source Node; sanitizers and leak check passed")
		})
	}
}
