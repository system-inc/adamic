package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// Check inference and the independent sequential witness; the main oracle runs native variants.
func TestConcurrencyAgreesWithNode(test *testing.T) {
	test.Parallel()
	paths, globError := filepath.Glob(filepath.Join(repository, "internal/oracle/testdata/concurrency/accepted/*.a"))
	if globError != nil || len(paths) < 9 {
		test.Fatalf("fixtures: %v (%d)", globError, len(paths))
	}
	for _, path := range paths {
		test.Run(filepath.Base(path), func(test *testing.T) {
			test.Parallel()
			absolute, absolutePathError := filepath.Abs(path)
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			started := time.Now()
			program, loweredError := lowered(test, absolute)
			if loweredError != nil {
				test.Fatal(loweredError)
			}
			if filepath.Base(path) == "recursive_tree.a" && time.Since(started) > 10*time.Second {
				test.Fatal("recursive readonly tree inference did not finish promptly")
			}
			if !strings.Contains(native.C(program), "adamic_parallel_map(") {
				test.Fatal("native output did not call the runtime ABI")
			}
			oracle, backend := onNode(test, absolute), onJavaScriptBackend(test, program)
			if difference := disagreement(oracle, backend); difference != "" {
				test.Fatalf(
					"%s: Node %d %q %q, backend %d %q %q",
					difference,
					oracle.exitCode,
					oracle.stdout,
					oracle.stderr,
					backend.exitCode,
					backend.stdout,
					backend.stderr,
				)
			}

		})
	}
}

func TestConcurrencyRefusals(test *testing.T) {
	test.Parallel()
	paths, globError := filepath.Glob(filepath.Join(repository, "internal/oracle/testdata/concurrency/refused/*.a"))
	if globError != nil || len(paths) < 12 {
		test.Fatalf("refusal fixtures: %v (%d)", globError, len(paths))
	}
	for _, path := range paths {
		test.Run(filepath.Base(path), func(test *testing.T) {
			test.Parallel()
			want, readError := os.ReadFile(strings.TrimSuffix(path, ".a") + ".what")
			if readError != nil {
				test.Fatal(readError)
			}
			absolute, absolutePathError := filepath.Abs(path)
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			_, loweredError := lowered(test, absolute)
			var refused *lower.Refused
			if !errors.As(loweredError, &refused) {
				test.Fatalf("want Refused %q, got %v", strings.TrimSpace(string(want)), loweredError)
			}
			if refused.What != strings.TrimSpace(string(want)) {
				test.Fatalf("want exact What %q, got %q", strings.TrimSpace(string(want)), refused.What)
			}
			if refused.Where == "" || refused.Fix == "" {
				test.Fatalf("incomplete diagnostic: %+v", refused)
			}
			if strings.Contains(refused.What, "not shareable") && !strings.Contains(refused.Fix, "part 2 (#p286ycm)") {
				test.Fatalf("mutable transfer fix omitted part 2: %s", refused.Fix)
			}
		})
	}
}
