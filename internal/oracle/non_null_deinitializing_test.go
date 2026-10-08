package oracle

import (
	"path/filepath"
	"testing"
)

func TestDeinitializingStatementsStopOnReadiness(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ fixture, expression, node string }{
		{"read", "variable 'value' in value", "undefined"},
		{"field", "field 'value' in this.value", "undefined"},
		{"alias", "field 'value' in box.value", "null"},
	} {
		t.Run(probe.fixture, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/non_null_deinitialize_"+probe.fixture+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			want := run{stdout: []byte("before\n"), stderr: []byte("adamic: panic: read before assignment: " + probe.expression + "\n"), exitCode: 70}
			native, _ := nativelyUncached(t, program)
			for _, got := range []run{native, onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Fatal(difference)
				}
			}
			node := onNode(t, path)
			if difference := disagreement(run{stdout: []byte("before\n" + probe.node + "\n")}, node); difference != "" {
				t.Fatal("source Node: " + difference)
			}
		})
	}
}
