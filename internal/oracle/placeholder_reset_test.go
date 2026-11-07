package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// A reset to a placeholder on a let local keeps its readiness flag on its declaration, so it
// agrees with Node natively, under the sanitizers and through JavaScript: the control for the
// refusals below.
func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/placeholder_reset_let.a", true, false})
}

// A reset to a placeholder (x = undefined!) of a binding whose readiness flag nothing declares
// where the reset reaches it is refused: a parameter, a for-of or for-in variable, an async
// function's local and a block local of a module that awaits at its top level. Admitted, each
// wrote an undeclared flag, C that clang rejects and JavaScript that throws a ReferenceError
// (stage3/scouts/step09/nonnull/03_checked_stop.a, slice 5). Each program's whole diagnostic is
// pinned in its .refused file. Mutant: placeholderResetRefusal returning nil, which lowers every one.
func TestPlaceholderResetsWithoutReadinessAreRefused(test *testing.T) {
	test.Parallel()
	directory := "internal/oracle/testdata/placeholder_reset_refused"
	entries, readDirectoryError := os.ReadDir(filepath.Join(repository, directory))
	if readDirectoryError != nil {
		test.Fatal(readDirectoryError)
	}
	programs := 0
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".a") {
			continue
		}
		programs++
		path := directory + "/" + name
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			absolute, absolutePathError := filepath.Abs(filepath.Join(repository, path))
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			_, loweredError := lowered(test, absolute)
			var refused *lower.Refused
			if !errors.As(loweredError, &refused) {
				test.Fatalf("want a refusal, got %v", loweredError)
			}
			expected, readError := os.ReadFile(strings.TrimSuffix(absolute, ".a") + ".refused")
			if readError != nil {
				test.Fatal(readError)
			}
			got := strings.ReplaceAll(refused.Error(), strings.TrimSuffix(absolute, path), "")
			if got != strings.TrimSpace(string(expected)) {
				test.Fatalf("diagnostic: got %q, want %q", got, strings.TrimSpace(string(expected)))
			}
		})
	}
	if programs != 5 {
		test.Fatalf(
			"%d programs in %s, want 5: a program added or removed needs this count changed with it",
			programs,
			directory,
		)
	}
}
