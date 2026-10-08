package oracle

import (
	"bytes"
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
		"internal/oracle/testdata/phantom_array_brands.a",
		"internal/oracle/testdata/native-sorted-array-brand.a",
		"internal/oracle/testdata/phantom_overload_results.a",
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

func TestPhantomSortedArrayProbe(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/native-sorted-array-brand.a"))
	if err != nil {
		t.Fatal(err)
	}
	original, err := os.ReadFile(filepath.Join(repository, "review/phantom-brands/array-evidence/native-sorted-array-brand.a"))
	if err != nil {
		t.Fatal(err)
	}
	fixture, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(original, fixture) {
		t.Fatal("parser row 5 probe was changed")
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	oracle := onNode(t, path)
	if string(oracle.stdout) != "2\n" || len(oracle.stderr) != 0 || oracle.exitCode != 0 {
		t.Fatalf("parser row 5 Node output changed: %#v", oracle)
	}
	native, sanitized := natively(t, program)
	for name, result := range map[string]run{"native": native, "JavaScript": onJavaScriptBackend(t, program), "release": released(t, program)} {
		if diff := disagreement(oracle, result); diff != "" {
			t.Errorf("%s: %s: got %#v, want %#v", name, diff, result, oracle)
		}
		t.Logf("%s: stdout %q, stderr %q, exit %d", name, result.stdout, result.stderr, result.exitCode)
	}
	if leaked := leaks(t, program, sanitized); leaked != "" {
		t.Errorf("leaks: %s", leaked)
	}
}
