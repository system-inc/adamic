package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// These incoming witnesses meet existing area boundaries, so keep their source
// and pin the complete diagnostic instead of silently excluding them from scans.
func TestClosureMergeRefusals(test *testing.T) {
	test.Parallel()
	paths := []string{
		"stage3/fixtures/nested-functions/refused/01_scanner_frame.a",
	}
	for _, name := range fileSystemFileRefusedFixtures {
		paths = append(paths, "internal/oracle/testdata/optional_widening_refused/node_fs_file_"+name+".a")
	}
	for _, path := range paths {
		test.Run(path, func(test *testing.T) {
			test.Parallel()
			absolute, pathError := filepath.Abs(filepath.Join(repository, path))
			if pathError != nil {
				test.Fatal(pathError)
			}
			var truth run
			if strings.Contains(path, "node_fs_file_") {
				truth = onNodeWith(test, fileSystemFilePrepare(test, sharedDirectory(test), "refused-node"), absolute)
			} else {
				truth = onNode(test, absolute)
			}
			if truth.exitCode != 0 || len(truth.stderr) != 0 {
				test.Fatalf("source Node: exit %d stderr %q", truth.exitCode, truth.stderr)
			}
			_, loweredError := lowered(test, absolute)
			var refused *lower.Refused
			if !errors.As(loweredError, &refused) {
				test.Fatalf("want pinned refusal, got %v", loweredError)
			}
			expected, readError := os.ReadFile(strings.TrimSuffix(absolute, ".a") + ".refused")
			if readError != nil {
				test.Fatal(readError)
			}
			got := strings.ReplaceAll(loweredError.Error(), strings.TrimSuffix(absolute, path), "")
			if got != strings.TrimSpace(string(expected)) {
				test.Fatalf("diagnostic: got %q, want %q", got, strings.TrimSpace(string(expected)))
			}
		})
	}
}
