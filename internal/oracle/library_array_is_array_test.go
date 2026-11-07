package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/library_array_is_array_unknown.a", true, false})
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/library_array_is_array_union.a", true, false})
}

// The prelude's isArray overload narrows unknown to readonly unknown[], so reading the narrowed value as a
// string[], or its element as a string, is the checker's error, pinned whole. Node runs the source and
// prints -1, then undefined. Mutant: the prelude's ArrayConstructor removed, which narrows to any[] and
// loads the program, so this test goes red.
func TestArrayIsArrayTypedReadIsRejected(test *testing.T) {
	test.Parallel()
	path := "internal/oracle/testdata/library_array_is_array_refused/typed_read.a"
	absolute, absolutePathError := filepath.Abs(filepath.Join(repository, path))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	contents, readError := os.ReadFile(absolute)
	if readError != nil {
		test.Fatal(readError)
	}
	// onNode loads the program to find its imports, and this one doesn't load: stock Node runs a .ts copy.
	stock := filepath.Join(test.TempDir(), "typed_read.ts")
	if writeError := os.WriteFile(stock, contents, 0o644); writeError != nil {
		test.Fatal(writeError)
	}
	if truth := execute(
		test,
		"node",
		"--disable-warning=ExperimentalWarning",
		stock,
	); disagreement(run{stdout: []byte("-1\nundefined\n")}, truth) != "" {
		test.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
	}
	_, loadError := load.Load([]string{absolute})
	var checkError *load.CheckError
	if !errors.As(loadError, &checkError) {
		test.Fatalf("want the checker to reject the typed reads, got %v", loadError)
	}
	expected, readError := os.ReadFile(strings.TrimSuffix(absolute, ".a") + ".refused")
	if readError != nil {
		test.Fatal(readError)
	}
	got := strings.ReplaceAll(checkError.Error(), strings.TrimSuffix(absolute, path), "")
	if got != strings.TrimSpace(string(expected)) {
		test.Fatalf("diagnostics: got %q, want %q", got, strings.TrimSpace(string(expected)))
	}
}
