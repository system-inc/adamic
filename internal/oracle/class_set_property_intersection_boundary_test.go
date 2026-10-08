package oracle

import (
	"errors"
	"github.com/system-inc/adamic/internal/lower"
	"path/filepath"
	"testing"
)

func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/class_set_property_intersection_boundary.a", false, false})
}
func TestClassSetPropertyIntersectionStorageBoundary(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/class_set_property_intersection_boundary.a"))
	if err != nil {
		t.Fatal(err)
	}
	node := onNode(t, path)
	if node.exitCode != 0 || string(node.stdout) != "kept\n" || len(node.stderr) != 0 {
		t.Fatalf("Node: %+v", node)
	}
	_, err = lowered(t, path)
	t.Logf("boundary: %v", err)
	var stopped *lower.NotYet
	if !errors.As(err, &stopped) {
		t.Fatalf("want unmatched intersection storage stop, got %v", err)
	}
}
