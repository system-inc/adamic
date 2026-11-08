package oracle

import (
	"path/filepath"
	"testing"
)

// asserted_method/absent.ts calls box.run!() after a call emptied the field. Since #0bxc4jh an asserted
// callee keeps its object, as an unasserted one does (ir.MethodCallee's Coalesce), and its check stays
// the non-null assertion's own: on every leg the panic names the assertion's site and text, as when it
// called the field's value with no object. Node throws box.run is not a function there, so the legs are
// held to the assertion's words, not to Node, and Node only to the stdout before it. The fixture keeps
// a directory of its own, out of the top of testdata, where its panic would be a disagreement.
func TestAssertedMethodCalleeKeepsTheAssertionsWords(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/asserted_method/absent.ts"))
	if err != nil {
		t.Fatal(err)
	}
	if node := onNode(t, path); node.exitCode == 0 || string(node.stdout) != "hello\n" {
		t.Fatalf("unexpected Node result %d %q %q", node.exitCode, node.stdout, node.stderr)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	want := run{stdout: []byte("hello\n"), stderr: []byte("adamic: panic: non-null assertion failed at " + path + ":17:1: box.run! is null or undefined\n"), exitCode: 70}
	native, _ := natively(t, program)
	for name, got := range map[string]run{"native": native, "release": released(t, program), "JavaScript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, got); difference != "" {
			t.Fatalf("%s: %s: exit %d, stdout %q, stderr %q", name, difference, got.exitCode, got.stdout, got.stderr)
		}
	}
}
