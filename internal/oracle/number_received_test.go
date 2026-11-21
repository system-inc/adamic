package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// fxspptb_oct8_nodehost_fs_numeric_error.a holds number.c's adamic_number_received, an ERR_OUT_OF_RANGE's Received as
// Node writes it, through readSync and closeSync (node_fs_file.c) and Buffer.allocUnsafe (node_buffer.c), #n6hgpm0.
// TestReviewProgramsAgreeWithNode runs it natively, on the JavaScript backend and in release; this runs it as a WASI
// command under node:wasi and logs each row it writes.
func TestNumberReceivedOnWASI(t *testing.T) {
	t.Parallel()
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	// Absolute, as reviewEach and runWASIFixtures hand their programs over: the cache keys a source by its path
	// relative to its own directory, which a relative path under ../.. can't give.
	path, err := filepath.Abs(filepath.Join(reviewRoot, "agree", "fxspptb_oct8_nodehost_fs_numeric_error.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "program.wasm")
	if reason := buildWASI(t, native.C(program), binary); reason != "" {
		t.Fatalf("refused on wasm32-wasi: %s", reason)
	}
	expected := onNode(t, path)
	actual := execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/wasi.mjs"), binary)
	if difference := disagreement(expected, actual); difference != "" {
		t.Fatalf("%s\nnode: exit %d, stdout %q, stderr %q\nwasi: exit %d, stdout %q, stderr %q", difference, expected.exitCode, expected.stdout, expected.stderr, actual.exitCode, actual.stdout, actual.stderr)
	}
	for _, row := range strings.Split(strings.TrimSuffix(string(actual.stdout), "\n"), "\n") {
		t.Log(row)
	}
}
