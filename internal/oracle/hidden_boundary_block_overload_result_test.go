package oracle

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// Backend registration follows the shared overload-results implementation.
// This witness pins the external observation without admitting the overload.
func TestHiddenBoundaryBlockOverloadResultNodeWitness(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/hidden_boundary_block_overload_result.a"))
	if err != nil {
		t.Fatal(err)
	}
	result := onNode(t, path)
	t.Logf("source Node: exit %d, stdout %q, stderr %q", result.exitCode, result.stdout, result.stderr)
	if difference := disagreement(run{stdout: []byte("7\n")}, result); difference != "" {
		t.Fatal(difference)
	}
}

func TestHiddenBoundaryBlockOverloadResultRefusal(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/refusals/hidden_boundary_block_overload_result.a"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = lowered(t, path)
	var refused *lower.Refused
	const want = "overload 1 of transformAsyncFunctionBody result Block cannot be served by implementation result ConciseBody"
	if !errors.As(err, &refused) || refused.What != want {
		t.Fatalf("Expression cannot satisfy the Block-only promise: want %q, got %v", want, err)
	}
	t.Logf("unsafe implementation: %v", err)
}
