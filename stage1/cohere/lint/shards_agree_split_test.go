package lint

import (
	"github.com/system-inc/adamic/stage1/cohere/lint/shards"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestShardsAgree requires the driver's output to be byte-identical however many processes share the
// manifest (#tj6d455): one, two, and the machine's cores. Equal counts cannot see a reordered or repeated
// case, so the whole output is compared, and the count mode too.
func TestShardsAgree(t *testing.T) {
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	oracle := goOracle(t)
	rows := generated(t)
	for _, row := range upstream(t) {
		if !strings.HasSuffix(row, "\tunsupported-recovery") {
			rows = append(rows, row)
		}
	}
	// The compiler files are the corpus with large files, where shards differ most in what they hold, so
	// the test refuses to run without them rather than passing on a smaller corpus.
	source := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
	if source == "" {
		t.Fatal("set ADAMIC_TYPESCRIPT_SOURCE to the pinned TypeScript checkout: TestShardsAgree needs its compiler files")
	}
	matches, err := filepath.Glob(filepath.Join(source, "src/compiler/*.ts"))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) == 0 {
		t.Fatalf("no compiler files under %s", source)
	}
	rows = append(rows, matches...)
	path := manifest(t, recoveryRows(t, oracle, rows))
	binary := buildPort(t, directory, false)
	want := execute(t, "", binary, "--manifest", path)
	wantCount := execute(t, "", binary, "--manifest", path, "--count")
	for _, count := range []int{1, 2, runtime.NumCPU()} {
		got, err := shards.Run(binary, path, count, false)
		if err != nil {
			t.Fatal(err)
		}
		if diff := difference(got, want.output); diff != "" {
			t.Fatalf("%d shards: %s", count, diff)
		}
		gotCount, err := shards.Run(binary, path, count, true)
		if err != nil {
			t.Fatal(err)
		}
		if string(gotCount) != string(wantCount.output) {
			t.Fatalf("%d shards count %q, want %q", count, gotCount, wantCount.output)
		}
	}
	t.Logf("%d rows: identical at 1, 2 and %d shards, %d bytes", len(rows), runtime.NumCPU(), len(want.output))
}
