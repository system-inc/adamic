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
func TestConcurrencyAgreesWithNode(t *testing.T) {
	t.Parallel()
	paths, err := filepath.Glob(filepath.Join(repository, "internal/oracle/testdata/concurrency/accepted/*.a"))
	if err != nil || len(paths) < 9 {
		t.Fatalf("fixtures: %v (%d)", err, len(paths))
	}
	for _, path := range paths {
		t.Run(filepath.Base(path), func(t *testing.T) {
			t.Parallel()
			absolute, err := filepath.Abs(path)
			if err != nil {
				t.Fatal(err)
			}
			started := time.Now()
			program, err := lowered(t, absolute)
			if err != nil {
				t.Fatal(err)
			}
			if filepath.Base(path) == "recursive_tree.a" && time.Since(started) > 10*time.Second {
				t.Fatal("recursive readonly tree inference did not finish promptly")
			}
			if !strings.Contains(native.C(program), "adamic_parallel_map(") {
				t.Fatal("native output did not call the runtime ABI")
			}
			oracle, backend := onNode(t, absolute), onJavaScriptBackend(t, program)
			if difference := disagreement(oracle, backend); difference != "" {
				t.Fatalf("%s: Node %d %q %q, backend %d %q %q", difference, oracle.exitCode, oracle.stdout, oracle.stderr, backend.exitCode, backend.stdout, backend.stderr)
			}

		})
	}
}

func TestConcurrencyRefusals(t *testing.T) {
	t.Parallel()
	paths, err := filepath.Glob(filepath.Join(repository, "internal/oracle/testdata/concurrency/refused/*.a"))
	if err != nil || len(paths) < 12 {
		t.Fatalf("refusal fixtures: %v (%d)", err, len(paths))
	}
	for _, path := range paths {
		t.Run(filepath.Base(path), func(t *testing.T) {
			t.Parallel()
			want, err := os.ReadFile(strings.TrimSuffix(path, ".a") + ".what")
			if err != nil {
				t.Fatal(err)
			}
			absolute, err := filepath.Abs(path)
			if err != nil {
				t.Fatal(err)
			}
			_, err = lowered(t, absolute)
			var refused *lower.Refused
			if !errors.As(err, &refused) {
				t.Fatalf("want Refused %q, got %v", strings.TrimSpace(string(want)), err)
			}
			if refused.What != strings.TrimSpace(string(want)) {
				t.Fatalf("want exact What %q, got %q", strings.TrimSpace(string(want)), refused.What)
			}
			if refused.Where == "" || refused.Fix == "" {
				t.Fatalf("incomplete diagnostic: %+v", refused)
			}
			if strings.Contains(refused.What, "not shareable") && !strings.Contains(refused.Fix, "part 2 (#p286ycm)") {
				t.Fatalf("mutable transfer fix omitted part 2: %s", refused.Fix)
			}
		})
	}
}
