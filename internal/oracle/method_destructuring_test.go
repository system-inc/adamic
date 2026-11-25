package oracle

import (
	"path/filepath"
	"testing"
)

func TestDetachableMethodDestructuringAgrees(t *testing.T) {
	t.Parallel()
	for _, name := range []string{
		"arrow", "method-bind", "fixed-bind", "getter", "generic-arrow",
		"assignment", "renamed", "class-arrow", "shared-arrow", "sys-wrapped", "write-fresh-parameter",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join("testdata", "method_destructuring", name+".ts"))
			if err != nil {
				t.Fatal(err)
			}
			want := onNode(t, path)
			if want.exitCode != 0 || len(want.stdout) == 0 {
				t.Fatalf("Node did not make an observation: %#v", want)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			for backend, got := range map[string]run{
				"JavaScript": onJavaScriptBackend(t, program),
				"native":     released(t, program),
			} {
				if diff := disagreement(want, got); diff != "" {
					t.Errorf("%s: %s", backend, diff)
				}
			}
		})
	}
}
