package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/void_widening_legal.a", true, false})
}

// Source Node keeps the value that a void native call has already discarded.
// Pin every diagnostic, including the site and the actionable return annotation.
func TestVoidWideningIsRefused(test *testing.T) {
	test.Parallel()
	directory := "internal/oracle/testdata/void_widening_refused"
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		test.Fatal(err)
	}
	programs := 0
	for _, entry := range entries {
		extension := filepath.Ext(entry.Name())
		if extension != ".a" && extension != ".ts" {
			continue
		}
		programs++
		test.Run(entry.Name(), func(test *testing.T) {
			test.Parallel()
			path := filepath.Join(directory, entry.Name())
			absolute, err := filepath.Abs(filepath.Join(repository, path))
			if err != nil {
				test.Fatal(err)
			}
			if truth := onNode(test, absolute); truth.exitCode != 0 || len(truth.stderr) != 0 {
				test.Fatalf("source Node: exit %d stderr %q", truth.exitCode, truth.stderr)
			}
			_, err = lowered(test, absolute)
			var refused *lower.Refused
			if !errors.As(err, &refused) {
				test.Fatalf("want a void widening refusal, got %v", err)
			}
			expected, err := os.ReadFile(absolute + ".refused")
			if err != nil {
				test.Fatal(err)
			}
			got := strings.ReplaceAll(refused.Error(), strings.TrimSuffix(absolute, path), "")
			if got != strings.TrimSpace(string(expected)) {
				test.Fatalf("diagnostic: got %q, want %q", got, strings.TrimSpace(string(expected)))
			}
			source, err := os.ReadFile(absolute)
			if err != nil {
				test.Fatal(err)
			}
			if header := "// a-check: refused " + refused.What + "\n"; !strings.HasPrefix(string(source), header) {
				test.Fatalf("want first line %q", header)
			}
		})
	}
	// A program lost from the directory must fail here, not shrink the walk.
	if programs != 39 {
		test.Fatalf("found %d refused programs, want 39", programs)
	}
}
