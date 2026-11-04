package oracle

import (
	"path/filepath"
	"testing"
)

// A probe for Lower's question on #1w8gvx9, never to land: a record handed to a generic whose
// constraint is a fixed, non-weak type. Each backend's output is logged beside Node's.
func TestProbeGenericRecordNonWeak(t *testing.T) {
	for _, name := range []string{"nonweak.ts", "nonweak.a"} {
		path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/probe_generic_record/"+name))
		if err != nil {
			t.Fatal(err)
		}
		node := onNode(t, path)
		t.Logf("PROBE %s node exit=%d stdout=%q stderr=%q", name, node.exitCode, node.stdout, node.stderr)
		program, err := lowered(t, path)
		if err != nil {
			t.Logf("PROBE %s lowering stops: %v", name, err)
			continue
		}
		native, _ := natively(t, program)
		t.Logf("PROBE %s native exit=%d stdout=%q stderr=%q", name, native.exitCode, native.stdout, native.stderr)
		javascript := onJavaScriptBackend(t, program)
		t.Logf("PROBE %s javascript exit=%d stdout=%q stderr=%q", name, javascript.exitCode, javascript.stdout, javascript.stderr)
	}
}
