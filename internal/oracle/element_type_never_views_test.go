package oracle

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestElementTypeNeverNeedsCheckedViews(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/element_type_never_views.a"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = lowered(t, path)
	if err == nil || !strings.Contains(err.Error(), "adamic/invariant-mutable") {
		t.Fatalf("wanted the incompatible mutable alias refusal, got %v", err)
	}
	got := onNode(t, path)
	if got.exitCode != 0 || string(got.stdout) != "number 1\n" || len(got.stderr) != 0 {
		t.Fatalf("Node alias observation changed: %+v", got)
	}
}
