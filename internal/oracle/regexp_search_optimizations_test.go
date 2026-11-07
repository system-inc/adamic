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

func runRegExpInputSearchNodeUnit(test *testing.T, target string) {
	test.Helper()
	if target == "wasm32-wasi" && os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		test.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	path, err := filepath.Abs(filepath.Join("testdata", "regexp_native_input_cache.a"))
	if err != nil {
		test.Fatal(err)
	}
	source, err := os.ReadFile(path)
	if err != nil {
		test.Fatal(err)
	}
	want := execute(test, "node", "-e", string(source))
	if want.exitCode != 0 {
		test.Fatalf("Node failed: %q", want.stderr)
	}
	program, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
	}
	var got run
	if target == "wasm32-wasi" {
		got = onWASI(test, native.C(program))
	} else {
		binary := filepath.Join(test.TempDir(), "program")
		if err := native.Build(native.C(program), binary, native.Options{Sanitize: target == "sanitized"}); err != nil {
			test.Fatal(err)
		}
		got = execute(test, binary)
	}
	if difference := disagreement(want, got); difference != "" {
		test.Fatalf("%s: Node=%q compiled=%q stderr=%q", difference, want.stdout, got.stdout, got.stderr)
	}
}
func TestRegExpInputSearchReleaseUnit(test *testing.T) {
	test.Parallel()
	runRegExpInputSearchNodeUnit(test, "native")
}
func TestRegExpInputSearchSanitizeUnit(test *testing.T) {
	test.Parallel()
	runRegExpInputSearchNodeUnit(test, "sanitized")
}
func TestRegExpInputSearchWASMUnit(test *testing.T) {
	test.Parallel()
	runRegExpInputSearchNodeUnit(test, "wasm32-wasi")
}
