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
	paths, err := filepath.Glob(filepath.Join(repository, "internal/oracle/testdata/concurrency/accepted/*.a"))
	if err != nil || len(paths) < 9 {
		test.Fatalf("fixtures: %v (%d)", err, len(paths))
	}
	for _, path := range paths {
		test.Run(filepath.Base(path), func(test *testing.T) {
			test.Parallel()
			absolute, err := filepath.Abs(path)
			if err != nil {
				test.Fatal(err)
			}
			started := time.Now()
			program, err := lowered(test, absolute)
			if err != nil {
				test.Fatal(err)
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
	paths, err := filepath.Glob(filepath.Join(repository, "internal/oracle/testdata/concurrency/refused/*.a"))
	if err != nil || len(paths) < 12 {
		test.Fatalf("refusal fixtures: %v (%d)", err, len(paths))
	}
	for _, path := range paths {
		test.Run(filepath.Base(path), func(test *testing.T) {
			test.Parallel()
			want, err := os.ReadFile(strings.TrimSuffix(path, ".a") + ".what")
			if err != nil {
				test.Fatal(err)
			}
			absolute, err := filepath.Abs(path)
			if err != nil {
				test.Fatal(err)
			}
			_, err = lowered(test, absolute)
			var refused *lower.Refused
			if !errors.As(err, &refused) {
				test.Fatalf("want Refused %q, got %v", strings.TrimSpace(string(want)), err)
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
