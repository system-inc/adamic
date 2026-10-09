package oracle

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func TestRegExpLiteralTestInputAgreesWithNode(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "literal.a")
	source := `console.log((/[\q{Ss|x}]/iv.test('sſ') ? 'true' : 'false') + ' ' + (/[\q{Ss|x}]/iv.test('SS') ? 'true' : 'false') + ' ' + (/[\q{Ss|x}]/iv.test('🌍ss') ? 'true' : 'false') + ' ' + (/[\q{Ss|x}]/iv.test('x') ? 'true' : 'false') + ' ' + (/[\q{Ss|x}]/iv.test('') ? 'true' : 'false') + ' ' + (/[\q{x}]/iv.test('x') ? 'true' : 'false'));`
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	want := onNode(t, path)
	if want.exitCode != 0 || len(want.stderr) != 0 || string(want.stdout) != "true true true true false true\n" {
		t.Fatalf("Node: exit=%d stdout=%q stderr=%q", want.exitCode, want.stdout, want.stderr)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	actual, binary := natively(t, program)
	for name, got := range map[string]run{"native": actual, "release": released(t, program), "javascript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, got); difference != "" {
			t.Errorf("%s: %s stdout=%q stderr=%q", name, difference, got.stdout, got.stderr)
		}
	}
	if report := leaks(t, program, binary); report != "" {
		t.Error(report)
	}
	if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
		if difference := disagreement(want, onWASI(t, native.C(program))); difference != "" {
			t.Fatal("wasm32: " + difference)
		}
	}
	t.Logf("Node and admitted literal inputs: %q", want.stdout)
}
func TestRegExpReviewFoldStringsWASI(t *testing.T) {
	t.Parallel()
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	path, err := filepath.Abs(filepath.Join(reviewRoot, "agree", "fxspptb_e23c7ab_fold_strings.a"))
	if err != nil {
		t.Fatal(err)
	}
	want := onNode(t, path)
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	actual := onWASI(t, native.C(program))
	if difference := disagreement(want, actual); difference != "" {
		t.Fatalf("%s: Node=%q wasm32=%q stderr=%q", difference, want.stdout, actual.stdout, actual.stderr)
	}
	t.Logf("Node and wasm32: exit=%d stdout=%q stderr=%q", actual.exitCode, actual.stdout, actual.stderr)
}
