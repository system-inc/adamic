package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

func TestNamespaceLiveExportBoundary(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "stage3/namespace-live-export/live.a"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "false:false\ntrue:true\nfalse:false\n" || len(truth.stderr) != 0 {
		t.Fatalf("unexpected live binding observation from Node: %+v", truth)
	}
	_, err = lowered(t, path)
	var gap *lower.NotYet
	if !errors.As(err, &gap) || gap.What != "a mutable namespace export; use a module or export functions around private state" || !strings.HasSuffix(gap.Where, ":8:5") {
		t.Fatalf("namespace live export boundary changed: %v", err)
	}
	t.Log("Node observes outside writes; delivery stops at mutable namespace export, live.a:8:5")
}
