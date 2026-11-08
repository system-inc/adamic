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
	path, absolutePathError := filepath.Abs(filepath.Join("testdata", "regexp_native_input_cache.a"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	source, readError := os.ReadFile(path)
	if readError != nil {
		test.Fatal(readError)
	}
	want := execute(test, "node", "-e", string(source))
	if want.exitCode != 0 {
		test.Fatalf("Node failed: %q", want.stderr)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	var got run
	if target == "wasm32-wasi" {
		got = onWasi(test, native.C(program))
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
func TestRegExpInputSearchReleaseUnit(test *testing.T) {
	test.Parallel()
	runRegExpInputSearchNodeUnit(test, "native")
}
func TestRegExpInputSearchSanitizeUnit(test *testing.T) {
	test.Parallel()
	runRegExpInputSearchNodeUnit(test, "sanitized")
}
func TestRegExpInputSearchWasmUnit(test *testing.T) {
	test.Parallel()
	runRegExpInputSearchNodeUnit(test, "wasm32-wasi")
}
