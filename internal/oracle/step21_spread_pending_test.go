package oracle

import (
	"os"
	"path/filepath"
	"testing"
)

// Main's object-spread loop must stop before the next getter even when the
// pending payload is undefined. A non-null Error cannot prove the pending bit.
func TestSpreadUndefinedThrowStopsBeforeNextGetter(t *testing.T) {
	t.Parallel()
	source := `const original = {
 get first(): number { throw undefined; },
 get second(): number { console.log('second getter'); return 2; }
};
try { const copy = {...original}; console.log(String(copy.first + copy.second)); }
catch (error) { console.log(error === undefined ? 'undefined' : 'wrong'); }
`
	path := filepath.Join(t.TempDir(), "spread.a")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	want := onNode(t, path)
	if string(want.stdout) != "undefined\n" || want.exitCode != 0 || len(want.stderr) != 0 {
		t.Fatalf("Node: %+v", want)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	got, binary := nativelyUncached(t, program)
	for name, observation := range map[string]run{"sanitized": got, "release": released(t, program), "JavaScript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, observation); difference != "" {
			t.Fatalf("%s spread: %s", name, difference)
		}
	}
	if leaked := leaks(t, program, binary); leaked != "" {
		t.Fatal(leaked)
	}
}
