package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

func TestAmbientDeclarationsAreRefused(test *testing.T) {
	test.Parallel()
	directory := "internal/oracle/testdata/ambient_refused"
	entries, readDirectoryError := os.ReadDir(filepath.Join(repository, directory))
	if readDirectoryError != nil {
		test.Fatal(readDirectoryError)
	}
	for _, entry := range entries {
		extension := filepath.Ext(entry.Name())
		if entry.IsDir() || strings.HasSuffix(entry.Name(), ".d.ts") || (extension != ".a" && extension != ".ts") {
			continue
		}
		test.Run(entry.Name(), func(test *testing.T) {
			test.Parallel()
			path := directory + "/" + entry.Name()
			absolute, absolutePathError := filepath.Abs(filepath.Join(repository, path))
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			truth := onNode(test, absolute)
			if truth.exitCode == 0 || !strings.Contains(string(truth.stderr), "ReferenceError") {
				test.Fatalf("source Node: want ReferenceError, exit %d stderr %q", truth.exitCode, truth.stderr)
			}
			_, loweredError := lowered(test, absolute)
			var refused *lower.Refused
			if !errors.As(loweredError, &refused) {
				test.Fatalf("want a refusal, got %v", loweredError)
			}
			expected, readError := os.ReadFile(strings.TrimSuffix(absolute, extension) + ".refused")
			if readError != nil {
				test.Fatal(readError)
			}
			got := strings.ReplaceAll(refused.Error(), strings.TrimSuffix(absolute, path), "")
			if got != strings.TrimSpace(string(expected)) {
				test.Fatalf("diagnostic: got %q, want %q", got, strings.TrimSpace(string(expected)))
			}
			source, readError := os.ReadFile(absolute)
			if readError != nil {
				test.Fatal(readError)
			}
			if !strings.HasPrefix(string(source), "// a-check: refused "+refused.What+"\n") {
				test.Fatalf("fixture header does not pin %q", refused.What)
			}
		})
	}
}
