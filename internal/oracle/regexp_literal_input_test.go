package oracle

import (
	"errors"
	"github.com/system-inc/adamic/internal/lower"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

const regexpLiteralInputSource = `console.log((/[\q{Ss|x}]/iv.test('sſ') ? 'true' : 'false') + ' ' + (/[\q{Ss|x}]/iv.test('SS') ? 'true' : 'false') + ' ' + (/[\q{Ss|x}]/iv.test('🌍ss') ? 'true' : 'false') + ' ' + (/[\q{Ss|x}]/iv.test('x') ? 'true' : 'false') + ' ' + (/[\q{Ss|x}]/iv.test('') ? 'true' : 'false') + ' ' + (/[\q{x}]/iv.test('x') ? 'true' : 'false'));`

func runRegExpLiteralInputNodeUnit(t *testing.T, target string) {
	if target == "wasm32-wasi" && os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	path := filepath.Join(t.TempDir(), "literal.a")
	if err := os.WriteFile(path, []byte(regexpLiteralInputSource), 0600); err != nil {
		t.Fatal(err)
	}
	want := execute(t, "node", "-e", regexpLiteralInputSource)
	if want.exitCode != 0 || len(want.stderr) != 0 || string(want.stdout) != "true true true true false true\n" {
		t.Fatalf("Node: %+v", want)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	var got run
	if target == "wasm32-wasi" {
		got = onWASI(t, native.C(program))
	} else {
		binary := filepath.Join(t.TempDir(), "program")
		if err := native.Build(native.C(program), binary, native.Options{Sanitize: target == "sanitized"}); err != nil {
			t.Fatal(err)
		}
		got = execute(t, binary)
	}
	if difference := disagreement(want, got); difference != "" {
		t.Fatalf("%s: Node=%q compiled=%q stderr=%q", difference, want.stdout, got.stdout, got.stderr)
	}
}
func TestRegExpLiteralInputReleaseUnit(t *testing.T) {
	t.Parallel()
	runRegExpLiteralInputNodeUnit(t, "native")
}
func TestRegExpLiteralInputSanitizeUnit(t *testing.T) {
	t.Parallel()
	runRegExpLiteralInputNodeUnit(t, "sanitized")
}
func TestRegExpLiteralInputWASMUnit(t *testing.T) {
	t.Parallel()
	runRegExpLiteralInputNodeUnit(t, "wasm32-wasi")
}

// A refusal is valid. Any admitted input must instead agree with Node.
func TestRegExpLiteralInputUnsafeNode(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "unsafe.a")
	if err := os.WriteFile(path, []byte("console.log( /[\\q{Ss|x}]/iv.test('X') ? 'true' : 'false');"), 0600); err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := execute(t, "node", "-e", string(source))
	program, err := lowered(t, path)
	if err != nil {
		var refusal *lower.NotYet
		if !errors.As(err, &refusal) {
			t.Fatal(err)
		}
		t.Logf("unsafe literal refused: %v", err)
		return
	}
	binary := filepath.Join(t.TempDir(), "program")
	if err := native.Build(native.C(program), binary, native.Options{}); err != nil {
		t.Fatal(err)
	}
	if difference := disagreement(want, execute(t, binary)); difference != "" {
		t.Fatalf("admitted unsafe literal: %s", difference)
	}
	if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
		if difference := disagreement(want, onWASI(t, native.C(program))); difference != "" {
			t.Fatalf("admitted unsafe wasm32 literal: %s", difference)
		}
	}
}
