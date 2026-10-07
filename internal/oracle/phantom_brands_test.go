package oracle

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

func init() {
	for _, path := range []string{
		"internal/oracle/testdata/phantom_brands.a",
		"internal/oracle/testdata/phantom_undefined.a",
		"internal/oracle/testdata/phantom_catch.a",
	} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path, true, false})
	}
}

// Keep the original review program as the Node oracle. String only adapts the console
// signature in Adamic's prelude; the property read throws before conversion can run.
func TestPhantomUndefinedReview(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "review/phantom-brands/undefined-read.a"))
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	normalized := strings.Replace(string(source), "console.log(value.__escapedIdentifier);", "console.log(String(value.__escapedIdentifier));", 1)
	checked, err := load.LoadOverlay([]string{path}, map[string]string{path: normalized})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), checked)
	if err != nil {
		t.Fatal(err)
	}
	oracle := execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle", "node.mjs"), path)
	native, _ := natively(t, program)
	for name, result := range map[string]run{"native": native, "JavaScript": onJavaScriptBackend(t, program), "release": released(t, program)} {
		if diff := disagreement(oracle, result); diff != "" {
			t.Errorf("%s: %s: Node exit %d stdout %q stderr %q; backend exit %d stdout %q stderr %q", name, diff, oracle.exitCode, oracle.stdout, oracle.stderr, result.exitCode, result.stdout, result.stderr)
		}
	}
	if oracle.exitCode != 70 || !strings.Contains(string(oracle.stderr), "TypeError: Cannot read properties of undefined (reading '__escapedIdentifier')") {
		t.Fatalf("review oracle changed: %#v", oracle)
	}
}
