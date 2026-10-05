package oracle

import (
	"path/filepath"
	"testing"
)

// A Weak read after what it points to was freed is where native and Node differ by design
// (docs/memory.md): natively the target is gone, and on Node the collector keeps it as long as the
// Weak points at it. So these programs are held to what each side is meant to do, not to each other:
// natively undefined, or a panic where the checker had proven the target present, and never a read
// of freed memory (the sanitizers are on); on Node, and through the JavaScript backend, the target.
func TestWeakReadsUndefinedOnceFreed(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		path           string
		native, onNode run
	}{
		{
			"internal/oracle/testdata/weak/freed.a",
			run{stdout: []byte("while held: bbb\nafter: gone true\n")},
			run{stdout: []byte("while held: bbb\nafter: bbb false\n")},
		},
		{
			"internal/oracle/testdata/weak/probe_chain.a",
			run{stdout: []byte("0\n")},
			run{stdout: []byte("100\n")},
		},
		{
			"internal/oracle/testdata/weak/reuse.a",
			run{stdout: []byte("gone 0\n")},
			run{stdout: []byte("t1 0\n")},
		},
		{
			"internal/oracle/testdata/weak/narrowed.a",
			run{stdout: []byte("before: cc\n"), stderr: []byte("adamic: panic: a weak reference was read after what it pointed to was freed\n"), exitCode: 70},
			run{stdout: []byte("before: cc\nafter: cc\n")},
		},
	} {
		t.Run(probe.path, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, probe.path))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatalf("Lower: %v", err)
			}
			native, sanitized := natively(t, program)
			if difference := disagreement(probe.native, native); difference != "" {
				t.Errorf("native: %s: exit %d, stdout %q, stderr %q", difference, native.exitCode, native.stdout, native.stderr)
			}
			for name, side := range map[string]run{"Node": onNode(t, path), "the JavaScript backend": onJavaScriptBackend(t, program)} {
				if difference := disagreement(probe.onNode, side); difference != "" {
					t.Errorf("%s: %s: exit %d, stdout %q, stderr %q", name, difference, side.exitCode, side.stdout, side.stderr)
				}
			}
			if native.exitCode == 0 {
				if leaked := leaks(t, program, sanitized); leaked != "" {
					t.Errorf("leaks:\n%s", leaked)
				}
			}
		})
	}
}
