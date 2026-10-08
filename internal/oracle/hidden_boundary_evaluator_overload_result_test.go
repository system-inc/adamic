package oracle

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// The sound witness awaits three-way registration after codex/overload-results
// supplies the shared mechanism. The negative remains a focused refusal probe.
func TestHiddenBoundaryEvaluatorOverloadResultSource(t *testing.T) {
	for _, probe := range []struct{ path, stdout string }{
		{"internal/oracle/testdata/hidden_boundary_evaluator_overload_result.a", "node\n"},
		{"internal/load/testdata/0.1/refuse/hidden_boundary_evaluator_overload_result.a", "7\n"},
	} {
		t.Run(filepath.Base(filepath.Dir(probe.path)), func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, probe.path))
			if err != nil {
				t.Fatal(err)
			}
			result := onNode(t, path)
			if result.exitCode != 0 || string(result.stdout) != probe.stdout || len(result.stderr) != 0 {
				t.Fatalf("source Node: stdout=%q stderr=%q exit=%d; want stdout=%q, empty stderr, exit 0", result.stdout, result.stderr, result.exitCode, probe.stdout)
			}
			t.Logf("stdout=%q stderr=%q exit=%d", result.stdout, result.stderr, result.exitCode)
		})
	}
}

func TestHiddenBoundaryEvaluatorOverloadResultNegative(t *testing.T) {
	path := filepath.Join(repository, "internal/load/testdata/0.1/refuse/hidden_boundary_evaluator_overload_result.a")
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	_, err = lower.Lower(context.Background(), program)
	var refused *lower.Refused
	if !errors.As(err, &refused) || !strings.Contains(refused.What, "overload 1 of evaluate result") {
		t.Fatalf("number-returning string overload must remain Refused at its result promise; got %v", err)
	}
	t.Log(err)
}
