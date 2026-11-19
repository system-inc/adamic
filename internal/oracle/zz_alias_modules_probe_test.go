package oracle

import (
	"path/filepath"
	"testing"
)

// Scratch for #1trc7d3, never cut: what an exported stdout alias does in an importing module.
func TestZZAliasModulesProbe(t *testing.T) {
	for _, name := range []string{"use_write.a", "use_value.a", "use_tty.a", "use_reexport.a", "use_namespace.a"} {
		path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/stdout_alias_modules", name))
		if err != nil {
			t.Fatal(err)
		}
		truth := onNode(t, path)
		t.Logf("PROBE %s node: exit %d stdout %q stderr %q", name, truth.exitCode, truth.stdout, truth.stderr)
		program, err := lowered(t, path)
		if err != nil {
			t.Logf("PROBE %s lower: %v", name, err)
			continue
		}
		release := released(t, program)
		t.Logf("PROBE %s release: exit %d stdout %q stderr %q", name, release.exitCode, release.stdout, release.stderr)
		script := onJavaScriptBackend(t, program)
		t.Logf("PROBE %s javascript: exit %d stdout %q stderr %q", name, script.exitCode, script.stdout, script.stderr)
	}
}
