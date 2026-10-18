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
	}{"internal/oracle/testdata/regexp_native_input_cache.a", true, false})
}

func runRegExpInputSearchNodeUnit(t *testing.T, target string) {
	t.Helper()
	if target == "wasm32-wasi" && os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	path, err := filepath.Abs(filepath.Join("testdata", "regexp_native_input_cache.a"))
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := execute(t, "node", "-e", string(source))
	if want.exitCode != 0 {
		t.Fatalf("Node failed: %q", want.stderr)
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
func TestRegExpInputSearchReleaseUnit(t *testing.T) {
	t.Parallel()
	runRegExpInputSearchNodeUnit(t, "native")
}
func TestRegExpInputSearchSanitizeUnit(t *testing.T) {
	t.Parallel()
	runRegExpInputSearchNodeUnit(t, "sanitized")
}
func TestRegExpInputSearchWASMUnit(t *testing.T) {
	t.Parallel()
	runRegExpInputSearchNodeUnit(t, "wasm32-wasi")
}
