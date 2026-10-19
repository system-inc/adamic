package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

func TestAmbientDeclarationsAreRefused(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/ambient_refused"
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		extension := filepath.Ext(entry.Name())
		if entry.IsDir() || strings.HasSuffix(entry.Name(), ".d.ts") || (extension != ".a" && extension != ".ts") {
			continue
		}
		t.Run(entry.Name(), func(t *testing.T) {
			t.Parallel()
			path := directory + "/" + entry.Name()
			absolute, err := filepath.Abs(filepath.Join(repository, path))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, absolute)
			if truth.exitCode == 0 || !strings.Contains(string(truth.stderr), "ReferenceError") {
				t.Fatalf("source Node: want ReferenceError, exit %d stderr %q", truth.exitCode, truth.stderr)
			}
			_, err = lowered(t, absolute)
			var refused *lower.Refused
			if !errors.As(err, &refused) {
				t.Fatalf("want a refusal, got %v", err)
			}
			expected, err := os.ReadFile(strings.TrimSuffix(absolute, extension) + ".refused")
			if err != nil {
				t.Fatal(err)
			}
			got := strings.ReplaceAll(refused.Error(), strings.TrimSuffix(absolute, path), "")
			if got != strings.TrimSpace(string(expected)) {
				t.Fatalf("diagnostic: got %q, want %q", got, strings.TrimSpace(string(expected)))
			}
			source, err := os.ReadFile(absolute)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.HasPrefix(string(source), "// a-check: refused "+refused.What+"\n") {
				t.Fatalf("fixture header does not pin %q", refused.What)
			}
		})
	}
}
