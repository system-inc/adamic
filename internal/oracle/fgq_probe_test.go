package oracle

import (
	"path/filepath"
	"testing"
)

// Scratch probe for #fgqj0tm, never for main: what each backend does when a statically dispatched method
// reads a field of a generic spread's copy.
func TestFgqProbe(t *testing.T) {
	for _, name := range []string{"private_hash.ts", "private_ts.ts", "public_field.ts", "control_instance.ts"} {
		t.Run(name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/fgq_probe", name))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			t.Logf("PROBE %s node exit %d stdout %q", name, truth.exitCode, truth.stdout)
			program, err := lowered(t, path)
			if err != nil {
				t.Logf("PROBE %s lowering: %v", name, err)
				return
			}
			native, _ := natively(t, program)
			t.Logf("PROBE %s sanitized exit %d stdout %q stderr %.600q", name, native.exitCode, native.stdout, native.stderr)
			release := released(t, program)
			t.Logf("PROBE %s release exit %d stdout %q stderr %.300q", name, release.exitCode, release.stdout, release.stderr)
			backend := onJavaScriptBackend(t, program)
			t.Logf("PROBE %s javascript exit %d stdout %q", name, backend.exitCode, backend.stdout)
		})
	}
}
