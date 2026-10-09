package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

func TestStep20ArrayViewVariance(t *testing.T) {
	t.Parallel()
	path, failure := filepath.Abs(filepath.Join(repository, "stage3/fixtures/iteration/array_view_variance.a"))
	if failure != nil {
		t.Fatal(failure)
	}
	node := onNode(t, path)
	if node.exitCode != 0 || len(node.stderr) != 0 || string(node.stdout) != "undefined\n" {
		t.Fatalf("Node counterexample: %+v", node)
	}
	_, err := lowered(t, path)
	var refused *lower.Refused
	if !errors.As(err, &refused) || !strings.Contains(refused.Fix, "adamic/invariant-mutable") {
		t.Fatalf("want nested mutable slot refusal through a readonly array view, got %v", err)
	}
}
