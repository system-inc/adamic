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

func runRegExpFixtureNodeUnit(test *testing.T, fixture string, target string) {
	if target == "wasm32-wasi" && os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		test.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	path, absolutePathError := filepath.Abs(filepath.Join("testdata", fixture+".a"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	source, readError := os.ReadFile(path)
	if readError != nil {
		test.Fatal(readError)
	}
	want := execute(test, "node", "-e", string(source))
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
func TestRegExpClassOctalNativeUnit(test *testing.T) {
	test.Parallel()
	runRegExpFixtureNodeUnit(test, "regexp_native_class_octal", "native")
}
func TestRegExpClassOctalWASMUnit(test *testing.T) {
	test.Parallel()
	runRegExpFixtureNodeUnit(test, "regexp_native_class_octal", "wasm32-wasi")
}
func TestRegExpFoldingNativeUnit(test *testing.T) {
	test.Parallel()
	runRegExpFixtureNodeUnit(test, "regexp_v8_folding", "native")
}
func TestRegExpFoldingWASMUnit(test *testing.T) {
	test.Parallel()
	runRegExpFixtureNodeUnit(test, "regexp_v8_folding", "wasm32-wasi")
}
func TestRegExpSurrogatesReleaseUnit(test *testing.T) {
	test.Parallel()
	runRegExpFixtureNodeUnit(test, "regexp_v8_surrogates", "native")
}
func TestRegExpSurrogatesWASMUnit(test *testing.T) {
	test.Parallel()
	runRegExpFixtureNodeUnit(test, "regexp_v8_surrogates", "wasm32-wasi")
}

func TestRegExpSurrogatesSanitizeUnit(test *testing.T) {
	test.Parallel()
	runRegExpFixtureNodeUnit(test, "regexp_v8_surrogates", "sanitized")
}
