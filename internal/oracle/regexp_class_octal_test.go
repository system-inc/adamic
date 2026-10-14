package oracle

import (
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"testing"
)

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/regexp_native_class_octal.a", true, false})
}

func runRegExpFixtureNodeUnit(t *testing.T, fixture string, target string) {
	if target == "wasm32-wasi" && os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	path, err := filepath.Abs(filepath.Join("testdata", fixture+".a"))
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := execute(t, "node", "-e", string(source))
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
func TestRegExpClassOctalNativeUnit(t *testing.T) {
	runRegExpFixtureNodeUnit(t, "regexp_native_class_octal", "native")
}
func TestRegExpClassOctalWASMUnit(t *testing.T) {
	runRegExpFixtureNodeUnit(t, "regexp_native_class_octal", "wasm32-wasi")
}
func TestRegExpFoldingNativeUnit(t *testing.T) {
	runRegExpFixtureNodeUnit(t, "regexp_v8_folding", "native")
}
func TestRegExpFoldingWASMUnit(t *testing.T) {
	runRegExpFixtureNodeUnit(t, "regexp_v8_folding", "wasm32-wasi")
}
func TestRegExpSurrogatesReleaseUnit(t *testing.T) {
	runRegExpFixtureNodeUnit(t, "regexp_v8_surrogates", "native")
}
func TestRegExpSurrogatesWASMUnit(t *testing.T) {
	runRegExpFixtureNodeUnit(t, "regexp_v8_surrogates", "wasm32-wasi")
}

func TestRegExpSurrogatesSanitizeUnit(t *testing.T) {
	runRegExpFixtureNodeUnit(t, "regexp_v8_surrogates", "sanitized")
}
