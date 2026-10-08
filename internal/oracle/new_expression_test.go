package oracle

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

func init() {
	for _, name := range []string{"uint16", "uint16_notyet", "class_value", "class_cache_or"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/new_expression_" + name + ".a", name == "class_value" || name == "class_cache_or", false})
	}
}

// Source Node demonstrates the conversion, while Adamic must stop before emission
// until the runtime-owned Uint16 storage and conversions are integrated.
func TestNewExpressionUint16RemainsNotYet(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/new_expression_uint16_notyet.a"))
	if err != nil {
		t.Fatal(err)
	}
	node := onNode(t, path)
	if node.exitCode != 0 || string(node.stdout) != "4464\n" || len(node.stderr) != 0 {
		t.Fatalf("Node conversion: exit %d, stdout %q, stderr %q", node.exitCode, node.stdout, node.stderr)
	}
	_, err = lowered(t, path)
	var notYet *lower.NotYet
	if !errors.As(err, &notYet) || notYet.What != "typed array element type Uint16Array" {
		t.Fatalf("want Uint16Array element-type NotYet, got %v", err)
	}
}
