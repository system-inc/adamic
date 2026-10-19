package oracle

import (
	"os"
	"path/filepath"
	"testing"
)

// Immortal prefix strings isolate nominal ancestry from the independent
// inherited-prefix leak guard in the larger subclass fixture.
func TestErrorSubclassAncestryAgreesWithNode(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "ancestry.a")
	source := "class Failure extends Error {} const value = new Failure('literal'); console.log(String(value instanceof Error));\n"
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	want := onNode(t, path)
	if string(want.stdout) != "true\n" || want.exitCode != 0 {
		t.Fatalf("Node: %+v", want)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	got, binary := nativelyUncached(t, program)
	for name, observation := range map[string]run{"sanitized": got, "release": released(t, program), "JavaScript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, observation); difference != "" {
			t.Fatalf("%s ancestry: %s", name, difference)
		}
	}
	if leaked := leaks(t, program, binary); leaked != "" {
		t.Fatal(leaked)
	}
}
