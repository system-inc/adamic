package css

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// The optimized artifact must match the full Go oracle too; sanitizer builds
// use a different optimizer setting and the timing checksum alone is weaker.
// Not parallel: native.Build writes the shared user cache directory adamic/runtime (and adamic/units when split builds are enabled). askedCases writes the shared ADAMIC_CSS_KEEP and ADAMIC_CSS_KEEP_RAW paths when configured.
func TestCSSPrinterOptimizedMatchesGo(t *testing.T) {
	cases, _ := askedCases(t)
	path, _ := filepath.Abs("print_main.ts")
	program := lowered(t, path)
	binary := filepath.Join(t.TempDir(), "printer")
	if err := native.Build(native.C(program), binary, native.Options{}); err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"default", "narrow"} {
		expected := printerAnswers(t, cases, mode)
		result := execute(t, nil, binary, cases, "output", "once", mode)
		if result.exitCode != 0 || len(result.stderr) != 0 {
			t.Fatalf("optimized %s: %d %s", mode, result.exitCode, result.stderr)
		}
		if difference := firstDifference(string(result.stdout), expected); difference != "" {
			t.Fatal(difference)
		}
		t.Logf("optimized %s: %d byte-identical Go results", mode, strings.Count(expected, "\n")/2)
	}
}

// Not parallel: native.Build writes the shared user cache directory adamic/runtime (and adamic/units when split builds are enabled). askedCases writes the shared ADAMIC_CSS_KEEP and ADAMIC_CSS_KEEP_RAW paths when configured.
func TestCSSParserOptimizedMatchesNode(t *testing.T) {
	cases, _ := askedCases(t)
	path, _ := filepath.Abs("compose_main.ts")
	program := lowered(t, path)
	binary := filepath.Join(t.TempDir(), "parser")
	if err := native.Build(native.C(program), binary, native.Options{}); err != nil {
		t.Fatal(err)
	}
	expected := onNode(t, path, cases)
	result := execute(t, nil, binary, cases)
	for _, side := range []struct {
		name   string
		result run
	}{{"Node", expected}, {"optimized native", result}} {
		if side.result.exitCode != 0 || len(side.result.stderr) != 0 {
			t.Fatalf("%s: %d %s", side.name, side.result.exitCode, side.result.stderr)
		}
	}
	if difference := firstDifference(string(result.stdout), string(expected.stdout)); difference != "" {
		t.Fatal(difference)
	}
	t.Logf("optimized parser: %d byte-identical Node trees/errors, held to Go by the composed-tree test", strings.Count(string(expected.stdout), "\n")/2)
}
