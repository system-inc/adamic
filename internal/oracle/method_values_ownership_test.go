package oracle

import (
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func TestLibraryMethodCaptureMutants(test *testing.T) {
	test.Parallel()
	for _, extraOwner := range []bool{false, true} {
		name := "borrowed capture"
		if extraOwner {
			name = "extra capture owner"
		}
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			path, absolutePathError := filepath.Abs(
				filepath.Join(repository, "internal/oracle/testdata/library_method_values.a"),
			)
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			program, loweredError := lowered(test, path)
			if loweredError != nil {
				test.Fatal(loweredError)
			}
			original := native.C(program)
			captures := regexp.MustCompile(`(->cells\[0\] = )adamic_retain\(([^;]+)\);`)
			replacement := `${1}${2};`
			if extraOwner {
				replacement = `${1}adamic_retain(adamic_retain(${2}));`
			}
			mutant := captures.ReplaceAllString(original, replacement)
			if mutant == original {
				test.Fatal("mutant changed no captured cell")
			}
			binary := filepath.Join(test.TempDir(), "mutant")
			if buildError := native.Build(mutant, binary, native.Options{Sanitize: true}); buildError != nil {
				test.Fatal(buildError)
			}
			result := executeWith(
				test,
				[]string{"ASAN_OPTIONS=detect_leaks=0:abort_on_error=1:halt_on_error=1"},
				binary,
			)
			if !extraOwner {
				if !strings.Contains(string(result.stderr), "AddressSanitizer: heap-use-after-free") {
					test.Fatalf(
						"want ASan alone to catch borrowed capture, got exit %d stderr %s",
						result.exitCode,
						result.stderr,
					)
				}
				test.Log("caught by ASan heap-use-after-free on a captured receiver cell")
			} else {
				if difference := disagreement(onNode(test, path), result); difference != "" {
					test.Fatalf("extra-owner mutant must otherwise agree with Node: %s", difference)
				}
				report := executeWith(test, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
				if !strings.Contains(string(report.stderr), "LeakSanitizer: detected memory leaks") {
					test.Fatalf("want LeakSanitizer alone, got exit %d stderr %s", report.exitCode, report.stderr)
				}
				test.Log("caught only by LeakSanitizer")
			}
		})
	}
}
