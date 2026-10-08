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
	}{"internal/oracle/testdata/class_set_property_array_dynamic.a", false, false})
}

// A dynamic length needs a no-holes proof; accepting arbitrary growth is unsound.
func TestClassSetPropertyDynamicLengthBoundary(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/class_set_property_array_dynamic.a"))
	if err != nil {
		t.Fatal(err)
	}
	node := onNode(t, path)
	if node.exitCode != 0 || string(node.stdout) != "heldheld\n" || len(node.stderr) != 0 {
		t.Fatalf("Node: %+v", node)
	}
	_, err = lowered(t, path)
	var stop *lower.NotYet
	if !errors.As(err, &stop) || stop.What != "assigning a field of a value" {
		t.Fatalf("want dynamic length boundary, got %v", err)
	}
}
