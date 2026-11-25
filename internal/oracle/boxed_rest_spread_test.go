package oracle

import (
	"path/filepath"
	"testing"
)

func TestBoxedRestSpread(t *testing.T) {
	t.Parallel()
	for _, name := range []string{
		"p1", "p2", "p3", "p4", "p5", "p6", "h02", "h03",
		"extra-push", "extra-forward-twice", "extra-function",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join("testdata", "boxed_rest_spread", name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			got, binary := natively(t, program)
			for name, run := range map[string]run{
				"sanitized": got, "release": released(t, program), "JavaScript": onJavaScriptBackend(t, program),
			} {
				if difference := disagreement(truth, run); difference != "" {
					t.Errorf("%s: %s; Node %q, got %q; stderr %q", name, difference, truth.stdout, run.stdout, run.stderr)
				}
			}
			if leaked := leaks(t, program, binary); leaked != "" {
				t.Fatal(leaked)
			}
		})
	}
}
