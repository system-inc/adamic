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

func runRegExpLiteralInputNodeUnit(test *testing.T, target string) {
	if target == "wasm32-wasi" && os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		test.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	path := filepath.Join(test.TempDir(), "literal.a")
	if writeError := os.WriteFile(path, []byte(regexpLiteralInputSource), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	want := execute(test, "node", "-e", regexpLiteralInputSource)
	if want.exitCode != 0 || len(want.stderr) != 0 || string(want.stdout) != "true true true true false true\n" {
		test.Fatalf("Node: %+v", want)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	var got run
	if target == "wasm32-wasi" {
		got = onWASI(test, native.C(program))
	} else {
		binary := filepath.Join(test.TempDir(), "program")
		if buildError := native.Build(
			native.C(program),
			binary,
			native.Options{Sanitize: target == "sanitized"},
		); buildError != nil {
			test.Fatal(buildError)
		}
		got = execute(test, binary)
	}
	if difference := disagreement(want, got); difference != "" {
		test.Fatalf("%s: Node=%q compiled=%q stderr=%q", difference, want.stdout, got.stdout, got.stderr)
	}
}
func TestRegExpLiteralInputReleaseUnit(test *testing.T) {
	test.Parallel()
	runRegExpLiteralInputNodeUnit(test, "native")
}
func TestRegExpLiteralInputSanitizeUnit(test *testing.T) {
	test.Parallel()
	runRegExpLiteralInputNodeUnit(test, "sanitized")
}
func TestRegExpLiteralInputWASMUnit(test *testing.T) {
	test.Parallel()
	runRegExpLiteralInputNodeUnit(test, "wasm32-wasi")
}

// A refusal is valid. Any admitted input must instead agree with Node.
func TestRegExpLiteralInputUnsafeNode(test *testing.T) {
	test.Parallel()
	path := filepath.Join(test.TempDir(), "unsafe.a")
	if writeError := os.WriteFile(
		path,
		[]byte("console.log( /[\\q{Ss|x}]/iv.test('X') ? 'true' : 'false');"),
		0600,
	); writeError != nil {
		test.Fatal(writeError)
	}
	source, readError := os.ReadFile(path)
	if readError != nil {
		test.Fatal(readError)
	}
	want := execute(test, "node", "-e", string(source))
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		var refusal *lower.NotYet
		if !errors.As(loweredError, &refusal) {
			test.Fatal(loweredError)
		}
		test.Logf("unsafe literal refused: %v", loweredError)
		return
	}
	binary := filepath.Join(test.TempDir(), "program")
	if buildError := native.Build(native.C(program), binary, native.Options{}); buildError != nil {
		test.Fatal(buildError)
	}
	if difference := disagreement(want, execute(test, binary)); difference != "" {
		test.Fatalf("admitted unsafe literal: %s", difference)
	}
	if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
		if difference := disagreement(want, onWASI(test, native.C(program))); difference != "" {
			test.Fatalf("admitted unsafe wasm32 literal: %s", difference)
		}
	}
}
