package oracle

import (
	"errors"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// An incoming graph cannot be shared with a task while its region counts are plain.
// Pin the ownership refusal and its move fix at the integrated parallelMap boundary.
func TestGraphParallelMapRefusesRegions(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/concurrency/refused/graph_region.a"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = lowered(t, path)
	var refused *lower.Refused
	if !errors.As(err, &refused) || !strings.Contains(refused.What, "Ring.next") || !strings.Contains(refused.Fix, "part 2 (#p286ycm)") {
		t.Fatalf("want named graph ownership rejection and move fix, got %v", err)
	}
	t.Logf("%v", err)
}

// Weak's target is hidden from LeakSanitizer. Counted teardown must independently
// prove that introducing the weak back edge did not keep its plain tree alive.
func TestWeakRegionReviewFreesWithoutRegions(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/weak_region_review.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	if len(program.GraphTypes) != 0 {
		t.Fatal("Weak tree acquired graph types")
	}
	binary := filepath.Join(t.TempDir(), "weak-counted")
	if err := native.Build(native.C(program), binary, native.Options{Count: true, Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	if report := leakChecked(t, native.C(program), binary); report != "" {
		t.Fatal(report)
	}
	result := execute(t, binary)
	counts := countsLine.FindSubmatch(result.stderr)
	if result.exitCode != 0 || counts == nil {
		t.Fatalf("%d %s", result.exitCode, result.stderr)
	}
	allocations, _ := strconv.Atoi(string(counts[1]))
	frees, _ := strconv.Atoi(string(counts[2]))
	arenas, _ := strconv.Atoi(string(counts[6]))
	if allocations != frees+arenas || graphRegionLine.Match(result.stderr) || graphCountsLine.Match(result.stderr) {
		t.Fatalf("Weak back edge kept a plain tree alive: %s", result.stderr)
	}
	t.Logf("%s", result.stderr)
}
