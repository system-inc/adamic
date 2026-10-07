package oracle

import (
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func TestLibraryMethodCaptureMutants(t *testing.T) {
	t.Parallel()
	for _, extraOwner := range []bool{false, true} {
		name := "borrowed capture"
		if extraOwner {
			name = "extra capture owner"
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/library_method_values.a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			original := native.C(program)
			captures := regexp.MustCompile(`(->cells\[0\] = )adamic_retain\(([^;]+)\);`)
			replacement := `${1}${2};`
			if extraOwner {
				replacement = `${1}adamic_retain(adamic_retain(${2}));`
			}
			mutant := captures.ReplaceAllString(original, replacement)
			if mutant == original {
				t.Fatal("mutant changed no captured cell")
			}
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(mutant, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			result := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=0:abort_on_error=1:halt_on_error=1"}, binary)
			if !extraOwner {
				if !strings.Contains(string(result.stderr), "AddressSanitizer: heap-use-after-free") {
					t.Fatalf("want ASan alone to catch borrowed capture, got exit %d stderr %s", result.exitCode, result.stderr)
				}
				t.Log("caught by ASan heap-use-after-free on a captured receiver cell")
			} else {
				if difference := disagreement(onNode(t, path), result); difference != "" {
					t.Fatalf("extra-owner mutant must otherwise agree with Node: %s", difference)
				}
				report := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
				if !strings.Contains(string(report.stderr), "LeakSanitizer: detected memory leaks") {
					t.Fatalf("want LeakSanitizer alone, got exit %d stderr %s", report.exitCode, report.stderr)
				}
				t.Log("caught only by LeakSanitizer")
			}
		})
	}
}
