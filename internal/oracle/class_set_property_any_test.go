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
	}{"internal/oracle/testdata/class_set_property_any.a", false, false})
}

// any erases the storage proof and remains a language-ruling refusal.
func TestClassSetPropertyAnyBoundary(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/class_set_property_any.a"))
	if err != nil {
		t.Fatal(err)
	}
	node := onNode(t, path)
	if node.exitCode != 0 || string(node.stdout) != "stored\n" || len(node.stderr) != 0 {
		t.Fatalf("Node: %+v", node)
	}
	_, err = lowered(t, path)
	var stopped *lower.NotYet
	if !errors.As(err, &stopped) || stopped.What != "storing any in a field" {
		t.Fatalf("want any storage boundary, got %v", err)
	}
}
